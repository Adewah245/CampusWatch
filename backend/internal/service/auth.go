package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/auth/session"
	"CampusWatch/backend/internal/model"
)

// ErrInvalidCredentials is returned when a login attempt fails because the user or password is invalid.
var ErrInvalidCredentials = errors.New("invalid credentials")

// UserRepository provides access to user records during authentication.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

// SessionRepository stores and loads authenticated sessions.
type SessionRepository interface {
	Save(ctx context.Context, session model.Session) error
	FindByToken(ctx context.Context, token string) (*model.Session, error)
	Delete(ctx context.Context, token string) error
}

// Logout invalidates an authentication session by its token.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("session token is required")
	}

	if err := s.sessionRepo.Delete(ctx, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

// AuthService manages authentication and session creation for users.
type AuthService struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
	sessionTTL  time.Duration
}

// NewAuthService creates a configured authentication service.
func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		sessionTTL:  24 * time.Hour,
	}
}

// Login authenticates a user and creates a new session token.
func (s *AuthService) Login(ctx context.Context, email, passwordValue string) (*model.Session, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if passwordValue == "" {
		return nil, errors.New("password is required")
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, ErrInvalidCredentials
	}
	if err := password.VerifyPassword(user.PasswordHash, passwordValue); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := session.GenerateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}

	createdSession := model.Session{
		Token:         token,
		UserID:        user.ID,
		InstitutionID: user.InstitutionID,
		Role:          user.Role,
		ExpiresAt:     time.Now().Add(s.sessionTTL),
		CreatedAt:     time.Now(),
	}
	if err := s.sessionRepo.Save(ctx, createdSession); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	return &createdSession, nil
}
