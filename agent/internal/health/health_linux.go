//go:build linux

package health

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Linux reports every metric this package collects through the kernel's virtual
// filesystems, so no external program is executed and no privileges beyond
// reading /proc and /sys are needed.

// sampleCPU derives CPU utilisation from the cumulative counters in /proc/stat.
// The previous reading is required because the file only publishes totals since
// boot, never an instantaneous percentage.
func sampleCPU(previous cpuTimes, hasPrevious bool) (float64, cpuTimes, bool) {
	content, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, previous, false
	}
	current, ok := parseProcStat(string(content))
	if !ok {
		return 0, previous, false
	}
	if !hasPrevious {
		return 0, current, false
	}
	percent, ok := cpuPercent(previous, current)
	return percent, current, ok
}

// readMemoryInfo reads total and available memory in kilobytes from
// /proc/meminfo.
func readMemoryInfo() (float64, float64, bool) {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, false
	}
	return parseProcMeminfo(string(content))
}

// readDiskPercent reports space used on the root filesystem.
//
// Free space is taken from Bavail, the blocks available to an unprivileged
// process, rather than Bfree. The difference is the reserved block pool that
// only root may use; counting it as free would understate usage and delay a
// genuinely needed disk warning.
func readDiskPercent() (float64, bool) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0, false
	}

	total := stat.Blocks * uint64(stat.Bsize)
	if total == 0 {
		return 0, false
	}
	available := stat.Bavail * uint64(stat.Bsize)
	if available > total {
		available = total
	}
	return clampPercent(float64(total-available) / float64(total) * 100), true
}

// readBatteryPercent reports battery charge where a battery exists.
//
// Kernels expose batteries as /sys/class/power_supply/BAT* on most laptops, but
// some drivers use other names, so the fallback scans every power supply and
// selects the one whose type is "Battery" rather than guessing from the name.
func readBatteryPercent() (float64, bool) {
	for _, candidate := range batteryCapacityPaths() {
		content, err := os.ReadFile(candidate)
		if err != nil {
			continue
		}
		if percent, ok := parseBatteryCapacity(string(content)); ok {
			return percent, true
		}
	}
	return 0, false
}

// batteryCapacityPaths returns candidate capacity files, most likely first.
func batteryCapacityPaths() []string {
	paths, _ := filepath.Glob("/sys/class/power_supply/BAT*/capacity")

	entries, err := filepath.Glob("/sys/class/power_supply/*")
	if err != nil {
		return paths
	}
	for _, entry := range entries {
		kind, err := os.ReadFile(filepath.Join(entry, "type"))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(kind)) != "Battery" {
			continue
		}
		paths = append(paths, filepath.Join(entry, "capacity"))
	}
	return paths
}
