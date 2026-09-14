package service

import (
	"context"
	"testing"

	"CampusWatch/backend/internal/model"
)

type systemRepoStub struct{ created *model.System }

func (r *systemRepoStub) Create(_ context.Context, value model.System) (*model.System, error) {
	r.created = &value
	return &value, nil
}
func (r *systemRepoStub) List(context.Context) ([]model.System, error)           { return nil, nil }
func (r *systemRepoStub) GetByID(context.Context, string) (*model.System, error) { return nil, nil }
func (r *systemRepoStub) Update(context.Context, string, model.UpdateSystemRequest) (*model.System, error) {
	return nil, nil
}
func (r *systemRepoStub) Approve(context.Context, string) (*model.System, error) { return nil, nil }
func (r *systemRepoStub) LatestHealth(context.Context, string) (*model.HealthReport, error) {
	return nil, nil
}

func TestSystemServiceCreateDefaultsLifecycle(t *testing.T) {
	repo := &systemRepoStub{}
	svc := NewSystemService(repo)

	system, err := svc.Create(context.Background(), model.CreateSystemRequest{
		LocationID: "location-1", Hostname: " PC-001 ", DeviceName: " Lab PC 1 ", OperatingSystem: " Windows ",
	})
	if err != nil {
		t.Fatalf("create system: %v", err)
	}
	if system.SystemStatus != "pending" || system.MonitorStatus != "offline" {
		t.Fatalf("unexpected lifecycle: %+v", system)
	}
	if system.Hostname != "PC-001" || system.OperatingSystem != "Windows" {
		t.Fatalf("expected trimmed fields: %+v", system)
	}
}

func TestSystemServiceRejectsInvalidMonitorStatus(t *testing.T) {
	repo := &systemRepoStub{}
	svc := NewSystemService(repo)
	status := "unknown"
	if _, err := svc.Update(context.Background(), "system-1", model.UpdateSystemRequest{MonitorStatus: &status}); err == nil {
		t.Fatal("expected invalid monitor status to fail")
	}
}
