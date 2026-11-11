package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
)

// Recoverer recovers from panics and responds with an HTTP 500 error.
// Also reports the panic to Sentry when a DSN is configured.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
					hub.Recover(rec)
				} else {
					sentry.CurrentHub().Recover(rec)
				}
				sentry.Flush(2 * time.Second)

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
