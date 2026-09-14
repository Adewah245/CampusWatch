package service

import (
	"context"
	"testing"
	"time"

	"CampusWatch/backend/internal/model"
)

type reportRepoStub struct{ period string }

func (r *reportRepoStub) Generate(_ context.Context, kind, period string, from, to time.Time) (*model.Report, error) {
	r.period = period
	return &model.Report{Kind: kind, Period: period, From: from, To: to}, nil
}

func TestReportServiceAcceptsWeeklyPeriod(t *testing.T) {
	repo := &reportRepoStub{}
	result, err := NewReportService(repo).Generate(context.Background(), "usage", "weekly", time.Now())
	if err != nil {
		t.Fatalf("generate report: %v", err)
	}
	if result.Period != "weekly" || repo.period != "weekly" {
		t.Fatalf("unexpected report: %+v", result)
	}
}

func TestReportServiceRejectsInvalidPeriod(t *testing.T) {
	if _, err := NewReportService(&reportRepoStub{}).Generate(context.Background(), "usage", "yearly", time.Now()); err == nil {
		t.Fatal("expected invalid period to fail")
	}
}
