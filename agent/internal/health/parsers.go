package health

import (
	"strconv"
	"strings"
)

// This file holds the pure parsing and arithmetic behind metric collection.
// It carries no build tags so the formats can be unit-tested on any machine,
// which matters because a change in a kernel's reporting format is the most
// likely way these collectors break in the field.

// cpuTimes holds cumulative CPU time counters. Linux and macOS both expose
// these as monotonically increasing tick counts, so a percentage is derived
// from the difference between two readings rather than from a single sample.
type cpuTimes struct {
	idle  float64
	total float64
}

// parseProcStat reads the aggregate "cpu" line of /proc/stat.
//
// The line looks like:
//
//	cpu  234513 1023 48211 8823123 512 0 811 0 0 0
//
// The fields after the label are user, nice, system, idle, iowait, irq,
// softirq, steal, guest and guest_nice. Idle time is counted as idle plus
// iowait, which is the convention used by every mainstream monitoring tool:
// a machine waiting on disk is not burning CPU.
func parseProcStat(content string) (cpuTimes, bool) {
	for _, line := range strings.Split(content, "\n") {
		// The aggregate line is labelled "cpu"; per-core lines are "cpu0",
		// "cpu1" and so on, so a trailing space rules them out.
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}

		fields := strings.Fields(line)[1:] // Drop the "cpu" label.
		if len(fields) < 5 {
			return cpuTimes{}, false
		}

		var times cpuTimes
		for index, field := range fields {
			value, err := strconv.ParseFloat(field, 64)
			if err != nil {
				return cpuTimes{}, false
			}
			times.total += value
			if index == 3 || index == 4 { // idle, iowait
				times.idle += value
			}
		}
		if times.total == 0 {
			return cpuTimes{}, false
		}
		return times, true
	}
	return cpuTimes{}, false
}

// cpuPercent computes utilisation between two readings. The second return value
// is false when the counters did not advance, which happens on the very first
// call and whenever the sampling interval was too short to accumulate a tick.
func cpuPercent(previous, current cpuTimes) (float64, bool) {
	totalDelta := current.total - previous.total
	if totalDelta <= 0 {
		return 0, false
	}
	idleDelta := current.idle - previous.idle
	if idleDelta < 0 {
		// The counters reset, which means the machine rebooted between samples.
		return 0, false
	}

	percent := (totalDelta - idleDelta) / totalDelta * 100
	return clampPercent(percent), true
}

// parseProcMeminfo extracts total and available memory in kilobytes from
// /proc/meminfo.
//
// MemAvailable is preferred over MemFree because it accounts for memory the
// kernel can reclaim from caches. Reporting MemFree would make every healthy
// machine look critically low on memory.
func parseProcMeminfo(content string) (totalKB, availableKB float64, ok bool) {
	var freeKB float64
	var hasAvailable bool

	for _, line := range strings.Split(content, "\n") {
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		fields := strings.Fields(value)
		if len(fields) == 0 {
			continue
		}
		amount, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			continue
		}

		switch strings.TrimSpace(key) {
		case "MemTotal":
			totalKB = amount
		case "MemAvailable":
			availableKB, hasAvailable = amount, true
		case "MemFree":
			if !hasAvailable {
				freeKB = amount
			}
		}
	}

	if totalKB <= 0 {
		return 0, 0, false
	}
	if !hasAvailable {
		// Kernels before 3.14 do not publish MemAvailable. MemFree plus the
		// reclaimable cache is a reasonable stand-in, but if only MemFree is
		// known the reading will overstate pressure.
		availableKB = freeKB
	}
	if availableKB < 0 {
		availableKB = 0
	}
	return totalKB, availableKB, true
}

// memoryPercent converts total and available memory into a used percentage.
func memoryPercent(totalKB, availableKB float64) (float64, bool) {
	if totalKB <= 0 {
		return 0, false
	}
	if availableKB > totalKB {
		availableKB = totalKB
	}
	return clampPercent((totalKB - availableKB) / totalKB * 100), true
}

// clampPercent confines a value to the 0-100 range so a surprising reading can
// never produce a nonsensical percentage on the dashboard.
func clampPercent(value float64) float64 {
	switch {
	case value < 0:
		return 0
	case value > 100:
		return 100
	default:
		return value
	}
}

