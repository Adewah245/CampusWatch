package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/service"
)

// LoginHandler handles user login requests and returns a session token on success.
func LoginHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req model.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		session, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				http.Error(w, "invalid email or password", http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := model.LoginResponse{
			Token:         session.Token,
			UserID:        session.UserID,
			InstitutionID: session.InstitutionID,
			Role:          session.Role,
			ExpiresAt:     session.ExpiresAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
