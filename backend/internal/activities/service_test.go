package activities_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"flow/internal/activities"
	"flow/internal/models"
)

// --- doubles ---

type stubCategoryFinder struct {
	cat *models.Category
	err error
}

func (s *stubCategoryFinder) GetCategoryByID(ctx context.Context, id, userID string) (*models.Category, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.cat, nil
}

type mockActivityStore struct {
	createCalls int
	lastCreated *models.Activity
	createErr   error
	getResult   *models.Activity
	getErr      error
}

func (m *mockActivityStore) CreateActivity(ctx context.Context, a *models.Activity) (*models.Activity, error) {
	m.createCalls++
	m.lastCreated = a
	if m.createErr != nil {
		return nil, m.createErr
	}
	out := *a
	out.ID = "act-1"
	out.CreatedAt = time.Now()
	out.UpdatedAt = out.CreatedAt
	return &out, nil
}

func (m *mockActivityStore) GetActivityByID(ctx context.Context, id, userID string) (*models.Activity, error) {
	return m.getResult, m.getErr
}
func (m *mockActivityStore) UpdateActivity(ctx context.Context, a *models.Activity) (*models.Activity, error) {
	return a, nil
}
func (m *mockActivityStore) DeleteActivity(ctx context.Context, id, userID string) error { return nil }
func (m *mockActivityStore) ToggleFavorite(ctx context.Context, id, userID string, favorite bool) error {
	return nil
}
func (m *mockActivityStore) ListActivities(ctx context.Context, userID string, filter models.ActivityFilter) ([]models.Activity, error) {
	return nil, nil
}

// --- Rule 1: duration must be > 0 (parameterized) ---

func TestCreateActivity_DurationValidation(t *testing.T) {
	cases := []struct {
		name     string
		duration int
		wantErr  error
	}{
		{"positive duration accepted", 3600, nil},
		{"zero duration rejected", 0, activities.ErrInvalidDuration},
		{"negative duration rejected", -120, activities.ErrInvalidDuration},
		{"one second accepted", 1, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			store := &mockActivityStore{}
			cats := &stubCategoryFinder{cat: &models.Category{ID: "cat-1", UserID: "user-1"}}
			svc := activities.NewService(store, cats)
			input := models.ActivityInput{
				Description:     "Coding",
				CategoryID:      "cat-1",
				DurationSeconds: tc.duration,
				ActivityDate:    "2026-08-19",
			}

			// Act
			_, err := svc.CreateActivity(context.Background(), "user-1", input)

			// Assert
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				if store.createCalls != 1 {
					t.Fatalf("expected CreateActivity called once, got %d", store.createCalls)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
			if store.createCalls != 0 {
				t.Fatalf("repo must not be called on validation failure")
			}
		})
	}
}

// --- Rule: empty description (error case) ---

func TestCreateActivity_EmptyDescriptionRejected(t *testing.T) {
	// Arrange
	store := &mockActivityStore{}
	cats := &stubCategoryFinder{cat: &models.Category{ID: "cat-1"}}
	svc := activities.NewService(store, cats)

	// Act
	_, err := svc.CreateActivity(context.Background(), "user-1", models.ActivityInput{
		Description:     "   ",
		CategoryID:      "cat-1",
		DurationSeconds: 600,
		ActivityDate:    "2026-08-19",
	})

	// Assert
	if !errors.Is(err, activities.ErrEmptyDescription) {
		t.Fatalf("expected ErrEmptyDescription, got %v", err)
	}
}

// --- Rule: invalid date format (error case) ---

func TestCreateActivity_InvalidDateRejected(t *testing.T) {
	store := &mockActivityStore{}
	cats := &stubCategoryFinder{cat: &models.Category{ID: "cat-1"}}
	svc := activities.NewService(store, cats)

	_, err := svc.CreateActivity(context.Background(), "user-1", models.ActivityInput{
		Description:     "Bad date",
		CategoryID:      "cat-1",
		DurationSeconds: 600,
		ActivityDate:    "19-08-2026",
	})

	if !errors.Is(err, activities.ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}
}

// --- Rule 2: category must belong to user (mock interaction) ---

func TestCreateActivity_InvalidCategoryRejected_AndRepoNotCalled(t *testing.T) {
	// Arrange: category finder fails → mock verifies store is never touched
	store := &mockActivityStore{}
	cats := &stubCategoryFinder{err: errors.New("not found")}
	svc := activities.NewService(store, cats)

	// Act
	_, err := svc.CreateActivity(context.Background(), "user-b", models.ActivityInput{
		Description:     "Stolen category",
		CategoryID:      "cat-user-a",
		DurationSeconds: 900,
		ActivityDate:    "2026-08-19",
	})

	// Assert
	if !errors.Is(err, activities.ErrInvalidCategory) {
		t.Fatalf("expected ErrInvalidCategory, got %v", err)
	}
	if store.createCalls != 0 {
		t.Fatalf("CreateActivity on repo must not run when category is invalid")
	}
}

func TestUpdateActivity_InvalidDurationRejected(t *testing.T) {
	store := &mockActivityStore{}
	cats := &stubCategoryFinder{cat: &models.Category{ID: "cat-1"}}
	svc := activities.NewService(store, cats)

	_, err := svc.UpdateActivity(context.Background(), "act-1", "user-1", models.ActivityInput{
		Description:     "Update",
		CategoryID:      "cat-1",
		DurationSeconds: 0,
		ActivityDate:    "2026-08-19",
	})
	if !errors.Is(err, activities.ErrInvalidDuration) {
		t.Fatalf("expected ErrInvalidDuration, got %v", err)
	}
}

func TestGetActivityByID_DelegatesToStore(t *testing.T) {
	store := &mockActivityStore{
		getResult: &models.Activity{ID: "act-1", Description: "Found"},
	}
	svc := activities.NewService(store, &stubCategoryFinder{})

	got, err := svc.GetActivityByID(context.Background(), "act-1", "user-1")
	if err != nil || got.Description != "Found" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestCreateActivity_PersistsWhenValid(t *testing.T) {
	store := &mockActivityStore{}
	cats := &stubCategoryFinder{cat: &models.Category{ID: "cat-1", UserID: "user-1", Name: "Work"}}
	svc := activities.NewService(store, cats)

	got, err := svc.CreateActivity(context.Background(), "user-1", models.ActivityInput{
		Description:     "  Ship TP5  ",
		CategoryID:      "cat-1",
		DurationSeconds: 1800,
		ActivityDate:    "2026-08-19",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "act-1" {
		t.Fatalf("expected id act-1, got %s", got.ID)
	}
	if store.lastCreated.Description != "Ship TP5" {
		t.Fatalf("description should be trimmed, got %q", store.lastCreated.Description)
	}
}
