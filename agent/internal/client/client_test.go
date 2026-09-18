package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testAgentCode    = "LAB-A-PC-001"
	testCredential   = "per-installation-secret"
	testAgentID      = "agent-0001"
	testSystemID     = "SYS-000001"
	testAgentVersion = "1.0.0"
)

func TestHeartbeatSendsCredentialsAndDecodesTheResponse(t *testing.T) {
	var (
		gotPath string
		gotCode string
		gotCred string
		gotBody HeartbeatRequest
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotCode = r.Header.Get(HeaderAgentCode)
		gotCred = r.Header.Get(HeaderAgentCredential)
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", contentType)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HeartbeatResponse{
			AgentID: testAgentID, SystemID: testSystemID, Status: "ONLINE", ReceivedAt: time.Now().UTC(),
		})
	}))
	defer server.Close()

	api := New(Config{ServerURL: server.URL, AgentCode: testAgentCode, Credential: testCredential})

	battery := 74.0
	response, err := api.Heartbeat(context.Background(), HeartbeatRequest{
		AgentID:      testAgentID,
		SystemID:     testSystemID,
		Timestamp:    time.Now().UTC(),
		AgentVersion: testAgentVersion,
		Health: &HealthReport{
			SystemID: testSystemID, CPUPercent: 42, MemoryPercent: 67, DiskPercent: 81,
			BatteryPercent: &battery, NetworkConnected: true, OperatingSystem: "linux",
			AgentHealth: "healthy", UptimeSeconds: 3600,
		},
	})
	if err != nil {
		t.Fatalf("Heartbeat() returned an unexpected error: %v", err)
	}

	if gotPath != pathHeartbeat {
		t.Errorf("request path = %q, want %q", gotPath, pathHeartbeat)
	}
	if gotCode != testAgentCode {
		t.Errorf("%s = %q, want %q", HeaderAgentCode, gotCode, testAgentCode)
	}
	if gotCred != testCredential {
		t.Errorf("%s = %q, want %q", HeaderAgentCredential, gotCred, testCredential)
	}
	if gotBody.SystemID != testSystemID {
		t.Errorf("request system_id = %q, want %q", gotBody.SystemID, testSystemID)
	}
	if gotBody.Health == nil || gotBody.Health.CPUPercent != 42 {
		t.Errorf("request health = %+v, want the CPU reading to survive encoding", gotBody.Health)
	}
	if response.Status != "ONLINE" {
		t.Errorf("response status = %q, want ONLINE", response.Status)
	}
}

func TestTrailingSlashInServerURLIsTolerated(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	api := New(Config{ServerURL: server.URL + "/", AgentCode: testAgentCode, Credential: testCredential})
	if err := api.SessionEvent(context.Background(), SessionEventRequest{Username: "student"}); err != nil {
		t.Fatalf("SessionEvent() returned an unexpected error: %v", err)
	}

	// A doubled slash would not match the backend's route.
	if gotPath != pathSessionEvent {
		t.Errorf("request path = %q, want %q", gotPath, pathSessionEvent)
	}
}

func TestUnauthorizedIsReportedAsASentinelError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid agent credentials", http.StatusUnauthorized)
	}))
	defer server.Close()

	api := New(Config{ServerURL: server.URL, AgentCode: testAgentCode, Credential: "wrong"})

	_, err := api.Heartbeat(context.Background(), HeartbeatRequest{})
	if err == nil {
		t.Fatal("Heartbeat() succeeded with rejected credentials, want an error")
	}

	// The agent must be able to tell a credential rejection apart from a
	// transient outage, because retrying the former never helps.
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("error = %v, want it to wrap ErrUnauthorized", err)
	}
}

func TestServerErrorIsReportedWithItsStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "database is unavailable", http.StatusInternalServerError)
	}))
	defer server.Close()

	api := New(Config{ServerURL: server.URL, AgentCode: testAgentCode, Credential: testCredential})

	err := api.Event(context.Background(), EventRequest{EventType: EventHealthWarning})
	if err == nil {
		t.Fatal("Event() succeeded against a failing backend, want an error")
	}

	var apiError *APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("error = %v, want an *APIError", err)
	}
	if apiError.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiError.StatusCode, http.StatusInternalServerError)
	}
}

func TestUnreachableServerIsReported(t *testing.T) {
	// Point at a closed port so the connection cannot be made.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close()

	api := New(Config{ServerURL: url, AgentCode: testAgentCode, Credential: testCredential, Timeout: time.Second})

	if _, err := api.Heartbeat(context.Background(), HeartbeatRequest{}); err == nil {
		t.Error("Heartbeat() succeeded against an unreachable backend, want an error")
	}
}

func TestEmptyTimeoutFallsBackToTheDefault(t *testing.T) {
	api := New(Config{ServerURL: "https://campuswatch.example.edu"})

	// A zero timeout would mean "no timeout at all", letting a stalled backend
	// block the agent indefinitely.
	if api.http.Timeout != DefaultTimeout {
		t.Errorf("client timeout = %s, want %s", api.http.Timeout, DefaultTimeout)
	}
}

func TestEventPayloadSurvivesEncoding(t *testing.T) {
	var gotBody EventRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	api := New(Config{ServerURL: server.URL, AgentCode: testAgentCode, Credential: testCredential})

	payload, err := json.Marshal(map[string]any{"metric": "disk_percent", "value": 92.5})
	if err != nil {
		t.Fatalf("could not encode the test payload: %v", err)
	}
	if err := api.Event(context.Background(), EventRequest{
		AgentID: testAgentID, SystemID: testSystemID,
		EventType: EventHealthWarning, Payload: payload, OccurredAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("Event() returned an unexpected error: %v", err)
	}

	if gotBody.EventType != EventHealthWarning {
		t.Errorf("event_type = %q, want %q", gotBody.EventType, EventHealthWarning)
	}
	if len(gotBody.Payload) == 0 {
		t.Error("payload was dropped during encoding")
	}
}
