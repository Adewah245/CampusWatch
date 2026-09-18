// Package identity describes the computer the agent is running on.
//
// Everything here is read-only discovery of information the operating system
// already publishes. Nothing in this package inspects user data, and per README
// sections 6 and 7 the agent never reads browser cookies, saved passwords,
// private files or any other personal content to work out who is using a
// machine.
package identity

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Device types inferred from hardware.
const (
	DeviceTypeDesktop = "Desktop"
	DeviceTypeLaptop  = "Laptop"
	DeviceTypeUnknown = "Unknown"
)

// Device is the agent's view of the monitored computer.
type Device struct {
	// Hostname is the machine's network name. It is descriptive only; README
	// section 11 warns that hostnames can change, so it must never be treated
	// as the system's stable identity.
	Hostname string
	// OperatingSystem is the Go platform name, for example "linux".
	OperatingSystem string
	// OSVersion is a human readable distribution or release name.
	OSVersion string
	// DeviceType is one of the DeviceType constants.
	DeviceType string
	// MachineID is the operating system's own machine identifier. It is more
	// stable than a hostname, but it is still not used as the CampusWatch
	// identity: the backend issues that, and the agent persists a local
	// installation id (see internal/system) for its own state.
	MachineID string
}

// Detect gathers the device identity. Every lookup degrades to an empty string
// or DeviceTypeUnknown rather than failing, so a partially understood platform
// can still report a heartbeat.
func Detect() Device {
	device := Device{
		OperatingSystem: runtime.GOOS,
		OSVersion:       osVersion(),
		MachineID:       machineID(),
		DeviceType:      detectDeviceType(),
	}
	if hostname, err := os.Hostname(); err == nil {
		device.Hostname = strings.TrimSpace(hostname)
	}
	return device
}

// Summary renders the identity as a single log line.
func (d Device) Summary() string {
	parts := []string{
		"hostname=" + orUnknown(d.Hostname),
		"os=" + orUnknown(d.OperatingSystem),
		"os_version=" + orUnknown(d.OSVersion),
		"device_type=" + orUnknown(d.DeviceType),
		"machine_id=" + orUnknown(d.MachineID),
	}
	return strings.Join(parts, " ")
}

// detectDeviceType classifies the machine using battery presence, which is the
// only portable hardware signal available without extra dependencies. Where the
// platform cannot answer, the type is reported as Unknown rather than guessed.
func detectDeviceType() string {
	count, supported := batteryCount()
	if !supported {
		return DeviceTypeUnknown
	}
	if count > 0 {
		return DeviceTypeLaptop
	}
	return DeviceTypeDesktop
}

func orUnknown(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}

// commandTimeout bounds the platform helper processes used on Windows and
// macOS. A hung shell-out must never stall the heartbeat loop.
const commandTimeout = 5 * time.Second

// runCommand executes a helper program and returns its trimmed stdout. It is
// used only by the Windows and macOS implementations; Linux reads /proc and
// /sys directly and shells out to nothing.
func runCommand(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	output, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
