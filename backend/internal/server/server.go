package server

import (
	"database/sql"
	"net/http"
	"time"

	"CampusWatch/backend/internal/health"
	"CampusWatch/backend/internal/handler"
	"CampusWatch/backend/internal/repository"
	"CampusWatch/backend/internal/service"
)

func NewRouter(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health.Handler)

	if db != nil {
		institutionRepo := repository.NewInstitutionRepository(db)
		institutionService := service.NewInstitutionService(institutionRepo)
		mux.HandleFunc("/api/v1/institutions", handler.InstitutionsHandler(institutionService))
		mux.HandleFunc("/api/v1/institutions/", handler.InstitutionByIDHandler(institutionService))

		campusRepo := repository.NewCampusRepository(db)
		campusService := service.NewCampusService(campusRepo)
		mux.HandleFunc("/api/v1/campuses", handler.CampusesHandler(campusService))
		mux.HandleFunc("/api/v1/campuses/", handler.CampusByIDHandler(campusService))
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
