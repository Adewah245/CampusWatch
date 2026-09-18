//go:build linux

package activity

import (
	"os/exec"
	"time"
)

// idleDuration reports how long the machine has been idle.
//
// Two sources are tried in order, because neither works everywhere:
//
//  1. xprintidle, when present, asks the X server for the time since the last
//     keyboard or mouse event. This is the accurate answer on an X11 desktop,
//     but it is an optional utility and does not exist on headless machines or
//     under Wayland.
//
//  2. w(1) reports the idle time of each terminal session. This is always
//     available, but it is a weaker signal: an X session's terminal records
//     almost no activity, so a user typing in a graphical application may still
//     appear idle. It is used only as a fallback, and the limitation is recorded
//     in the agent README.
//
// When neither source can answer, the result is reported as unavailable and the
// agent emits no activity transitions rather than inventing them.
func idleDuration() (time.Duration, bool) {
	if idle, ok := x11IdleDuration(); ok {
		return idle, true
	}
	return wIdleDuration()
}

// x11IdleDuration asks the X server for the idle time via xprintidle.
func x11IdleDuration() (time.Duration, bool) {
	path, err := exec.LookPath("xprintidle")
	if err != nil {
		return 0, false
	}
	output, err := exec.Command(path).Output()
	if err != nil {
		return 0, false
	}
	return parseXprintidle(string(output))
}

// wIdleDuration derives idle time from the standard w(1) listing.
func wIdleDuration() (time.Duration, bool) {
	output, err := exec.Command("w", "-h").Output()
	if err != nil {
		return 0, false
	}
	return parseWIdle(string(output))
}
