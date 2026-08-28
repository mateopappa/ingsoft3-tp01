package models

// DashboardMetrics aggregates analytics data for the dashboard screen.
type DashboardMetrics struct {
	TodaySeconds         int                    `json:"today_seconds"`
	TodayFormatted       string                 `json:"today_formatted"`
	WeekSeconds          int                    `json:"week_seconds"`
	WeekFormatted        string                 `json:"week_formatted"`
	CategoryDistribution []CategoryDistribution `json:"category_distribution"`
	DailyActivity        []DailyActivity        `json:"daily_activity"`
}

// CategoryDistribution represents duration and percentage spent on a category.
type CategoryDistribution struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Seconds    int    `json:"seconds"`
	Percentage int    `json:"percentage"`
}

// DailyActivity represents tracked seconds for a specific day of the current week.
type DailyActivity struct {
	Day     string `json:"day"`     // "Mon", "Tue", etc.
	Date    string `json:"date"`    // YYYY-MM-DD
	Seconds int    `json:"seconds"` // Total duration in seconds
}
