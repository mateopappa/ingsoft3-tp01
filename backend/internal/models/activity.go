package models

import (
	"time"
)

// Activity represents a tracked unit of completed time.
type Activity struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	CategoryID      string    `json:"category_id"`
	CategoryName    string    `json:"category_name,omitempty"`
	CategoryColor   string    `json:"category_color,omitempty"`
	Description     string    `json:"description"`
	DurationSeconds int       `json:"duration_seconds"`
	ActivityDate    string    `json:"activity_date"` // YYYY-MM-DD
	Note            *string   `json:"note,omitempty"`
	Favorite        bool      `json:"favorite"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ActivityInput represents payload for creating or editing an activity.
type ActivityInput struct {
	Description     string  `json:"description"`
	CategoryID      string  `json:"category_id"`
	DurationSeconds int     `json:"duration_seconds"`
	ActivityDate    string  `json:"activity_date"`
	Note            *string `json:"note"`
	Favorite        bool    `json:"favorite"`
}

// ActivityFilter holds criteria for searching and filtering activities.
type ActivityFilter struct {
	Search     string
	CategoryID string
	FromDate   string
	ToDate     string
	Favorite   *bool
}
