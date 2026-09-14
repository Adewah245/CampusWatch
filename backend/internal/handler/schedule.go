package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"CampusWatch/backend/internal/model"
	"CampusWatch/backend/internal/service"
)

func SchedulesHandler(svc *service.ScheduleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			institutionID := r.URL.Query().Get("institution_id")
			schedules, err := svc.List(r.Context(), institutionID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(schedules)
		case http.MethodPost:
			var request model.CreateScheduleRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			schedule, err := svc.Create(r.Context(), request)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(schedule)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func ScheduleCheckHandler(svc *service.ScheduleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		institutionID := strings.TrimSpace(r.URL.Query().Get("institution_id"))
		if institutionID == "" {
			http.Error(w, "institution id is required", http.StatusBadRequest)
			return
		}
		at := time.Now()
		if raw := r.URL.Query().Get("timestamp"); raw != "" {
			parsed, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				http.Error(w, "timestamp must be RFC3339", http.StatusBadRequest)
				return
			}
			at = parsed
		}
		schedules, err := svc.List(r.Context(), institutionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, schedule := range schedules {
			if schedule.DayOfWeek == int(at.Weekday()) {
				result, err := service.CheckSchedule(schedule, at)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(result)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.ScheduleCheck{Allowed: false, State: "closed", DayOfWeek: int(at.Weekday())})
	}
}
