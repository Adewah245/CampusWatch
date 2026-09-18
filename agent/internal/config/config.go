// Package config loads and validates the CampusWatch agent configuration.
//
// Every setting is sourced from the environment so the agent can be driven
// directly by the deployment scripts. This is deliberate: the variable names
// below are part of the deployment contract. installer/linux/install.sh writes
// them into /etc/campuswatch-agent.env (loaded by systemd via EnvironmentFile)
// and installer/windows/install.ps1 writes the same names into agent.env.
// Keeping the single source of truth here means neither installer needs to
// change when the agent gains a setting.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Environment variable names. These must stay in sync with the values written
// by installer/linux/install.sh and installer/windows/install.ps1.
const (
	EnvServerURL            = "CAMPUSWATCH_SERVER_URL"
	EnvAgentID              = "CAMPUSWATCH_AGENT_ID"
	EnvAgentCode            = "CAMPUSWATCH_AGENT_CODE"
	EnvAgentCredential      = "CAMPUSWATCH_AGENT_CREDENTIAL"
	EnvSystemID             = "CAMPUSWATCH_SYSTEM_ID"
	EnvAgentVersion         = "CAMPUSWATCH_AGENT_VERSION"
	EnvHeartbeatSeconds     = "CAMPUSWATCH_HEARTBEAT_INTERVAL_SECONDS"
	EnvSessionPollSeconds   = "CAMPUSWATCH_SESSION_POLL_SECONDS"
	EnvIdleThresholdSeconds = "CAMPUSWATCH_IDLE_THRESHOLD_SECONDS"
	EnvStateDir             = "CAMPUSWATCH_STATE_DIR"
)

// Defaults and bounds for the optional settings.
//
// The heartbeat interval mirrors the value proposed in README section 13 (30
// seconds, with a 2 minute backend offline threshold). The minimum exists so a
// misconfiguration cannot turn the agent into a request flood against the API.
const (
	DefaultHeartbeatInterval  = 30 * time.Second
	DefaultSessionPollInterval = 15 * time.Second
	DefaultIdleThreshold      = 5 * time.Minute
	MinimumHeartbeatInterval  = 5 * time.Second
	MinimumSessionPollInterval = 5 * time.Second
	DefaultAgentVersion       = "1.0.0"
)

// Config holds the fully validated settings required to run the agent.
type Config struct {
	// ServerURL is the CampusWatch backend base URL, without a trailing slash.
	ServerURL string
	// AgentID and AgentCode are issued by the backend when the agent registers.
	AgentID string
	// AgentCode is sent as the X-Agent-Code header on every authenticated call.
	AgentCode string
	// Credential is the per-installation secret issued by the backend. It is
	// sent as X-Agent-Credential and must never be identical across machines.
	Credential string
	// SystemID links this agent to its registered system record.
	SystemID string
	// AgentVersion is reported to the backend on every heartbeat.
	AgentVersion string
	// HeartbeatInterval controls how often a heartbeat is posted.
	HeartbeatInterval time.Duration
	// SessionPollInterval controls how often login/logout and idle state are checked.
	SessionPollInterval time.Duration
	// IdleThreshold is how long a session must be inactive before it is idle.
	IdleThreshold time.Duration
	// StateDir is a writable directory used for locally persisted agent state.
	StateDir string
}

// Load reads configuration from the process environment, first applying the
// optional environment file at path when one is supplied.
//
// Values already present in the environment always win over the file, matching
// the behaviour of the backend loader in backend/internal/config/config.go.
func Load(path string) (Config, error) {
	if path != "" {
		if err := applyEnvFile(path); err != nil {
			return Config{}, err
		}
	}
	return FromEnv()
}

