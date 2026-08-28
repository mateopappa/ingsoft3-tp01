package models

import (
	"time"
)

// User represents an application user account.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session represents an active authenticated session for a user.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterInput represents registration payload.
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginInput represents login payload.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
