package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"CampusWatch/backend/internal/model"
)

type SystemRepository interface {
	Create(context.Context, model.System) (*model.System, error)
	List(context.Context) ([]model.System, error)
	GetByID(context.Context, string) (*model.System, error)
	Update(context.Context, string, model.UpdateSystemRequest) (*model.System, error)
	Approve(context.Context, string) (*model.System, error)
	LatestHealth(context.Context, string) (*model.HealthReport, error)
}

type SystemService struct{ repo SystemRepository }

func NewSystemService(repo SystemRepository) *SystemService { return &SystemService{repo: repo} }

func (s *SystemService) Create(ctx context.Context, input model.CreateSystemRequest) (*model.System, error) {
	input.LocationID = strings.TrimSpace(input.LocationID)
	if input.LocationID == "" {
		return nil, errors.New("location id is required")
	}
	input.Hostname = strings.TrimSpace(input.Hostname)
	if input.Hostname == "" {
		return nil, errors.New("hostname is required")
	}
	input.DeviceName = strings.TrimSpace(input.DeviceName)
	if input.DeviceName == "" {
		return nil, errors.New("device name is required")
	}
	input.OperatingSystem = strings.TrimSpace(input.OperatingSystem)
	if input.OperatingSystem == "" {
		return nil, errors.New("operating system is required")
	}
	return s.repo.Create(ctx, model.System{LocationID: input.LocationID, Hostname: input.Hostname, DeviceName: input.DeviceName, OperatingSystem: input.OperatingSystem, OSVersion: strings.TrimSpace(input.OSVersion), SystemStatus: "pending", MonitorStatus: "offline"})
}

func (s *SystemService) List(ctx context.Context) ([]model.System, error) { return s.repo.List(ctx) }

func (s *SystemService) GetByID(ctx context.Context, id string) (*model.System, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("system id is required")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *SystemService) Update(ctx context.Context, id string, input model.UpdateSystemRequest) (*model.System, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("system id is required")
	}
	if input.LocationID != nil {
		value := strings.TrimSpace(*input.LocationID)
		if value == "" {
			return nil, errors.New("location id cannot be empty")
		}
		input.LocationID = &value
	}
	if input.Hostname != nil {
		value := strings.TrimSpace(*input.Hostname)
		if value == "" {
			return nil, errors.New("hostname cannot be empty")
		}
		input.Hostname = &value
	}
	if input.DeviceName != nil {
		value := strings.TrimSpace(*input.DeviceName)
		if value == "" {
			return nil, errors.New("device name cannot be empty")
		}
		input.DeviceName = &value
	}
	if input.OperatingSystem != nil {
		value := strings.TrimSpace(*input.OperatingSystem)
		if value == "" {
			return nil, errors.New("operating system cannot be empty")
		}
		input.OperatingSystem = &value
	}
	if input.SystemStatus != nil {
		value := strings.TrimSpace(*input.SystemStatus)
		if !systemStatusValid(value) {
			return nil, fmt.Errorf("invalid system status %q", value)
		}
		input.SystemStatus = &value
	}
	if input.MonitorStatus != nil {
		value := strings.TrimSpace(*input.MonitorStatus)
		if !monitorStatusValid(value) {
			return nil, fmt.Errorf("invalid monitor status %q", value)
		}
		input.MonitorStatus = &value
	}
	return s.repo.Update(ctx, id, input)
}

func (s *SystemService) Approve(ctx context.Context, id string) (*model.System, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("system id is required")
	}
	return s.repo.Approve(ctx, id)
}

func (s *SystemService) LatestHealth(ctx context.Context, id string) (*model.HealthReport, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("system id is required")
	}
	return s.repo.LatestHealth(ctx, id)
}

func systemStatusValid(value string) bool {
	switch value {
	case "pending", "active", "suspended", "retired":
		return true
	default:
		return false
	}
}
func monitorStatusValid(value string) bool {
	switch value {
	case "online", "offline", "inactive", "maintenance", "suspended", "retired":
		return true
	default:
		return false
	}
}
