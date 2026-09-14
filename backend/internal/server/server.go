package server

import (
	"database/sql"
	"net/http"
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
		mux.Handle("/api/v1/systems/", admin(handler.SystemByIDHandler(systemService, userSessionService, eventService)))
		mux.HandleFunc("/api/v1/agents", handler.AgentsHandler(agentService))
		mux.HandleFunc("/api/v1/agents/", handler.AgentByIDHandler(agentService))
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

	return mux
}

func NewHttpServer(Port string, db *sql.DB) *http.Server {
	router := NewRouter(db)
	secured := middleware.SecurityHeaders(middleware.NewRateLimiter(120, time.Minute).Handler(router))
	return &http.Server{
		Addr:         ":" + Port,
		Handler:      secured,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
