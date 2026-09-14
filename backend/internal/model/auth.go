package model

import "time"

// User represents an authenticated application user for a single institution.
type User struct {
	ID            string    `json:"id"`
	InstitutionID string    `json:"institution_id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Role          string    `json:"role"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Session represents a valid authentication session for a user.
type Session struct {
	Token         string    `json:"token"`
	UserID        string    `json:"user_id"`
	InstitutionID string    `json:"institution_id"`
	Role          string    `json:"role"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// LoginRequest carries the email and password for a login attempt.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse contains the generated session metadata for a successful login.
type LoginResponse struct {
	Token         string    `json:"token"`
	UserID        string    `json:"user_id"`
	InstitutionID string    `json:"institution_id"`
	Role          string    `json:"role"`
	ExpiresAt     time.Time `json:"expires_at"`
}
