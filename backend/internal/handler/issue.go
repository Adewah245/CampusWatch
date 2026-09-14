package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/service"
)

func IssuesHandler(svc *service.IssueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			issues, err := svc.List(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(issues)
		case http.MethodPost:
			var request model.CreateIssueRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			issue, err := svc.Create(r.Context(), request)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(issue)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func IssueByIDHandler(svc *service.IssueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/issues/"), "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "issue id is required", http.StatusBadRequest)
			return
		}
		id := parts[0]
		if len(parts) == 2 && parts[1] == "updates" {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var request model.CreateIssueUpdateRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			update, err := svc.AddUpdate(r.Context(), id, request)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(update)
			return
		}
		if len(parts) != 1 {
			http.Error(w, "invalid issue path", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			issue, err := svc.GetByID(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if issue == nil {
				http.Error(w, "issue not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(issue)
		case http.MethodPatch:
			var request model.UpdateIssueRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			issue, err := svc.Update(r.Context(), id, request)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if issue == nil {
				http.Error(w, "issue not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(issue)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
