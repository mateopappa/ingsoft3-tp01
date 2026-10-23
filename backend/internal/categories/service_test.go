package categories_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"flow/internal/categories"
	"flow/internal/models"
)

type mockCategoryStore struct {
	getResult     *models.Category
	getErr        error
	activityCount int
	countErr      error
	deleteCalls   int
	createCalls   int
	lastColor     string
	createResult  *models.Category
	createErr     error
}

func (m *mockCategoryStore) ListCategories(ctx context.Context, userID string) ([]models.Category, error) {
	return nil, nil
}
func (m *mockCategoryStore) GetCategoryByID(ctx context.Context, id, userID string) (*models.Category, error) {
	return m.getResult, m.getErr
}
func (m *mockCategoryStore) CreateCategory(ctx context.Context, userID, name, color string) (*models.Category, error) {
	m.createCalls++
	m.lastColor = color
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.createResult != nil {
		return m.createResult, nil
	}
	return &models.Category{ID: "c1", UserID: userID, Name: name, Color: color}, nil
}
func (m *mockCategoryStore) UpdateCategory(ctx context.Context, id, userID, name, color string) (*models.Category, error) {
	return &models.Category{ID: id, UserID: userID, Name: name, Color: color}, nil
}
func (m *mockCategoryStore) DeleteCategory(ctx context.Context, id, userID string) error {
	m.deleteCalls++
	return nil
}
func (m *mockCategoryStore) CountActivitiesByCategoryID(ctx context.Context, categoryID, userID string) (int, error) {
	return m.activityCount, m.countErr
}
func (m *mockCategoryStore) SeedDefaultCategories(ctx context.Context, userID string) error {
	return nil
}

// --- Rule 4: cannot delete category referenced by activities ---

func TestDeleteCategory_InUseRejected(t *testing.T) {
	cases := []struct {
		name          string
		activityCount int
		wantErr       bool
		wantDelete    int
	}{
		{"zero activities allows delete", 0, false, 1},
		{"five activities blocks delete", 5, true, 0},
		{"one activity blocks delete", 1, true, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			store := &mockCategoryStore{
				getResult:     &models.Category{ID: "c1", UserID: "u1", Name: "Work"},
				activityCount: tc.activityCount,
			}
			svc := categories.NewService(store)

			// Act
			err := svc.DeleteCategory(context.Background(), "c1", "u1")

			// Assert
			if tc.wantErr {
				if !errors.Is(err, categories.ErrCategoryInUse) {
					t.Fatalf("expected ErrCategoryInUse, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if store.deleteCalls != tc.wantDelete {
				t.Fatalf("deleteCalls=%d want %d", store.deleteCalls, tc.wantDelete)
			}
		})
	}
}

func TestDeleteCategory_NotFoundPropagates(t *testing.T) {
	store := &mockCategoryStore{getErr: categories.ErrCategoryNotFound}
	svc := categories.NewService(store)

	err := svc.DeleteCategory(context.Background(), "missing", "u1")
	if !errors.Is(err, categories.ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
	if store.deleteCalls != 0 {
		t.Fatalf("must not delete when category is missing")
	}
}

// --- Rule: category name length (parameterized + error) ---

func TestCreateCategory_NameValidation(t *testing.T) {
	cases := []struct {
		name    string
		catName string
		wantErr error
	}{
		{"valid standard name", "Programming", nil},
		{"valid short name", "Go", nil},
		{"empty string invalid", "", categories.ErrInvalidCategoryName},
		{"whitespace only invalid", "   ", categories.ErrInvalidCategoryName},
		{"over 50 characters invalid", strings.Repeat("A", 51), categories.ErrInvalidCategoryName},
		{"exactly 50 characters valid", strings.Repeat("B", 50), nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockCategoryStore{}
			svc := categories.NewService(store)

			_, err := svc.CreateCategory(context.Background(), "u1", tc.catName, "#10B981")

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
			if store.createCalls != 0 {
				t.Fatalf("repo must not be called on invalid name")
			}
		})
	}
}

func TestListAndGetCategory_DelegateToStore(t *testing.T) {
	store := &mockCategoryStore{
		getResult: &models.Category{ID: "c1", UserID: "u1", Name: "Work"},
	}
	svc := categories.NewService(store)

	if _, err := svc.ListCategories(context.Background(), "u1"); err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	got, err := svc.GetCategoryByID(context.Background(), "c1", "u1")
	if err != nil || got.Name != "Work" {
		t.Fatalf("GetCategoryByID: got=%v err=%v", got, err)
	}
}

func TestUpdateCategory_InvalidNameRejected(t *testing.T) {
	store := &mockCategoryStore{}
	svc := categories.NewService(store)

	_, err := svc.UpdateCategory(context.Background(), "c1", "u1", "", "#fff")
	if !errors.Is(err, categories.ErrInvalidCategoryName) {
		t.Fatalf("expected ErrInvalidCategoryName, got %v", err)
	}
}

func TestCreateCategory_DefaultColorWhenEmpty(t *testing.T) {
	store := &mockCategoryStore{}
	svc := categories.NewService(store)

	_, err := svc.CreateCategory(context.Background(), "u1", "Work", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.lastColor != "#4F46E5" {
		t.Fatalf("expected default color #4F46E5, got %s", store.lastColor)
	}
}