// parseBatteryCapacity reads a battery charge percentage from a
// /sys/class/power_supply/*/capacity file, which contains a bare integer.
func parseBatteryCapacity(content string) (float64, bool) {
	value, err := strconv.ParseFloat(strings.TrimSpace(content), 64)
	if err != nil {
		return 0, false
	}
	return clampPercent(value), true
}

// parseSysctlCPTime reads the cumulative CPU ticks reported by macOS under
// kern.cp_time, which are the user, nice, system, interrupt and idle counters
// in that order. Only the total and the idle counter are needed to derive a
// utilisation percentage.
func parseSysctlCPTime(content string) (cpuTimes, bool) {
	fields := strings.Fields(strings.TrimSpace(content))
	if len(fields) < 5 {
		return cpuTimes{}, false
	}

	var times cpuTimes
	for index, field := range fields {
		value, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return cpuTimes{}, false
		}
		times.total += value
		if index == 4 { // idle
			times.idle = value
		}
	}
	if times.total == 0 {
		return cpuTimes{}, false
	}
	return times, true
}

// parseVMStat extracts the page size and the free, inactive and speculative
// page counts from macOS vm_stat output.
//
// Available memory is taken as free plus inactive plus speculative pages.
// Counting only free pages would report a healthy Mac as critically low on
// memory, because macOS puts almost everything it can into cache.
func parseVMStat(content string) (availableKB, totalPageKB float64, ok bool) {
	var pageSizeBytes float64
	var availablePages float64

	for _, line := range strings.Split(content, "\n") {
		if _, rest, found := strings.Cut(line, "page size of"); found {
			fields := strings.Fields(rest)
			if len(fields) > 0 {
				if size, err := strconv.ParseFloat(fields[0], 64); err == nil && size > 0 {
					pageSizeBytes = size
				}
			}
			continue
		}

		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		switch strings.TrimSpace(key) {
		case "Pages free", "Pages inactive", "Pages speculative":
			amount := strings.TrimSuffix(strings.TrimSpace(value), ".")
			pages, err := strconv.ParseFloat(amount, 64)
			if err != nil {
				continue
			}
			availablePages += pages
		}
	}

	if pageSizeBytes <= 0 {
		return 0, 0, false
	}
	return availablePages * pageSizeBytes / 1024, pageSizeBytes / 1024, true
}

// parseDF reads the used and total capacity in kilobytes from the output of
// df -k, which is reported the same way on macOS and Linux.
func parseDF(content string) (usedKB, totalKB float64, ok bool) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 2 {
		return 0, 0, false
	}

	// Columns are Filesystem, 1024-blocks, Used, Available, Capacity, Mounted on.
	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 3 {
		return 0, 0, false
	}
	total, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, false
	}
	used, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, 0, false
	}
	if total <= 0 {
		return 0, 0, false
	}
	return used, total, true
}

// parseTwoNumbers reads two whitespace-separated numbers, which is the shape
// the PowerShell collectors emit.
func parseTwoNumbers(output string) (float64, float64, bool) {
	fields := strings.Fields(strings.TrimSpace(output))
	if len(fields) < 2 {
		return 0, 0, false
	}
	first, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, false
	}
	second, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, false
	}
	return first, second, true
}

// parseSingleNumber reads one number, which is the shape used by the Windows
// CPU and battery collectors.
func parseSingleNumber(output string) (float64, bool) {
	value, err := strconv.ParseFloat(strings.TrimSpace(output), 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

// parsePMsetBattery extracts the charge percentage from macOS pmset output,
// which reports lines such as:
//
//	-InternalBattery-0 (id=1234567)	85%; discharging; 4:12 remaining present: true
//
// A machine with no battery simply does not contain a percentage, which is
// reported as "no battery" rather than as an error.
func parsePMsetBattery(output string) (float64, bool) {
	before, _, found := strings.Cut(output, "%")
	if !found {
		return 0, false
	}

	// Walk back from the percent sign to the start of the number. Taking the
	// last whitespace-separated token would fail, because the value is attached
	// to the preceding field.
	end := len(before)
	start := end
	for start > 0 {
		character := before[start-1]
		if (character >= '0' && character <= '9') || character == '.' {
			start--
			continue
		}
		break
	}
	if start == end {
		return 0, false
	}

	value, err := strconv.ParseFloat(before[start:end], 64)
	if err != nil {
		return 0, false
	}
	return clampPercent(value), true
}
