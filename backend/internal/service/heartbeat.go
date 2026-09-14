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

var ErrInvalidAgentCredentials = errors.New("invalid agent credentials")

type AgentHeartbeatRepository interface {
	FindAuth(context.Context, string) (*model.AgentAuth, error)
	RecordHeartbeat(context.Context, string, time.Time) error
}

type SystemHeartbeatRepository interface {
	RecordHeartbeat(context.Context, string, time.Time) error
	SaveHealth(context.Context, model.HealthReport) error
}

type HeartbeatService struct {
	agents  AgentHeartbeatRepository
	systems SystemHeartbeatRepository
}

func NewHeartbeatService(agents AgentHeartbeatRepository, systems SystemHeartbeatRepository) *HeartbeatService {
	return &HeartbeatService{agents: agents, systems: systems}
}

func (s *HeartbeatService) Receive(ctx context.Context, agentCode, credential string, input model.HeartbeatRequest) (*model.HeartbeatResponse, error) {
	auth, err := s.agents.FindAuth(ctx, strings.TrimSpace(agentCode))
	if err != nil {
		return nil, fmt.Errorf("find agent: %w", err)
	}
	if auth == nil || auth.Agent.Status != "approved" || password.VerifyPassword(auth.CredentialHash, credential) != nil {
		return nil, ErrInvalidAgentCredentials
	}
	if strings.TrimSpace(input.AgentID) == "" || input.AgentID != auth.Agent.ID {
		return nil, errors.New("agent id does not match credentials")
	}
	if strings.TrimSpace(input.SystemID) == "" || input.SystemID != auth.Agent.SystemID {
		return nil, errors.New("system id does not match agent")
	}
	if input.Timestamp.IsZero() {
		input.Timestamp = time.Now().UTC()
	}
	if input.Timestamp.After(time.Now().Add(time.Minute)) {
		return nil, errors.New("heartbeat timestamp is in the future")
	}
	if err := s.agents.RecordHeartbeat(ctx, auth.Agent.ID, input.Timestamp); err != nil {
		return nil, err
	}
	if err := s.systems.RecordHeartbeat(ctx, auth.Agent.SystemID, input.Timestamp); err != nil {
		return nil, err
	}
	if input.Health != nil {
		if err := validateHealth(*input.Health); err != nil {
			return nil, err
		}
		input.Health.SystemID = auth.Agent.SystemID
		input.Health.RecordedAt = input.Timestamp
		if err := s.systems.SaveHealth(ctx, *input.Health); err != nil {
			return nil, err
		}
	}
	return &model.HeartbeatResponse{AgentID: auth.Agent.ID, SystemID: auth.Agent.SystemID, Status: "online", ReceivedAt: time.Now().UTC()}, nil
}

func validateHealth(report model.HealthReport) error {
	if report.CPUPercent < 0 || report.CPUPercent > 100 {
		return errors.New("cpu percent must be between 0 and 100")
	}
	if report.MemoryPercent < 0 || report.MemoryPercent > 100 {
		return errors.New("memory percent must be between 0 and 100")
	}
	if report.DiskPercent < 0 || report.DiskPercent > 100 {
		return errors.New("disk percent must be between 0 and 100")
	}
	if report.BatteryPercent != nil && (*report.BatteryPercent < 0 || *report.BatteryPercent > 100) {
		return errors.New("battery percent must be between 0 and 100")
	}
	if report.UptimeSeconds < 0 {
		return errors.New("uptime seconds cannot be negative")
	}
	switch report.AgentHealth {
	case "healthy", "warning", "unhealthy":
		return nil
	default:
		return fmt.Errorf("invalid agent health %q", report.AgentHealth)
	}
}
