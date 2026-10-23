package dashboard_test

import (
	"context"
	"testing"
	"time"

	"flow/internal/dashboard"
	"flow/internal/models"
)

type mockMetricsStore struct {
	todaySeconds int
	weekSeconds  int
	dist         []models.CategoryDistribution
	daily        map[string]int
}

func (m *mockMetricsStore) GetTodaySeconds(ctx context.Context, userID, todayDate string) (int, error) {
	return m.todaySeconds, nil
}
func (m *mockMetricsStore) GetWeekSeconds(ctx context.Context, userID, startDate, endDate string) (int, error) {
	return m.weekSeconds, nil
}
func (m *mockMetricsStore) GetCategoryDistribution(ctx context.Context, userID, startDate, endDate string) ([]models.CategoryDistribution, error) {
	return m.dist, nil
}
func (m *mockMetricsStore) GetDailyActivity(ctx context.Context, userID, startDate, endDate string) (map[string]int, error) {
	if m.daily == nil {
		return map[string]int{}, nil
	}
	return m.daily, nil
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
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
	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			if got := dashboard.FormatDuration(tc.seconds); got != tc.expected {
				t.Fatalf("FormatDuration(%d)=%q want %q", tc.seconds, got, tc.expected)
			}
		})
	}
}

func TestGetMetrics_CategoryPercentages(t *testing.T) {
	// Arrange — Wednesday 2026-08-19 → week Mon 17 .. Sun 23
	store := &mockMetricsStore{
		todaySeconds: 4200,
		weekSeconds:  10000,
		dist: []models.CategoryDistribution{
			{CategoryID: "1", Name: "Programming", Seconds: 4200},
			{CategoryID: "2", Name: "Work", Seconds: 3100},
			{CategoryID: "3", Name: "Study", Seconds: 1800},
			{CategoryID: "4", Name: "Training", Seconds: 900},
		},
		daily: map[string]int{"2026-08-19": 4200},
	}
	svc := dashboard.NewService(store)
	ref := time.Date(2026, 8, 19, 14, 0, 0, 0, time.UTC)

	// Act
	metrics, err := svc.GetMetrics(context.Background(), "u1", ref)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.TodaySeconds != 4200 || metrics.WeekSeconds != 10000 {
		t.Fatalf("unexpected totals today=%d week=%d", metrics.TodaySeconds, metrics.WeekSeconds)
	}
	want := []int{42, 31, 18, 9}
	for i, p := range want {
		if metrics.CategoryDistribution[i].Percentage != p {
			t.Fatalf("category %d percentage=%d want %d", i, metrics.CategoryDistribution[i].Percentage, p)
		}
	}
	if len(metrics.DailyActivity) != 7 {
		t.Fatalf("expected 7 daily slots, got %d", len(metrics.DailyActivity))
	}
	if metrics.DailyActivity[0].Date != "2026-08-17" || metrics.DailyActivity[6].Date != "2026-08-23" {
		t.Fatalf("week bounds wrong: %s .. %s", metrics.DailyActivity[0].Date, metrics.DailyActivity[6].Date)
	}
}

func TestGetMetrics_ZeroWeekSeconds_PercentagesAreZero(t *testing.T) {
	// Uncovered-path exercise target: the `else` branch when weekSeconds == 0
	store := &mockMetricsStore{
		weekSeconds: 0,
		dist: []models.CategoryDistribution{
			{CategoryID: "1", Name: "Programming", Seconds: 0},
		},
	}
	svc := dashboard.NewService(store)
	ref := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC) // Sunday → daysFromMonday < 0 path

	metrics, err := svc.GetMetrics(context.Background(), "u1", ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.CategoryDistribution[0].Percentage != 0 {
		t.Fatalf("expected 0%% when week total is 0, got %d", metrics.CategoryDistribution[0].Percentage)
	}
	if metrics.DailyActivity[0].Date != "2026-08-10" {
		t.Fatalf("Sunday ref should yield Monday 2026-08-10, got %s", metrics.DailyActivity[0].Date)
	}
}
