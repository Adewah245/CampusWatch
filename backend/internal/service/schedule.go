package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"CampusWatch/backend/internal/model"
)

type ScheduleRepository interface {
	Create(context.Context, model.Schedule) (*model.Schedule, error)
	List(context.Context, string) ([]model.Schedule, error)
}

type ScheduleService struct{ repo ScheduleRepository }

func NewScheduleService(repo ScheduleRepository) *ScheduleService {
	return &ScheduleService{repo: repo}
}

func (s *ScheduleService) Create(ctx context.Context, input model.CreateScheduleRequest) (*model.Schedule, error) {
	input.InstitutionID = strings.TrimSpace(input.InstitutionID)
	if input.InstitutionID == "" {
		return nil, errors.New("institution id is required")
	}
	if input.DayOfWeek < 0 || input.DayOfWeek > 6 {
		return nil, errors.New("day of week must be between 0 and 6")
	}
	if _, err := parseClock(input.OpeningTime); err != nil {
		return nil, fmt.Errorf("opening time: %w", err)
	}
	if _, err := parseClock(input.ClosingTime); err != nil {
		return nil, fmt.Errorf("closing time: %w", err)
	}
	if input.BreakStart == nil && input.BreakEnd != nil || input.BreakStart != nil && input.BreakEnd == nil {
		return nil, errors.New("break start and break end must be provided together")
	}
	if input.BreakStart != nil {
		if _, err := parseClock(*input.BreakStart); err != nil {
			return nil, fmt.Errorf("break start: %w", err)
		}
		if _, err := parseClock(*input.BreakEnd); err != nil {
			return nil, fmt.Errorf("break end: %w", err)
		}
	}
	return s.repo.Create(ctx, model.Schedule{InstitutionID: input.InstitutionID, DayOfWeek: input.DayOfWeek, OpeningTime: input.OpeningTime, BreakStart: input.BreakStart, BreakEnd: input.BreakEnd, ClosingTime: input.ClosingTime})
}

func (s *ScheduleService) List(ctx context.Context, institutionID string) ([]model.Schedule, error) {
	if strings.TrimSpace(institutionID) == "" {
		return nil, errors.New("institution id is required")
	}
	return s.repo.List(ctx, institutionID)
}

func CheckSchedule(schedule model.Schedule, at time.Time) (model.ScheduleCheck, error) {
	if schedule.DayOfWeek != int(at.Weekday()) {
		return model.ScheduleCheck{Allowed: false, State: "closed", DayOfWeek: int(at.Weekday())}, nil
	}
	opening, err := parseClock(schedule.OpeningTime)
	if err != nil {
		return model.ScheduleCheck{}, err
	}
	closing, err := parseClock(schedule.ClosingTime)
	if err != nil {
		return model.ScheduleCheck{}, err
	}
	current := at.Hour()*60 + at.Minute()
	if current < opening {
		return model.ScheduleCheck{Allowed: false, State: "after_hours", DayOfWeek: int(at.Weekday())}, nil
	}
	if schedule.BreakStart != nil && schedule.BreakEnd != nil {
		start, _ := parseClock(*schedule.BreakStart)
		end, _ := parseClock(*schedule.BreakEnd)
		if current >= start && current < end {
			return model.ScheduleCheck{Allowed: false, State: "break", DayOfWeek: int(at.Weekday())}, nil
		}
	}
	if current >= closing {
		return model.ScheduleCheck{Allowed: false, State: "after_hours", DayOfWeek: int(at.Weekday())}, nil
	}
	return model.ScheduleCheck{Allowed: true, State: "open", DayOfWeek: int(at.Weekday())}, nil
}

func parseClock(value string) (int, error) {
	parsed, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, errors.New("must use HH:MM")
	}
	return parsed.Hour()*60 + parsed.Minute(), nil
}
