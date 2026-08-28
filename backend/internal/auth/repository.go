package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"flow/internal/models"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrSessionNotFound = errors.New("session not found or expired")
	ErrDuplicateEmail  = errors.New("email already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	query := `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, email, created_at, updated_at
	`
	user := &models.User{PasswordHash: passwordHash}
	err := r.db.QueryRowContext(ctx, query, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

func (r *Repository) CreateSession(ctx context.Context, sessionID, userID string, expiresAt time.Time) error {
	query := `
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := r.db.ExecContext(ctx, query, sessionID, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *Repository) GetSession(ctx context.Context, sessionID string) (*models.Session, *models.User, error) {
	query := `
		SELECT s.id, s.user_id, s.expires_at, s.created_at, u.id, u.email, u.created_at, u.updated_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = $1 AND s.expires_at > NOW()
	`
	session := &models.Session{}
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrSessionNotFound
		}
		return nil, nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, user, nil
}

func (r *Repository) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	// pq error code 23505 = unique_violation
	return err != nil && (contains(err.Error(), "23505") || contains(err.Error(), "unique constraint") || contains(err.Error(), "duplicate key"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
