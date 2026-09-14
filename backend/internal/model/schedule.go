package model

import "time"

type Schedule struct {
	ID            string    `json:"id"`
	InstitutionID string    `json:"institution_id"`
	DayOfWeek     int       `json:"day_of_week"`
	OpeningTime   string    `json:"opening_time"`
	BreakStart    *string   `json:"break_start,omitempty"`
	BreakEnd      *string   `json:"break_end,omitempty"`
	ClosingTime   string    `json:"closing_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ScheduleCheck struct {
	Allowed   bool   `json:"allowed"`
	State     string `json:"state"`
	DayOfWeek int    `json:"day_of_week"`
}

type CreateScheduleRequest struct {
	InstitutionID string  `json:"institution_id"`
	DayOfWeek     int     `json:"day_of_week"`
	OpeningTime   string  `json:"opening_time"`
	BreakStart    *string `json:"break_start,omitempty"`
	BreakEnd      *string `json:"break_end,omitempty"`
	ClosingTime   string  `json:"closing_time"`
}
