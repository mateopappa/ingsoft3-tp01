package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"

	"flow/internal/activities"
	"flow/internal/auth"
	"flow/internal/categories"
	"flow/internal/config"
	"flow/internal/dashboard"
	"flow/internal/database"
	internalhttp "flow/internal/http"
	"flow/internal/timer"
	"flow/migrations"
	"flow/web"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "Run database migrations and exit")
	flag.Parse()

	// 1. Setup Structured Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting Flow Activity Tracker...")

	// 2. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Configuration failure", "error", err)
		os.Exit(1)
	}

	// 3. Connect to Database
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Database connection failure", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Connected to PostgreSQL database")

	// 4. Run Migrations
	if cfg.AutoMigrate || *migrateOnly {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := database.RunMigrations(ctx, db, migrations.FS); err != nil {
			slog.Error("Migration execution failure", "error", err)
			os.Exit(1)
		}
		slog.Info("Database migrations verified and up to date")

		if *migrateOnly {
			slog.Info("Migrations complete, exiting")
			os.Exit(0)
		}
	}

	// 5. Initialize Repositories & Services
	categoriesRepo := categories.NewRepository(db)
	categoriesService := categories.NewService(categoriesRepo)
	categoriesHandler := categories.NewHandler(categoriesService)

	sessionDuration := time.Duration(cfg.SessionDurationSeconds) * time.Second
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, categoriesService, sessionDuration)
	authHandler := auth.NewHandler(authService, cfg.CookieSecure)

	activitiesRepo := activities.NewRepository(db)
	activitiesService := activities.NewService(activitiesRepo, categoriesService)
	activitiesHandler := activities.NewHandler(activitiesService)

	timerRepo := timer.NewRepository(db)
	timerService := timer.NewService(timerRepo, categoriesService)
	timerHandler := timer.NewHandler(timerService)

	dashboardRepo := dashboard.NewRepository(db)
	dashboardService := dashboard.NewService(dashboardRepo)
	dashboardHandler := dashboard.NewHandler(dashboardService)

	// 6. Sentry (optional — empty SENTRY_DSN leaves the SDK idle)
	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			Environment:      cfg.AppEnv,
			TracesSampleRate: 1.0,
		}); err != nil {
			slog.Error("Sentry init failed", "error", err)
			os.Exit(1)
		}
		defer sentry.Flush(2 * time.Second)
		slog.Info("Sentry enabled", "environment", cfg.AppEnv)
	}

	// 7. Setup HTTP Router & Middleware
	router := internalhttp.NewRouter(internalhttp.RouterConfig{
		AuthHandler:      authHandler,
		AuthService:      authService,
		CategoryHandler:  categoriesHandler,
		ActivityHandler:  activitiesHandler,
		TimerHandler:     timerHandler,
		DashboardHandler: dashboardHandler,
		DB:               db,
		WebFS:            web.FS,
		EnableDemoFault:  cfg.EnableDemoFault,
	})

	var handler http.Handler = router
	if cfg.SentryDSN != "" {
		handler = sentryhttp.New(sentryhttp.Options{Repanic: true}).Handle(router)
	}

	// 8. Setup HTTP Server
	serverAddr := fmt.Sprintf(":%d", cfg.AppPort)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Start Server
	serverErrChan := make(chan error, 1)
	go func() {
		slog.Info("Flow server listening", "port", cfg.AppPort, "url", fmt.Sprintf("http://localhost:%d", cfg.AppPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- err
		}
	}()

	// 10. Graceful Shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		slog.Info("Shutdown signal received", "signal", sig.String())
	case err := <-serverErrChan:
		slog.Error("Server encountered unexpected runtime error", "error", err)
		os.Exit(1)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	} else {
		slog.Info("Server stopped gracefully")
	}
}
