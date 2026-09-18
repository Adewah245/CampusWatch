// Package activity decides whether a logged-in computer is currently in use.
//
// This is deliberately separate from system status and from user session state.
// README section 9 gives the example of a machine that is ONLINE but
// LOGGED_OUT: powered on, agent reporting, nobody using it. Activity answers
// only the narrower question of whether the person at the keyboard is still
// interacting with it, and it never inspects what they are doing.
package activity

import (
	"time"

	"CampusWatch/agent/internal/client"
)

// States reported as user session activity.
const (
	StateActive = "ACTIVE"
	StateIdle   = "IDLE"
)

// Tracker converts a stream of idle durations into activity state changes.
//
// It is not safe for concurrent use; the agent drives it from a single polling
// goroutine.
type Tracker struct {
	threshold time.Duration
	current   string
	seeded    bool
}

// NewTracker returns a tracker that treats an idle duration of at least
// threshold as idle. A non-positive threshold falls back to DefaultThreshold.
func NewTracker(threshold time.Duration) *Tracker {
	if threshold <= 0 {
		threshold = DefaultThreshold
	}
	return &Tracker{threshold: threshold}
}

// DefaultThreshold is used when no idle threshold is configured.
const DefaultThreshold = 5 * time.Minute

// Observe records how long the machine has been idle and returns the resulting
// state together with whether that state changed.
//
// The first call seeds the tracker and always reports no change, for the same
// reason the session tracker does: the agent is typically started by the
// service manager, and announcing "USER_ACTIVE" merely because the agent
// started would fill the event log with noise that says nothing about the user.
func (t *Tracker) Observe(idleFor time.Duration) (string, bool) {
	state := StateActive
	if idleFor >= t.threshold {
		state = StateIdle
	}

	if !t.seeded {
		t.current, t.seeded = state, true
		return state, false
	}
	if state == t.current {
		return state, false
	}

	t.current = state
	return state, true
}

// Current returns the state the tracker last settled on, or an empty string
// before the first observation.
func (t *Tracker) Current() string {
	return t.current
}

// EventName maps an activity state onto the event the backend records.
func EventName(state string) string {
	if state == StateIdle {
		return client.EventUserIdle
	}
	return client.EventUserActive
}

// IdleDuration reports how long the machine has been idle.
//
// A false second return value means the platform could not answer. The caller
// must then skip this poll rather than assume activity, because guessing would
// emit transitions that never happened.
func IdleDuration() (time.Duration, bool) {
	return idleDuration()
}
