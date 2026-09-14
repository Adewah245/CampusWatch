package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/auth/session"
	"CampusWatch/backend/internal/model"
)

type AgentRepository interface {
	Create(context.Context, model.Agent, string) (*model.Agent, error)
	List(context.Context) ([]model.Agent, error)
	Approve(context.Context, string, string) (*model.Agent, error)
}

type AgentService struct{ repo AgentRepository }

func NewAgentService(repo AgentRepository) *AgentService { return &AgentService{repo: repo} }

func (s *AgentService) Register(ctx context.Context, input model.AgentRegistrationRequest) (*model.AgentCredentialResponse, error) {
	input.SystemID = strings.TrimSpace(input.SystemID)
	input.AgentCode = strings.TrimSpace(input.AgentCode)
	input.AgentVersion = strings.TrimSpace(input.AgentVersion)
	if input.SystemID == "" {
		return nil, errors.New("system id is required")
	}
	if input.AgentCode == "" {
		return nil, errors.New("agent code is required")
	}
	if input.AgentVersion == "" {
		return nil, errors.New("agent version is required")
	}
	credential, err := session.GenerateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generate agent credential: %w", err)
	}
	hash, err := password.HashPassword(credential)
	if err != nil {
		return nil, fmt.Errorf("hash agent credential: %w", err)
	}
	agent, err := s.repo.Create(ctx, model.Agent{SystemID: input.SystemID, AgentCode: input.AgentCode, AgentVersion: input.AgentVersion}, hash)
	if err != nil {
		return nil, err
	}
	return &model.AgentCredentialResponse{Agent: *agent, Credential: credential}, nil
}

func (s *AgentService) List(ctx context.Context) ([]model.Agent, error) { return s.repo.List(ctx) }

func (s *AgentService) Approve(ctx context.Context, id string) (*model.AgentCredentialResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("agent id is required")
	}
	credential, err := session.GenerateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generate agent credential: %w", err)
	}
	hash, err := password.HashPassword(credential)
	if err != nil {
		return nil, fmt.Errorf("hash agent credential: %w", err)
	}
	agent, err := s.repo.Approve(ctx, id, hash)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, nil
	}
	return &model.AgentCredentialResponse{Agent: *agent, Credential: credential}, nil
}
