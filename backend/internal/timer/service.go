package timer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"flow/internal/categories"
	"flow/internal/models"
)

var (
	ErrEmptyTimerDescription = errors.New("timer description cannot be empty")
	ErrInvalidTimerCategory  = errors.New("category does not exist or does not belong to you")
)

type Service struct {
	repo        *Repository
	categorySvc *categories.Service
}

func NewService(repo *Repository, categorySvc *categories.Service) *Service {
	return &Service{
		repo:        repo,
		categorySvc: categorySvc,
	}
}

func (s *Service) GetActiveTimer(ctx context.Context, userID string) (*models.ActiveTimer, error) {
	t, err := s.repo.GetActiveTimer(ctx, userID)
	if err != nil {
		return nil, err
	}
	t.ElapsedSeconds = int(time.Since(t.StartedAt).Seconds())
	if t.ElapsedSeconds < 0 {
		t.ElapsedSeconds = 0
	}
	return t, nil
}

func (s *Service) StartTimer(ctx context.Context, userID, description, categoryID string) (*models.ActiveTimer, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return nil, ErrEmptyTimerDescription
	}

	// Validate category exists and belongs to user
	if _, err := s.categorySvc.GetCategoryByID(ctx, categoryID, userID); err != nil {
		return nil, ErrInvalidTimerCategory
	}

	// Business Rule 5: Check if user already has an active timer
	existing, err := s.repo.GetActiveTimer(ctx, userID)
	if err == nil && existing != nil {
		existing.ElapsedSeconds = int(time.Since(existing.StartedAt).Seconds())
		return nil, fmt.Errorf("%w: '%s' (running for %ds)", ErrTimerAlreadyActive, existing.Description, existing.ElapsedSeconds)
	}

	t := &models.ActiveTimer{
		UserID:      userID,
		CategoryID:  categoryID,
		Description: description,
	}

	created, err := s.repo.CreateActiveTimer(ctx, t)
	if err != nil {
		return nil, err
	}
	created.ElapsedSeconds = 0
	return created, nil
}

func (s *Service) StopTimer(ctx context.Context, userID string, note *string, favorite bool) (*models.Activity, error) {
	// Business Rule 6: Stop timer calculates duration from server timestamps and atomically converts to activity
	return s.repo.StopTimerTx(ctx, userID, note, favorite)
}

func (s *Service) DiscardTimer(ctx context.Context, userID string) error {
	return s.repo.DeleteActiveTimer(ctx, userID)
}
