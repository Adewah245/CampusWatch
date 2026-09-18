package identity

import "strings"

// This file holds the pure parsing helpers used by the platform-specific
// discovery code. They are kept free of build tags on purpose: parsing is the
// part most likely to break when an operating system changes its output
// format, so it must be unit-testable on every development machine rather than
// only on the platform it targets.

// parseOSRelease extracts PRETTY_NAME from the contents of an /etc/os-release
// file, falling back to NAME when PRETTY_NAME is absent.
func parseOSRelease(content string) string {
	var name string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), "\"'")
		switch strings.TrimSpace(key) {
		case "PRETTY_NAME":
			if value != "" {
				return value
			}
		case "NAME":
			name = value
		}
	}
	return name
}

// parseSwVers extracts the product name and version from sw_vers output, which
// is a list of tab-separated "Key:\tValue" lines.
func parseSwVers(output string) string {
	var name, version string
	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "ProductName":
			name = value
		case "ProductVersion":
			version = value
		}
	}
	return strings.TrimSpace(name + " " + version)
}

// parseIOPlatformUUID pulls the IOPlatformUUID value out of ioreg output. The
// line looks like:
//
//	"IOPlatformUUID" = "A1B2C3D4-E5F6-7890-ABCD-EF1234567890"
//
// so the value is whatever follows the assignment, stripped of quoting.
func parseIOPlatformUUID(output string) string {
	for _, line := range strings.Split(output, "\n") {
		_, value, ok := strings.Cut(line, "IOPlatformUUID")
		if !ok {
			continue
		}
		_, value, ok = strings.Cut(value, "=")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), "\"")
		if value != "" {
			return value
		}
	}
	return ""
}
