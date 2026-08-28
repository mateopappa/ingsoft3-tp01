package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"flow/internal/models"
)

type Handler struct {
	service      *Service
	cookieSecure bool
}

func NewHandler(service *Service, cookieSecure bool) *Handler {
	return &Handler{
		service:      service,
		cookieSecure: cookieSecure,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpError(w, http.StatusBadRequest, "Invalid JSON request body", "INVALID_REQUEST")
		return
	}

	user, sessionID, err := h.service.Register(r.Context(), input.Email, input.Password)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			httpError(w, http.StatusConflict, "Email address is already registered", "EMAIL_EXISTS")
			return
		}
		if errors.Is(err, ErrInvalidEmail) || errors.Is(err, ErrPasswordTooShort) {
			httpError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
			return
		}
		httpError(w, http.StatusInternalServerError, "Failed to register user", "INTERNAL_ERROR")
		return
	}

	h.setSessionCookie(w, sessionID)
	jsonResponse(w, http.StatusCreated, map[string]any{
		"user": user,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpError(w, http.StatusBadRequest, "Invalid JSON request body", "INVALID_REQUEST")
		return
	}

	user, sessionID, err := h.service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpError(w, http.StatusUnauthorized, "Invalid email or password", "INVALID_CREDENTIALS")
			return
		}
		httpError(w, http.StatusInternalServerError, "Failed to authenticate", "INTERNAL_ERROR")
		return
	}

	h.setSessionCookie(w, sessionID)
	jsonResponse(w, http.StatusOK, map[string]any{
		"user": user,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		_ = h.service.Logout(r.Context(), cookie.Value)
	}

	h.clearSessionCookie(w)
	jsonResponse(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	if !ok || user == nil {
		httpError(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"user": user,
	})
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.service.sessionDuration.Seconds()),
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func httpError(w http.ResponseWriter, status int, message, code string) {
	jsonResponse(w, status, map[string]string{
		"error": message,
		"code":  code,
	})
}
