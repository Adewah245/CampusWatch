package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"CampusWatch/backend/internal/model"
)

type UserSessionRepository struct{ db *sql.DB }

func NewUserSessionRepository(db *sql.DB) *UserSessionRepository {
	return &UserSessionRepository{db: db}
}

const userSessionColumns = "id, system_id, username, state, login_at, logout_at, last_activity_at, duration_seconds, created_at"

func scanUserSession(scanner interface{ Scan(...any) error }) (*model.UserSession, error) {
	var value model.UserSession
	err := scanner.Scan(&value.ID, &value.SystemID, &value.Username, &value.State, &value.LoginAt, &value.LogoutAt, &value.LastActivityAt, &value.DurationSeconds, &value.CreatedAt)
	return &value, err
}

func (r *UserSessionRepository) Create(ctx context.Context, value model.UserSession) (*model.UserSession, error) {
	row := r.db.QueryRowContext(ctx, `INSERT INTO user_sessions (system_id, username, state, login_at, last_activity_at) VALUES ($1, $2, $3, $4, $4) RETURNING `+userSessionColumns, value.SystemID, value.Username, value.State, value.LoginAt)
	created, err := scanUserSession(row)
	if err != nil {
		return nil, fmt.Errorf("create user session: %w", err)
	}
	return created, nil
}

func (r *UserSessionRepository) UpdateEvent(ctx context.Context, id, state string, at time.Time) (*model.UserSession, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE user_sessions
		SET state = $1, last_activity_at = $2,
		    logout_at = CASE WHEN $1 = 'logged_out' THEN $2 ELSE logout_at END,
		    duration_seconds = CASE WHEN $1 = 'logged_out' THEN EXTRACT(EPOCH FROM ($2 - login_at))::BIGINT ELSE duration_seconds END
		WHERE id = $3 RETURNING `+userSessionColumns, state, at, id)
	value, err := scanUserSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update user session: %w", err)
	}
	return value, nil
}

func (r *UserSessionRepository) ListBySystem(ctx context.Context, systemID string) ([]model.UserSession, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+userSessionColumns+` FROM user_sessions WHERE system_id = $1 ORDER BY login_at DESC`, systemID)
	if err != nil {
		return nil, fmt.Errorf("query user sessions: %w", err)
	}
	defer rows.Close()
	var values []model.UserSession
	for rows.Next() {
		value, err := scanUserSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user session: %w", err)
		}
		values = append(values, *value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user sessions: %w", err)
	}
	return values, nil
}
