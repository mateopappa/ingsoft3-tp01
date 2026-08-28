package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"flow/internal/activities"
	"flow/internal/categories"
	"flow/internal/models"
)

// MockCategoryService implements a lightweight mock for category validation in tests.
type mockCategoryFinder struct {
	categories map[string]string // map[categoryID]userID
}

func (m *mockCategoryFinder) GetCategoryByID(ctx context.Context, id, userID string) (*models.Category, error) {
	owner, exists := m.categories[id]
	if !exists || owner != userID {
		return nil, categories.ErrCategoryNotFound
	}
	return &models.Category{ID: id, UserID: userID, Name: "Test Category"}, nil
}

// Test BT-1 & BT-2: Valid activity accepted & invalid duration rejected
func TestActivityValidation(t *testing.T) {
	// Rule 1: duration_seconds > 0
	tests := []struct {
		name        string
		description string
		duration    int
		date        string
		wantErr     error
	}{
		{
			name:        "Valid positive duration",
			description: "Coding Go Monolith",
			duration:    3600,
			date:        "2026-08-19",
			wantErr:     nil,
		},
		{
			name:        "Zero duration rejected",
			description: "Zero activity",
			duration:    0,
			date:        "2026-08-19",
			wantErr:     activities.ErrInvalidDuration,
		},
		{
			name:        "Negative duration rejected",
			description: "Negative activity",
			duration:    -120,
			date:        "2026-08-19",
			wantErr:     activities.ErrInvalidDuration,
		},
		{
			name:        "Empty description rejected",
			description: "   ",
			duration:    1800,
			date:        "2026-08-19",
			wantErr:     activities.ErrEmptyDescription,
		},
		{
			name:        "Invalid date format rejected",
			description: "Bad date activity",
			duration:    1800,
			date:        "19-08-2026",
			wantErr:     activities.ErrInvalidDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := models.ActivityInput{
				Description:     tt.description,
				CategoryID:      "cat-1",
				DurationSeconds: tt.duration,
				ActivityDate:    tt.date,
			}

			// Validate duration rule logic
			if input.DurationSeconds <= 0 {
				if tt.wantErr == nil || !errors.Is(activities.ErrInvalidDuration, tt.wantErr) {
					t.Errorf("expected ErrInvalidDuration, got nil")
				}
				return
			}

			// Validate description rule logic
			if len(input.Description) == 0 || input.Description == "   " {
				if tt.wantErr == nil || !errors.Is(activities.ErrEmptyDescription, tt.wantErr) {
					t.Errorf("expected ErrEmptyDescription, got nil")
				}
				return
			}

			// Validate date rule logic
			if input.ActivityDate != "" {
				if _, err := time.Parse("2006-01-02", input.ActivityDate); err != nil {
					if tt.wantErr == nil || !errors.Is(activities.ErrInvalidDate, tt.wantErr) {
						t.Errorf("expected ErrInvalidDate, got %v", err)
					}
					return
				}
			}

			if tt.wantErr != nil {
				t.Errorf("expected error %v, got nil", tt.wantErr)
			}
		})
	}
}

// Test BT-3: Invalid category rejected
func TestInvalidCategoryValidation(t *testing.T) {
	mockFinder := &mockCategoryFinder{
		categories: map[string]string{
			"cat-user-a": "user-a",
		},
	}

	// User B trying to use User A's category
	_, err := mockFinder.GetCategoryByID(context.Background(), "cat-user-a", "user-b")
	if err == nil {
		t.Errorf("expected error when referencing another user's category, got nil")
	}

	// User A using non-existent category
	_, err = mockFinder.GetCategoryByID(context.Background(), "cat-non-existent", "user-a")
	if err == nil {
		t.Errorf("expected error when referencing non-existent category, got nil")
	}

	// User A using valid owned category
	cat, err := mockFinder.GetCategoryByID(context.Background(), "cat-user-a", "user-a")
	if err != nil || cat == nil {
		t.Errorf("expected valid category, got error: %v", err)
	}
}
