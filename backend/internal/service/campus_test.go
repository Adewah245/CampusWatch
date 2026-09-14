package service

import (
	"context"
	"testing"

	"CampusWatch/backend/internal/model"
)

type campusRepoStub struct{}

func (r *campusRepoStub) Create(ctx context.Context, campus model.Campus) (*model.Campus, error) {
	return &campus, nil
}

func (r *campusRepoStub) List(ctx context.Context) ([]model.Campus, error) {
	return []model.Campus{}, nil
}

func (r *campusRepoStub) GetByID(ctx context.Context, id string) (*model.Campus, error) {
	return &model.Campus{ID: id}, nil
}

func (r *campusRepoStub) Update(ctx context.Context, id string, input model.UpdateCampusRequest) (*model.Campus, error) {
	return &model.Campus{ID: id, Name: "Updated Campus"}, nil
}

func (r *campusRepoStub) Delete(ctx context.Context, id string) (bool, error) {
	return true, nil
}

func TestCampusServiceCreate(t *testing.T) {
	repo := &campusRepoStub{}
	svc := NewCampusService(repo)

	campus, err := svc.Create(context.Background(), "institution-123", "Main Campus", "main-campus", "active")
	if err != nil {
		t.Fatalf("create campus: %v", err)
	}
	if campus == nil {
		t.Fatal("expected created campus")
	}
	if campus.Name != "Main Campus" {
		t.Fatalf("expected name to be preserved, got %q", campus.Name)
	}
	if campus.Slug != "main-campus" {
		t.Fatalf("expected slug to be normalized, got %q", campus.Slug)
	}
}

func TestCampusServiceCreateRejectsEmptyName(t *testing.T) {
	repo := &campusRepoStub{}
	svc := NewCampusService(repo)

	if _, err := svc.Create(context.Background(), "institution-123", "   ", "main-campus", "active"); err == nil {
		t.Fatal("expected empty campus name to fail")
	}
}
