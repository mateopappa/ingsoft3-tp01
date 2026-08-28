package middleware

import (
	"context"
	"net/http"

	"flow/internal/auth"
)

// Authenticate checks for a valid session cookie and attaches the user to the request context.
func Authenticate(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err == nil && cookie.Value != "" {
				user, err := authService.ValidateSession(r.Context(), cookie.Value)
				if err == nil && user != nil {
					ctx := context.WithValue(r.Context(), auth.UserContextKey, user)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth enforces that a valid user is present in the request context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(auth.UserContextKey)
		if user == nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Authentication required","code":"UNAUTHORIZED"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
