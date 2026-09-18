//go:build darwin

package health

import "os/exec"

// macOS publishes most metrics through sysctl and a few purpose-built tools,
// none of which require elevated privileges.

// sampleCPU derives CPU utilisation from the cumulative tick counters reported
// by kern.cp_time. As on Linux, the counters are totals since boot, so a
// percentage needs two readings.
func sampleCPU(previous cpuTimes, hasPrevious bool) (float64, cpuTimes, bool) {
	output, err := exec.Command("sysctl", "-n", "kern.cp_time").Output()
	if err != nil {
		return 0, previous, false
	}
	current, ok := parseSysctlCPTime(string(output))
	if !ok {
		return 0, previous, false
	}
	if !hasPrevious {
		return 0, current, false
	}
	percent, ok := cpuPercent(previous, current)
	return percent, current, ok
}

// readMemoryInfo combines the installed memory reported by hw.memsize with the
// available page counts from vm_stat.
func readMemoryInfo() (float64, float64, bool) {
	sizeOutput, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 0, 0, false
	}
	totalBytes, ok := parseSingleNumber(string(sizeOutput))
	if !ok || totalBytes <= 0 {
		return 0, 0, false
	}

	statOutput, err := exec.Command("vm_stat").Output()
	if err != nil {
		return 0, 0, false
	}
	availableKB, _, ok := parseVMStat(string(statOutput))
	if !ok {
		return 0, 0, false
	}

	return totalBytes / 1024, availableKB, true
}

// readDiskPercent reports space used on the root volume.
func readDiskPercent() (float64, bool) {
	output, err := exec.Command("df", "-k", "/").Output()
	if err != nil {
		return 0, false
	}
	usedKB, totalKB, ok := parseDF(string(output))
	if !ok {
		return 0, false
	}
	return clampPercent(usedKB / totalKB * 100), true
}

// readBatteryPercent reports battery charge on portable Macs.
func readBatteryPercent() (float64, bool) {
	output, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return 0, false
	}
	return parsePMsetBattery(string(output))
}
