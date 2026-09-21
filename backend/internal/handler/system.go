package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/service"
)

func SystemsHandler(svc *service.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			systems, err := svc.List(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(systems)
		case http.MethodPost:
			var request model.CreateSystemRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			system, err := svc.Create(r.Context(), request)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(system)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func SystemByIDHandler(svc *service.SystemService, extras ...any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/systems/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "system id is required", http.StatusBadRequest)
			return
		}
		id := parts[0]
		var sessionService *service.UserSessionService
		var eventService *service.EventService
		var agentService *service.AgentService
		for _, extra := range extras {
			switch value := extra.(type) {
			case *service.UserSessionService:
				sessionService = value
			case *service.EventService:
				eventService = value
			case *service.AgentService:
				agentService = value
			}
		}
		if len(parts) == 2 && parts[1] == "agents" && agentService != nil {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			agents, err := agentService.ListBySystem(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(agents)
			return
		}
		if len(parts) == 2 && parts[1] == "sessions" && sessionService != nil {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			sessions, err := sessionService.ListBySystem(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sessions)
			return
		}
		if len(parts) == 2 && parts[1] == "events" && eventService != nil {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			events, err := eventService.ListBySystem(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(events)
			return
		}
		if len(parts) == 2 && parts[1] == "health" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			report, err := svc.LatestHealth(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if report == nil {
				http.Error(w, "system health not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(report)
			return
		}
		if len(parts) == 2 && parts[1] == "approve" {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			system, err := svc.Approve(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if system == nil {
				http.Error(w, "system not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(system)
			return
		}
		if len(parts) != 1 {
			http.Error(w, "invalid system path", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			system, err := svc.GetByID(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if system == nil {
				http.Error(w, "system not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(system)
		case http.MethodPatch:
			var request model.UpdateSystemRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			system, err := svc.Update(r.Context(), id, request)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if system == nil {
				http.Error(w, "system not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(system)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
