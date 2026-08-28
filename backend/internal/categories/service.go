package categories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"flow/internal/models"
)

var (
	ErrInvalidCategoryName = errors.New("category name must be between 1 and 50 characters")
	ErrCategoryInUse       = errors.New("category cannot be deleted because it is referenced by existing activities")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListCategories(ctx context.Context, userID string) ([]models.Category, error) {
	return s.repo.ListCategories(ctx, userID)
}

func (s *Service) GetCategoryByID(ctx context.Context, id, userID string) (*models.Category, error) {
	return s.repo.GetCategoryByID(ctx, id, userID)
}

func (s *Service) CreateCategory(ctx context.Context, userID, name, color string) (*models.Category, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 || len(name) > 50 {
		return nil, ErrInvalidCategoryName
	}
	if color == "" {
		color = "#4F46E5"
	}
	return s.repo.CreateCategory(ctx, userID, name, color)
}

func (s *Service) UpdateCategory(ctx context.Context, id, userID, name, color string) (*models.Category, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 || len(name) > 50 {
		return nil, ErrInvalidCategoryName
	}
	if color == "" {
		color = "#4F46E5"
	}
	return s.repo.UpdateCategory(ctx, id, userID, name, color)
}

func (s *Service) DeleteCategory(ctx context.Context, id, userID string) error {
	// Check if category exists
	_, err := s.repo.GetCategoryByID(ctx, id, userID)
	if err != nil {
		return err
	}

	// Business Rule 4: Category Protection
	activityCount, err := s.repo.CountActivitiesByCategoryID(ctx, id, userID)
	if err != nil {
		return err
	}
	if activityCount > 0 {
		return fmt.Errorf("%w (%d activities linked)", ErrCategoryInUse, activityCount)
	}

	return s.repo.DeleteCategory(ctx, id, userID)
}

func (s *Service) SeedDefaultCategories(ctx context.Context, userID string) error {
	return s.repo.SeedDefaultCategories(ctx, userID)
}
