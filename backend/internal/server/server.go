package server

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"CampusWatch/backend/internal/handler"
	"CampusWatch/backend/internal/health"
	"CampusWatch/backend/internal/middleware"
	"CampusWatch/backend/internal/repository"
	"CampusWatch/backend/internal/service"
)

func NewRouter(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health.Handler)

	if db != nil {
		userRepo := repository.NewUserRepository(db)
		sessionRepo := repository.NewSessionRepository(db)
		authService := service.NewAuthService(userRepo, sessionRepo)
		mux.HandleFunc("/api/v1/auth/login", handler.AuthHandler(authService))
		// Registration is deliberately outside the authenticated block below:
		// it exists for the state where no account exists, so there is nobody to
		// authenticate. Its own first-run gate in the service is what limits it,
		// and the rate limiter in NewHttpServer blunts repeated attempts.
		mux.HandleFunc("/api/v1/auth/register", handler.RegisterHandler(authService))
		authMiddleware := middleware.NewMiddleware(sessionRepo)
		auditRepo := repository.NewAuditRepository(db)
		admin := func(next http.HandlerFunc) http.Handler {
			protected := middleware.RequireRole("admin", "manager")(authMiddleware.RequireAuth(next))
			return middleware.Audit(auditRepo, protected)
		}
		mux.Handle("/api/v1/auth/me", authMiddleware.RequireAuth(handler.AuthMeHandler))
		mux.Handle("/api/v1/auth/logout", authMiddleware.RequireAuth(handler.AuthLogoutHandler(authService)))

		institutionRepo := repository.NewInstitutionRepository(db)
		institutionService := service.NewInstitutionService(institutionRepo)
		mux.Handle("/api/v1/institutions", admin(handler.InstitutionsHandler(institutionService)))
		mux.Handle("/api/v1/institutions/", admin(handler.InstitutionByIDHandler(institutionService)))

		campusRepo := repository.NewCampusRepository(db)
		campusService := service.NewCampusService(campusRepo)
		mux.Handle("/api/v1/campuses", admin(handler.CampusesHandler(campusService)))
		mux.Handle("/api/v1/campuses/", admin(handler.CampusByIDHandler(campusService)))

		clusterRepo := repository.NewClusterRepository(db)
		clusterService := service.NewClusterService(clusterRepo)
		mux.Handle("/api/v1/clusters", admin(handler.ClustersHandler(clusterService)))
		mux.Handle("/api/v1/clusters/", admin(handler.ClusterByIDHandler(clusterService)))

		locationRepo := repository.NewLocationRepository(db)
		locationService := service.NewLocationService(locationRepo)
		mux.Handle("/api/v1/locations", admin(handler.LocationsHandler(locationService)))
		mux.Handle("/api/v1/locations/", admin(handler.LocationByIDHandler(locationService)))

		systemRepo := repository.NewSystemRepository(db)
		systemService := service.NewSystemService(systemRepo)
		mux.Handle("/api/v1/systems", admin(handler.SystemsHandler(systemService)))

		agentRepo := repository.NewAgentRepository(db)
		agentService := service.NewAgentService(agentRepo)
		sessionRepoForUsers := repository.NewUserSessionRepository(db)
		userSessionService := service.NewUserSessionService(agentRepo, sessionRepoForUsers)
		eventRepo := repository.NewEventRepository(db)
		eventService := service.NewEventService(agentRepo, eventRepo)
		mux.Handle("/api/v1/systems/", admin(handler.SystemByIDHandler(systemService, userSessionService, eventService, agentService)))
		// Registering an agent mints a credential and listing them exposes the
		// whole fleet's agent codes and last heartbeats, so both are operator
		// actions and are guarded as such. `/api/v1/agents/` is a catch-all for the
		// reporting routes below, which authenticate as agents rather than as
		// operators, so it is registered separately from the administration pair.
		mux.Handle("/api/v1/agents", admin(handler.AgentsHandler(agentService)))
		// Approval is an operator action and is guarded as one. It is what flips a
		// machine from pending to reporting, and it issues the credential the agent
		// then uses, so leaving it open would make the pending state decorative:
		// anyone able to reach the port could register an agent and immediately
		// approve it, then report heartbeats, logins and events as that system —
		// which is the data every question this product answers is drawn from.
		mux.Handle("/api/v1/agents/", admin(handler.AgentByIDHandler(agentService)))
		heartbeatService := service.NewHeartbeatService(agentRepo, systemRepo)
		mux.HandleFunc("/api/v1/agents/heartbeat", handler.HeartbeatHandler(heartbeatService))
		mux.HandleFunc("/api/v1/agents/session", handler.UserSessionEventHandler(userSessionService))
		mux.HandleFunc("/api/v1/agents/event", handler.EventHandler(eventService))

		issueRepo := repository.NewIssueRepository(db)
		issueService := service.NewIssueService(issueRepo)
		mux.Handle("/api/v1/issues", admin(handler.IssuesHandler(issueService)))
		mux.Handle("/api/v1/issues/", admin(handler.IssueByIDHandler(issueService)))

		dashboardRepo := repository.NewDashboardRepository(db)
		dashboardService := service.NewDashboardService(dashboardRepo)
		mux.Handle("/api/v1/dashboard/summary", admin(handler.DashboardSummaryHandler(dashboardService)))
		mux.Handle("/api/v1/dashboard/systems", admin(handler.DashboardSystemsHandler(dashboardService)))
		mux.Handle("/api/v1/dashboard/issues", admin(handler.IssuesHandler(issueService)))

		reportRepo := repository.NewReportRepository(db)
		reportService := service.NewReportService(reportRepo)
		mux.Handle("/api/v1/reports/", admin(handler.ReportsHandler(reportService)))

		scheduleRepo := repository.NewScheduleRepository(db)
		scheduleService := service.NewScheduleService(scheduleRepo)
		mux.Handle("/api/v1/schedules/check", admin(handler.ScheduleCheckHandler(scheduleService)))
		mux.Handle("/api/v1/schedules", admin(handler.SchedulesHandler(scheduleService)))
	}

	// A path under /api/ that no route claims must not reach the static handler.
	// The file server would answer it with "404 page not found" after probing the
	// dashboard directory for a file of that name, which is indistinguishable
	// from a route that exists and rejected the request. That ambiguity is worth
	// removing: the usual reasons to land here are a frontend calling a route
	// this build does not have, and a server started before that route was added,
	// and both are obvious once the response names the path.
	//
	// ServeMux prefers the longer patterns registered above, so this only sees
	// paths nothing else wanted.
	unclaimedAPI := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no such API route: "+r.URL.Path, http.StatusNotFound)
	}
	mux.HandleFunc("/api", unclaimedAPI)
	mux.HandleFunc("/api/", unclaimedAPI)

	// The dashboard is served from the same origin as the API, which is the
	// arrangement the frontend is written for: it defaults its API base to
	// same-origin (frontend/js/config.js) because the API sends no CORS headers,
	// so a dashboard on another port could not read its responses anyway.
	// ServeMux prefers the longer patterns registered above, so "/" never
	// shadows an API route.
	mux.Handle("/", staticHandler(resolveFrontendDir()))

	return mux
}

