package model

import "time"

type AuditLog struct {
	ID            string    `json:"id"`
	UserID        *string   `json:"user_id,omitempty"`
	InstitutionID *string   `json:"institution_id,omitempty"`
	Method        string    `json:"method"`
	Path          string    `json:"path"`
	CreatedAt     time.Time `json:"created_at"`
}
