package model

import "time"

// Campus represents a campus inside an institution.
type Campus struct {
	ID            string    `json:"id"`
	InstitutionID string    `json:"institution_id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateCampusRequest contains the input required to create a campus.
type CreateCampusRequest struct {
	InstitutionID string `json:"institution_id"`
	Name          string `json:"name"`
	Slug          string `json:"slug,omitempty"`
	Status        string `json:"status,omitempty"`
}

// UpdateCampusRequest contains optional values for updating a campus record.
type UpdateCampusRequest struct {
	InstitutionID *string `json:"institution_id,omitempty"`
	Name          *string `json:"name,omitempty"`
	Slug          *string `json:"slug,omitempty"`
	Status        *string `json:"status,omitempty"`
}