// resolveFrontendDir locates the static dashboard directory.
//
// This mirrors database.ResolveMigrationsDir: an explicit override wins, then the
// candidate paths are tried in turn, because the directory exists at a different
// relative path depending on whether the process was started from the repository
// root or from inside backend/.
//
// Unlike migrations, a missing dashboard is not fatal, so this returns "" rather
// than an error. Two cases depend on that: the tests run with the package
// directory as the working directory and have no dashboard beside them, and a
// deployment may legitimately serve the dashboard from nginx and never want the
// binary to serve it at all. Callers get a 404 explaining the override instead.
func resolveFrontendDir() string {
	if dir := os.Getenv("FRONTEND_DIR"); dir != "" {
		return dir
	}

	for _, candidate := range []string{"frontend", filepath.Join("..", "frontend")} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return ""
}

// dashboardPath is where the root of the site sends a visitor.
//
// The dashboard is a page in the pages/ directory rather than an index.html at
// the frontend root, because every page there links to its siblings by bare
// filename (systems.html, system.html). Serving that page's bytes at "/" would
// resolve those links against the root and 404, so "/" redirects instead of
// rewriting.
const dashboardPath = "/pages/dashboard.html"

// staticHandler serves the dashboard from dir.
//
// An empty dir yields a 404 that names the override, so a misconfigured
// deployment says what to fix rather than looking like a missing page.
func staticHandler(dir string) http.Handler {
	if dir == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "dashboard directory not found; set FRONTEND_DIR to the frontend directory", http.StatusNotFound)
		})
	}

	files := http.FileServer(http.Dir(dir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Opening the site lands on the dashboard, which in turn sends an
		// unauthenticated visitor to the sign-in page. A 302 rather than a 301:
		// where the root points is a property of this frontend's layout, and a
		// permanently cached redirect would outlive any change to it.
		if r.URL.Path == "/" {
			http.Redirect(w, r, dashboardPath, http.StatusFound)
			return
		}

		// Refuse directory listings. http.FileServer answers a directory that has
		// no index.html with a generated listing of its contents, and the frontend
		// has no index.html anywhere — so /js/, /pages/ and the rest would each
		// hand out an inventory of the tree to anyone who asked. Nothing in the
		// dashboard links to a directory, so a 404 is the honest answer.
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		files.ServeHTTP(w, r)
	})
}

// isAPIPath reports whether a request belongs to the API rather than the dashboard.
func isAPIPath(path string) bool {
	return path == "/health" || path == "/api" || strings.HasPrefix(path, "/api/")
}

func NewHttpServer(Port string, db *sql.DB) *http.Server {
	router := NewRouter(db)
	limited := middleware.NewRateLimiter(120, time.Minute).Handler(router)

	// The limiter covers the API only. A dashboard page load fetches several
	// static files and then polls, so metering those against the same 120 per
	// minute budget as the API would let the dashboard rate-limit its own
	// stylesheet. Static requests take the un-metered path through the same
	// router, which serves them from the "/" handler.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAPIPath(r.URL.Path) {
			limited.ServeHTTP(w, r)
			return
		}
		router.ServeHTTP(w, r)
	})

	secured := middleware.SecurityHeaders(handler)
	return &http.Server{
		Addr:         ":" + Port,
		Handler:      secured,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
