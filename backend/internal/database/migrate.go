package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RunMigrations executes all SQL migration files in the provided filesystem in sorted order.
// It uses a `schema_migrations` tracking table to ensure each migration is executed once.
func RunMigrations(ctx context.Context, db *sql.DB, migrationsFS fs.FS) error {
	// Create schema_migrations table if not exists
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Read all SQL files from migrationsFS
	var files []string
	err = fs.WalkDir(migrationsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	sort.Strings(files)

	for _, file := range files {
		version := filepath.Base(file)

		var exists bool
		err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", version, err)
		}

		if exists {
			continue // already applied
		}

		content, err := fs.ReadFile(migrationsFS, file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		slog.Info("Applying database migration", "version", version)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", version, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)", version, time.Now()); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", version, err)
		}

		slog.Info("Successfully applied migration", "version", version)
	}

	return nil
}
