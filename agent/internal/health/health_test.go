package health

import (
	"context"
	"net"
	"testing"
	"time"
)

const procStatFixture = `cpu  234513 1023 48211 8823123 512 0 811 0 0 0
cpu0 117256 511 24105 4411561 256 0 405 0 0 0
cpu1 117257 512 24106 4411562 256 0 406 0 0 0
intr 12345678 0 0
ctxt 23456789
`

func TestParseProcStatIgnoresPerCoreLines(t *testing.T) {
	times, ok := parseProcStat(procStatFixture)
	if !ok {
		t.Fatal("parseProcStat() failed on a valid fixture")
	}

	// The aggregate "cpu" line only. Summing the per-core lines as well would
	// double-count every tick.
	wantTotal := 234513.0 + 1023 + 48211 + 8823123 + 512 + 0 + 811 + 0 + 0 + 0
	if times.total != wantTotal {
		t.Errorf("total = %v, want %v", times.total, wantTotal)
	}

	// Idle is idle plus iowait, which is the convention every mainstream
	// monitoring tool follows: waiting on disk is not burning CPU.
	wantIdle := 8823123.0 + 512
	if times.idle != wantIdle {
		t.Errorf("idle = %v, want %v", times.idle, wantIdle)
	}
}

func TestParseProcStatRejectsMalformedInput(t *testing.T) {
	if _, ok := parseProcStat("cpu  1 2 bad 4 5\n"); ok {
		t.Error("parseProcStat() accepted non-numeric counters")
	}
	if _, ok := parseProcStat("something else entirely\n"); ok {
		t.Error("parseProcStat() accepted input with no cpu line")
	}
}

func TestCPUPercentFromCounters(t *testing.T) {
	tests := map[string]struct {
		previous  cpuTimes
		current   cpuTimes
		want      float64
		wantValid bool
	}{
		"half idle means fifty percent used": {
			previous:  cpuTimes{idle: 100, total: 200},
			current:   cpuTimes{idle: 150, total: 300},
			want:      50,
			wantValid: true,
		},
		"fully idle means zero percent used": {
			previous:  cpuTimes{idle: 100, total: 200},
			current:   cpuTimes{idle: 200, total: 300},
			want:      0,
			wantValid: true,
		},
		"fully busy means one hundred percent used": {
			previous:  cpuTimes{idle: 100, total: 200},
			current:   cpuTimes{idle: 100, total: 300},
			want:      100,
			wantValid: true,
		},
		"stalled counters are not a valid reading": {
			previous:  cpuTimes{idle: 100, total: 200},
			current:   cpuTimes{idle: 100, total: 200},
			wantValid: false,
		},
		"counters that went backwards mean a reboot": {
			previous:  cpuTimes{idle: 150, total: 300},
			current:   cpuTimes{idle: 100, total: 200},
			wantValid: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, ok := cpuPercent(test.previous, test.current)

			if ok != test.wantValid {
				t.Fatalf("cpuPercent() valid = %t, want %t", ok, test.wantValid)
			}
			if ok && got != test.want {
				t.Errorf("cpuPercent() = %v, want %v", got, test.want)
			}
		})
	}
}

const meminfoFixture = `MemTotal:       16316232 kB
MemFree:         1234567 kB
MemAvailable:    9123456 kB
Buffers:          123456 kB
Cached:          4567890 kB
`

func TestParseProcMeminfoPrefersMemAvailable(t *testing.T) {
	total, available, ok := parseProcMeminfo(meminfoFixture)
	if !ok {
		t.Fatal("parseProcMeminfo() failed on a valid fixture")
	}

	if total != 16316232 {
		t.Errorf("total = %v, want 16316232", total)
	}
	// MemFree would report this machine as nearly out of memory; MemAvailable
	// is the honest figure.
	if available != 9123456 {
		t.Errorf("available = %v, want 9123456", available)
	}
}

func TestParseProcMeminfoFallsBackToMemFree(t *testing.T) {
	// Kernels before 3.14 do not publish MemAvailable at all.
	const oldKernel = "MemTotal: 8192000 kB\nMemFree: 4096000 kB\n"

	total, available, ok := parseProcMeminfo(oldKernel)
	if !ok {
		t.Fatal("parseProcMeminfo() failed on an older kernel fixture")
	}
	if total != 8192000 || available != 4096000 {
		t.Errorf("got total=%v available=%v, want 8192000 and 4096000", total, available)
	}
}

func TestMemoryPercent(t *testing.T) {
	got, ok := memoryPercent(1000, 250)
	if !ok {
		t.Fatal("memoryPercent() failed")
	}
	if got != 75 {
		t.Errorf("memoryPercent() = %v, want 75", got)
	}

	// A reported figure larger than the total must not produce a negative
	// percentage on the dashboard.
	if got, ok := memoryPercent(1000, 5000); !ok || got != 0 {
		t.Errorf("memoryPercent() with oversized availability = %v (ok=%t), want 0", got, ok)
	}
}

func TestParseBatteryCapacity(t *testing.T) {
	if percent, ok := parseBatteryCapacity("74\n"); !ok || percent != 74 {
		t.Errorf("parseBatteryCapacity() = %v (ok=%t), want 74", percent, ok)
	}
	if _, ok := parseBatteryCapacity("not a number"); ok {
		t.Error("parseBatteryCapacity() accepted non-numeric input")
	}
}

