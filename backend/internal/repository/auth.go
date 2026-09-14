package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"CampusWatch/backend/internal/model"
)

// UserRepository persists and loads user accounts for authentication.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a PostgreSQL-backed user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail loads an active user's account by email address.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, institution_id, email, password_hash, first_name, last_name,
		       role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
		ORDER BY created_at ASC
		LIMIT 1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.InstitutionID, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}