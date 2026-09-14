package service

import (
	"context"

	"CampusWatch/backend/internal/model"
)

type DashboardRepository interface {
	Summary(context.Context) (*model.DashboardSummary, error)
	Systems(context.Context) ([]model.System, error)
}

type DashboardService struct{ repo DashboardRepository }

func NewDashboardService(repo DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) Summary(ctx context.Context) (*model.DashboardSummary, error) {
	return s.repo.Summary(ctx)
}
func (s *DashboardService) Systems(ctx context.Context) ([]model.System, error) {
	return s.repo.Systems(ctx)
}
