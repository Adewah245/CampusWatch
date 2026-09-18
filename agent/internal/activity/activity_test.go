package activity

import (
	"testing"
	"time"

	"CampusWatch/agent/internal/client"
)

func TestParseIdleField(t *testing.T) {
	tests := map[string]struct {
		field     string
		want      time.Duration
		wantValid bool
	}{
		"fractional seconds below a minute": {"0.00s", 0, true},
		"seconds with a fraction":            {"12.50s", 12500 * time.Millisecond, true},
		"minutes and seconds":                {"5:23", 5*time.Minute + 23*time.Second, true},
		"hours and minutes marked with m":    {"2:52m", 2*time.Hour + 52*time.Minute, true},
		"a dot means no measurable idle":     {".", 0, true},
		"a dash means the same":              {"-", 0, true},
		"old means a very long idle":         {"old", 24 * time.Hour, true},
		"garbage is rejected":                {"whenever", 0, false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, ok := parseIdleField(test.field)

			if ok != test.wantValid {
				t.Fatalf("parseIdleField(%q) valid = %t, want %t", test.field, ok, test.wantValid)
			}
			if ok && got != test.want {
				t.Errorf("parseIdleField(%q) = %v, want %v", test.field, got, test.want)
			}
		})
	}
}

func TestParseWIdlePicksTheShortestIdleTime(t *testing.T) {
	// If any session shows recent activity then somebody is at the machine, so
	// the shortest idle time is the right answer, not the longest.
	output := `student  tty7     :0               07:34    2:52m  4:02   0.19s /usr/lib/xorg/Xorg
james    pts/0    192.168.1.5      09:15    0.30s  0.12s  0.05s -zsh
`

	idle, ok := parseWIdle(output)
	if !ok {
		t.Fatal("parseWIdle() failed on a valid fixture")
	}
	if idle != 300*time.Millisecond {
		t.Errorf("parseWIdle() = %v, want 300ms", idle)
	}
}

func TestParseWIdleIgnoresShortLines(t *testing.T) {
	if _, ok := parseWIdle("james pts/0\n"); ok {
		t.Error("parseWIdle() accepted a line with no IDLE column")
	}
}

func TestTrackerSeedsWithoutReporting(t *testing.T) {
	tracker := NewTracker(time.Minute)

	// Announcing a state merely because the agent started would fill the event
	// log with noise that says nothing about the user.
	state, changed := tracker.Observe(30 * time.Minute)
	if changed {
		t.Fatalf("first Observe() reported a change to %s, want none", state)
	}
	if state != StateIdle {
		t.Errorf("Observe() = %s, want %s", state, StateIdle)
	}
}

func TestTrackerReportsTransitionsIntoAndOutOfIdle(t *testing.T) {
	tracker := NewTracker(time.Minute)
	tracker.Observe(0) // Seed as active.

	state, changed := tracker.Observe(90 * time.Second)
	if !changed || state != StateIdle {
		t.Fatalf("Observe() while idle = (%s, %t), want (%s, true)", state, changed, StateIdle)
	}

	state, changed = tracker.Observe(2 * time.Second)
	if !changed || state != StateActive {
		t.Fatalf("Observe() after activity = (%s, %t), want (%s, true)", state, changed, StateActive)
	}
}

func TestTrackerReportsNoChangeWhileTheStateHolds(t *testing.T) {
	tracker := NewTracker(time.Minute)
	tracker.Observe(0)
	tracker.Observe(90 * time.Second) // Now idle.

	if _, changed := tracker.Observe(5 * time.Minute); changed {
		t.Error("Observe() reported a change while the machine stayed idle")
	}
}

func TestNewTrackerFallsBackToTheDefaultThreshold(t *testing.T) {
	// A misconfigured threshold of zero would make every machine permanently
	// idle, so it must fall back rather than be taken literally.
	if got := NewTracker(0).threshold; got != DefaultThreshold {
		t.Errorf("threshold = %s, want %s", got, DefaultThreshold)
	}
}

func TestEventName(t *testing.T) {
	if got := EventName(StateIdle); got != client.EventUserIdle {
		t.Errorf("EventName(%s) = %q, want %q", StateIdle, got, client.EventUserIdle)
	}
	if got := EventName(StateActive); got != client.EventUserActive {
		t.Errorf("EventName(%s) = %q, want %q", StateActive, got, client.EventUserActive)
	}
}

func TestParseIOHIDIdleTime(t *testing.T) {
	// IOKit reports this value in nanoseconds.
	output := `+-o IOHIDSystem  <class IOHIDSystem, id 0x100000abc>
    {
      "HIDIdleTime" = 45000000000
    }
`

	idle, ok := parseIOHIDIdleTime(output)
	if !ok {
		t.Fatal("parseIOHIDIdleTime() failed on valid output")
	}
	if idle != 45*time.Second {
		t.Errorf("parseIOHIDIdleTime() = %v, want 45s", idle)
	}
}

func TestParseIOHIDIdleTimeRejectsMissingValue(t *testing.T) {
	if _, ok := parseIOHIDIdleTime("no such key here\n"); ok {
		t.Error("parseIOHIDIdleTime() accepted output with no HIDIdleTime")
	}
}

func TestParseXprintidle(t *testing.T) {
	idle, ok := parseXprintidle("1500\n")
	if !ok {
		t.Fatal("parseXprintidle() failed on valid output")
	}
	if idle != 1500*time.Millisecond {
		t.Errorf("parseXprintidle() = %v, want 1.5s", idle)
	}
}
