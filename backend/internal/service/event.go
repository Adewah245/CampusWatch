package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/model"
)

type EventRepository interface {
	Create(context.Context, model.Event) (*model.Event, error)
	ListBySystem(context.Context, string) ([]model.Event, error)
}

type EventService struct {
	agents AgentHeartbeatRepository
	events EventRepository
}

func NewEventService(agents AgentHeartbeatRepository, events EventRepository) *EventService {
	return &EventService{agents: agents, events: events}
}

func (s *EventService) Record(ctx context.Context, agentCode, credential string, input model.EventRequest) (*model.Event, error) {
	auth, err := s.agents.FindAuth(ctx, strings.TrimSpace(agentCode))
	if err != nil {
		return nil, fmt.Errorf("find agent: %w", err)
	}
	if auth == nil || auth.Agent.Status != "approved" || password.VerifyPassword(auth.CredentialHash, credential) != nil {
		return nil, ErrInvalidAgentCredentials
	}
	if input.AgentID != auth.Agent.ID || input.SystemID != auth.Agent.SystemID {
		return nil, errors.New("event identity does not match agent")
	}
	input.EventType = strings.ToUpper(strings.TrimSpace(input.EventType))
	if !validEventType(input.EventType) {
		return nil, fmt.Errorf("invalid event type %q", input.EventType)
	}
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	}
	if len(input.Payload) == 0 {
		input.Payload = json.RawMessage(`{}`)
	}
	if !json.Valid(input.Payload) {
		return nil, errors.New("event payload must be valid JSON")
	}
	return s.events.Create(ctx, model.Event{SystemID: input.SystemID, AgentID: &input.AgentID, EventType: input.EventType, Payload: input.Payload, OccurredAt: input.OccurredAt})
}

func (s *EventService) ListBySystem(ctx context.Context, systemID string) ([]model.Event, error) {
	if strings.TrimSpace(systemID) == "" {
		return nil, errors.New("system id is required")
	}
	return s.events.ListBySystem(ctx, systemID)
}

func validEventType(value string) bool {
	switch value {
	case "SYSTEM_REGISTERED", "SYSTEM_APPROVED", "SYSTEM_ONLINE", "SYSTEM_OFFLINE", "USER_LOGIN", "USER_LOGOUT", "USER_IDLE", "USER_ACTIVE", "AFTER_HOURS_USAGE", "HEALTH_WARNING", "FAULT_DETECTED", "ISSUE_REPORTED":
		return true
	default:
		return false
	}
}
