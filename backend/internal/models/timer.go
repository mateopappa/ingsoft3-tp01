package models

import (
	"time"
)

// ActiveTimer represents an ongoing time-tracking session persisted on the server.
type ActiveTimer struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	CategoryID     string    `json:"category_id"`
	CategoryName   string    `json:"category_name,omitempty"`
	CategoryColor  string    `json:"category_color,omitempty"`
	Description    string    `json:"description"`
	StartedAt      time.Time `json:"started_at"`
	CreatedAt      time.Time `json:"created_at"`
	ElapsedSeconds int       `json:"elapsed_seconds"` // Dynamically calculated on server
}

// StartTimerInput represents payload to start a persistent timer.
type StartTimerInput struct {
	Description string `json:"description"`
	CategoryID  string `json:"category_id"`
}

// StopTimerInput represents optional payload when stopping a timer.
type StopTimerInput struct {
	Note     *string `json:"note"`
	Favorite bool    `json:"favorite"`
}
