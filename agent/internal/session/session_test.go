package session

import (
	"reflect"
	"testing"
	"time"

	"CampusWatch/agent/internal/client"
)

func TestParseWhoExtractsUniqueSortedUsernames(t *testing.T) {
	// One person may hold several terminal sessions, so duplicates must collapse,
	// and the order must be stable so log output and API calls are predictable.
	output := `student  tty7     2024-07-02 07:34 (:0)
james    pts/0    2024-07-02 09:15 (192.168.1.5)
student  pts/1    2024-07-02 09:20 (192.168.1.5)
`

	got := parseWho(output)
	want := []string{"james", "student"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseWho() = %v, want %v", got, want)
	}
}

func TestParseWhoOnEmptyOutput(t *testing.T) {
	if got := parseWho(""); len(got) != 0 {
		t.Errorf("parseWho(\"\") = %v, want an empty list", got)
	}
}

func TestParseQUserSkipsTheHeaderRow(t *testing.T) {
	output := ` USERNAME              SESSIONNAME        ID  STATE   IDLE TIME   LOGON TIME
 student               console             1  Active          .   7/2/2024 7:34 AM
`

	got := parseQUser(output)
	want := []string{"student"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseQUser() = %v, want %v", got, want)
	}
}

func TestParseQUserHandlesTheNoSessionsNotice(t *testing.T) {
	// quser reports this when nobody is logged in, and it must not be mistaken
	// for a username.
	got := parseQUser("No User exists for *\n")

	if len(got) != 0 {
		t.Errorf("parseQUser() = %v, want an empty list", got)
	}
}

func TestParseQUserStripsTheCurrentSessionMarker(t *testing.T) {
	output := `>student    console    1  Active   7/2/2024 7:34 AM`

	got := parseQUser(output)
	want := []string{"student"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseQUser() = %v, want %v", got, want)
	}
}

func TestTrackerSeedsWithoutReporting(t *testing.T) {
	tracker := NewTracker()
	at := time.Now()

	// The agent is normally started by the service manager, so whoever is
	// already logged in did not just log in. Reporting them would fabricate
	// session history.
	if transitions := tracker.Observe([]string{"student"}, at); transitions != nil {
		t.Fatalf("first Observe() = %v, want no transitions", transitions)
	}
	if got := tracker.Current(); !reflect.DeepEqual(got, []string{"student"}) {
		t.Errorf("Current() = %v, want [student]", got)
	}
}

func TestTrackerReportsLoginAndLogout(t *testing.T) {
	tracker := NewTracker()
	at := time.Now()
	tracker.Observe([]string{"student"}, at) // Seed.

	login := tracker.Observe([]string{"student", "james"}, at)
	if len(login) != 1 || login[0].Username != "james" || login[0].Event != client.EventUserLogin {
		t.Fatalf("Observe() after a new login = %v, want a single USER_LOGIN for james", login)
	}

	logout := tracker.Observe([]string{"james"}, at)
	if len(logout) != 1 || logout[0].Username != "student" || logout[0].Event != client.EventUserLogout {
		t.Fatalf("Observe() after a logout = %v, want a single USER_LOGOUT for student", logout)
	}
}

func TestTrackerReportsNoTransitionWhenNothingChanges(t *testing.T) {
	tracker := NewTracker()
	at := time.Now()
	tracker.Observe([]string{"student"}, at)

	if transitions := tracker.Observe([]string{"student"}, at); transitions != nil {
		t.Errorf("Observe() with an unchanged set = %v, want no transitions", transitions)
	}
}

func TestTrackerOrdersTransitionsByUsername(t *testing.T) {
	tracker := NewTracker()
	at := time.Now()
	tracker.Observe([]string{"zoe"}, at)

	// Two users arrive at once; the result must be deterministic rather than
	// following Go's randomised map iteration order.
	transitions := tracker.Observe([]string{"zoe", "adam", "michael"}, at)
	if len(transitions) != 2 {
		t.Fatalf("Observe() = %v, want two transitions", transitions)
	}
	if transitions[0].Username != "adam" || transitions[1].Username != "michael" {
		t.Errorf("Observe() = %v, want adam before michael", transitions)
	}
}
