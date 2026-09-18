package identity

import "testing"

func TestParseOSReleasePrefersPrettyName(t *testing.T) {
	content := `NAME="Ubuntu"
VERSION="24.04.1 LTS (Noble Numbat)"
ID=ubuntu
PRETTY_NAME="Ubuntu 24.04.1 LTS"
VERSION_ID="24.04"
`

	if got := parseOSRelease(content); got != "Ubuntu 24.04.1 LTS" {
		t.Errorf("parseOSRelease() = %q, want %q", got, "Ubuntu 24.04.1 LTS")
	}
}

func TestParseOSReleaseFallsBackToName(t *testing.T) {
	// Older distributions may not ship PRETTY_NAME at all.
	content := "NAME=\"Debian GNU/Linux\"\nID=debian\n"

	if got := parseOSRelease(content); got != "Debian GNU/Linux" {
		t.Errorf("parseOSRelease() = %q, want %q", got, "Debian GNU/Linux")
	}
}

func TestParseOSReleaseIgnoresCommentsAndBlanks(t *testing.T) {
	content := "# a comment\n\nPRETTY_NAME=\"Alpine Linux v3.20\"\n"

	if got := parseOSRelease(content); got != "Alpine Linux v3.20" {
		t.Errorf("parseOSRelease() = %q, want %q", got, "Alpine Linux v3.20")
	}
}

func TestParseOSReleaseOnEmptyContent(t *testing.T) {
	if got := parseOSRelease(""); got != "" {
		t.Errorf("parseOSRelease(\"\") = %q, want an empty string", got)
	}
}

func TestParseSwVers(t *testing.T) {
	output := "ProductName:\tmacOS\nProductVersion:\t14.5\nBuildVersion:\t23F79\n"

	if got := parseSwVers(output); got != "macOS 14.5" {
		t.Errorf("parseSwVers() = %q, want %q", got, "macOS 14.5")
	}
}

func TestParseIOPlatformUUID(t *testing.T) {
	output := `+-o J316sAP  <class IOPlatformExpertDevice, id 0x1000001>
    {
      "IOPlatformUUID" = "A1B2C3D4-E5F6-7890-ABCD-EF1234567890"
    }
`

	got := parseIOPlatformUUID(output)
	if got != "A1B2C3D4-E5F6-7890-ABCD-EF1234567890" {
		t.Errorf("parseIOPlatformUUID() = %q, want the platform UUID", got)
	}
}

func TestParseIOPlatformUUIDRejectsMissingValue(t *testing.T) {
	if got := parseIOPlatformUUID("nothing useful here\n"); got != "" {
		t.Errorf("parseIOPlatformUUID() = %q, want an empty string", got)
	}
}
