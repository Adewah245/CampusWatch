package config

import (
	"strings"
	"testing"
	"time"
)

// setRequired populates the five mandatory settings with valid values and
// clears the optional ones, so each test starts from a known baseline
// regardless of the developer's own environment.
func setRequired(t *testing.T) {
	t.Helper()

	t.Setenv(EnvServerURL, "https://campuswatch.example.edu")
	t.Setenv(EnvAgentID, "agent-0001")
	t.Setenv(EnvAgentCode, "LAB-A-PC-001")
	t.Setenv(EnvAgentCredential, "per-installation-secret")
	t.Setenv(EnvSystemID, "SYS-000001")

	t.Setenv(EnvAgentVersion, "")
	t.Setenv(EnvHeartbeatSeconds, "")
	t.Setenv(EnvSessionPollSeconds, "")
	t.Setenv(EnvIdleThresholdSeconds, "")
	t.Setenv(EnvStateDir, "")
}

func TestFromEnvAppliesDefaults(t *testing.T) {
	setRequired(t)

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv() returned an unexpected error: %v", err)
	}

	if cfg.AgentVersion != DefaultAgentVersion {
		t.Errorf("AgentVersion = %q, want %q", cfg.AgentVersion, DefaultAgentVersion)
	}
	if cfg.HeartbeatInterval != DefaultHeartbeatInterval {
		t.Errorf("HeartbeatInterval = %s, want %s", cfg.HeartbeatInterval, DefaultHeartbeatInterval)
	}
	if cfg.SessionPollInterval != DefaultSessionPollInterval {
		t.Errorf("SessionPollInterval = %s, want %s", cfg.SessionPollInterval, DefaultSessionPollInterval)
	}
	if cfg.IdleThreshold != DefaultIdleThreshold {
		t.Errorf("IdleThreshold = %s, want %s", cfg.IdleThreshold, DefaultIdleThreshold)
	}
	if cfg.StateDir == "" {
		t.Error("StateDir is empty; a usable state directory must always be resolved")
	}
}

func TestFromEnvStripsTrailingSlashFromServerURL(t *testing.T) {
	setRequired(t)
	t.Setenv(EnvServerURL, "https://campuswatch.example.edu/")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv() returned an unexpected error: %v", err)
	}

	// A trailing slash would produce a doubled slash once the API path is
	// appended, so it must be removed here.
	if cfg.ServerURL != "https://campuswatch.example.edu" {
		t.Errorf("ServerURL = %q, want the trailing slash removed", cfg.ServerURL)
	}
}

func TestFromEnvReportsEveryMissingSetting(t *testing.T) {
	// Clear everything, then rely on the error naming the absent variables.
	for _, name := range []string{
		EnvServerURL, EnvAgentID, EnvAgentCode,
		EnvAgentCredential, EnvSystemID, EnvStateDir,
	} {
		t.Setenv(name, "")
	}

	_, err := FromEnv()
	if err == nil {
		t.Fatal("FromEnv() succeeded with no configuration, want an error")
	}

	// Fixing a deployment should take one pass, so every missing variable is
	// named rather than only the first one found.
	for _, name := range []string{EnvServerURL, EnvAgentID, EnvAgentCode, EnvAgentCredential, EnvSystemID} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not mention the missing setting %s", err, name)
		}
	}
}

func TestFromEnvRejectsIntervalsBelowTheMinimum(t *testing.T) {
	setRequired(t)
	// One second would flood the API; the floor exists to prevent that.
	t.Setenv(EnvHeartbeatSeconds, "1")

	if _, err := FromEnv(); err == nil {
		t.Fatal("FromEnv() accepted a heartbeat interval below the minimum, want an error")
	}
}

func TestFromEnvRejectsNonNumericInterval(t *testing.T) {
	setRequired(t)
	t.Setenv(EnvSessionPollSeconds, "often")

	if _, err := FromEnv(); err == nil {
		t.Fatal("FromEnv() accepted a non-numeric interval, want an error")
	}
}

func TestFromEnvHonoursConfiguredIntervals(t *testing.T) {
	setRequired(t)
	t.Setenv(EnvHeartbeatSeconds, "45")
	t.Setenv(EnvSessionPollSeconds, "20")
	t.Setenv(EnvIdleThresholdSeconds, "600")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv() returned an unexpected error: %v", err)
	}

	if cfg.HeartbeatInterval != 45*time.Second {
		t.Errorf("HeartbeatInterval = %s, want 45s", cfg.HeartbeatInterval)
	}
	if cfg.SessionPollInterval != 20*time.Second {
		t.Errorf("SessionPollInterval = %s, want 20s", cfg.SessionPollInterval)
	}
	if cfg.IdleThreshold != 10*time.Minute {
		t.Errorf("IdleThreshold = %s, want 10m", cfg.IdleThreshold)
	}
}

func TestValidateAcceptsOnlyHTTPAndHTTPS(t *testing.T) {
	base := Config{
		ServerURL:  "https://campuswatch.example.edu",
		AgentID:    "agent-0001",
		AgentCode:  "LAB-A-PC-001",
		Credential: "secret",
		SystemID:   "SYS-000001",
	}

	tests := map[string]struct {
		serverURL string
		wantError bool
	}{
		"https is accepted":        {"https://campuswatch.example.edu", false},
		"http is accepted for dev": {"http://localhost:8080", false},
		"ftp is rejected":          {"ftp://campuswatch.example.edu", true},
		"missing scheme":           {"campuswatch.example.edu", true},
		"missing host":             {"https://", true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			candidate := base
			candidate.ServerURL = test.serverURL

			err := candidate.Validate()
			if test.wantError && err == nil {
				t.Fatalf("Validate() accepted %q, want an error", test.serverURL)
			}
			if !test.wantError && err != nil {
				t.Fatalf("Validate() rejected %q: %v", test.serverURL, err)
			}
		})
	}
}

func TestIsSecure(t *testing.T) {
	tests := map[string]bool{
		"https://campuswatch.example.edu": true,
		"HTTPS://campuswatch.example.edu": true,
		"http://localhost:8080":           false,
	}

	for serverURL, want := range tests {
		if got := (Config{ServerURL: serverURL}).IsSecure(); got != want {
			t.Errorf("IsSecure() for %q = %t, want %t", serverURL, got, want)
		}
	}
}
