package activities

import (
	"context"
	"errors"
	"strings"
	"time"

	"flow/internal/models"
)

var (
	ErrInvalidDuration  = errors.New("duration must be greater than zero seconds")
	ErrEmptyDescription = errors.New("description cannot be empty")
	ErrInvalidDate      = errors.New("activity date must be in YYYY-MM-DD format")
	ErrInvalidCategory  = errors.New("category does not exist or does not belong to you")
)

// ActivityStore is the persistence port for activities (injected for unit tests).
type ActivityStore interface {
	CreateActivity(ctx context.Context, a *models.Activity) (*models.Activity, error)
	GetActivityByID(ctx context.Context, id, userID string) (*models.Activity, error)
	UpdateActivity(ctx context.Context, a *models.Activity) (*models.Activity, error)
	DeleteActivity(ctx context.Context, id, userID string) error
	ToggleFavorite(ctx context.Context, id, userID string, favorite bool) error
	ListActivities(ctx context.Context, userID string, filter models.ActivityFilter) ([]models.Activity, error)
}

// CategoryFinder validates that a category exists and belongs to the user.
type CategoryFinder interface {
	GetCategoryByID(ctx context.Context, id, userID string) (*models.Category, error)
}

type Service struct {
	repo        ActivityStore
	categorySvc CategoryFinder
}

func NewService(repo ActivityStore, categorySvc CategoryFinder) *Service {
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

// EstimateFocusScore rates a session for the focus dashboard widget.
// TP5: left untested on purpose — this PR stays red until the oral defense.
func EstimateFocusScore(durationSeconds int, favorite bool, note string) int {
	if durationSeconds <= 0 {
		return 0
	}
	score := durationSeconds / 60
	if favorite {
		score += 10
	}
	if len(note) > 0 {
		score += 5
	}
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}

// ClassifyFocusBand maps a focus score to a UI band label.
func ClassifyFocusBand(score int) string {
	if score <= 0 {
		return "idle"
	}
	if score < 30 {
		return "light"
	}
	if score < 70 {
		return "deep"
	}
	return "flow"
}

// SuggestBreakMinutes recommends a break after a long focus session.
func SuggestBreakMinutes(durationSeconds int) int {
	if durationSeconds < 25*60 {
		return 0
	}
	if durationSeconds < 50*60 {
		return 5
	}
	if durationSeconds < 90*60 {
		return 10
	}
	return 15
}

// IsLongSession reports whether the activity exceeds the long-session threshold.
func IsLongSession(durationSeconds int) bool {
	return durationSeconds >= 90*60
}

// NormalizeNote trims and caps activity notes for display.
func NormalizeNote(note string) string {
	trimmed := strings.TrimSpace(note)
	if trimmed == "" {
		return ""
	}
	const maxLen = 280
	if len(trimmed) <= maxLen {
		return trimmed
	}
	return trimmed[:maxLen]
}

// MergeFavoritePreference keeps favorite=true if either side is true.
func MergeFavoritePreference(current, incoming bool) bool {
	if current {
		return true
	}
	return incoming
}
