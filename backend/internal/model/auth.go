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

// RegisterRequest carries the fields needed to create the first administrator.
//
// There is no Role field: the account this creates is an administrator by
// definition, and accepting a role from an unauthenticated caller would let the
// first caller choose their own privileges beyond what the bootstrap needs.
type RegisterRequest struct {
	Institution string `json:"institution"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

// RegisterResponse reports the account that was created.
//
// It deliberately carries no session token. Registration is followed by a
// sign-in, so the password is proven to work before the account is trusted with
// a session — and so that a response body is never the only thing standing
// between a caller and an authenticated session.
type RegisterResponse struct {
	UserID        string `json:"user_id"`
	InstitutionID string `json:"institution_id"`
	Role          string `json:"role"`
}
