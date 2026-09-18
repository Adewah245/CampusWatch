//go:build windows

package identity

import (
	"strconv"
	"strings"
)

// RunPowerShell runs a PowerShell expression and returns its standard output.
//
// PowerShell is used instead of a native syscall dependency so the agent can be
// cross-compiled as a single static binary with no external modules and no CGO,
// which keeps large-scale deployment simple (README section 25). The queries
// below are read-only CIM lookups of information the operating system already
// exposes.
//
// It is exported so the health collectors can reuse the same execution policy.
func RunPowerShell(script string) (string, error) {
	return runCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
}

// osVersion reports the Windows edition and version, for example
// "Microsoft Windows 11 Pro 10.0.22631".
func osVersion() string {
	output, err := RunPowerShell(
		`$os = Get-CimInstance Win32_OperatingSystem; "{0} {1} {2}" -f $os.Caption, $os.Version, $os.BuildNumber`)
	if err != nil {
		return ""
	}
	return strings.Join(strings.Fields(output), " ")
}

// machineID reports the SMBIOS product UUID. Windows keeps this stable across
// reboots, which makes it a better anchor than a hostname or IP address.
func machineID() string {
	output, err := RunPowerShell(`(Get-CimInstance Win32_ComputerSystemProduct).UUID`)
	if err != nil {
		return ""
	}
	id := strings.TrimSpace(output)
	// Some firmware reports all-zero or all-F placeholders instead of a real
	// UUID; treat those as absent rather than as an identity.
	if id == "" || strings.Trim(id, "0") == "" || strings.Trim(id, "-F") == "" {
		return ""
	}
	return id
}

// batteryCount counts installed batteries. A laptop reports at least one; a
// desktop reports none.
func batteryCount() (int, bool) {
	output, err := RunPowerShell(`@(Get-CimInstance Win32_Battery).Count`)
	if err != nil {
		return 0, false
	}
	count, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0, false
	}
	return count, true
}
