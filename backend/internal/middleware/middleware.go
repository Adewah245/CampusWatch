package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/service"
)

type contextKey string

const sessionContextKey contextKey = "session"

// Middleware validates bearer tokens and stores the session in the request context.
type Middleware struct {
	sessionRepo service.SessionRepository
}

// NewMiddleware creates a middleware instance for validating bearer tokens.
func NewMiddleware(sessionRepo service.SessionRepository) *Middleware {
	return &Middleware{sessionRepo: sessionRepo}
}

// RequireAuth ensures that a request contains a valid bearer token before continuing.
func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			http.Error(w, "authorization header is required", http.StatusUnauthorized)
			return
		}

		scheme, token, ok := strings.Cut(authorization, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		session, err := m.sessionRepo.FindByToken(r.Context(), token)
		if err != nil || session == nil {
			http.Error(w, "invalid or expired session", http.StatusUnauthorized)
			return
		}
		if session.ExpiresAt.Before(time.Now()) {
			http.Error(w, "session expired", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), sessionContextKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// SessionFromContext returns the authenticated session stored on the request.
func SessionFromContext(ctx context.Context) (*model.Session, error) {
	session, ok := ctx.Value(sessionContextKey).(*model.Session)
	if !ok || session == nil {
		return nil, errors.New("session not found in context")
	}
	return session, nil
}
