// Command campuswatch-agent is the CampusWatch monitoring agent.
//
// It runs on each monitored computer, collects the operational information the
// platform is permitted to gather, and reports it to the CampusWatch backend
// over the agent API. It is installed and supervised by the scripts in
// installer/, which supply its configuration through the environment.
//
// The agent is deliberately narrow in what it looks at. README section 6 sets
// the privacy boundary, and nothing here reads passwords, keystrokes, browser
// cookies, private files or screen contents. It reports aggregate system
// counters, the operating system accounts that are logged in, and whether the
// keyboard and mouse have recently been used.
//
// # Building
//
// This file sits at the module root so the agent is built from the directory
// named after it:
//
//	cd agent
//	go build -o campuswatch-agent .
//
// The -o flag matters. Without it, Go names the binary after the directory and
// produces "agent", but the installers in ../installer/ look for a binary
// called "campuswatch-agent". Build with -o, or set CAMPUSWATCH_AGENT_BINARY
// (Linux) / -Binary (Windows) to point the installer at a differently named
// binary.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math"
	"net"
	"net/url"
	"os/signal"
	"syscall"
	"time"

	"CampusWatch/agent/internal/activity"
	"CampusWatch/agent/internal/client"
	"CampusWatch/agent/internal/config"
	"CampusWatch/agent/internal/health"
	"CampusWatch/agent/internal/identity"
	"CampusWatch/agent/internal/session"
	"CampusWatch/agent/internal/system"
)

// networkProbeTimeout bounds the connectivity check performed with each health
// sample. It is short so an unreachable backend cannot stall the heartbeat loop.
const networkProbeTimeout = 3 * time.Second

func main() {
	configPath := flag.String("config", "",
		"optional path to an environment file holding the CAMPUSWATCH_* settings; "+
			"values already present in the environment take precedence")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("campuswatch-agent: %v", err)
	}

	// README section 37 requires HTTPS in production. A plain HTTP endpoint is
	// allowed so the agent can be pointed at a development backend, but saying
	// so plainly in the log beats letting it pass unnoticed.
	if !cfg.IsSecure() {
		log.Printf("warning: %s does not use HTTPS; agent credentials and monitoring data will be sent in clear text",
			config.EnvServerURL)
	}

	device := identity.Detect()
	log.Printf("device identity: %s", device.Summary())

	// The installation identifier is persisted so the agent keeps a stable
	// local identity across restarts, independent of the hostname or address,
	// as README section 11 requires.
	installationID, err := system.InstallationID(cfg.StateDir)
	if err != nil {
		// Not fatal: the identifier is diagnostic state, and refusing to
		// monitor a machine because a file could not be written would be worse
		// than running without it.
		log.Printf("warning: could not persist installation id in %s: %v", cfg.StateDir, err)
	} else {
		log.Printf("installation id: %s (state directory %s)", installationID, cfg.StateDir)
	}

	api := client.New(client.Config{
		ServerURL:  cfg.ServerURL,
		AgentCode:  cfg.AgentCode,
		Credential: cfg.Credential,
	})

	// Connectivity is judged by whether the backend itself is reachable, which
	// is the connection the heartbeat actually depends on.
	collector := health.NewCollector(device.OperatingSystem, health.TCPProbe(probeTarget(cfg.ServerURL), networkProbeTimeout))
	thresholds := health.DefaultThresholds()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("campuswatch-agent %s starting: system_id=%s heartbeat=%s session_poll=%s idle_threshold=%s",
		cfg.AgentVersion, cfg.SystemID, cfg.HeartbeatInterval, cfg.SessionPollInterval, cfg.IdleThreshold)

	run(ctx, cfg, api, collector, thresholds)

	log.Printf("campuswatch-agent stopped")
}

