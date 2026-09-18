// Package health collects the operational metrics CampusWatch reports for a
// monitored computer: CPU, memory, disk, battery, network reachability and
// uptime, as described in README section 17.
//
// Two rules govern everything in this package:
//
//   - Collection never fails the heartbeat. A metric that cannot be read is
//     reported as absent, because README section 17 requires the platform to
//     distinguish a health warning from a confirmed hardware failure, and an
//     unreadable sensor is neither.
//   - Only aggregate system counters are read. No process names, command lines,
//     window titles or file contents are collected, in line with the privacy
//     boundary in README section 6.
package health

import (
	"context"
	"net"
	"time"

	"CampusWatch/agent/internal/system"
)

// Agent health values reported in the heartbeat. This describes the agent's own
// ability to observe the machine, not the machine's condition.
const (
	AgentHealthHealthy  = "healthy"
	AgentHealthDegraded = "degraded"
)

// Probe reports whether the monitored computer currently has usable network
// connectivity. It is injected so the caller decides what "connected" means.
type Probe func(ctx context.Context) bool

// Sample is one reading of the machine's operational state.
type Sample struct {
	// RecordedAt is when the reading was taken, in UTC.
	RecordedAt time.Time
	// CPUPercent, MemoryPercent and DiskPercent are percentages in the 0-100
	// range.
	CPUPercent    float64
	MemoryPercent float64
	DiskPercent   float64
	// BatteryPercent is nil when the machine has no battery, which is
	// meaningfully different from a battery reading zero percent.
	BatteryPercent *float64
	// NetworkConnected reports whether the backend was reachable.
	NetworkConnected bool
	// Uptime is how long the machine has been running.
	Uptime time.Duration
	// OperatingSystem is the Go platform name, for example "linux".
	OperatingSystem string
	// AgentHealth is one of the AgentHealth constants.
	AgentHealth string

	// The Measured flags record which optional readings succeeded, so agent
	// health can be reported honestly instead of always claiming "healthy".
	CPUMeasured    bool
	MemoryMeasured bool
}

// Collector gathers metrics, holding the previous CPU counters needed to derive
// a utilisation percentage from two samples.
//
// A Collector is not safe for concurrent use; the agent calls it from its single
// collection loop.
type Collector struct {
	previous        cpuTimes
	hasPrevious     bool
	probe           Probe
	operatingSystem string
}

// NewCollector builds a Collector. Passing a nil probe leaves NetworkConnected
// permanently false, which is the honest answer when no probe is configured.
func NewCollector(operatingSystem string, probe Probe) *Collector {
	return &Collector{probe: probe, operatingSystem: operatingSystem}
}

// Collect takes a fresh reading. It always returns a usable Sample: individual
// collectors that fail simply leave their field at zero.
//
// The first call cannot report CPU utilisation, because a percentage requires
// the difference between two counter readings. That call establishes the
// baseline and reports zero CPU, which is correct rather than a failure.
func (c *Collector) Collect(ctx context.Context) Sample {
	sample := Sample{
		RecordedAt:      time.Now().UTC(),
		OperatingSystem: c.operatingSystem,
	}

	percent, current, measured := sampleCPU(c.previous, c.hasPrevious)
	c.previous, c.hasPrevious = current, true
	if measured {
		sample.CPUPercent = percent
		sample.CPUMeasured = true
	}

	if totalKB, availableKB, ok := readMemoryInfo(); ok {
		if memory, ok := memoryPercent(totalKB, availableKB); ok {
			sample.MemoryPercent = memory
			sample.MemoryMeasured = true
		}
	}

	if disk, ok := readDiskPercent(); ok {
		sample.DiskPercent = disk
	}

	if battery, ok := readBatteryPercent(); ok {
		sample.BatteryPercent = &battery
	}

	if c.probe != nil {
		sample.NetworkConnected = c.probe(ctx)
	}

	if uptime, ok := system.Uptime(); ok {
		sample.Uptime = uptime
	}

	sample.AgentHealth = reportAgentHealth(sample)
	return sample
}

// reportAgentHealth describes the agent's ability to observe the machine. If
// neither CPU nor memory could be read, the agent is not seeing enough to be
// trusted, and says so rather than reporting a cheerful "healthy".
func reportAgentHealth(sample Sample) string {
	if !sample.CPUMeasured && !sample.MemoryMeasured {
		return AgentHealthDegraded
	}
	return AgentHealthHealthy
}

// TCPProbe returns a Probe that considers the machine connected when a TCP
// connection to address succeeds. Dialling the backend is the most honest test
// available: it proves the path the heartbeat actually needs, rather than just
// whether a network interface is up.
func TCPProbe(address string, timeout time.Duration) Probe {
	return func(ctx context.Context) bool {
		if address == "" {
			return false
		}
		dialCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		var dialer net.Dialer
		connection, err := dialer.DialContext(dialCtx, "tcp", address)
		if err != nil {
			return false
		}
		_ = connection.Close()
		return true
	}
}

// Thresholds holds the configured warning levels. README section 17 requires
// these to be configurable and to generate warnings rather than confirmed
// faults.
type Thresholds struct {
	CPUPercent     float64
	MemoryPercent  float64
	DiskPercent    float64
	BatteryPercent float64
}

// DefaultThresholds returns the initial warning levels.
//
// Battery is a low-water mark, so it warns when the reading falls *below* it;
// the other three warn when the reading rises above them.
func DefaultThresholds() Thresholds {
	return Thresholds{
		CPUPercent:     90,
		MemoryPercent:  90,
		DiskPercent:    90,
		BatteryPercent: 15,
	}
}

// Breach describes a single threshold crossing.
type Breach struct {
	// Metric is a stable identifier suitable for the event payload.
	Metric string
	// Value is the reading that crossed the threshold.
	Value float64
	// Threshold is the configured limit that was crossed.
	Threshold float64
}

// Metric identifiers used in health warning payloads.
const (
	MetricCPU     = "cpu_percent"
	MetricMemory  = "memory_percent"
	MetricDisk    = "disk_percent"
	MetricBattery = "battery_percent"
)

// Breaches returns every threshold the sample crosses. A zero threshold means
// "not configured" and is skipped, which lets a deployment disable a check by
// setting it to zero.
func (t Thresholds) Breaches(sample Sample) []Breach {
	var breaches []Breach

	if t.CPUPercent > 0 && sample.CPUMeasured && sample.CPUPercent >= t.CPUPercent {
		breaches = append(breaches, Breach{MetricCPU, sample.CPUPercent, t.CPUPercent})
	}
	if t.MemoryPercent > 0 && sample.MemoryMeasured && sample.MemoryPercent >= t.MemoryPercent {
		breaches = append(breaches, Breach{MetricMemory, sample.MemoryPercent, t.MemoryPercent})
	}
	if t.DiskPercent > 0 && sample.DiskPercent >= t.DiskPercent {
		breaches = append(breaches, Breach{MetricDisk, sample.DiskPercent, t.DiskPercent})
	}
	if t.BatteryPercent > 0 && sample.BatteryPercent != nil && *sample.BatteryPercent <= t.BatteryPercent {
		breaches = append(breaches, Breach{MetricBattery, *sample.BatteryPercent, t.BatteryPercent})
	}

	return breaches
}
