// Package session detects who is logged in to the monitored computer.
//
// The agent reports operating system accounts, not CampusWatch identities.
// README section 7 explains why this separation exists: a laboratory may run
// every machine as a shared "student" account, so the OS account alone cannot
// say who is actually present. Mapping the OS account onto an institutional
// identity is the backend's job, and the agent deliberately does not attempt it.
//
// Detection is by observation of the operating system's own login records. No
// command line, window title, keystroke or file content is ever inspected.
package session

import (
	"sort"
	"strings"
	"time"

	"CampusWatch/agent/internal/client"
)

// Transition records a user arriving at or leaving the machine.
type Transition struct {
	// Username is the operating system account.
	Username string
	// Event is one of the client event name constants.
	Event string
	// At is when the change was observed.
	At time.Time
}

// Tracker turns a series of login snapshots into login and logout transitions.
//
// It is not safe for concurrent use; the agent drives it from a single polling
// goroutine.
type Tracker struct {
	known  map[string]struct{}
	seeded bool
}

// NewTracker returns an empty tracker.
func NewTracker() *Tracker {
	return &Tracker{known: make(map[string]struct{})}
}

// Observe compares the current set of logged-in users against the previous
// snapshot and returns the transitions between them, ordered by username so the
// agent's log output and API calls are deterministic.
//
// The first call seeds the tracker and deliberately reports nothing. The agent
// is normally started by the service manager at boot and may also be restarted
// at any moment, so users already logged in at that instant did not just log in.
// Reporting them as fresh logins would fabricate history in a platform whose
// whole purpose is keeping accurate session records. The trade-off is that a
// login occurring while the agent is stopped is not reported; the sessions the
// backend already holds remain the source of truth for that window.
func (t *Tracker) Observe(usernames []string, at time.Time) []Transition {
	current := make(map[string]struct{}, len(usernames))
	for _, username := range usernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}
		current[username] = struct{}{}
	}

	if !t.seeded {
		t.known = current
		t.seeded = true
		return nil
	}

	var transitions []Transition
	for username := range current {
		if _, seen := t.known[username]; !seen {
			transitions = append(transitions, Transition{Username: username, Event: client.EventUserLogin, At: at})
		}
	}
	for username := range t.known {
		if _, present := current[username]; !present {
			transitions = append(transitions, Transition{Username: username, Event: client.EventUserLogout, At: at})
		}
	}

	t.known = current

	sort.Slice(transitions, func(i, j int) bool {
		if transitions[i].Username != transitions[j].Username {
			return transitions[i].Username < transitions[j].Username
		}
		return transitions[i].Event < transitions[j].Event
	})
	return transitions
}

// Current lists the users the tracker currently believes are logged in.
func (t *Tracker) Current() []string {
	usernames := make([]string, 0, len(t.known))
	for username := range t.known {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)
	return usernames
}

// LoggedInUsers returns the operating system accounts currently logged in.
//
// A non-nil error means the platform query failed, in which case the caller
// should skip this poll rather than treat the empty list as "everyone logged
// out", which would fabricate logout events.
func LoggedInUsers() ([]string, error) {
	return listLoggedInUsers()
}
