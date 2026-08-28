package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"flow/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetTodaySeconds(ctx context.Context, userID, todayDate string) (int, error) {
	query := `
		SELECT COALESCE(SUM(duration_seconds), 0)
		FROM activities
		WHERE user_id = $1 AND activity_date = $2
	`
	var total int
	err := r.db.QueryRowContext(ctx, query, userID, todayDate).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get today total: %w", err)
	}
	return total, nil
}

func (r *Repository) GetWeekSeconds(ctx context.Context, userID, startDate, endDate string) (int, error) {
	query := `
		SELECT COALESCE(SUM(duration_seconds), 0)
		FROM activities
		WHERE user_id = $1 AND activity_date >= $2 AND activity_date <= $3
	`
	var total int
	err := r.db.QueryRowContext(ctx, query, userID, startDate, endDate).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get week total: %w", err)
	}
	return total, nil
}

func (r *Repository) GetCategoryDistribution(ctx context.Context, userID, startDate, endDate string) ([]models.CategoryDistribution, error) {
	query := `
		SELECT c.id, c.name, c.color, COALESCE(SUM(a.duration_seconds), 0) as total_seconds
		FROM categories c
		LEFT JOIN activities a ON a.category_id = c.id AND a.user_id = $1 AND a.activity_date >= $2 AND a.activity_date <= $3
		WHERE c.user_id = $1
		GROUP BY c.id, c.name, c.color
		HAVING COALESCE(SUM(a.duration_seconds), 0) > 0
		ORDER BY total_seconds DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get category distribution: %w", err)
	}
	defer rows.Close()

	var list []models.CategoryDistribution
	for rows.Next() {
		var cd models.CategoryDistribution
		if err := rows.Scan(&cd.CategoryID, &cd.Name, &cd.Color, &cd.Seconds); err != nil {
			return nil, fmt.Errorf("failed to scan category distribution: %w", err)
		}
		list = append(list, cd)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.CategoryDistribution{}
	}
	return list, nil
}

func (r *Repository) GetDailyActivity(ctx context.Context, userID, startDate, endDate string) (map[string]int, error) {
	query := `
		SELECT activity_date, COALESCE(SUM(duration_seconds), 0)
		FROM activities
		WHERE user_id = $1 AND activity_date >= $2 AND activity_date <= $3
		GROUP BY activity_date
	`
	rows, err := r.db.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily activity: %w", err)
	}
	defer rows.Close()

	res := make(map[string]int)
	for rows.Next() {
		var dateStr string
		var seconds int
		// scan date string
		var t time.Time
		if err := rows.Scan(&t, &seconds); err == nil {
			dateStr = t.Format("2006-01-02")
		} else {
			// fallback to string scan if driver returned string
			if err2 := rows.Scan(&dateStr, &seconds); err2 != nil {
				return nil, fmt.Errorf("failed to scan daily activity: %w", err)
			}
		}
		res[dateStr] = seconds
	}
	return res, nil
}
