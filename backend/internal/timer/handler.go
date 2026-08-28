package timer

import (
	"encoding/json"
	"errors"
	"net/http"

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

	t, err := h.service.GetActiveTimer(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, ErrNoActiveTimer) {
			jsonResponse(w, http.StatusOK, map[string]any{
				"active_timer": nil,
			})
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to get active timer", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"active_timer": t,
	})
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	var input models.StartTimerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_REQUEST")
		return
	}

	t, err := h.service.StartTimer(r.Context(), user.ID, input.Description, input.CategoryID)
	if err != nil {
		if errors.Is(err, ErrTimerAlreadyActive) || (err != nil && errors.Unwrap(err) == ErrTimerAlreadyActive) {
			jsonError(w, http.StatusConflict, err.Error(), "TIMER_ALREADY_ACTIVE")
			return
		}
		if errors.Is(err, ErrEmptyTimerDescription) || errors.Is(err, ErrInvalidTimerCategory) {
			jsonError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to start timer", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]any{
		"active_timer": t,
	})
}

func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	var input models.StopTimerInput
	_ = json.NewDecoder(r.Body).Decode(&input) // optional body

	activity, err := h.service.StopTimer(r.Context(), user.ID, input.Note, input.Favorite)
	if err != nil {
		if errors.Is(err, ErrNoActiveTimer) {
			jsonError(w, http.StatusNotFound, "No active timer is currently running", "NO_ACTIVE_TIMER")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to stop timer", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"activity": activity,
	})
}

func (h *Handler) Discard(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	err := h.service.DiscardTimer(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, ErrNoActiveTimer) {
			jsonError(w, http.StatusNotFound, "No active timer is currently running", "NO_ACTIVE_TIMER")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to discard timer", "INTERNAL_ERROR")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, message, code string) {
	jsonResponse(w, status, map[string]string{
		"error": message,
		"code":  code,
	})
}
