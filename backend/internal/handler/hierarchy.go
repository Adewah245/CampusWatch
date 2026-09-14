package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/service"
)

func ClustersHandler(svc *service.ClusterService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			clusters, err := svc.List(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(clusters)
		case http.MethodPost:
			var request model.CreateClusterRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			cluster, err := svc.Create(r.Context(), request.CampusID, request.Name, request.Slug, request.Status)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(cluster)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func ClusterByIDHandler(svc *service.ClusterService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/clusters/")
		if id == "" || id == r.URL.Path {
			http.Error(w, "cluster id is required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			cluster, err := svc.GetByID(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if cluster == nil {
				http.Error(w, "cluster not found", http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(cluster)
		case http.MethodPatch:
			var request model.UpdateClusterRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			cluster, err := svc.Update(r.Context(), id, request)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if cluster == nil {
				http.Error(w, "cluster not found", http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(cluster)
		case http.MethodDelete:
			deleted, err := svc.Delete(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !deleted {
				http.Error(w, "cluster not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func LocationsHandler(svc *service.LocationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			locations, err := svc.List(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(locations)
		case http.MethodPost:
			var request model.CreateLocationRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			location, err := svc.Create(r.Context(), request.ClusterID, request.Name, request.Slug, request.LocationType, request.Status)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(location)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func LocationByIDHandler(svc *service.LocationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/locations/")
		if id == "" || id == r.URL.Path {
			http.Error(w, "location id is required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			location, err := svc.GetByID(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if location == nil {
				http.Error(w, "location not found", http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(location)
		case http.MethodPatch:
			var request model.UpdateLocationRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			location, err := svc.Update(r.Context(), id, request)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if location == nil {
				http.Error(w, "location not found", http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(location)
		case http.MethodDelete:
			deleted, err := svc.Delete(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !deleted {
				http.Error(w, "location not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
