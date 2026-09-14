package service

import (
	"context"
	"testing"
	"time"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/model"
)

type eventRepoStub struct{ event *model.Event }

func (r *eventRepoStub) Create(_ context.Context, value model.Event) (*model.Event, error) {
	r.event = &value
	return &value, nil
}
func (r *eventRepoStub) ListBySystem(context.Context, string) ([]model.Event, error) { return nil, nil }

func TestEventServiceRecordsAllowedEvent(t *testing.T) {
	hash, err := password.HashPassword("secret")
	if err != nil {
		t.Fatalf("hash credential: %v", err)
	}
	agents := &heartbeatAgentRepoStub{auth: &model.AgentAuth{Agent: model.Agent{ID: "agent-1", SystemID: "system-1", Status: "approved"}, CredentialHash: hash}}
	events := &eventRepoStub{}
	result, err := NewEventService(agents, events).Record(context.Background(), "agent-code", "secret", model.EventRequest{AgentID: "agent-1", SystemID: "system-1", EventType: "user_login", OccurredAt: time.Now(), Payload: []byte(`{"username":"student"}`)})
	if err != nil {
		t.Fatalf("record event: %v", err)
	}
	if result.EventType != "USER_LOGIN" {
		t.Fatalf("expected normalized event type, got %q", result.EventType)
	}
}
