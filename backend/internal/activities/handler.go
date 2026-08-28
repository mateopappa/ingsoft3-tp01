package activities

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"flow/internal/auth"
	"flow/internal/models"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	query := r.URL.Query()
	filter := models.ActivityFilter{
		Search:     query.Get("search"),
		CategoryID: query.Get("category_id"),
		FromDate:   query.Get("from_date"),
		ToDate:     query.Get("to_date"),
	}

	if favStr := query.Get("favorite"); favStr != "" {
		fav, err := strconv.ParseBool(favStr)
		if err == nil {
			filter.Favorite = &fav
		}
	}

	list, err := h.service.ListActivities(r.Context(), user.ID, filter)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to list activities", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"activities": list,
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	var input models.ActivityInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_REQUEST")
		return
	}

	activity, err := h.service.CreateActivity(r.Context(), user.ID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidDuration) || errors.Is(err, ErrEmptyDescription) ||
			errors.Is(err, ErrInvalidDate) || errors.Is(err, ErrInvalidCategory) {
			jsonError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to create activity", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]any{
		"activity": activity,
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, http.StatusBadRequest, "Activity ID is required", "INVALID_ID")
		return
	}

	activity, err := h.service.GetActivityByID(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			jsonError(w, http.StatusNotFound, "Activity not found", "NOT_FOUND")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to get activity", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"activity": activity,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, http.StatusBadRequest, "Activity ID is required", "INVALID_ID")
		return
	}

	var input models.ActivityInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_REQUEST")
		return
	}

	activity, err := h.service.UpdateActivity(r.Context(), id, user.ID, input)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			jsonError(w, http.StatusNotFound, "Activity not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrInvalidDuration) || errors.Is(err, ErrEmptyDescription) ||
			errors.Is(err, ErrInvalidDate) || errors.Is(err, ErrInvalidCategory) {
			jsonError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to update activity", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"activity": activity,
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, http.StatusBadRequest, "Activity ID is required", "INVALID_ID")
		return
	}

	err := h.service.DeleteActivity(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			jsonError(w, http.StatusNotFound, "Activity not found", "NOT_FOUND")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to delete activity", "INTERNAL_ERROR")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ToggleFavorite(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, http.StatusBadRequest, "Activity ID is required", "INVALID_ID")
		return
	}

	var body struct {
		Favorite bool `json:"favorite"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_REQUEST")
		return
	}

	err := h.service.ToggleFavorite(r.Context(), id, user.ID, body.Favorite)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			jsonError(w, http.StatusNotFound, "Activity not found", "NOT_FOUND")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to toggle favorite", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"id":       id,
		"favorite": body.Favorite,
	})
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
