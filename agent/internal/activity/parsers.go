package activity

import (
	"strconv"
	"strings"
	"time"
)

// This file holds the pure parsers for idle time. They are untagged so the
// formats can be tested on any machine.

// parseWIdle extracts the shortest idle time from w(1) output.
//
// The shortest value is used rather than the longest or an average because the
// question being answered is whether anyone is using the computer: if any
// session shows recent activity, somebody is present.
//
// w reports idle time in one of three shapes, which is why this is a best-effort
// fallback rather than the primary source:
//
//	0.00s   less than a minute, as seconds with a trailing "s"
//	5:23    minutes and seconds
//	2:52m   hours and minutes, marked with a trailing "m"
//
// The column position assumed here is the standard layout produced by
// procps-ng: USER, TTY, FROM, LOGIN@, IDLE, JCPU, PCPU, WHAT.
func parseWIdle(output string) (time.Duration, bool) {
	var shortest time.Duration
	found := false

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		idle, ok := parseIdleField(fields[4])
		if !ok {
			continue
		}
		if !found || idle < shortest {
			shortest, found = idle, true
		}
	}
	return shortest, found
}

// parseIdleField converts a single w(1) IDLE cell into a duration.
func parseIdleField(value string) (time.Duration, bool) {
	switch value {
	case "", ".", "-":
		// No session has been idle at all, or the field is not reported.
		return 0, true
	case "old":
		// w uses "old" for sessions idle for more than a day.
		return 24 * time.Hour, true
	}

	if strings.HasSuffix(value, "m") {
		// Hours and minutes, for example "2:52m".
		return parseColonClock(strings.TrimSuffix(value, "m"), time.Hour, time.Minute)
	}
	if strings.HasSuffix(value, "s") {
		// Seconds with a fractional part, for example "0.00s".
		seconds, err := strconv.ParseFloat(strings.TrimSuffix(value, "s"), 64)
		if err != nil || seconds < 0 {
			return 0, false
		}
		return time.Duration(seconds * float64(time.Second)), true
	}
	// Minutes and seconds, for example "5:23".
	return parseColonClock(value, time.Minute, time.Second)
}

// parseColonClock converts a "first:second" pair into a duration, scaling each
// half by the supplied unit.
func parseColonClock(value string, firstUnit, secondUnit time.Duration) (time.Duration, bool) {
	first, second, found := strings.Cut(value, ":")
	if !found {
		return 0, false
	}
	high, err := strconv.Atoi(strings.TrimSpace(first))
	if err != nil || high < 0 {
		return 0, false
	}
	low, err := strconv.Atoi(strings.TrimSpace(second))
	if err != nil || low < 0 {
		return 0, false
	}
	return time.Duration(high)*firstUnit + time.Duration(low)*secondUnit, true
}

// parseXprintidle converts the millisecond value printed by xprintidle into a
// duration.
func parseXprintidle(output string) (time.Duration, bool) {
	milliseconds, err := strconv.ParseInt(strings.TrimSpace(output), 10, 64)
	if err != nil || milliseconds < 0 {
		return 0, false
	}
	return time.Duration(milliseconds) * time.Millisecond, true
}

// parseSingleInteger reads a whole number of milliseconds, which is the shape
// the Windows GetLastInputInfo collector produces.
func parseSingleInteger(output string) (int64, bool) {
	value, err := strconv.ParseInt(strings.TrimSpace(output), 10, 64)
	if err != nil || value < 0 {
		return 0, false
	}
	return value, true
}

// parseIOHIDIdleTime extracts the HIDIdleTime value that the macOS IOKit
// registry reports, which is expressed in nanoseconds. The line looks like:
//
//	"HIDIdleTime" = 12345678900
func parseIOHIDIdleTime(output string) (time.Duration, bool) {
	for _, line := range strings.Split(output, "\n") {
		_, value, found := strings.Cut(line, "HIDIdleTime")
		if !found {
			continue
		}
		_, value, found = strings.Cut(value, "=")
		if !found {
			continue
		}
		nanoseconds, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || nanoseconds < 0 {
			continue
		}
		return time.Duration(nanoseconds), true
	}
	return 0, false
}
