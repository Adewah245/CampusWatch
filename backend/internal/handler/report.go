package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"CampusWatch/backend/internal/service"
)

func ReportsHandler(svc *service.ReportService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		kind := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/reports/"), "/")
		period := r.URL.Query().Get("period")
		if period == "" {
			period = "daily"
		}
		value, err := svc.Generate(r.Context(), kind, period, time.Now())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(value)
	}
}
