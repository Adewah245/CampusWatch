package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"CampusWatch/backend/internal/middleware"
	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/service"
)

// AuthHandler handles authentication endpoints.
func AuthHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/auth/login" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request model.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		createdSession, err := svc.Login(r.Context(), request.Email, request.Password)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := model.LoginResponse{
			Token: createdSession.Token, UserID: createdSession.UserID,
			InstitutionID: createdSession.InstitutionID, Role: createdSession.Role,
			ExpiresAt: createdSession.ExpiresAt,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// RegisterHandler creates the deployment's first administrator.
//
// This is the only unauthenticated write route in the platform. The authorization
// is not a role check — there is nobody to check yet — but the service's
// first-run gate: the request succeeds only while no account exists at all, and
// every later attempt is answered with 409. That makes a second call harmless
// rather than a way in.
//
// A successful response is 201 with no session token, so the frontend sends the
// operator to sign in and the new credentials are exercised before being trusted.
func RegisterHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/auth/register" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request model.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		created, _, err := svc.Register(r.Context(), request)
		if err != nil {
			if errors.Is(err, service.ErrRegistrationClosed) {
				// Not an error the caller caused, and not one they can fix by
				// changing the request: the deployment is already set up.
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := model.RegisterResponse{
			UserID:        created.ID,
			InstitutionID: created.InstitutionID,
			Role:          created.Role,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// AuthMeHandler returns the session associated with the authenticated request.
func AuthMeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentSession, err := middleware.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(currentSession)
}

// AuthLogoutHandler invalidates the session associated with the authenticated request.
func AuthLogoutHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		currentSession, err := middleware.SessionFromContext(r.Context())
		if err != nil {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		if err := svc.Logout(r.Context(), currentSession.Token); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}