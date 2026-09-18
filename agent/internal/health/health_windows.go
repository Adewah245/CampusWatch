//go:build windows

package health

import "CampusWatch/agent/internal/identity"

// Windows exposes its performance counters through CIM. Collecting them via
// PowerShell keeps the agent a single static binary with no CGO and no external
// Go modules, which is what makes unattended deployment to a large fleet
// practical (README section 25).

// The formatted performance counter already reports a percentage, so unlike
// Linux and macOS there is no pair of cumulative readings to difference. The
// previous reading is accepted only to satisfy the shared collector signature.
const cpuQueryScript = `(Get-CimInstance Win32_PerfFormattedData_PerfOS_Processor | Where-Object { $_.Name -eq '_Total' }).PercentProcessorTime`

// sampleCPU reports CPU utilisation from the total processor counter.
func sampleCPU(previous cpuTimes, hasPrevious bool) (float64, cpuTimes, bool) {
	output, err := identity.RunPowerShell(cpuQueryScript)
	if err != nil {
		return 0, previous, false
	}
	percent, ok := parseSingleNumber(output)
	if !ok {
		return 0, previous, false
	}
	return clampPercent(percent), previous, true
}

// readMemoryInfo reports total and available physical memory in kilobytes.
//
// FreePhysicalMemory is used directly rather than deriving "available" from the
// standby list, which is the closest equivalent Windows offers to the
// MemAvailable figure Linux publishes.
func readMemoryInfo() (float64, float64, bool) {
	output, err := identity.RunPowerShell(
		`$os = Get-CimInstance Win32_OperatingSystem; "{0} {1}" -f $os.TotalVisibleMemorySize, $os.FreePhysicalMemory`)
	if err != nil {
		return 0, 0, false
	}
	totalKB, availableKB, ok := parseTwoNumbers(output)
	if !ok || totalKB <= 0 {
		return 0, 0, false
	}
	return totalKB, availableKB, true
}

// readDiskPercent reports space used on the system drive, which is where the
// operating system and installed software live.
func readDiskPercent() (float64, bool) {
	output, err := identity.RunPowerShell(
		`$d = Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='C:'"; "{0} {1}" -f $d.Size, $d.FreeSpace`)
	if err != nil {
		return 0, false
	}

	total, free, ok := parseTwoNumbers(output)
	if !ok || total <= 0 {
		return 0, false
	}
	if free > total {
		free = total
	}
	return clampPercent((total - free) / total * 100), true
}

// readBatteryPercent reports battery charge on machines that have a battery.
//
// Win32_Battery is absent on desktops, which is reported as "no battery" rather
// than as an error, matching how the battery field is treated everywhere else.
func readBatteryPercent() (float64, bool) {
	output, err := identity.RunPowerShell(
		`(Get-CimInstance Win32_Battery | Select-Object -First 1).EstimatedChargeRemaining`)
	if err != nil {
		return 0, false
	}
	percent, ok := parseSingleNumber(output)
	if !ok {
		return 0, false
	}
	return clampPercent(percent), true
}
