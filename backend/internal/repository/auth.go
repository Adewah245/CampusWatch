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

// bootstrapLockKey is the advisory lock key that serialises first-run
// registration.
//
// The value is arbitrary. It only has to be the same for every bootstrap, so
// that concurrent attempts contend instead of racing. It is spelled out here so
// that a later advisory lock elsewhere in the codebase does not reuse it by
// accident.
const bootstrapLockKey int64 = 0x0c4d5057 // "CMPW"

// RegisterFirstAdmin creates the deployment's first institution and its
// administrator account in one transaction.
//
// This is the only unauthenticated write path in the platform, so the gate that
// makes it a *first-run* bootstrap is the entire security control on it, and it
// has to hold when two requests arrive at once. A plain "count, then insert"
// would let both callers pass the count before either commits, leaving the
// deployment with two administrators. The advisory lock closes that window: it
// is taken before the count, so the second caller waits, then sees the first
// caller's committed row and is refused. The lock is transaction-scoped, so
// commit or rollback releases it and there is no key to leak.
//
// It touches institutions and users in one transaction because the concept it
// implements — a first administrator — does not exist without both, and a
// half-applied bootstrap would leave an institution nobody can sign in to.
//
// user.InstitutionID is ignored: this creates the institution the account
// belongs to. created is false when an account already exists, which the caller
// reports as "registration is closed" rather than as an error, because that is
// the expected outcome on every deployment after the first.
func (r *UserRepository) RegisterFirstAdmin(
	ctx context.Context,
	institutionName, institutionSlug string,
	user model.User,
) (*model.User, *model.Institution, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, false, fmt.Errorf("begin bootstrap transaction: %w", err)
	}
	// Rollback after a successful commit reports ErrTxDone, which is expected
	// and therefore ignored rather than checked.
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, bootstrapLockKey); err != nil {
		return nil, nil, false, fmt.Errorf("lock bootstrap: %w", err)
	}

	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM users`).Scan(&existing); err != nil {
		return nil, nil, false, fmt.Errorf("count users: %w", err)
	}
	if existing > 0 {
		return nil, nil, false, nil
	}

	var institution model.Institution
	if err := tx.QueryRowContext(ctx, `
			INSERT INTO institutions (name, slug, status)
			VALUES ($1, $2, 'active')
			RETURNING id, name, slug, status, created_at, updated_at
		`, institutionName, institutionSlug).Scan(
		&institution.ID,
		&institution.Name,
		&institution.Slug,
		&institution.Status,
		&institution.CreatedAt,
		&institution.UpdatedAt,
	); err != nil {
		return nil, nil, false, fmt.Errorf("insert institution: %w", err)
	}

	var created model.User
	if err := tx.QueryRowContext(ctx, `
			INSERT INTO users (institution_id, email, password_hash, first_name, last_name, role, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, institution_id, email, password_hash, first_name, last_name,
			          role, is_active, created_at, updated_at
		`,
		institution.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.IsActive,
	).Scan(
		&created.ID, &created.InstitutionID, &created.Email, &created.PasswordHash,
		&created.FirstName, &created.LastName, &created.Role, &created.IsActive,
		&created.CreatedAt, &created.UpdatedAt,
	); err != nil {
		return nil, nil, false, fmt.Errorf("insert first administrator: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, false, fmt.Errorf("commit bootstrap: %w", err)
	}

	return &created, &institution, true, nil
}