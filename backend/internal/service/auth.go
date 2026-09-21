package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/auth/session"
	"CampusWatch/backend/internal/model"
)

// ErrInvalidCredentials is returned when a login attempt fails because the user or password is invalid.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrRegistrationClosed is returned when first-run registration is attempted on
// a deployment that already has an account.
//
// It is a distinct error rather than a validation failure because it is the
// expected answer on every deployment after the first, and the frontend says so
// with a link to sign-in instead of showing it as a mistake.
var ErrRegistrationClosed = errors.New("registration is closed: this deployment already has an account")

// MinPasswordLength is the shortest password Register accepts.
const MinPasswordLength = 12

// maxPasswordBytes matches bcrypt's input limit.
//
// bcrypt ignores everything past 72 bytes, so without this limit two different
// long passwords would be interchangeable at sign-in. The limit is enforced
// here rather than inherited silently from the hash function.
const maxPasswordBytes = 72

// UserRepository provides access to user records during authentication.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	RegisterFirstAdmin(ctx context.Context, institutionName, institutionSlug string, user model.User) (*model.User, *model.Institution, bool, error)
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

// Register creates the deployment's first administrator.
//
// The account is always an administrator of the institution it also creates,
// because the only thing this path exists to unblock is a deployment with no
// accounts at all — the dashboard refuses every route without the admin or
// manager role, so any lesser role would leave the operator unable to proceed.
// It is not a general sign-up: the repository refuses as soon as any account
// exists, and that refusal surfaces as ErrRegistrationClosed.
//
// No session is created. The caller signs in afterwards, which proves the
// password was hashed and stored correctly before the account is relied on.
func (s *AuthService) Register(ctx context.Context, request model.RegisterRequest) (*model.User, *model.Institution, error) {
	institutionName := strings.TrimSpace(request.Institution)
	email := strings.TrimSpace(request.Email)
	firstName := strings.TrimSpace(request.FirstName)
	lastName := strings.TrimSpace(request.LastName)

	switch {
	case institutionName == "":
		return nil, nil, errors.New("institution is required")
	case email == "":
		return nil, nil, errors.New("email is required")
	case firstName == "":
		return nil, nil, errors.New("first name is required")
	case lastName == "":
		return nil, nil, errors.New("last name is required")
	}

	if err := validateNewPassword(request.Password); err != nil {
		return nil, nil, err
	}

	// The slug rule is the institution service's, so the name stored here
	// resolves through GetBySlug later just as one created anywhere else would.
	slug := slugify(institutionName)
	if slug == "" {
		return nil, nil, errors.New("institution name must contain a letter or digit")
	}

	hash, err := password.HashPassword(request.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	// The address is stored as typed. Sign-in matches the exact string
	// (repository.FindByEmail), so normalising the case here would reject the
	// very address the operator had just registered with. Case-folding belongs
	// on all three paths at once, not on this one alone.
	created, institution, registered, err := s.userRepo.RegisterFirstAdmin(
		ctx, institutionName, slug, model.User{
			Email:        email,
			PasswordHash: hash,
			FirstName:    firstName,
			LastName:     lastName,
			Role:         "admin",
			IsActive:     true,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	if !registered {
		return nil, nil, ErrRegistrationClosed
	}

	return created, institution, nil
}

// validateNewPassword enforces the rules for a password chosen through the
// registration form.
//
// createuser accepts any non-empty password, on the grounds that whoever runs
// it has chosen the value deliberately and is responsible for it. This path is
// an unauthenticated web form, so the same assumption does not hold and a
// minimum length is enforced here. The two rules differ as a result, which is
// deliberate but unresolved: whichever way it is settled, the decision belongs
// to both paths together.
//
// Length is measured in bytes, which is what bcrypt limits; the minimum is
// therefore a floor in ASCII characters and reached sooner by multi-byte text.
func validateNewPassword(value string) error {
	if value == "" {
		return errors.New("password is required")
	}
	if len(value) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	if len(value) > maxPasswordBytes {
		return fmt.Errorf("password must be at most %d bytes", maxPasswordBytes)
	}
	return nil
}
