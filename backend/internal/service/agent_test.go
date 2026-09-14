package service

import (
	"context"
	"testing"

	"CampusWatch/backend/internal/model"
)

type agentRepoStub struct {
	agent          *model.Agent
	credentialHash string
}

func (r *agentRepoStub) Create(_ context.Context, value model.Agent, credentialHash string) (*model.Agent, error) {
	r.agent, r.credentialHash = &value, credentialHash
	return &value, nil
}
func (r *agentRepoStub) List(context.Context) ([]model.Agent, error) { return nil, nil }
func (r *agentRepoStub) Approve(_ context.Context, _ string, credentialHash string) (*model.Agent, error) {
	r.credentialHash = credentialHash
	return r.agent, nil
}

func TestAgentServiceRegisterCreatesPendingCredential(t *testing.T) {
	repo := &agentRepoStub{}
	svc := NewAgentService(repo)
	result, err := svc.Register(context.Background(), model.AgentRegistrationRequest{SystemID: "system-1", AgentCode: "agent-1", AgentVersion: "1.0.0"})
	if err != nil {
		t.Fatalf("register agent: %v", err)
	}
	if result.Credential == "" || repo.credentialHash == "" {
		t.Fatal("expected generated credential and stored hash")
	}
	if result.Credential == repo.credentialHash {
		t.Fatal("credential must not be stored as plaintext")
	}
}

func TestAgentServiceRegisterRequiresSystem(t *testing.T) {
	if _, err := NewAgentService(&agentRepoStub{}).Register(context.Background(), model.AgentRegistrationRequest{AgentCode: "agent-1", AgentVersion: "1.0.0"}); err == nil {
		t.Fatal("expected missing system id to fail")
	}
}
