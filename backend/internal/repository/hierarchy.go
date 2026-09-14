package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"CampusWatch/backend/internal/model"
)

type ClusterRepository struct{ db *sql.DB }

func NewClusterRepository(db *sql.DB) *ClusterRepository { return &ClusterRepository{db: db} }

func (r *ClusterRepository) Create(ctx context.Context, cluster model.Cluster) (*model.Cluster, error) {
	var created model.Cluster
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO clusters (campus_id, name, slug, status) VALUES ($1, $2, $3, $4)
		RETURNING id, campus_id, name, slug, status, created_at, updated_at
	`, cluster.CampusID, cluster.Name, cluster.Slug, cluster.Status).Scan(
		&created.ID, &created.CampusID, &created.Name, &created.Slug, &created.Status,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert cluster: %w", err)
	}
	return &created, nil
}

func (r *ClusterRepository) List(ctx context.Context) ([]model.Cluster, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, campus_id, name, slug, status, created_at, updated_at FROM clusters ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query clusters: %w", err)
	}
	defer rows.Close()
	var clusters []model.Cluster
	for rows.Next() {
		var cluster model.Cluster
		if err := rows.Scan(&cluster.ID, &cluster.CampusID, &cluster.Name, &cluster.Slug, &cluster.Status, &cluster.CreatedAt, &cluster.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan cluster: %w", err)
		}
		clusters = append(clusters, cluster)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate clusters: %w", err)
	}
	return clusters, nil
}

func (r *ClusterRepository) GetByID(ctx context.Context, id string) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.QueryRowContext(ctx, `SELECT id, campus_id, name, slug, status, created_at, updated_at FROM clusters WHERE id = $1`, id).Scan(
		&cluster.ID, &cluster.CampusID, &cluster.Name, &cluster.Slug, &cluster.Status, &cluster.CreatedAt, &cluster.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select cluster: %w", err)
	}
	return &cluster, nil
}

func (r *ClusterRepository) Update(ctx context.Context, id string, input model.UpdateClusterRequest) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.QueryRowContext(ctx, `
		UPDATE clusters SET campus_id = COALESCE($1, campus_id), name = COALESCE($2, name),
		 slug = COALESCE($3, slug), status = COALESCE($4, status), updated_at = NOW()
		WHERE id = $5 RETURNING id, campus_id, name, slug, status, created_at, updated_at
	`, input.CampusID, input.Name, input.Slug, input.Status, id).Scan(
		&cluster.ID, &cluster.CampusID, &cluster.Name, &cluster.Slug, &cluster.Status, &cluster.CreatedAt, &cluster.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update cluster: %w", err)
	}
	return &cluster, nil
}

func (r *ClusterRepository) Delete(ctx context.Context, id string) (bool, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM clusters WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete cluster: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read deletion result: %w", err)
	}
	return affected > 0, nil
}

type LocationRepository struct{ db *sql.DB }

func NewLocationRepository(db *sql.DB) *LocationRepository { return &LocationRepository{db: db} }

func (r *LocationRepository) Create(ctx context.Context, location model.Location) (*model.Location, error) {
	var created model.Location
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO locations (cluster_id, name, slug, location_type, status) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, cluster_id, name, slug, location_type, status, created_at, updated_at
	`, location.ClusterID, location.Name, location.Slug, location.LocationType, location.Status).Scan(
		&created.ID, &created.ClusterID, &created.Name, &created.Slug, &created.LocationType, &created.Status,
		&created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert location: %w", err)
	}
	return &created, nil
}

func (r *LocationRepository) List(ctx context.Context) ([]model.Location, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, cluster_id, name, slug, location_type, status, created_at, updated_at FROM locations ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query locations: %w", err)
	}
	defer rows.Close()
	var locations []model.Location
	for rows.Next() {
		var location model.Location
		if err := rows.Scan(&location.ID, &location.ClusterID, &location.Name, &location.Slug, &location.LocationType, &location.Status, &location.CreatedAt, &location.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		locations = append(locations, location)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locations: %w", err)
	}
	return locations, nil
}

func (r *LocationRepository) GetByID(ctx context.Context, id string) (*model.Location, error) {
	var location model.Location
	err := r.db.QueryRowContext(ctx, `SELECT id, cluster_id, name, slug, location_type, status, created_at, updated_at FROM locations WHERE id = $1`, id).Scan(
		&location.ID, &location.ClusterID, &location.Name, &location.Slug, &location.LocationType, &location.Status, &location.CreatedAt, &location.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select location: %w", err)
	}
	return &location, nil
}

func (r *LocationRepository) Update(ctx context.Context, id string, input model.UpdateLocationRequest) (*model.Location, error) {
	var location model.Location
	err := r.db.QueryRowContext(ctx, `
		UPDATE locations SET cluster_id = COALESCE($1, cluster_id), name = COALESCE($2, name),
		 slug = COALESCE($3, slug), location_type = COALESCE($4, location_type), status = COALESCE($5, status), updated_at = NOW()
		WHERE id = $6 RETURNING id, cluster_id, name, slug, location_type, status, created_at, updated_at
	`, input.ClusterID, input.Name, input.Slug, input.LocationType, input.Status, id).Scan(
		&location.ID, &location.ClusterID, &location.Name, &location.Slug, &location.LocationType, &location.Status, &location.CreatedAt, &location.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update location: %w", err)
	}
	return &location, nil
}

func (r *LocationRepository) Delete(ctx context.Context, id string) (bool, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM locations WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete location: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read deletion result: %w", err)
	}
	return affected > 0, nil
}
