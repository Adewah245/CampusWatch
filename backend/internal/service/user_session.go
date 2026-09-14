package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/model"
)

type UserSessionRepository interface {
	Create(context.Context, model.UserSession) (*model.UserSession, error)
	UpdateEvent(context.Context, string, string, time.Time) (*model.UserSession, error)
	ListBySystem(context.Context, string) ([]model.UserSession, error)
}

type UserSessionService struct {
	agents   AgentHeartbeatRepository
	sessions UserSessionRepository
}

func NewUserSessionService(agents AgentHeartbeatRepository, sessions UserSessionRepository) *UserSessionService {
	return &UserSessionService{agents: agents, sessions: sessions}
}

func (s *UserSessionService) Record(ctx context.Context, agentCode, credential string, input model.UserSessionEventRequest) (*model.UserSession, error) {
	auth, err := s.agents.FindAuth(ctx, strings.TrimSpace(agentCode))
	if err != nil {
		return nil, fmt.Errorf("find agent: %w", err)
	}
	if auth == nil || auth.Agent.Status != "approved" || password.VerifyPassword(auth.CredentialHash, credential) != nil {
		return nil, ErrInvalidAgentCredentials
	}
	if input.AgentID != auth.Agent.ID || input.SystemID != auth.Agent.SystemID {
		return nil, errors.New("session identity does not match agent")
	}
	input.Event = strings.ToLower(strings.TrimSpace(input.Event))
	if input.Timestamp.IsZero() {
		input.Timestamp = time.Now().UTC()
	}
	switch input.Event {
	case "login":
		input.Username = strings.TrimSpace(input.Username)
		if input.Username == "" {
			return nil, errors.New("username is required")
		}
		return s.sessions.Create(ctx, model.UserSession{SystemID: input.SystemID, Username: input.Username, State: "logged_in_active", LoginAt: input.Timestamp})
	case "logout", "active", "idle":
		if strings.TrimSpace(input.SessionID) == "" {
			return nil, errors.New("session id is required")
		}
		state := map[string]string{"logout": "logged_out", "active": "logged_in_active", "idle": "logged_in_idle"}[input.Event]
		value, err := s.sessions.UpdateEvent(ctx, input.SessionID, state, input.Timestamp)
		if err != nil {
			return nil, err
		}
		if value == nil {
			return nil, errors.New("user session not found")
		}
		return value, nil
	default:
		return nil, fmt.Errorf("invalid session event %q", input.Event)
	}
}

func (s *UserSessionService) ListBySystem(ctx context.Context, systemID string) ([]model.UserSession, error) {
	if strings.TrimSpace(systemID) == "" {
		return nil, errors.New("system id is required")
	}
	return s.sessions.ListBySystem(ctx, systemID)
}
