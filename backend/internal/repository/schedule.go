package repository

import (
	"context"
	"database/sql"
	"fmt"

	"CampusWatch/backend/internal/model"
)

type ScheduleRepository struct{ db *sql.DB }

func NewScheduleRepository(db *sql.DB) *ScheduleRepository { return &ScheduleRepository{db: db} }

const scheduleColumns = "id, institution_id, day_of_week, opening_time::text, break_start::text, break_end::text, closing_time::text, created_at, updated_at"

func scanSchedule(scanner interface{ Scan(...any) error }) (*model.Schedule, error) {
	var value model.Schedule
	err := scanner.Scan(&value.ID, &value.InstitutionID, &value.DayOfWeek, &value.OpeningTime, &value.BreakStart, &value.BreakEnd, &value.ClosingTime, &value.CreatedAt, &value.UpdatedAt)
	return &value, err
}

func (r *ScheduleRepository) Create(ctx context.Context, input model.Schedule) (*model.Schedule, error) {
	value, err := scanSchedule(r.db.QueryRowContext(ctx, `INSERT INTO schedules (institution_id, day_of_week, opening_time, break_start, break_end, closing_time) VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+scheduleColumns, input.InstitutionID, input.DayOfWeek, input.OpeningTime, input.BreakStart, input.BreakEnd, input.ClosingTime))
	if err != nil {
		return nil, fmt.Errorf("create schedule: %w", err)
	}
	return value, nil
}

func (r *ScheduleRepository) List(ctx context.Context, institutionID string) ([]model.Schedule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+scheduleColumns+` FROM schedules WHERE institution_id = $1 ORDER BY day_of_week`, institutionID)
	if err != nil {
		return nil, fmt.Errorf("query schedules: %w", err)
	}
	defer rows.Close()
	var values []model.Schedule
	for rows.Next() {
		value, err := scanSchedule(rows)
		if err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}
		values = append(values, *value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schedules: %w", err)
	}
	return values, nil
}
