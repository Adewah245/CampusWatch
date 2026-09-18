package system

import (
	"strconv"
	"strings"
	"time"
)

// This file holds the pure parsers backing Uptime on each platform. Keeping
// them untagged means the formats can be tested on any machine, not only on the
// operating system that produces them.

// parseProcUptime reads the first field of /proc/uptime, which is the number of
// seconds since boot expressed as a floating point value.
func parseProcUptime(content string) (time.Duration, bool) {
	fields := strings.Fields(strings.TrimSpace(content))
	if len(fields) == 0 {
		return 0, false
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || seconds < 0 {
		return 0, false
	}
	return time.Duration(seconds * float64(time.Second)), true
}

// parseSysctlBootTime reads the kern.boottime value printed by sysctl, which
// looks like: { sec = 1719999999, usec = 0 } Tue Jul 2 12:00:00 2024
func parseSysctlBootTime(output string) (time.Time, bool) {
	_, rest, ok := strings.Cut(output, "sec =")
	if !ok {
		return time.Time{}, false
	}
	value, _, ok := strings.Cut(rest, ",")
	if !ok {
		return time.Time{}, false
	}
	seconds, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || seconds <= 0 {
		return time.Time{}, false
	}
	return time.Unix(seconds, 0), true
}

// parseWindowsBootTime reads the timestamp printed by PowerShell for
// Win32_OperatingSystem.LastBootUpTime, for example "07/02/2024 12:00:00".
//
// The value is requested from CIM as a DateTime and formatted with an explicit
// ISO-8601 pattern, so only one layout needs to be understood here.
func parseWindowsBootTime(output string) (time.Time, bool) {
	value := strings.TrimSpace(output)
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"01/02/2006 15:04:05",
	} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}
