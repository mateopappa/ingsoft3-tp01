package dashboard

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"flow/internal/auth"
	"flow/internal/models"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	// Client timezone offset in minutes (optional)
	clientTime := time.Now().UTC()
	if offsetStr := r.URL.Query().Get("tz_offset"); offsetStr != "" {
		if offsetMins, err := strconv.Atoi(offsetStr); err == nil {
			// JS getTimezoneOffset() returns minutes west of UTC (positive for Americas, negative for Asia/Europe)
			clientTime = time.Now().UTC().Add(-time.Duration(offsetMins) * time.Minute)
		}
	}

	metrics, err := h.service.GetMetrics(r.Context(), user.ID, clientTime)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to load dashboard metrics",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(metrics)
}
