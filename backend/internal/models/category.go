package models

import (
	"time"
)

// Category represents a user-managed classification for activities.
type Category struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CategoryInput represents category creation/update payload.
type CategoryInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}
