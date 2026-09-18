//go:build darwin

package identity

import (
	"strings"
)

// osVersion reports the macOS product name and version, for example
// "macOS 14.5", read from sw_vers.
func osVersion() string {
	output, err := runCommand("sw_vers")
	if err != nil {
		return ""
	}
	return parseSwVers(output)
}

// machineID reports the platform UUID published by the IOKit registry.
func machineID() string {
	output, err := runCommand("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	if err != nil {
		return ""
	}
	return parseIOPlatformUUID(output)
}

// batteryCount reports whether the machine has an internal battery, which on
// macOS distinguishes a MacBook from a desktop Mac.
func batteryCount() (int, bool) {
	output, err := runCommand("pmset", "-g", "batt")
	if err != nil {
		return 0, false
	}
	if strings.Contains(output, "InternalBattery") {
		return 1, true
	}
	return 0, true
}
