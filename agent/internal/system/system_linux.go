//go:build linux

package system

import (
	"os"
	"time"
)

// Uptime reports how long the machine has been running, read from /proc/uptime.
// The kernel maintains this counter, so no privileges are required.
func Uptime() (time.Duration, bool) {
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, false
	}
	return parseProcUptime(string(content))
}
