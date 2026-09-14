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
		mux.Handle("/api/v1/auth/me", authMiddleware.RequireAuth(handler.AuthMeHandler))
		mux.Handle("/api/v1/auth/logout", authMiddleware.RequireAuth(handler.AuthLogoutHandler(authService)))

		institutionRepo := repository.NewInstitutionRepository(db)
		institutionService := service.NewInstitutionService(institutionRepo)
		mux.HandleFunc("/api/v1/institutions", handler.InstitutionsHandler(institutionService))
		mux.HandleFunc("/api/v1/institutions/", handler.InstitutionByIDHandler(institutionService))

		campusRepo := repository.NewCampusRepository(db)
		campusService := service.NewCampusService(campusRepo)
		mux.HandleFunc("/api/v1/campuses", handler.CampusesHandler(campusService))
		mux.HandleFunc("/api/v1/campuses/", handler.CampusByIDHandler(campusService))

		clusterRepo := repository.NewClusterRepository(db)
		clusterService := service.NewClusterService(clusterRepo)
		mux.HandleFunc("/api/v1/clusters", handler.ClustersHandler(clusterService))
		mux.HandleFunc("/api/v1/clusters/", handler.ClusterByIDHandler(clusterService))

		locationRepo := repository.NewLocationRepository(db)
		locationService := service.NewLocationService(locationRepo)
		mux.HandleFunc("/api/v1/locations", handler.LocationsHandler(locationService))
		mux.HandleFunc("/api/v1/locations/", handler.LocationByIDHandler(locationService))

		systemRepo := repository.NewSystemRepository(db)
		systemService := service.NewSystemService(systemRepo)
		mux.HandleFunc("/api/v1/systems", handler.SystemsHandler(systemService))

		agentRepo := repository.NewAgentRepository(db)
		agentService := service.NewAgentService(agentRepo)
		sessionRepoForUsers := repository.NewUserSessionRepository(db)
		userSessionService := service.NewUserSessionService(agentRepo, sessionRepoForUsers)
		eventRepo := repository.NewEventRepository(db)
		eventService := service.NewEventService(agentRepo, eventRepo)
		mux.HandleFunc("/api/v1/systems/", handler.SystemByIDHandler(systemService, userSessionService, eventService))
		mux.HandleFunc("/api/v1/agents", handler.AgentsHandler(agentService))
		mux.HandleFunc("/api/v1/agents/", handler.AgentByIDHandler(agentService))
		heartbeatService := service.NewHeartbeatService(agentRepo, systemRepo)
		mux.HandleFunc("/api/v1/agents/heartbeat", handler.HeartbeatHandler(heartbeatService))
		mux.HandleFunc("/api/v1/agents/session", handler.UserSessionEventHandler(userSessionService))
		mux.HandleFunc("/api/v1/agents/event", handler.EventHandler(eventService))
	}

	return mux
}

func NewHttpServer(Port string, db *sql.DB) *http.Server {
	return &http.Server{
		Addr:         ":" + Port,
		Handler:      NewRouter(db),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
