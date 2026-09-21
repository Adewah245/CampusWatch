package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/model"
)

type authUserRepoStub struct {
	user *model.User

	// open reports whether the first-run gate lets a registration through.
	// Closed is the interesting default: it is what every deployment reaches
	// once an account exists.
	open bool

	// Recorded by RegisterFirstAdmin so a test can assert what the service asked
	// to be stored, rather than only what it returned.
	receivedName string
	receivedSlug string
	receivedUser model.User
}

func (r *authUserRepoStub) FindByEmail(context.Context, string) (*model.User, error) {
	return r.user, nil
}

func (r *authUserRepoStub) RegisterFirstAdmin(
	_ context.Context, institutionName, institutionSlug string, user model.User,
) (*model.User, *model.Institution, bool, error) {
	r.receivedName, r.receivedSlug, r.receivedUser = institutionName, institutionSlug, user
	if !r.open {
		return nil, nil, false, nil
	}

	created := user
	created.ID = "user-created"
	created.InstitutionID = "institution-created"
	return &created, &model.Institution{
		ID:   "institution-created",
		Name: institutionName,
		Slug: institutionSlug,
	}, true, nil
}

type authSessionRepoStub struct {
	session *model.Session
}

func (r *authSessionRepoStub) Save(_ context.Context, value model.Session) error {
	r.session = &value
	return nil
}

func (r *authSessionRepoStub) FindByToken(context.Context, string) (*model.Session, error) {
	return r.session, nil
}

func (r *authSessionRepoStub) Delete(_ context.Context, token string) error {
	if r.session != nil && r.session.Token == token {
		r.session = nil
	}
	return nil
}

func TestAuthServiceLoginCreatesSession(t *testing.T) {
	hash, err := password.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	sessionRepo := &authSessionRepoStub{}
	svc := NewAuthService(&authUserRepoStub{user: &model.User{
		ID: "user-1", InstitutionID: "institution-1", Email: "user@example.com",
		PasswordHash: hash, Role: "operator", IsActive: true,
	}}, sessionRepo)

	created, err := svc.Login(context.Background(), "user@example.com", "correct-password")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if created.Token == "" || sessionRepo.session == nil {
		t.Fatal("expected a generated and persisted session")
	}
	if created.UserID != "user-1" || created.Role != "operator" {
		t.Fatalf("unexpected session identity: %+v", created)
	}
}

func TestAuthServiceLoginRejectsInactiveUser(t *testing.T) {
	svc := NewAuthService(&authUserRepoStub{user: &model.User{
		Email: "inactive@example.com", IsActive: false,
	}}, &authSessionRepoStub{})

	if _, err := svc.Login(context.Background(), "inactive@example.com", "password"); err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestAuthServiceLogoutDeletesSession(t *testing.T) {
	sessionRepo := &authSessionRepoStub{session: &model.Session{Token: "token-1"}}
	svc := NewAuthService(&authUserRepoStub{}, sessionRepo)

	if err := svc.Logout(context.Background(), "token-1"); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if sessionRepo.session != nil {
		t.Fatal("expected session to be deleted")
	}
}

func TestAuthServiceRegisterCreatesAdministrator(t *testing.T) {
	userRepo := &authUserRepoStub{open: true}
	svc := NewAuthService(userRepo, &authSessionRepoStub{})

	created, institution, err := svc.Register(context.Background(), model.RegisterRequest{
		Institution: "  Adewah University  ",
		FirstName:   "Ada",
		LastName:    "Wah",
		Email:       "ada@adewah.edu",
		Password:    "a-long-enough-password",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if created.Role != "admin" {
		t.Fatalf("expected the first account to be an admin, got %q", created.Role)
	}
	if !created.IsActive {
		t.Fatal("expected the account to be active")
	}
	if institution == nil || institution.Slug != "adewah-university" {
		t.Fatalf("expected a slug derived from the name, got %+v", institution)
	}
	if userRepo.receivedName != "Adewah University" {
		t.Fatalf("expected the name to be trimmed, got %q", userRepo.receivedName)
	}

	// The plaintext password must never reach the repository, and the hash that
	// does has to verify — otherwise the account is created but unusable.
	if userRepo.receivedUser.PasswordHash == "a-long-enough-password" {
		t.Fatal("expected the password to be hashed before storage")
	}
	if err := password.VerifyPassword(userRepo.receivedUser.PasswordHash, "a-long-enough-password"); err != nil {
		t.Fatalf("stored hash does not verify: %v", err)
	}
}

func TestAuthServiceRegisterRefusedWhenClosed(t *testing.T) {
	// The default stub is closed, which is the state of every deployment that
	// already has an account.
	svc := NewAuthService(&authUserRepoStub{}, &authSessionRepoStub{})

	_, _, err := svc.Register(context.Background(), model.RegisterRequest{
		Institution: "Adewah University",
		FirstName:   "Ada",
		LastName:    "Wah",
		Email:       "ada@adewah.edu",
		Password:    "a-long-enough-password",
	})
	if !errors.Is(err, ErrRegistrationClosed) {
		t.Fatalf("expected ErrRegistrationClosed, got %v", err)
	}
}

func TestAuthServiceRegisterValidatesInput(t *testing.T) {
	valid := model.RegisterRequest{
		Institution: "Adewah University",
		FirstName:   "Ada",
		LastName:    "Wah",
		Email:       "ada@adewah.edu",
		Password:    "a-long-enough-password",
	}

	cases := map[string]func(request model.RegisterRequest) model.RegisterRequest{
		"no institution": func(r model.RegisterRequest) model.RegisterRequest { r.Institution = "  "; return r },
		"no email":       func(r model.RegisterRequest) model.RegisterRequest { r.Email = ""; return r },
		"no first name":  func(r model.RegisterRequest) model.RegisterRequest { r.FirstName = ""; return r },
		"no last name":   func(r model.RegisterRequest) model.RegisterRequest { r.LastName = ""; return r },
		"short password": func(r model.RegisterRequest) model.RegisterRequest { r.Password = "short"; return r },
		// bcrypt ignores anything past 72 bytes, so a longer password must be
		// refused rather than silently truncated.
		"overlong password": func(r model.RegisterRequest) model.RegisterRequest {
			r.Password = strings.Repeat("a", maxPasswordBytes+1)
			return r
		},
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			// Open, so a rejection can only come from validation: with the gate
			// closed every case would return ErrRegistrationClosed and pass
			// without proving anything.
			svc := NewAuthService(&authUserRepoStub{open: true}, &authSessionRepoStub{})

			if _, _, err := svc.Register(context.Background(), mutate(valid)); err == nil {
				t.Fatal("expected a validation error")
			} else if errors.Is(err, ErrRegistrationClosed) {
				t.Fatalf("expected a validation error, got the closed gate: %v", err)
			}
		})
	}
}

func TestAuthServiceRegisterRejectsNamelessInstitution(t *testing.T) {
	// "!!!" trims to nothing once slugified, so there would be no slug to store.
	svc := NewAuthService(&authUserRepoStub{open: true}, &authSessionRepoStub{})

	_, _, err := svc.Register(context.Background(), model.RegisterRequest{
		Institution: "!!!",
		FirstName:   "Ada",
		LastName:    "Wah",
		Email:       "ada@adewah.edu",
		Password:    "a-long-enough-password",
	})
	if err == nil || errors.Is(err, ErrRegistrationClosed) {
		t.Fatalf("expected a slug validation error, got %v", err)
	}
}