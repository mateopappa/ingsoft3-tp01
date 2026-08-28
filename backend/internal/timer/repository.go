package timer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"flow/internal/models"
)

var (
	ErrNoActiveTimer      = errors.New("no active timer found")
	ErrTimerAlreadyActive = errors.New("a timer is already active for this user")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetActiveTimer(ctx context.Context, userID string) (*models.ActiveTimer, error) {
	query := `
		SELECT t.id, t.user_id, t.category_id, c.name, c.color, t.description, t.started_at, t.created_at
		FROM active_timers t
		JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1
	`
	t := &models.ActiveTimer{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&t.ID, &t.UserID, &t.CategoryID, &t.CategoryName, &t.CategoryColor,
		&t.Description, &t.StartedAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoActiveTimer
		}
		return nil, fmt.Errorf("failed to get active timer: %w", err)
	}
	return t, nil
}

func (r *Repository) CreateActiveTimer(ctx context.Context, t *models.ActiveTimer) (*models.ActiveTimer, error) {
	query := `
		INSERT INTO active_timers (user_id, category_id, description, started_at, created_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, started_at, created_at
	`
	err := r.db.QueryRowContext(ctx, query, t.UserID, t.CategoryID, t.Description).Scan(
		&t.ID, &t.StartedAt, &t.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrTimerAlreadyActive
		}
		return nil, fmt.Errorf("failed to create active timer: %w", err)
	}

	catQuery := `SELECT name, color FROM categories WHERE id = $1`
	_ = r.db.QueryRowContext(ctx, catQuery, t.CategoryID).Scan(&t.CategoryName, &t.CategoryColor)

	return t, nil
}

func (r *Repository) DeleteActiveTimer(ctx context.Context, userID string) error {
	query := `DELETE FROM active_timers WHERE user_id = $1`
	res, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete active timer: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNoActiveTimer
	}
	return nil
}

// StopTimerTx atomically converts the active timer into an activity and deletes the timer.
func (r *Repository) StopTimerTx(ctx context.Context, userID string, note *string, favorite bool) (*models.Activity, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin stop timer transaction: %w", err)
	}
	defer tx.Rollback()

	// Select active timer with lock
	query := `
		SELECT id, category_id, description, started_at
		FROM active_timers
		WHERE user_id = $1
		FOR UPDATE
	`
	var timerID, categoryID, description string
	var startedAt time.Time
	err = tx.QueryRowContext(ctx, query, userID).Scan(&timerID, &categoryID, &description, &startedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoActiveTimer
		}
		return nil, fmt.Errorf("failed to fetch active timer in transaction: %w", err)
	}

	now := time.Now()
	durationSeconds := int(now.Sub(startedAt).Seconds())
	if durationSeconds < 1 {
		durationSeconds = 1 // Minimum 1 second duration
	}
	activityDate := now.Format("2006-01-02")

	// Insert activity
	insertQuery := `
		INSERT INTO activities (user_id, category_id, description, duration_seconds, activity_date, note, favorite, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	activity := &models.Activity{
		UserID:          userID,
		CategoryID:      categoryID,
		Description:     description,
		DurationSeconds: durationSeconds,
		ActivityDate:    activityDate,
		Note:            note,
		Favorite:        favorite,
	}
	err = tx.QueryRowContext(
		ctx, insertQuery,
		activity.UserID, activity.CategoryID, activity.Description,
		activity.DurationSeconds, activity.ActivityDate, activity.Note, activity.Favorite,
	).Scan(&activity.ID, &activity.CreatedAt, &activity.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert activity in stop timer transaction: %w", err)
	}

	// Delete active timer
	_, err = tx.ExecContext(ctx, `DELETE FROM active_timers WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete active timer in transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit stop timer transaction: %w", err)
	}

	catQuery := `SELECT name, color FROM categories WHERE id = $1`
	_ = r.db.QueryRowContext(ctx, catQuery, activity.CategoryID).Scan(&activity.CategoryName, &activity.CategoryColor)

	return activity, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key"))
}
