package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Open establishes a configured connection pool to PostgreSQL.
func Open(databaseURL string) (*sql.DB, error) {
	dsn, err := withNeonPoolerCompat(databaseURL)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database handle: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return db, nil
}

// withNeonPoolerCompat adds binary_parameters=yes for Neon/PgBouncer transaction
// poolers. Without it, lib/pq prepared statements fail at migrate-check time with:
// "pq: unnamed prepared statement does not exist".
func withNeonPoolerCompat(databaseURL string) (string, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid DATABASE_URL: %w", err)
	}
	q := u.Query()
	if q.Get("binary_parameters") == "" {
		q.Set("binary_parameters", "yes")
		u.RawQuery = q.Encode()
	}
	// url.Parse lowercases scheme; postgres URLs must keep userinfo intact.
	out := u.String()
	// Prefer postgres:// over postgresql:// if the input used the short form.
	if strings.HasPrefix(databaseURL, "postgres://") && strings.HasPrefix(out, "postgresql://") {
		out = "postgres://" + strings.TrimPrefix(out, "postgresql://")
	}
	return out, nil
}