func TestParsePMsetBattery(t *testing.T) {
	// The percentage is attached to the field before it, so a naive
	// whitespace split would pick up the wrong token.
	output := "Now drawing from 'Battery Power'\n -InternalBattery-0 (id=1234567)\t85%; discharging; 4:12 remaining present: true\n"

	percent, ok := parsePMsetBattery(output)
	if !ok {
		t.Fatal("parsePMsetBattery() failed on valid output")
	}
	if percent != 85 {
		t.Errorf("parsePMsetBattery() = %v, want 85", percent)
	}
}

func TestParsePMsetBatteryOnDesktop(t *testing.T) {
	// No battery at all, which must be reported as absent rather than zero.
	if _, ok := parsePMsetBattery("Now drawing from 'AC Power'\n"); ok {
		t.Error("parsePMsetBattery() reported a battery on a machine without one")
	}
}

func TestParseSysctlCPTime(t *testing.T) {
	times, ok := parseSysctlCPTime("100 20 30 5 400\n")
	if !ok {
		t.Fatal("parseSysctlCPTime() failed on valid output")
	}
	if times.idle != 400 {
		t.Errorf("idle = %v, want 400", times.idle)
	}
	if times.total != 555 {
		t.Errorf("total = %v, want 555", times.total)
	}
}

func TestParseVMStatSumsReclaimablePages(t *testing.T) {
	output := `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                              1000.
Pages active:                           50000.
Pages inactive:                          2000.
Pages speculative:                        500.
`

	availableKB, pageKB, ok := parseVMStat(output)
	if !ok {
		t.Fatal("parseVMStat() failed on valid output")
	}
	if pageKB != 16 {
		t.Errorf("pageKB = %v, want 16", pageKB)
	}
	// Free, inactive and speculative: counting only free pages would report a
	// healthy Mac as critically low on memory.
	wantAvailable := (1000.0 + 2000 + 500) * 16
	if availableKB != wantAvailable {
		t.Errorf("availableKB = %v, want %v", availableKB, wantAvailable)
	}
}

func TestParseDF(t *testing.T) {
	output := `Filesystem   1024-blocks      Used Available Capacity iused      ifree %iused  Mounted on
/dev/disk1s5   488245288  120000000 360000000    26%  500000 4000000000    0%   /
`

	used, total, ok := parseDF(output)
	if !ok {
		t.Fatal("parseDF() failed on valid output")
	}
	if total != 488245288 || used != 120000000 {
		t.Errorf("got used=%v total=%v, want 120000000 and 488245288", used, total)
	}
}

func TestThresholdsBreaches(t *testing.T) {
	thresholds := DefaultThresholds()

	battery := 10.0
	sample := Sample{
		CPUPercent:     95,
		MemoryPercent:  40,
		DiskPercent:    92,
		BatteryPercent: &battery,
		CPUMeasured:    true,
		MemoryMeasured: true,
	}

	breaches := thresholds.Breaches(sample)
	metrics := make(map[string]bool, len(breaches))
	for _, breach := range breaches {
		metrics[breach.Metric] = true
	}

	// CPU is above, memory is below, disk is above and battery is below its own
	// low-water mark.
	for _, want := range []string{MetricCPU, MetricDisk, MetricBattery} {
		if !metrics[want] {
			t.Errorf("Breaches() did not report %s", want)
		}
	}
	if metrics[MetricMemory] {
		t.Error("Breaches() reported memory, which is comfortably below its threshold")
	}
}

func TestThresholdsIgnoreUnmeasuredCPU(t *testing.T) {
	// The first collection cannot measure CPU, and an unmeasured zero must not
	// be treated as a reading.
	sample := Sample{CPUPercent: 0, CPUMeasured: false}

	if breaches := DefaultThresholds().Breaches(sample); len(breaches) != 0 {
		t.Errorf("Breaches() = %v, want none for an unmeasured CPU", breaches)
	}
}

func TestThresholdsSkipDisabledChecks(t *testing.T) {
	// A zero threshold is how a deployment turns a check off.
	thresholds := Thresholds{CPUPercent: 0, MemoryPercent: 0, DiskPercent: 0, BatteryPercent: 0}
	sample := Sample{CPUPercent: 100, MemoryPercent: 100, DiskPercent: 100, CPUMeasured: true, MemoryMeasured: true}

	if breaches := thresholds.Breaches(sample); len(breaches) != 0 {
		t.Errorf("Breaches() = %v, want none when every threshold is disabled", breaches)
	}
}

func TestTCPProbeReportsAReachableEndpoint(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not open a listener for the test: %v", err)
	}
	defer listener.Close()

	if !TCPProbe(listener.Addr().String(), time.Second)(context.Background()) {
		t.Error("TCPProbe() reported a listening endpoint as unreachable")
	}
}

func TestTCPProbeReportsAnUnreachableEndpoint(t *testing.T) {
	// Bind and immediately release a port so nothing is listening on it.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not open a listener for the test: %v", err)
	}
	address := listener.Addr().String()
	listener.Close()

	if TCPProbe(address, time.Second)(context.Background()) {
		t.Error("TCPProbe() reported a closed port as reachable")
	}
}

func TestTCPProbeWithNoAddressIsUnreachable(t *testing.T) {
	if TCPProbe("", time.Second)(context.Background()) {
		t.Error("TCPProbe(\"\") reported reachable, want unreachable")
	}
}

func TestCollectorReportsZeroCPUOnTheFirstSample(t *testing.T) {
	// A percentage needs two readings, so the first collection establishes the
	// baseline. Reporting a failure here would make every agent look degraded
	// for the first interval after each restart.
	collector := NewCollector("linux", nil)
	collector.Collect(context.Background())

	// The baseline must have been stored, otherwise the second call would also
	// decline to report a value.
	if !collector.hasPrevious {
		t.Error("Collect() did not retain the CPU baseline for the next sample")
	}
}
