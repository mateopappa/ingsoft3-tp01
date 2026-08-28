package http

import (
	"encoding/json"
	"io/fs"
	"net/http"

	"flow/internal/activities"
	"flow/internal/auth"
	"flow/internal/categories"
	"flow/internal/dashboard"
	"flow/internal/http/middleware"
	"flow/internal/timer"
)

type RouterConfig struct {
	AuthHandler       *auth.Handler
	AuthService       *auth.Service
	CategoryHandler   *categories.Handler
	ActivityHandler   *activities.Handler
	TimerHandler      *timer.Handler
	DashboardHandler  *dashboard.Handler
	WebFS             fs.FS
}

// NewRouter sets up the application routes, middleware, and static asset handlers.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// 1. Healthcheck
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"database": "connected",
		})
	})

	// 2. Public Auth Routes
	mux.HandleFunc("POST /api/auth/register", cfg.AuthHandler.Register)
	mux.HandleFunc("POST /api/auth/login", cfg.AuthHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", cfg.AuthHandler.Logout)

	// 3. Protected Routes
	protect := func(h http.HandlerFunc) http.Handler {
		return middleware.RequireAuth(h)
	}

	// Auth me
	mux.Handle("GET /api/auth/me", protect(cfg.AuthHandler.Me))

	// Categories
	mux.Handle("GET /api/categories", protect(cfg.CategoryHandler.List))
	mux.Handle("POST /api/categories", protect(cfg.CategoryHandler.Create))
	mux.Handle("PUT /api/categories/{id}", protect(cfg.CategoryHandler.Update))
	mux.Handle("DELETE /api/categories/{id}", protect(cfg.CategoryHandler.Delete))

	// Activities
	mux.Handle("GET /api/activities", protect(cfg.ActivityHandler.List))
	mux.Handle("POST /api/activities", protect(cfg.ActivityHandler.Create))
	mux.Handle("GET /api/activities/{id}", protect(cfg.ActivityHandler.GetByID))
	mux.Handle("PUT /api/activities/{id}", protect(cfg.ActivityHandler.Update))
	mux.Handle("DELETE /api/activities/{id}", protect(cfg.ActivityHandler.Delete))
	mux.Handle("PATCH /api/activities/{id}/favorite", protect(cfg.ActivityHandler.ToggleFavorite))

	// Timer
	mux.Handle("GET /api/timer", protect(cfg.TimerHandler.Get))
	mux.Handle("POST /api/timer/start", protect(cfg.TimerHandler.Start))
	mux.Handle("POST /api/timer/stop", protect(cfg.TimerHandler.Stop))
	mux.Handle("DELETE /api/timer/discard", protect(cfg.TimerHandler.Discard))

	// Dashboard
	mux.Handle("GET /api/dashboard", protect(cfg.DashboardHandler.Get))

	// 4. Static & Web Frontend Handler
	if cfg.WebFS != nil {
		staticFS, err := fs.Sub(cfg.WebFS, "static")
		if err == nil {
			mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
		}

		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			indexBytes, err := fs.ReadFile(cfg.WebFS, "templates/index.html")
			if err != nil {
				http.Error(w, "Failed to load application interface", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(indexBytes)
		})
	}

	// Apply global middleware chain: Recovery -> Security -> Logging -> Auth Context
	var handler http.Handler = mux
	handler = middleware.Authenticate(cfg.AuthService)(handler)
	handler = middleware.SecurityHeaders(handler)
	handler = middleware.Logger(handler)
	handler = middleware.Recoverer(handler)

	return handler
}
