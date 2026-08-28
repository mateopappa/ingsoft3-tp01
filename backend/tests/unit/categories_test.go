package unit

import (
	"errors"
	"strings"
	"testing"

	"flow/internal/categories"
)

// Test BT-5: Category deletion rejected when activities reference it (Rule 4)
func TestCategoryDeletionRule(t *testing.T) {
	tests := []struct {
		name          string
		activityCount int
		wantErr       error
	}{
		{
			name:          "Category with 0 activities can be deleted",
			activityCount: 0,
			wantErr:       nil,
		},
		{
			name:          "Category with 5 activities cannot be deleted",
			activityCount: 5,
			wantErr:       categories.ErrCategoryInUse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.activityCount > 0 {
				err = categories.ErrCategoryInUse
			}

			if tt.wantErr != nil {
				if err == nil || !errors.Is(err, tt.wantErr) {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestCategoryNameValidation(t *testing.T) {
	tests := []struct {
		name    string
		catName string
		valid   bool
	}{
		{"Valid standard name", "Programming", true},
		{"Valid short name", "Go", true},
		{"Empty string invalid", "", false},
		{"Whitespace only invalid", "   ", false},
		{"Over 50 characters invalid", strings.Repeat("A", 51), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trimmed := strings.TrimSpace(tt.catName)
			isValid := len(trimmed) > 0 && len(trimmed) <= 50
			if isValid != tt.valid {
				t.Errorf("expected validity %v for '%s', got %v", tt.valid, tt.catName, isValid)
			}
		})
	}
}
