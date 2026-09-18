//go:build linux

package identity

import (
	"os"
	"path/filepath"
	"strings"
)

// osVersion reports the Linux distribution name, for example
// "Ubuntu 24.04.1 LTS", read from /etc/os-release.
func osVersion() string {
	content, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	return parseOSRelease(string(content))
}

// machineID reads the machine identifier systemd and D-Bus publish, preferring
// /etc/machine-id and falling back to the older D-Bus location.
func machineID() string {
	for _, path := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if id := strings.TrimSpace(string(content)); id != "" {
			return id
		}
	}
	return ""
}

// batteryCount counts battery devices exposed by the kernel. The second return
// value reports whether the platform can answer the question at all, which
// distinguishes "definitely no battery" (a desktop) from "cannot tell".
func batteryCount() (int, bool) {
	entries, err := filepath.Glob("/sys/class/power_supply/*")
	if err != nil || entries == nil {
		return 0, false
	}

	count := 0
	for _, entry := range entries {
		// Only devices whose type is literally "Battery" count. Mains adapters
		// and USB power supplies appear in the same directory and must not be
		// mistaken for a battery.
		kind, err := os.ReadFile(filepath.Join(entry, "type"))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(kind)) == "Battery" {
			count++
		}
	}
	return count, true
}
