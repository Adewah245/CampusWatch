package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/service"
)

// InstitutionsHandler handles creation and listing of institutions.
func InstitutionsHandler(svc *service.InstitutionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			institutions, err := svc.List(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(institutions)
			return
		case http.MethodPost:
			var request model.CreateInstitutionRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			institution, err := svc.Create(r.Context(), request.Name, request.Slug, request.Status)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(institution)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// InstitutionByIDHandler handles fetching and updating a specific institution.
func InstitutionByIDHandler(svc *service.InstitutionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathID := strings.TrimPrefix(r.URL.Path, "/api/v1/institutions/")
		if pathID == "" || pathID == r.URL.Path {
			http.Error(w, "institution id is required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			institution, err := svc.GetByID(r.Context(), pathID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if institution == nil {
				http.Error(w, "institution not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(institution)
			return
		case http.MethodPatch:
			var request model.UpdateInstitutionRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			institution, err := svc.Update(r.Context(), pathID, request)
			if err != nil {
				if strings.Contains(err.Error(), "institution id is required") || strings.Contains(err.Error(), "cannot be empty") || strings.Contains(err.Error(), "invalid institution status") {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if institution == nil {
				http.Error(w, "institution not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(institution)
			return
		case http.MethodDelete:
			deleted, err := svc.Delete(r.Context(), pathID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !deleted {
				http.Error(w, "institution not found", http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