// FromEnv builds a Config from the current process environment.
func FromEnv() (Config, error) {
	cfg := Config{
		ServerURL:    strings.TrimRight(strings.TrimSpace(os.Getenv(EnvServerURL)), "/"),
		AgentID:      strings.TrimSpace(os.Getenv(EnvAgentID)),
		AgentCode:    strings.TrimSpace(os.Getenv(EnvAgentCode)),
		Credential:   strings.TrimSpace(os.Getenv(EnvAgentCredential)),
		SystemID:     strings.TrimSpace(os.Getenv(EnvSystemID)),
		AgentVersion: strings.TrimSpace(os.Getenv(EnvAgentVersion)),
	}

	if cfg.AgentVersion == "" {
		cfg.AgentVersion = DefaultAgentVersion
	}

	interval, err := durationSetting(EnvHeartbeatSeconds, DefaultHeartbeatInterval, MinimumHeartbeatInterval)
	if err != nil {
		return Config{}, err
	}
	cfg.HeartbeatInterval = interval

	poll, err := durationSetting(EnvSessionPollSeconds, DefaultSessionPollInterval, MinimumSessionPollInterval)
	if err != nil {
		return Config{}, err
	}
	cfg.SessionPollInterval = poll

	idle, err := durationSetting(EnvIdleThresholdSeconds, DefaultIdleThreshold, time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.IdleThreshold = idle

	stateDir, err := resolveStateDir(os.Getenv(EnvStateDir))
	if err != nil {
		return Config{}, err
	}
	cfg.StateDir = stateDir

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate reports the first missing or malformed required setting.
//
// The error names every missing variable at once rather than failing on the
// first one, so a misconfigured deployment can be fixed in a single pass. This
// matches the behaviour of the previous stub agent.
func (c Config) Validate() error {
	var missing []string
	for _, required := range []struct {
		name  string
		value string
	}{
		{EnvServerURL, c.ServerURL},
		{EnvAgentID, c.AgentID},
		{EnvAgentCode, c.AgentCode},
		{EnvAgentCredential, c.Credential},
		{EnvSystemID, c.SystemID},
	} {
		if required.value == "" {
			missing = append(missing, required.name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	parsed, err := url.Parse(c.ServerURL)
	if err != nil {
		return fmt.Errorf("%s is not a valid URL: %w", EnvServerURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s must use the http or https scheme", EnvServerURL)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%s must include a host", EnvServerURL)
	}

	return nil
}

// IsSecure reports whether the configured backend URL uses TLS. A plain HTTP
// endpoint is accepted so agents can talk to a development backend, but the
// caller uses this to log a warning: README section 37 requires HTTPS in
// production.
func (c Config) IsSecure() bool {
	return strings.HasPrefix(strings.ToLower(c.ServerURL), "https://")
}

// durationSetting reads a whole number of seconds from the environment.
func durationSetting(name string, fallback, minimum time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number of seconds: %w", name, err)
	}
	value := time.Duration(seconds) * time.Second
	if value < minimum {
		return 0, fmt.Errorf("%s must be at least %s", name, minimum)
	}
	return value, nil
}

// resolveStateDir picks a writable directory for persisted agent state.
//
// When not set explicitly this defaults to the directory holding the running
// executable. That choice is not arbitrary: the shipped systemd unit sets
// ProtectSystem=strict with ReadWritePaths=$INSTALL_DIR, so the install
// directory is the only place the service is permitted to write. Deriving the
// path from the executable therefore works both when installed (/opt/... or
// %ProgramFiles%\...) and when run straight from a build tree during
// development, with no platform-specific directory guessing.
func resolveStateDir(configured string) (string, error) {
	if configured != "" {
		return configured, nil
	}

	if executable, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(executable); err == nil {
			executable = resolved
		}
		return filepath.Dir(executable), nil
	}

	fallback, err := os.UserConfigDir()
	if err != nil {
		return "", errors.New("cannot determine a state directory; set " + EnvStateDir)
	}
	return filepath.Join(fallback, "campuswatch-agent"), nil
}

// applyEnvFile loads KEY=VALUE pairs from path without overwriting variables
// that are already set. Export prefixes, blank lines and # comments are
// tolerated, and surrounding quotes are stripped, so the same file works
// whether it is loaded by systemd or read directly here.
func applyEnvFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %s: %w", path, err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), "\"'")
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set %s from %s: %w", key, path, err)
			}
		}
	}
	return nil
}
