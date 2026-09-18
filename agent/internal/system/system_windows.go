//go:build windows

package system

import (
	"os/exec"
	"time"
)

// Uptime reports how long the machine has been running, derived from the last
// boot timestamp Windows records for the operating system.
func Uptime() (time.Duration, bool) {
	output, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command",
		`(Get-CimInstance Win32_OperatingSystem).LastBootUpTime.ToString("yyyy-MM-ddTHH:mm:ss")`).Output()
	if err != nil {
		return 0, false
	}
	bootTime, ok := parseWindowsBootTime(string(output))
	if !ok {
		return 0, false
	}
	elapsed := time.Since(bootTime)
	if elapsed < 0 {
		return 0, false
	}
	return elapsed, true
}
