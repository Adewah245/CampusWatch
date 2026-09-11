package server

import (
	"CampusWatch/backend/internal/health"
	"net/http"
	"time"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health.Handler)
	return mux
}

func NewHttpServer(Port string) *http.Server {
	return &http.Server{
		Addr:         ":" + Port,
		Handler:      NewRouter(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
