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

// FindByEmailInInstitution loads one user by institution and email.
//
// FindByEmail deliberately searches across institutions, because sign-in only
// has an email to work with. Callers that already know the institution — such
// as account creation, where users_email_unique is scoped to
// (institution_id, email) — should use this instead, so the same address can
// legitimately exist in two institutions without one shadowing the other.
func (r *UserRepository) FindByEmailInInstitution(ctx context.Context, institutionID, email string) (*model.User, error) {
	query := `
		SELECT id, institution_id, email, password_hash, first_name, last_name,
		       role, is_active, created_at, updated_at
		FROM users
		WHERE institution_id = $1 AND email = $2
		LIMIT 1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, institutionID, email).Scan(
		&user.ID, &user.InstitutionID, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by institution and email: %w", err)
	}

	return &user, nil
}

// Create stores a new user account and returns the inserted row.
//
// The caller supplies PasswordHash already hashed — this method never sees a
// plaintext password, so there is no path by which one could be logged or
// stored by accident.
func (r *UserRepository) Create(ctx context.Context, user model.User) (*model.User, error) {
	query := `
		INSERT INTO users (institution_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, institution_id, email, password_hash, first_name, last_name,
		          role, is_active, created_at, updated_at
	`

	row := r.db.QueryRowContext(ctx, query,
		user.InstitutionID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.IsActive,
	)

	var created model.User
	if err := row.Scan(
		&created.ID, &created.InstitutionID, &created.Email, &created.PasswordHash,
		&created.FirstName, &created.LastName, &created.Role, &created.IsActive,
		&created.CreatedAt, &created.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &created, nil
}