package categories

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	categories, err := h.service.ListCategories(r.Context(), user.ID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to list categories", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"categories": categories,
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)

	var input models.CategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_REQUEST")
		return
	}

	category, err := h.service.CreateCategory(r.Context(), user.ID, input.Name, input.Color)
	if err != nil {
		if errors.Is(err, ErrDuplicateCategory) {
			jsonError(w, http.StatusConflict, err.Error(), "CATEGORY_EXISTS")
			return
		}
		if errors.Is(err, ErrInvalidCategoryName) {
			jsonError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to create category", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]any{
		"category": category,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, http.StatusBadRequest, "Category ID is required", "INVALID_ID")
		return
	}

	var input models.CategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_REQUEST")
		return
	}

	category, err := h.service.UpdateCategory(r.Context(), id, user.ID, input.Name, input.Color)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			jsonError(w, http.StatusNotFound, "Category not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrDuplicateCategory) {
			jsonError(w, http.StatusConflict, err.Error(), "CATEGORY_EXISTS")
			return
		}
		if errors.Is(err, ErrInvalidCategoryName) {
			jsonError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to update category", "INTERNAL_ERROR")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"category": category,
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*models.User)
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, http.StatusBadRequest, "Category ID is required", "INVALID_ID")
		return
	}

	err := h.service.DeleteCategory(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			jsonError(w, http.StatusNotFound, "Category not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrCategoryInUse) || (err != nil && errors.Unwrap(err) == ErrCategoryInUse) {
			jsonError(w, http.StatusConflict, err.Error(), "CATEGORY_IN_USE")
			return
		}
		jsonError(w, http.StatusInternalServerError, "Failed to delete category", "INTERNAL_ERROR")
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
