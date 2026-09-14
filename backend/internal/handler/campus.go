package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/service"
)

// CampusesHandler handles creation and listing of campuses.
func CampusesHandler(svc *service.CampusService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			campuses, err := svc.List(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(campuses)
			return
		case http.MethodPost:
			var request model.CreateCampusRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			campus, err := svc.Create(r.Context(), request.InstitutionID, request.Name, request.Slug, request.Status)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(campus)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// CampusByIDHandler handles fetching, updating, and deleting a specific campus.
func CampusByIDHandler(svc *service.CampusService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathID := strings.TrimPrefix(r.URL.Path, "/api/v1/campuses/")
		if pathID == "" || pathID == r.URL.Path {
			http.Error(w, "campus id is required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			campus, err := svc.GetByID(r.Context(), pathID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if campus == nil {
				http.Error(w, "campus not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(campus)
			return
		case http.MethodPatch:
			var request model.UpdateCampusRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			campus, err := svc.Update(r.Context(), pathID, request)
			if err != nil {
				if strings.Contains(err.Error(), "campus id is required") || strings.Contains(err.Error(), "cannot be empty") || strings.Contains(err.Error(), "invalid campus status") {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if campus == nil {
				http.Error(w, "campus not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(campus)
			return
		case http.MethodDelete:
			deleted, err := svc.Delete(r.Context(), pathID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !deleted {
				http.Error(w, "campus not found", http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
