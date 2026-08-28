package dashboard

import (
	"context"
	"fmt"
	"time"

	"flow/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMetrics(ctx context.Context, userID string, clientNow time.Time) (*models.DashboardMetrics, error) {
	todayStr := clientNow.Format("2006-01-02")

	// Calculate Monday of current week
	weekday := clientNow.Weekday()
	daysFromMonday := int(weekday - time.Monday)
	if daysFromMonday < 0 {
		daysFromMonday = 6 // Sunday is day 6 after Monday
	}

	monday := clientNow.AddDate(0, 0, -daysFromMonday)
	sunday := monday.AddDate(0, 0, 6)

	mondayStr := monday.Format("2006-01-02")
	sundayStr := sunday.Format("2006-01-02")

	// 1. Today total
	todaySeconds, err := s.repo.GetTodaySeconds(ctx, userID, todayStr)
	if err != nil {
		return nil, err
	}

	// 2. Week total
	weekSeconds, err := s.repo.GetWeekSeconds(ctx, userID, mondayStr, sundayStr)
	if err != nil {
		return nil, err
	}

	// 3. Category distribution
	catDist, err := s.repo.GetCategoryDistribution(ctx, userID, mondayStr, sundayStr)
	if err != nil {
		return nil, err
	}

	// Calculate percentage for each category
	for i := range catDist {
		if weekSeconds > 0 {
			catDist[i].Percentage = int((float64(catDist[i].Seconds) / float64(weekSeconds)) * 100)
		} else {
			catDist[i].Percentage = 0
		}
	}

	// 4. Daily activity for Monday through Sunday
	dailyMap, err := s.repo.GetDailyActivity(ctx, userID, mondayStr, sundayStr)
	if err != nil {
		return nil, err
	}

	dayNames := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	dailyActivities := make([]models.DailyActivity, 7)
	for i := 0; i < 7; i++ {
		dayDate := monday.AddDate(0, 0, i)
		dStr := dayDate.Format("2006-01-02")
		dailyActivities[i] = models.DailyActivity{
			Day:     dayNames[i],
			Date:    dStr,
			Seconds: dailyMap[dStr],
		}
	}

	metrics := &models.DashboardMetrics{
		TodaySeconds:         todaySeconds,
		TodayFormatted:       FormatDuration(todaySeconds),
		WeekSeconds:          weekSeconds,
		WeekFormatted:        FormatDuration(weekSeconds),
		CategoryDistribution: catDist,
		DailyActivity:        dailyActivities,
	}

	return metrics, nil
}

// FormatDuration converts seconds into human-readable e.g. "5h 42m" or "45m" or "0m".
func FormatDuration(seconds int) string {
	if seconds <= 0 {
		return "0m"
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	if hours > 0 {
		return fmt.Sprintf("%dh %02dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
