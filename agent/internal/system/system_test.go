package system

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseProcUptime(t *testing.T) {
	// /proc/uptime reports seconds since boot, then idle seconds.
	uptime, ok := parseProcUptime("12345.67 98765.43\n")
	if !ok {
		t.Fatal("parseProcUptime() failed on valid input")
	}
	if uptime != 12345*time.Second+670*time.Millisecond {
		t.Errorf("parseProcUptime() = %v, want 12345.67s", uptime)
	}
}

func TestParseProcUptimeRejectsMalformedInput(t *testing.T) {
	if _, ok := parseProcUptime(""); ok {
		t.Error("parseProcUptime() accepted empty input")
	}
	if _, ok := parseProcUptime("not a number"); ok {
		t.Error("parseProcUptime() accepted non-numeric input")
	}
}

func TestParseSysctlBootTime(t *testing.T) {
	output := "{ sec = 1719999999, usec = 0 } Tue Jul  2 12:00:00 2024\n"

	bootTime, ok := parseSysctlBootTime(output)
	if !ok {
		t.Fatal("parseSysctlBootTime() failed on valid output")
	}
	if bootTime.Unix() != 1719999999 {
		t.Errorf("parseSysctlBootTime() = %v, want unix time 1719999999", bootTime.Unix())
	}
}

func TestParseSysctlBootTimeRejectsMalformedInput(t *testing.T) {
	if _, ok := parseSysctlBootTime("no boot time here"); ok {
		t.Error("parseSysctlBootTime() accepted output with no boot time")
	}
}

func TestParseWindowsBootTimeAcceptsTheFormatsPowerShellProduces(t *testing.T) {
	// The collector asks PowerShell for an explicit ISO-8601 pattern, but the
	// fallbacks keep the agent working if a locale overrides the format.
	formats := []string{
		"2024-07-02T12:00:00",
		"2024-07-02T12:00:00Z",
		"2024-07-02 12:00:00",
		"07/02/2024 12:00:00",
	}

	for _, value := range formats {
		if _, ok := parseWindowsBootTime(value); !ok {
			t.Errorf("parseWindowsBootTime(%q) failed, want it accepted", value)
		}
	}
}

func TestParseWindowsBootTimeRejectsGarbage(t *testing.T) {
	if _, ok := parseWindowsBootTime("whenever the machine started"); ok {
		t.Error("parseWindowsBootTime() accepted non-timestamp input")
	}
}

func TestInstallationIDIsPersistedAndStable(t *testing.T) {
	stateDir := t.TempDir()

	first, err := InstallationID(stateDir)
	if err != nil {
		t.Fatalf("InstallationID() returned an unexpected error: %v", err)
	}
	if first == "" {
		t.Fatal("InstallationID() returned an empty identifier")
	}

	// A restart must reuse the same identifier, otherwise the agent's local
	// state would change identity every time the service restarted.
	second, err := InstallationID(stateDir)
	if err != nil {
		t.Fatalf("second InstallationID() returned an unexpected error: %v", err)
	}
	if second != first {
		t.Errorf("InstallationID() = %q on the second call, want the persisted %q", second, first)
	}
}

func TestInstallationIDIsWrittenWithRestrictivePermissions(t *testing.T) {
	stateDir := t.TempDir()

	if _, err := InstallationID(stateDir); err != nil {
		t.Fatalf("InstallationID() returned an unexpected error: %v", err)
	}

	info, err := os.Stat(filepath.Join(stateDir, installationIDFile))
	if err != nil {
		t.Fatalf("could not stat the installation id file: %v", err)
	}

	// The file is local diagnostic state, not a secret, but nothing else on the
	// machine needs to read it.
	if permissions := info.Mode().Perm(); permissions != 0o600 {
		t.Errorf("installation id permissions = %o, want 600", permissions)
	}
}

func TestInstallationIDDiffersBetweenMachines(t *testing.T) {
	first, err := InstallationID(t.TempDir())
	if err != nil {
		t.Fatalf("InstallationID() returned an unexpected error: %v", err)
	}
	second, err := InstallationID(t.TempDir())
	if err != nil {
		t.Fatalf("InstallationID() returned an unexpected error: %v", err)
	}

	if first == second {
		t.Error("two separate installations produced the same identifier")
	}
}

func TestInstallationIDRejectsAnEmptyStateDirectory(t *testing.T) {
	if _, err := InstallationID(""); err == nil {
		t.Error("InstallationID(\"\") succeeded, want an error")
	}
}