// run drives the heartbeat and session polling loops until the context is
// cancelled. A failure to reach the backend is logged and retried on the next
// tick rather than terminating the agent, because a monitoring agent that quits
// whenever the network hiccups would be useless on a large fleet.
func run(ctx context.Context, cfg config.Config, api *client.Client, collector *health.Collector, thresholds health.Thresholds) {
	sessions := session.NewTracker()
	activityTracker := activity.NewTracker(cfg.IdleThreshold)
	breaches := newBreachTracker()

	sendHeartbeat := func() {
		sample := collector.Collect(ctx)

		report := &client.HealthReport{
			SystemID:         cfg.SystemID,
			RecordedAt:       sample.RecordedAt,
			CPUPercent:       sample.CPUPercent,
			MemoryPercent:    sample.MemoryPercent,
			DiskPercent:      sample.DiskPercent,
			BatteryPercent:   sample.BatteryPercent,
			NetworkConnected: sample.NetworkConnected,
			OperatingSystem:  sample.OperatingSystem,
			AgentHealth:      sample.AgentHealth,
			UptimeSeconds:    int64(sample.Uptime.Seconds()),
		}
		request := client.HeartbeatRequest{
			AgentID:      cfg.AgentID,
			SystemID:     cfg.SystemID,
			Timestamp:    sample.RecordedAt,
			AgentVersion: cfg.AgentVersion,
			Health:       report,
		}

		if _, err := api.Heartbeat(ctx, request); err != nil {
			log.Printf("heartbeat failed: %v", err)
			return
		}
		log.Printf("heartbeat sent: cpu=%.1f%% memory=%.1f%% disk=%.1f%% network=%t agent=%s",
			sample.CPUPercent, sample.MemoryPercent, sample.DiskPercent,
			sample.NetworkConnected, sample.AgentHealth)

		reportBreaches(ctx, api, cfg, breaches, thresholds.Breaches(sample))
	}

	pollSession := func() {
		now := time.Now().UTC()

		users, err := session.LoggedInUsers()
		if err != nil {
			// Skipping the poll matters: treating an unreadable login list as
			// "nobody is logged in" would fabricate logout events for every
			// user on the machine.
			log.Printf("session poll failed, skipping this cycle: %v", err)
		} else {
			for _, transition := range sessions.Observe(users, now) {
				event := client.SessionEventRequest{
					AgentID:   cfg.AgentID,
					SystemID:  cfg.SystemID,
					Username:  transition.Username,
					Event:     transition.Event,
					Timestamp: transition.At,
				}
				if err := api.SessionEvent(ctx, event); err != nil {
					log.Printf("report %s for %s failed: %v", transition.Event, transition.Username, err)
					continue
				}
				log.Printf("reported %s for %s", transition.Event, transition.Username)
			}
		}

		reportActivity(ctx, api, cfg, sessions, activityTracker, now)
	}

	// Report once at start-up so the backend sees the machine immediately
	// rather than waiting for the first interval to elapse.
	sendHeartbeat()

	heartbeatTicker := time.NewTicker(cfg.HeartbeatInterval)
	defer heartbeatTicker.Stop()
	sessionTicker := time.NewTicker(cfg.SessionPollInterval)
	defer sessionTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeatTicker.C:
			sendHeartbeat()
		case <-sessionTicker.C:
			pollSession()
		}
	}
}

// reportActivity translates the idle state into a session event.
//
// Activity is attributed to a logged-in account because the backend records
// activity against a user session. With nobody logged in there is no session to
// attach the change to, so the state is tracked silently and reported if and
// when somebody logs in.
func reportActivity(
	ctx context.Context,
	api *client.Client,
	cfg config.Config,
	sessions *session.Tracker,
	tracker *activity.Tracker,
	now time.Time,
) {
	idleFor, ok := activity.IdleDuration()
	if !ok {
		// The platform cannot answer, so no transition is invented.
		return
	}

	state, changed := tracker.Observe(idleFor)
	if !changed {
		return
	}

	loggedIn := sessions.Current()
	if len(loggedIn) == 0 {
		log.Printf("activity is now %s but no user is logged in; not reporting", state)
		return
	}

	for _, username := range loggedIn {
		event := client.SessionEventRequest{
			AgentID:   cfg.AgentID,
			SystemID:  cfg.SystemID,
			Username:  username,
			Event:     activity.EventName(state),
			Timestamp: now,
		}
		if err := api.SessionEvent(ctx, event); err != nil {
			log.Printf("report %s for %s failed: %v", activity.EventName(state), username, err)
			continue
		}
		log.Printf("reported %s for %s (idle for %s)", activity.EventName(state), username, idleFor.Round(time.Second))
	}
}

// reportBreaches posts a HEALTH_WARNING event for each newly crossed threshold.
func reportBreaches(ctx context.Context, api *client.Client, cfg config.Config, tracker *breachTracker, current []health.Breach) {
	for _, breach := range tracker.newlyBreached(current) {
		payload, err := json.Marshal(map[string]any{
			"metric":    breach.Metric,
			"value":     math.Round(breach.Value*10) / 10,
			"threshold": breach.Threshold,
		})
		if err != nil {
			log.Printf("encode health warning payload: %v", err)
			continue
		}

		event := client.EventRequest{
			AgentID:    cfg.AgentID,
			SystemID:   cfg.SystemID,
			EventType:  client.EventHealthWarning,
			Payload:    payload,
			OccurredAt: time.Now().UTC(),
		}
		if err := api.Event(ctx, event); err != nil {
			log.Printf("report health warning (%s) failed: %v", breach.Metric, err)
			continue
		}
		log.Printf("reported HEALTH_WARNING: %s is %.1f (threshold %.1f)",
			breach.Metric, breach.Value, breach.Threshold)
	}
}

// breachTracker remembers which metrics are currently over threshold.
//
// README section 17 asks for warnings when thresholds are reached. Reporting the
// breach on every heartbeat would produce a warning every thirty seconds for as
// long as a disk stayed full, drowning the event history the platform exists to
// keep. Only the transition into breach is reported.
type breachTracker struct {
	active map[string]bool
}

func newBreachTracker() *breachTracker {
	return &breachTracker{active: make(map[string]bool)}
}

// newlyBreached returns the breaches that were not already active, and records
// the current set so recovery is detected on a later cycle.
func (b *breachTracker) newlyBreached(current []health.Breach) []health.Breach {
	var fresh []health.Breach
	seen := make(map[string]bool, len(current))

	for _, breach := range current {
		seen[breach.Metric] = true
		if !b.active[breach.Metric] {
			fresh = append(fresh, breach)
		}
	}
	b.active = seen
	return fresh
}

// probeTarget converts the configured backend URL into a host:port pair for the
// connectivity probe.
func probeTarget(serverURL string) string {
	parsed, err := url.Parse(serverURL)
	if err != nil || parsed.Hostname() == "" {
		return ""
	}

	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return net.JoinHostPort(parsed.Hostname(), port)
}
