package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouterHealthEndpoint(t *testing.T) {
	mux := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	if body := res.Body.String(); body != "CampusWatch is healthy" {
		t.Fatalf("expected body %q, got %q", "CampusWatch is healthy", body)
	}
}

func TestNewHttpServer(t *testing.T) {
	srv := NewHttpServer("0")
	if srv == nil {
		t.Fatal("expected HTTP server instance")
	}
	if srv.Addr != ":0" {
		t.Fatalf("expected address :0, got %q", srv.Addr)
	}
	if srv.Handler == nil {
		t.Fatal("expected HTTP handler to be configured")
	}
}
