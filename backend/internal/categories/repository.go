package categories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"flow/internal/models"
)

var (
	ErrCategoryNotFound  = errors.New("category not found")
	ErrDuplicateCategory = errors.New("a category with this name already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListCategories(ctx context.Context, userID string) ([]models.Category, error) {
	query := `
		SELECT id, user_id, name, color, created_at, updated_at
		FROM categories
		WHERE user_id = $1
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var list []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Color, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.Category{}
	}
	return list, nil
}

func (r *Repository) GetCategoryByID(ctx context.Context, id, userID string) (*models.Category, error) {
	query := `
		SELECT id, user_id, name, color, created_at, updated_at
		FROM categories
		WHERE id = $1 AND user_id = $2
	`
	c := &models.Category{}
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Color, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return c, nil
}

func (r *Repository) CreateCategory(ctx context.Context, userID, name, color string) (*models.Category, error) {
	query := `
		INSERT INTO categories (user_id, name, color, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, user_id, name, color, created_at, updated_at
	`
	c := &models.Category{}
	err := r.db.QueryRowContext(ctx, query, userID, name, color).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Color, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateCategory
		}
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return c, nil
}

func (r *Repository) UpdateCategory(ctx context.Context, id, userID, name, color string) (*models.Category, error) {
	query := `
		UPDATE categories
		SET name = $1, color = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, name, color, created_at, updated_at
	`
	c := &models.Category{}
	err := r.db.QueryRowContext(ctx, query, name, color, id, userID).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Color, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		if isUniqueViolation(err) {
			return nil, ErrDuplicateCategory
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return c, nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id, userID string) error {
	query := `DELETE FROM categories WHERE id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *Repository) CountActivitiesByCategoryID(ctx context.Context, categoryID, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM activities WHERE category_id = $1 AND user_id = $2`
	var count int
	err := r.db.QueryRowContext(ctx, query, categoryID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count activities for category: %w", err)
	}
	return count, nil
}

func (r *Repository) SeedDefaultCategories(ctx context.Context, userID string) error {
	defaults := []struct {
		Name  string
		Color string
	}{
		{"Work", "#3B82F6"},
		{"Study", "#F59E0B"},
		{"Programming", "#10B981"},
		{"Training", "#EF4444"},
		{"Personal", "#8B5CF6"},
	}

	for _, d := range defaults {
		_, _ = r.CreateCategory(ctx, userID, d.Name, d.Color)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key"))
}
