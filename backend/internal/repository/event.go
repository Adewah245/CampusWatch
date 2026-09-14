package repository

import (
	"context"
	"database/sql"
	"fmt"

	"CampusWatch/backend/internal/model"
)

type EventRepository struct{ db *sql.DB }

func NewEventRepository(db *sql.DB) *EventRepository { return &EventRepository{db: db} }

const eventColumns = "id, system_id, agent_id, event_type, payload, occurred_at, created_at"

func scanEvent(scanner interface{ Scan(...any) error }) (*model.Event, error) {
	var value model.Event
	err := scanner.Scan(&value.ID, &value.SystemID, &value.AgentID, &value.EventType, &value.Payload, &value.OccurredAt, &value.CreatedAt)
	return &value, err
}

func (r *EventRepository) Create(ctx context.Context, value model.Event) (*model.Event, error) {
	row := r.db.QueryRowContext(ctx, `INSERT INTO events (system_id, agent_id, event_type, payload, occurred_at) VALUES ($1, $2, $3, $4, $5) RETURNING `+eventColumns, value.SystemID, value.AgentID, value.EventType, value.Payload, value.OccurredAt)
	created, err := scanEvent(row)
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	return created, nil
}

func (r *EventRepository) ListBySystem(ctx context.Context, systemID string) ([]model.Event, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+eventColumns+` FROM events WHERE system_id = $1 ORDER BY occurred_at DESC`, systemID)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()
	var values []model.Event
	for rows.Next() {
		value, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		values = append(values, *value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	return values, nil
}
