package middleware

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/service"
)

type contextKey string

const sessionContextKey contextKey = "session"

type AuditRepository interface {
	Record(context.Context, model.AuditLog) error
}

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

// RequireRole permits only sessions with one of the supplied roles.
func RequireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[strings.ToLower(role)] = struct{}{}
	}
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			session, err := SessionFromContext(r.Context())
			if err != nil {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}
			if _, ok := allowed[strings.ToLower(session.Role)]; !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next(w, r)
		}
	}
}

// RequireInstitution restricts a request to the institution in the session.
func RequireInstitution(institutionID func(*http.Request) string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			session, err := SessionFromContext(r.Context())
			if err != nil {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}
			if institutionID(r) != session.InstitutionID {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next(w, r)
		}
	}
}

// SecurityHeaders adds baseline browser protections to every response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func Audit(repo AuditRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if session, err := SessionFromContext(r.Context()); err == nil {
			userID, institutionID := session.UserID, session.InstitutionID
			_ = repo.Record(r.Context(), model.AuditLog{UserID: &userID, InstitutionID: &institutionID, Method: r.Method, Path: r.URL.Path})
		}
		next.ServeHTTP(w, r)
	})
}

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]rateLimitEntry
}

type rateLimitEntry struct {
	started time.Time
	count   int
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{limit: limit, window: window, clients: make(map[string]rateLimitEntry)}
}

// Handler limits requests per remote address within a fixed time window.
func (l *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		address, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			address = r.RemoteAddr
		}
		now := time.Now()
		l.mu.Lock()
		entry := l.clients[address]
		if entry.started.IsZero() || now.Sub(entry.started) >= l.window {
			entry = rateLimitEntry{started: now}
		}
		entry.count++
		l.clients[address] = entry
		blocked := entry.count > l.limit
		l.mu.Unlock()
		if blocked {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
