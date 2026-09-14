package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"CampusWatch/backend/internal/model"
)

type ReportRepository interface {
	Generate(context.Context, string, string, time.Time, time.Time) (*model.Report, error)
}

type ReportService struct{ repo ReportRepository }

func NewReportService(repo ReportRepository) *ReportService { return &ReportService{repo: repo} }

func (s *ReportService) Generate(ctx context.Context, kind, period string, now time.Time) (*model.Report, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	period = strings.ToLower(strings.TrimSpace(period))
	if kind != "usage" && kind != "uptime" && kind != "sessions" && kind != "issues" && kind != "after-hours" {
		return nil, errors.New("invalid report kind")
	}
	var duration time.Duration
	switch period {
	case "daily":
		duration = 24 * time.Hour
	case "weekly":
		duration = 7 * 24 * time.Hour
	case "monthly":
		duration = 30 * 24 * time.Hour
	default:
		return nil, errors.New("period must be daily, weekly, or monthly")
	}
	to := now.UTC()
	from := to.Add(-duration)
	return s.repo.Generate(ctx, kind, period, from, to)
}
