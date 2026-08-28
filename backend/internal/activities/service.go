package activities

import (
	"context"
	"errors"
	"strings"
	"time"

	"flow/internal/categories"
	"flow/internal/models"
)

var (
	ErrInvalidDuration    = errors.New("duration must be greater than zero seconds")
	ErrEmptyDescription   = errors.New("description cannot be empty")
	ErrInvalidDate        = errors.New("activity date must be in YYYY-MM-DD format")
	ErrInvalidCategory    = errors.New("category does not exist or does not belong to you")
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

func (s *Service) CreateActivity(ctx context.Context, userID string, input models.ActivityInput) (*models.Activity, error) {
	input.Description = strings.TrimSpace(input.Description)
	if input.Description == "" {
		return nil, ErrEmptyDescription
	}

	// Business Rule 1: Valid duration (> 0)
	if input.DurationSeconds <= 0 {
		return nil, ErrInvalidDuration
	}

	// Business Rule 8: Valid date format
	if input.ActivityDate == "" {
		input.ActivityDate = time.Now().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", input.ActivityDate); err != nil {
			return nil, ErrInvalidDate
		}
	}

	// Business Rule 2: Valid category owned by user
	if _, err := s.categorySvc.GetCategoryByID(ctx, input.CategoryID, userID); err != nil {
		return nil, ErrInvalidCategory
	}

	activity := &models.Activity{
		UserID:          userID,
		CategoryID:      input.CategoryID,
		Description:     input.Description,
		DurationSeconds: input.DurationSeconds,
		ActivityDate:    input.ActivityDate,
		Note:            input.Note,
		Favorite:        input.Favorite,
	}

	return s.repo.CreateActivity(ctx, activity)
}

func (s *Service) GetActivityByID(ctx context.Context, id, userID string) (*models.Activity, error) {
	return s.repo.GetActivityByID(ctx, id, userID)
}

func (s *Service) UpdateActivity(ctx context.Context, id, userID string, input models.ActivityInput) (*models.Activity, error) {
	input.Description = strings.TrimSpace(input.Description)
	if input.Description == "" {
		return nil, ErrEmptyDescription
	}

	if input.DurationSeconds <= 0 {
		return nil, ErrInvalidDuration
	}

	if input.ActivityDate != "" {
		if _, err := time.Parse("2006-01-02", input.ActivityDate); err != nil {
			return nil, ErrInvalidDate
		}
	}

	if _, err := s.categorySvc.GetCategoryByID(ctx, input.CategoryID, userID); err != nil {
		return nil, ErrInvalidCategory
	}

	activity := &models.Activity{
		ID:              id,
		UserID:          userID,
		CategoryID:      input.CategoryID,
		Description:     input.Description,
		DurationSeconds: input.DurationSeconds,
		ActivityDate:    input.ActivityDate,
		Note:            input.Note,
		Favorite:        input.Favorite,
	}

	return s.repo.UpdateActivity(ctx, activity)
}

func (s *Service) DeleteActivity(ctx context.Context, id, userID string) error {
	return s.repo.DeleteActivity(ctx, id, userID)
}

func (s *Service) ToggleFavorite(ctx context.Context, id, userID string, favorite bool) error {
	return s.repo.ToggleFavorite(ctx, id, userID, favorite)
}

func (s *Service) ListActivities(ctx context.Context, userID string, filter models.ActivityFilter) ([]models.Activity, error) {
	return s.repo.ListActivities(ctx, userID, filter)
}
