//go:build darwin

package activity

import (
	"os/exec"
	"time"
)

// idleDuration reports how long the machine has been idle, read from the
// HIDIdleTime value that IOKit publishes for the human interface subsystem.
//
// This is the same value macOS itself uses to decide when to dim the display,
// so it reflects genuine keyboard and pointing device activity rather than
// terminal activity.
func idleDuration() (time.Duration, bool) {
	output, err := exec.Command("ioreg", "-c", "IOHIDSystem").Output()
	if err != nil {
		return 0, false
	}
	return parseIOHIDIdleTime(string(output))
}
