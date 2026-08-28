package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recoverer recovers from panics and responds with an HTTP 500 error.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("Panic recovered in HTTP handler",
					"error", rec,
					"stack", string(debug.Stack()),
					"path", r.URL.Path,
				)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"Internal Server Error","code":"INTERNAL_ERROR"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
