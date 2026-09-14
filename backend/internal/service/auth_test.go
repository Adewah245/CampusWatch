package service

import (
	"context"
	"testing"

	"CampusWatch/backend/internal/auth/password"
	"CampusWatch/backend/internal/model"
)

type authUserRepoStub struct {
	user *model.User
}

func (r *authUserRepoStub) FindByEmail(context.Context, string) (*model.User, error) {
	return r.user, nil
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