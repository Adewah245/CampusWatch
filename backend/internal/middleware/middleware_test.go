package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"CampusWatch/backend/internal/model"
)

type sessionRepoStub struct {
	session *model.Session
}

func (r *sessionRepoStub) Save(context.Context, model.Session) error { return nil }

func (r *sessionRepoStub) FindByToken(context.Context, string) (*model.Session, error) {
	return r.session, nil
}

func (r *sessionRepoStub) Delete(context.Context, string) error { return nil }

func TestRequireAuthRejectsMissingHeader(t *testing.T) {
	handler := NewMiddleware(&sessionRepoStub{}).RequireAuth(func(http.ResponseWriter, *http.Request) {
		t.Fatal("expected request to be rejected")
	})
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	response := httptest.NewRecorder()

	handler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestRequireAuthStoresValidSession(t *testing.T) {
	expected := &model.Session{Token: "token-1", ExpiresAt: time.Now().Add(time.Hour)}
	var received *model.Session
	handler := NewMiddleware(&sessionRepoStub{session: expected}).RequireAuth(func(_ http.ResponseWriter, request *http.Request) {
		received, _ = SessionFromContext(request.Context())
	})
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer token-1")
	response := httptest.NewRecorder()

	handler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if received != expected {
		t.Fatal("expected authenticated session in request context")
	}
}
