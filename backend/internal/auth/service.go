package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"flow/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail       = errors.New("invalid email address format")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters long")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type CategorySeeder interface {
	SeedDefaultCategories(ctx context.Context, userID string) error
}

type Service struct {
	repo            *Repository
	categorySeeder  CategorySeeder
	sessionDuration time.Duration
}

func NewService(repo *Repository, categorySeeder CategorySeeder, sessionDuration time.Duration) *Service {
	return &Service{
		repo:            repo,
		categorySeeder:  categorySeeder,
		sessionDuration: sessionDuration,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) (*models.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if err := validateEmail(email); err != nil {
		return nil, "", err
	}
	if len(password) < 8 {
		return nil, "", ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, email, string(hash))
	if err != nil {
		return nil, "", err
	}

	// Automatically seed default categories for new user
	if s.categorySeeder != nil {
		_ = s.categorySeeder.SeedDefaultCategories(ctx, user.ID)
	}

	// Create session
	sessionID, err := s.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, sessionID, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	sessionID, err := s.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, sessionID, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return s.repo.DeleteSession(ctx, sessionID)
}

func (s *Service) ValidateSession(ctx context.Context, sessionID string) (*models.User, error) {
	if sessionID == "" {
		return nil, ErrSessionNotFound
	}
	_, user, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) createSession(ctx context.Context, userID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	sessionID := hex.EncodeToString(b)
	expiresAt := time.Now().Add(s.sessionDuration)

	if err := s.repo.CreateSession(ctx, sessionID, userID, expiresAt); err != nil {
		return "", err
	}

	return sessionID, nil
}

func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}
