package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"CampusWatch/backend/internal/model"
)

// SessionRepository persists authentication sessions.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a PostgreSQL-backed session repository.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Save stores a new authentication session.
func (r *SessionRepository) Save(ctx context.Context, value model.Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (token, user_id, institution_id, role, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, value.Token, value.UserID, value.InstitutionID, value.Role, value.ExpiresAt, value.CreatedAt)
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	return nil
}

// FindByToken loads an unexpired session.
func (r *SessionRepository) FindByToken(ctx context.Context, token string) (*model.Session, error) {
	var value model.Session
	err := r.db.QueryRowContext(ctx, `
		SELECT token, user_id, institution_id, role, expires_at, created_at
		FROM sessions
		WHERE token = $1 AND expires_at > NOW()
	`, token).Scan(
		&value.Token, &value.UserID, &value.InstitutionID, &value.Role,
		&value.ExpiresAt, &value.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}

	return &value, nil
}

// Delete removes an authentication session.
func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = $1`, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}