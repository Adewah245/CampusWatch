package service

import (
	"context"
	"testing"
	"time"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/model"
)

type heartbeatAgentRepoStub struct {
	auth     *model.AgentAuth
	recorded bool
}

func (r *heartbeatAgentRepoStub) FindAuth(context.Context, string) (*model.AgentAuth, error) {
	return r.auth, nil
}
func (r *heartbeatAgentRepoStub) RecordHeartbeat(context.Context, string, time.Time) error {
	r.recorded = true
	return nil
}

type heartbeatSystemRepoStub struct{ recorded bool }

func (r *heartbeatSystemRepoStub) RecordHeartbeat(context.Context, string, time.Time) error {
	r.recorded = true
	return nil
}
func (r *heartbeatSystemRepoStub) SaveHealth(context.Context, model.HealthReport) error { return nil }

func TestHeartbeatServiceMarksApprovedAgentOnline(t *testing.T) {
	credential := "agent-secret"
	hash, err := password.HashPassword(credential)
	if err != nil {
		t.Fatalf("hash credential: %v", err)
	}
	agents := &heartbeatAgentRepoStub{auth: &model.AgentAuth{Agent: model.Agent{ID: "agent-1", SystemID: "system-1", Status: "approved"}, CredentialHash: hash}}
	systems := &heartbeatSystemRepoStub{}
	svc := NewHeartbeatService(agents, systems)

	result, err := svc.Receive(context.Background(), "agent-code", credential, model.HeartbeatRequest{AgentID: "agent-1", SystemID: "system-1"})
	if err != nil {
		t.Fatalf("receive heartbeat: %v", err)
	}
	if result.Status != "online" || !agents.recorded || !systems.recorded {
		t.Fatalf("unexpected heartbeat result: %+v", result)
	}
}

func TestHeartbeatServiceRejectsPendingAgent(t *testing.T) {
	agents := &heartbeatAgentRepoStub{auth: &model.AgentAuth{Agent: model.Agent{ID: "agent-1", SystemID: "system-1", Status: "pending"}, CredentialHash: "unused"}}
	_, err := NewHeartbeatService(agents, &heartbeatSystemRepoStub{}).Receive(context.Background(), "agent-code", "secret", model.HeartbeatRequest{AgentID: "agent-1", SystemID: "system-1"})
	if err != ErrInvalidAgentCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}
