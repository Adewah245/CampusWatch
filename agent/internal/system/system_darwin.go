//go:build darwin

package system

import (
	"os/exec"
	"time"
)

// Uptime reports how long the machine has been running, derived from the kernel
// boot timestamp exposed by sysctl.
func Uptime() (time.Duration, bool) {
	output, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
	if err != nil {
		return 0, false
	}
	bootTime, ok := parseSysctlBootTime(string(output))
	if !ok {
		return 0, false
	}
	elapsed := time.Since(bootTime)
	if elapsed < 0 {
		// The clock moved backwards (an NTP correction, for example), so the
		// derived uptime would be meaningless.
		return 0, false
	}
	return elapsed, true
}
