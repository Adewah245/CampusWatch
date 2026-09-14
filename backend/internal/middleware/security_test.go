package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Header().Get("X-Content-Type-Options") != "nosniff" || response.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("expected security headers, got %v", response.Header())
	}
}

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	first := httptest.NewRequest(http.MethodGet, "/", nil)
	first.RemoteAddr = "127.0.0.1:1000"
	second := httptest.NewRequest(http.MethodGet, "/", nil)
	second.RemoteAddr = "127.0.0.1:1000"
	firstResponse := httptest.NewRecorder()
	secondResponse := httptest.NewRecorder()
	handler.ServeHTTP(firstResponse, first)
	handler.ServeHTTP(secondResponse, second)
	if firstResponse.Code != http.StatusNoContent || secondResponse.Code != http.StatusTooManyRequests {
		t.Fatalf("unexpected statuses: %d and %d", firstResponse.Code, secondResponse.Code)
	}
}
