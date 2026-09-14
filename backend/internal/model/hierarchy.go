package model

import "time"

// Cluster represents a group of systems inside a campus.
type Cluster struct {
	ID        string    `json:"id"`
	CampusID  string    `json:"campus_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateClusterRequest struct {
	CampusID string `json:"campus_id"`
	Name     string `json:"name"`
	Slug     string `json:"slug,omitempty"`
	Status   string `json:"status,omitempty"`
}

type UpdateClusterRequest struct {
	CampusID *string `json:"campus_id,omitempty"`
	Name     *string `json:"name,omitempty"`
	Slug     *string `json:"slug,omitempty"`
	Status   *string `json:"status,omitempty"`
}

// Location represents a physical space inside a cluster.
type Location struct {
	ID           string    `json:"id"`
	ClusterID    string    `json:"cluster_id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	LocationType string    `json:"location_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateLocationRequest struct {
	ClusterID    string `json:"cluster_id"`
	Name         string `json:"name"`
	Slug         string `json:"slug,omitempty"`
	LocationType string `json:"location_type,omitempty"`
	Status       string `json:"status,omitempty"`
}

type UpdateLocationRequest struct {
	ClusterID    *string `json:"cluster_id,omitempty"`
	Name         *string `json:"name,omitempty"`
	Slug         *string `json:"slug,omitempty"`
	LocationType *string `json:"location_type,omitempty"`
	Status       *string `json:"status,omitempty"`
}
