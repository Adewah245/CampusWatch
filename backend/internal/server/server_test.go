package server

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRouterHealthEndpoint(t *testing.T) {
	mux := NewRouter(nil)
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
	srv := NewHttpServer("0", &sql.DB{})
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

// TestNewRouterUnclaimedAPIPath pins what an unknown API path answers with.
//
// With a nil database the router registers no API routes at all, so this is also
// the shape of a real failure: a frontend calling a route the running build does
// not have — a build from before that route was added, most of the time. The
// response must name the path, because "404 page not found" is what a static
// file server says and gives no hint that a route was ever expected.
func TestNewRouterUnclaimedAPIPath(t *testing.T) {
	mux := NewRouter(nil)

	for _, path := range []string{"/api", "/api/", "/api/v1/auth/register", "/api/v1/nope"} {
		res := get(mux, path)
		if res.Code != http.StatusNotFound {
			t.Errorf("GET %s: expected status %d, got %d", path, http.StatusNotFound, res.Code)
			continue
		}
		if body := res.Body.String(); !strings.Contains(body, path) {
			t.Errorf("GET %s: expected the response to name the path, got %q", path, body)
		}
	}
}

// TestNewRouterServesTheDashboard covers the arrangement the frontend is written
// for: one origin serving both the API and the static dashboard.
//
// The path list matters more than it looks. `/` is the catch-all pattern, so a
// regression here would not fail loudly — it would silently answer API requests
// with an HTML page, and the dashboard would report that the API base URL was
// wrong. Checking a route that only exists in the router (`/health`) and one that
// only exists on disk keeps the two apart.
func TestNewRouterServesTheDashboard(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pages", "dashboard.html"), "<html>dashboard</html>")
	writeFile(t, filepath.Join(dir, "pages", "systems.html"), "<html>systems</html>")
	writeFile(t, filepath.Join(dir, "css", "main.css"), "body{margin:0}")
	t.Setenv("FRONTEND_DIR", dir)

	mux := NewRouter(nil)

	cases := []struct {
		path     string
		status   int
		contains string
	}{
		{"/pages/dashboard.html", http.StatusOK, "dashboard"},
		{"/pages/systems.html", http.StatusOK, "systems"},
		{"/css/main.css", http.StatusOK, "margin"},
		{"/missing.html", http.StatusNotFound, ""},
		// Served by the router, not from disk. If "/" ever shadowed the API
		// routes this would come back 404 with an HTML body instead.
		{"/health", http.StatusOK, "CampusWatch is healthy"},
	}

	for _, testCase := range cases {
		res := get(mux, testCase.path)
		if res.Code != testCase.status {
			t.Errorf("GET %s: expected status %d, got %d", testCase.path, testCase.status, res.Code)
			continue
		}
		if testCase.contains != "" && !strings.Contains(res.Body.String(), testCase.contains) {
			t.Errorf("GET %s: expected body to contain %q, got %q", testCase.path, testCase.contains, res.Body.String())
		}
	}
}

// TestNewRouterRootRedirectsToTheDashboard pins where opening the site lands.
//
// A redirect rather than serving dashboard.html's contents at "/", because the
// dashboard's own links are bare sibling filenames: rewritten to the root, every
// one of them would resolve to /systems.html and 404.
func TestNewRouterRootRedirectsToTheDashboard(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pages", "dashboard.html"), "<html>dashboard</html>")
	t.Setenv("FRONTEND_DIR", dir)

	res := get(NewRouter(nil), "/")

	if res.Code != http.StatusFound {
		t.Fatalf("GET /: expected status %d, got %d", http.StatusFound, res.Code)
	}
	if got := res.Header().Get("Location"); got != dashboardPath {
		t.Fatalf("GET /: expected Location %q, got %q", dashboardPath, got)
	}
}

// TestNewRouterRefusesDirectoryListings pins what a directory path answers with.
//
// http.FileServer generates a listing for any directory that has no index.html,
// and this frontend has no index.html anywhere — the root is a redirect and the
// rest are named pages. Left alone, /js/ and /pages/ would each answer with an
// inventory of the tree. Nothing links to a directory, so 404 is the whole
// intent.
func TestNewRouterRefusesDirectoryListings(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pages", "dashboard.html"), "<html>dashboard</html>")
	writeFile(t, filepath.Join(dir, "js", "config.js"), "export const API_BASE = '';")
	t.Setenv("FRONTEND_DIR", dir)

	mux := NewRouter(nil)

	for _, path := range []string{"/js/", "/pages/", "/css/"} {
		res := get(mux, path)
		if res.Code != http.StatusNotFound {
			t.Errorf("GET %s: expected status %d, got %d", path, http.StatusNotFound, res.Code)
			continue
		}
		// A listing would name the files inside; the 404 must not.
		if body := res.Body.String(); strings.Contains(body, "config.js") || strings.Contains(body, "dashboard.html") {
			t.Errorf("GET %s: expected no listing, got %q", path, body)
		}
	}
}

// TestNewHttpServerWithoutADashboard pins the deliberate degradation: a missing
// dashboard directory must not stop the server from starting.
//
// Two callers depend on it — these tests run from the package directory, which
// has no frontend/ beside it, and a deployment behind nginx may never want the
// binary to serve static files at all. The 404 names the override so a
// misconfiguration reads as a misconfiguration rather than a missing page.
func TestNewHttpServerWithoutADashboard(t *testing.T) {
	if dir := resolveFrontendDir(); dir != "" {
		t.Skipf("a dashboard directory is resolvable at %q, so the missing-dashboard path cannot be exercised", dir)
	}

	res := get(NewHttpServer("0", nil).Handler, "/")
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.Code)
	}
	if body := res.Body.String(); !strings.Contains(body, "FRONTEND_DIR") {
		t.Fatalf("expected the 404 to name FRONTEND_DIR, got %q", body)
	}
}

