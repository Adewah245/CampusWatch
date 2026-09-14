package model

import "time"

// Institution represents a tenant institution in the CampusWatch platform.
type Institution struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateInstitutionRequest contains the data needed to create a new institution.
type CreateInstitutionRequest struct {
	Name   string `json:"name"`
	Slug   string `json:"slug,omitempty"`
	Status string `json:"status,omitempty"`
}

// UpdateInstitutionRequest contains optional fields for updating an institution record.
type UpdateInstitutionRequest struct {
	Name   *string `json:"name,omitempty"`
	Slug   *string `json:"slug,omitempty"`
	Status *string `json:"status,omitempty"`
}
