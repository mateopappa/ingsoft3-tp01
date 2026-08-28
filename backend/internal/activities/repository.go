package activities

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"flow/internal/models"
)

var (
	ErrActivityNotFound = errors.New("activity not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateActivity(ctx context.Context, a *models.Activity) (*models.Activity, error) {
	query := `
		INSERT INTO activities (user_id, category_id, description, duration_seconds, activity_date, note, favorite, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx, query,
		a.UserID, a.CategoryID, a.Description, a.DurationSeconds, a.ActivityDate, a.Note, a.Favorite,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	// Fetch category info for response
	catQuery := `SELECT name, color FROM categories WHERE id = $1`
	_ = r.db.QueryRowContext(ctx, catQuery, a.CategoryID).Scan(&a.CategoryName, &a.CategoryColor)

	return a, nil
}

func (r *Repository) GetActivityByID(ctx context.Context, id, userID string) (*models.Activity, error) {
	query := `
		SELECT a.id, a.user_id, a.category_id, c.name, c.color, a.description, a.duration_seconds,
		       a.activity_date, a.note, a.favorite, a.created_at, a.updated_at
		FROM activities a
		JOIN categories c ON a.category_id = c.id
		WHERE a.id = $1 AND a.user_id = $2
	`
	a := &models.Activity{}
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&a.ID, &a.UserID, &a.CategoryID, &a.CategoryName, &a.CategoryColor, &a.Description,
		&a.DurationSeconds, &a.ActivityDate, &a.Note, &a.Favorite, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrActivityNotFound
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}
	return a, nil
}

func (r *Repository) UpdateActivity(ctx context.Context, a *models.Activity) (*models.Activity, error) {
	query := `
		UPDATE activities
		SET category_id = $1, description = $2, duration_seconds = $3, activity_date = $4,
		    note = $5, favorite = $6, updated_at = NOW()
		WHERE id = $7 AND user_id = $8
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(
		ctx, query,
		a.CategoryID, a.Description, a.DurationSeconds, a.ActivityDate, a.Note, a.Favorite, a.ID, a.UserID,
	).Scan(&a.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrActivityNotFound
		}
		return nil, fmt.Errorf("failed to update activity: %w", err)
	}

	catQuery := `SELECT name, color FROM categories WHERE id = $1`
	_ = r.db.QueryRowContext(ctx, catQuery, a.CategoryID).Scan(&a.CategoryName, &a.CategoryColor)

	return a, nil
}

func (r *Repository) DeleteActivity(ctx context.Context, id, userID string) error {
	query := `DELETE FROM activities WHERE id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrActivityNotFound
	}
	return nil
}

func (r *Repository) ToggleFavorite(ctx context.Context, id, userID string, favorite bool) error {
	query := `UPDATE activities SET favorite = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`
	res, err := r.db.ExecContext(ctx, query, favorite, id, userID)
	if err != nil {
		return fmt.Errorf("failed to toggle favorite: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrActivityNotFound
	}
	return nil
}

func (r *Repository) ListActivities(ctx context.Context, userID string, filter models.ActivityFilter) ([]models.Activity, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("a.user_id = $%d", argIdx))
	args = append(args, userID)
	argIdx++

	if filter.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("a.category_id = $%d", argIdx))
		args = append(args, filter.CategoryID)
		argIdx++
	}

	if filter.FromDate != "" {
		conditions = append(conditions, fmt.Sprintf("a.activity_date >= $%d", argIdx))
		args = append(args, filter.FromDate)
		argIdx++
	}

	if filter.ToDate != "" {
		conditions = append(conditions, fmt.Sprintf("a.activity_date <= $%d", argIdx))
		args = append(args, filter.ToDate)
		argIdx++
	}

	if filter.Favorite != nil {
		conditions = append(conditions, fmt.Sprintf("a.favorite = $%d", argIdx))
		args = append(args, *filter.Favorite)
		argIdx++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(a.description ILIKE $%d OR COALESCE(a.note, '') ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT a.id, a.user_id, a.category_id, c.name, c.color, a.description, a.duration_seconds,
		       a.activity_date, a.note, a.favorite, a.created_at, a.updated_at
		FROM activities a
		JOIN categories c ON a.category_id = c.id
		WHERE %s
		ORDER BY a.activity_date DESC, a.created_at DESC
	`, strings.Join(conditions, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}
	defer rows.Close()

	var list []models.Activity
	for rows.Next() {
		var a models.Activity
		err := rows.Scan(
			&a.ID, &a.UserID, &a.CategoryID, &a.CategoryName, &a.CategoryColor, &a.Description,
			&a.DurationSeconds, &a.ActivityDate, &a.Note, &a.Favorite, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.Activity{}
	}
	return list, nil
}
