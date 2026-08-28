package unit

import (
	"testing"
	"time"

	"flow/internal/dashboard"
	"flow/internal/models"
)

// Test BT-8: Dashboard calculations & percentage math
func TestDashboardCalculationMath(t *testing.T) {
	t.Run("FormatDuration", func(t *testing.T) {
		tests := []struct {
			seconds  int
			expected string
		}{
			{0, "0m"},
			{-10, "0m"},
			{45, "0m"},
			{60, "1m"},
			{3600, "1h 00m"},
			{5022, "1h 23m"},
			{20520, "5h 42m"},
			{98100, "27h 15m"},
		}

		for _, tt := range tests {
			res := dashboard.FormatDuration(tt.seconds)
			if res != tt.expected {
				t.Errorf("FormatDuration(%d) = '%s', expected '%s'", tt.seconds, res, tt.expected)
			}
		}
	})

	t.Run("CategoryPercentageCalculation", func(t *testing.T) {
		weekTotalSeconds := 10000

		dist := []models.CategoryDistribution{
			{CategoryID: "1", Name: "Programming", Seconds: 4200},
			{CategoryID: "2", Name: "Work", Seconds: 3100},
			{CategoryID: "3", Name: "Study", Seconds: 1800},
			{CategoryID: "4", Name: "Training", Seconds: 900},
		}

		for i := range dist {
			dist[i].Percentage = int((float64(dist[i].Seconds) / float64(weekTotalSeconds)) * 100)
		}

		if dist[0].Percentage != 42 {
			t.Errorf("expected Programming 42%%, got %d%%", dist[0].Percentage)
		}
		if dist[1].Percentage != 31 {
			t.Errorf("expected Work 31%%, got %d%%", dist[1].Percentage)
		}
		if dist[2].Percentage != 18 {
			t.Errorf("expected Study 18%%, got %d%%", dist[2].Percentage)
		}
		if dist[3].Percentage != 9 {
			t.Errorf("expected Training 9%%, got %d%%", dist[3].Percentage)
		}
	})

	t.Run("WeekMondaySundayCalculation", func(t *testing.T) {
		// Reference Wednesday: 2026-08-19
		refDate := time.Date(2026, 8, 19, 14, 0, 0, 0, time.UTC)
		weekday := refDate.Weekday()
		daysFromMonday := int(weekday - time.Monday)

		monday := refDate.AddDate(0, 0, -daysFromMonday)
		sunday := monday.AddDate(0, 0, 6)

		if monday.Format("2006-01-02") != "2026-08-17" {
			t.Errorf("expected Monday 2026-08-17, got %s", monday.Format("2006-01-02"))
		}
		if sunday.Format("2006-01-02") != "2026-08-23" {
			t.Errorf("expected Sunday 2026-08-23, got %s", sunday.Format("2006-01-02"))
		}
	})
}
