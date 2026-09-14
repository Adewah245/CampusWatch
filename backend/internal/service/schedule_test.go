package service

import (
	"testing"
	"time"

	"CampusWatch/backend/internal/model"
)

func TestCheckScheduleOpen(t *testing.T) {
	schedule := model.Schedule{DayOfWeek: 1, OpeningTime: "08:00", ClosingTime: "17:00"}
	result, err := CheckSchedule(schedule, time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC))
	if err != nil || !result.Allowed || result.State != "open" {
		t.Fatalf("expected open schedule, got %+v, %v", result, err)
	}
}

func TestCheckScheduleBreakAndAfterHours(t *testing.T) {
	start, end := "12:00", "13:00"
	schedule := model.Schedule{DayOfWeek: 1, OpeningTime: "08:00", BreakStart: &start, BreakEnd: &end, ClosingTime: "17:00"}
	breakResult, _ := CheckSchedule(schedule, time.Date(2026, 9, 14, 12, 30, 0, 0, time.UTC))
	lateResult, _ := CheckSchedule(schedule, time.Date(2026, 9, 14, 18, 0, 0, 0, time.UTC))
	if breakResult.Allowed || breakResult.State != "break" {
		t.Fatalf("expected break, got %+v", breakResult)
	}
	if lateResult.Allowed || lateResult.State != "after_hours" {
		t.Fatalf("expected after-hours, got %+v", lateResult)
	}
}
