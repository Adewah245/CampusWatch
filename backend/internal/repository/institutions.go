package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"CampusWatch/backend/internal/model"
)

// InstitutionRepository persists institution records in PostgreSQL.
type InstitutionRepository struct {
	db *sql.DB
}

// NewInstitutionRepository creates a PostgreSQL-backed institution repository.
func NewInstitutionRepository(db *sql.DB) *InstitutionRepository {
	return &InstitutionRepository{db: db}
}

// Create stores a new institution record and returns the inserted row.
func (r *InstitutionRepository) Create(ctx context.Context, institution model.Institution) (*model.Institution, error) {
	query := `
		INSERT INTO institutions (name, slug, status)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, status, created_at, updated_at
	`

	row := r.db.QueryRowContext(ctx, query, institution.Name, institution.Slug, institution.Status)
	var created model.Institution
	if err := row.Scan(
		&created.ID,
		&created.Name,
		&created.Slug,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("insert institution: %w", err)
	}

	return &created, nil
}

// List returns all institutions in the platform.
func (r *InstitutionRepository) List(ctx context.Context) ([]model.Institution, error) {
	query := `
		SELECT id, name, slug, status, created_at, updated_at
		FROM institutions
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query institutions: %w", err)
	}
	defer rows.Close()

	var institutions []model.Institution
	for rows.Next() {
		var institution model.Institution
		if err := rows.Scan(
			&institution.ID,
			&institution.Name,
			&institution.Slug,
			&institution.Status,
			&institution.CreatedAt,
			&institution.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan institution: %w", err)
		}
		institutions = append(institutions, institution)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate institutions: %w", err)
	}

	return institutions, nil
}

// GetByID loads one institution by primary key.
func (r *InstitutionRepository) GetByID(ctx context.Context, id string) (*model.Institution, error) {
	query := `
		SELECT id, name, slug, status, created_at, updated_at
		FROM institutions
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	var institution model.Institution
	if err := row.Scan(
		&institution.ID,
		&institution.Name,
		&institution.Slug,
		&institution.Status,
		&institution.CreatedAt,
		&institution.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select institution: %w", err)
	}

	return &institution, nil
}

// Update changes the allowed fields for an institution and returns the updated record.
func (r *InstitutionRepository) Update(ctx context.Context, id string, input model.UpdateInstitutionRequest) (*model.Institution, error) {
	query := `
		UPDATE institutions
		SET name = COALESCE($1, name),
		    slug = COALESCE($2, slug),
		    status = COALESCE($3, status),
		    updated_at = NOW()
		WHERE id = $4
		RETURNING id, name, slug, status, created_at, updated_at
	`

	var nameValue, slugValue, statusValue *string
	if input.Name != nil {
		nameValue = input.Name
	}
	if input.Slug != nil {
		slugValue = input.Slug
	}
	if input.Status != nil {
		statusValue = input.Status
	}

	row := r.db.QueryRowContext(ctx, query, nameValue, slugValue, statusValue, id)
	var institution model.Institution
	if err := row.Scan(
		&institution.ID,
		&institution.Name,
		&institution.Slug,
		&institution.Status,
		&institution.CreatedAt,
		&institution.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update institution: %w", err)
	}

	return &institution, nil
}

// Delete removes an institution from the database and returns whether any row was affected.
func (r *InstitutionRepository) Delete(ctx context.Context, id string) (bool, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM institutions WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete institution: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read deletion result: %w", err)
	}

	return affected > 0, nil
}
