package service

import (
	"context"
	"testing"
	"time"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/model"
)

type userSessionRepoStub struct{ created *model.UserSession }

func (r *userSessionRepoStub) Create(_ context.Context, value model.UserSession) (*model.UserSession, error) {
	value.ID = "session-1"
	r.created = &value
	return &value, nil
}
func (r *userSessionRepoStub) UpdateEvent(context.Context, string, string, time.Time) (*model.UserSession, error) {
	return r.created, nil
}
func (r *userSessionRepoStub) ListBySystem(context.Context, string) ([]model.UserSession, error) {
	return nil, nil
}

func TestUserSessionServiceRecordsLogin(t *testing.T) {
	hash, err := password.HashPassword("secret")
	if err != nil {
		t.Fatalf("hash credential: %v", err)
	}
	agents := &heartbeatAgentRepoStub{auth: &model.AgentAuth{Agent: model.Agent{ID: "agent-1", SystemID: "system-1", Status: "approved"}, CredentialHash: hash}}
	sessions := &userSessionRepoStub{}
	result, err := NewUserSessionService(agents, sessions).Record(context.Background(), "agent-code", "secret", model.UserSessionEventRequest{AgentID: "agent-1", SystemID: "system-1", Username: "student", Event: "login"})
	if err != nil {
		t.Fatalf("record login: %v", err)
	}
	if result.State != "logged_in_active" || result.Username != "student" {
		t.Fatalf("unexpected session: %+v", result)
	}
}
