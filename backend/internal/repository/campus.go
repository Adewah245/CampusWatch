package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"CampusWatch/backend/internal/model"
)

// CampusRepository persists campus records in PostgreSQL.
type CampusRepository struct {
	db *sql.DB
}

// NewCampusRepository creates a PostgreSQL-backed campus repository.
func NewCampusRepository(db *sql.DB) *CampusRepository {
	return &CampusRepository{db: db}
}

// Create stores a new campus record and returns the inserted row.
func (r *CampusRepository) Create(ctx context.Context, campus model.Campus) (*model.Campus, error) {
	query := `
		INSERT INTO campuses (institution_id, name, slug, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, institution_id, name, slug, status, created_at, updated_at
	`

	row := r.db.QueryRowContext(ctx, query, campus.InstitutionID, campus.Name, campus.Slug, campus.Status)
	var created model.Campus
	if err := row.Scan(
		&created.ID,
		&created.InstitutionID,
		&created.Name,
		&created.Slug,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("insert campus: %w", err)
	}

	return &created, nil
}

// List returns all campuses in the platform.
func (r *CampusRepository) List(ctx context.Context) ([]model.Campus, error) {
	query := `
		SELECT id, institution_id, name, slug, status, created_at, updated_at
		FROM campuses
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query campuses: %w", err)
	}
	defer rows.Close()

	var campuses []model.Campus
	for rows.Next() {
		var campus model.Campus
		if err := rows.Scan(
			&campus.ID,
			&campus.InstitutionID,
			&campus.Name,
			&campus.Slug,
			&campus.Status,
			&campus.CreatedAt,
			&campus.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan campus: %w", err)
		}
		campuses = append(campuses, campus)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate campuses: %w", err)
	}

	return campuses, nil
}

// GetByID loads one campus by primary key.
func (r *CampusRepository) GetByID(ctx context.Context, id string) (*model.Campus, error) {
	query := `
		SELECT id, institution_id, name, slug, status, created_at, updated_at
		FROM campuses
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	var campus model.Campus
	if err := row.Scan(
		&campus.ID,
		&campus.InstitutionID,
		&campus.Name,
		&campus.Slug,
		&campus.Status,
		&campus.CreatedAt,
		&campus.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select campus: %w", err)
	}

	return &campus, nil
}

// Update changes the allowed fields for a campus and returns the updated record.
func (r *CampusRepository) Update(ctx context.Context, id string, input model.UpdateCampusRequest) (*model.Campus, error) {
	query := `
		UPDATE campuses
		SET institution_id = COALESCE($1, institution_id),
		    name = COALESCE($2, name),
		    slug = COALESCE($3, slug),
		    status = COALESCE($4, status),
		    updated_at = NOW()
		WHERE id = $5
		RETURNING id, institution_id, name, slug, status, created_at, updated_at
	`

	var institutionIDValue *string
	if input.InstitutionID != nil {
		institutionIDValue = input.InstitutionID
	}
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

	row := r.db.QueryRowContext(ctx, query, institutionIDValue, nameValue, slugValue, statusValue, id)
	var campus model.Campus
	if err := row.Scan(
		&campus.ID,
		&campus.InstitutionID,
		&campus.Name,
		&campus.Slug,
		&campus.Status,
		&campus.CreatedAt,
		&campus.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update campus: %w", err)
	}

	return &campus, nil
}

// Delete removes a campus from the database and returns whether any row was affected.
func (r *CampusRepository) Delete(ctx context.Context, id string) (bool, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM campuses WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete campus: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read deletion result: %w", err)
	}

	return affected > 0, nil
}