// TestNewHttpServerRateLimitsOnlyTheAPI pins the split between the metered API
// and the unmetered dashboard.
//
// Metering both would be a self-inflicted outage: one dashboard page load
// fetches several files and then polls, so the dashboard would consume the same
// budget as the API and eventually rate-limit its own stylesheet.
func TestNewHttpServerRateLimitsOnlyTheAPI(t *testing.T) {
	handler := NewHttpServer("0", nil).Handler

	// Comfortably more than the 120 request limit: none may be refused, whatever
	// the static handler answers with.
	for i := 1; i <= 300; i++ {
		if code := get(handler, "/").Code; code == http.StatusTooManyRequests {
			t.Fatalf("static request %d was rate limited; the limiter must cover the API only", i)
		}
	}

	// The API is still metered. Requests 1 to 120 fit in the window; the 121st
	// does not, because the limiter blocks once the count exceeds the limit.
	for i := 1; i <= 120; i++ {
		if code := get(handler, "/api/v1/auth/me").Code; code == http.StatusTooManyRequests {
			t.Fatalf("API request %d was rate limited before the limit was reached", i)
		}
	}
	if code := get(handler, "/api/v1/auth/me").Code; code != http.StatusTooManyRequests {
		t.Fatalf("expected the 121st API request to be rate limited, got status %d", code)
	}
}

// TestIsAPIPath guards the dispatch the two tests above rely on.
//
// The case that earns its keep is "/apiary.html": a page whose name merely
// begins with "api" must not be metered with the API, which a naive prefix
// check on "/api" would do.
func TestIsAPIPath(t *testing.T) {
	cases := map[string]bool{
		"/health":               true,
		"/api":                  true,
		"/api/":                 true,
		"/api/v1/auth/me":       true,
		"/api/v1/agents/event":  true,
		"/":                     false,
		"/pages/login.html":     false,
		"/pages/dashboard.html": false,
		"/apiary.html":          false,
	}

	for path, want := range cases {
		if got := isAPIPath(path); got != want {
			t.Errorf("isAPIPath(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestNewRouterAgentApprovalRequiresAnOperator pins the auth on agent approval.
//
// Approving an agent is what moves a machine from pending to reporting, and it
// hands back the credential the agent then uses. Unauthenticated, the pending
// state would mean nothing: register an agent, approve it yourself, and the
// backend would accept fabricated heartbeats, logins and events as that system.
//
// A zero *sql.DB is enough to tell the two apart. RequireAuth rejects a request
// with no Authorization header before it reaches the repository, so a guarded
// route answers 401 and never touches the database — the 400 an unguarded one
// gives for a malformed path would be indistinguishable from the handler's own
// validation, which is exactly the confusion this test exists to prevent.
func TestNewRouterAgentApprovalRequiresAnOperator(t *testing.T) {
	res := post(NewRouter(&sql.DB{}), "/api/v1/agents/agent-1/approve")

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("POST approve without a token: expected status %d, got %d (%q)",
			http.StatusUnauthorized, res.Code, strings.TrimSpace(res.Body.String()))
	}
}

// get runs one request through handler and returns the recorded response.
func get(handler http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

// post runs one bodyless POST through handler and returns the recorded response.
func post(handler http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

// writeFile writes a test fixture, creating the parent directory as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
