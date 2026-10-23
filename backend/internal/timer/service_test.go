package timer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"flow/internal/models"
	"flow/internal/timer"
)

type stubCategoryFinder struct {
	err error
}

func (s *stubCategoryFinder) GetCategoryByID(ctx context.Context, id, userID string) (*models.Category, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &models.Category{ID: id, UserID: userID, Name: "Work"}, nil
}

type mockTimerStore struct {
	active        *models.ActiveTimer
	getErr        error
	createCalls   int
	createErr     error
	stopResult    *models.Activity
	stopErr       error
	discardCalls  int
	lastCreated   *models.ActiveTimer
}

func (m *mockTimerStore) GetActiveTimer(ctx context.Context, userID string) (*models.ActiveTimer, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.active, nil
}
func (m *mockTimerStore) CreateActiveTimer(ctx context.Context, t *models.ActiveTimer) (*models.ActiveTimer, error) {
	m.createCalls++
	m.lastCreated = t
	if m.createErr != nil {
		return nil, m.createErr
	}
	out := *t
	out.ID = "tm-1"
	out.StartedAt = time.Now().UTC()
	return &out, nil
}
func (m *mockTimerStore) DeleteActiveTimer(ctx context.Context, userID string) error {
	m.discardCalls++
	return nil
}
func (m *mockTimerStore) StopTimerTx(ctx context.Context, userID string, note *string, favorite bool) (*models.Activity, error) {
	return m.stopResult, m.stopErr
}

// --- Rule 5: second active timer rejected (mock verifies Create is not called) ---

func TestStartTimer_SecondActiveRejected(t *testing.T) {
	// Arrange
	started := time.Now().UTC().Add(-10 * time.Minute)
	store := &mockTimerStore{
		active: &models.ActiveTimer{
			ID:          "tm-old",
			UserID:      "u1",
			Description: "Already running",
			StartedAt:   started,
		},
	}
	svc := timer.NewService(store, &stubCategoryFinder{})

	// Act
	_, err := svc.StartTimer(context.Background(), "u1", "Second attempt", "cat-1")

	// Assert
	if !errors.Is(err, timer.ErrTimerAlreadyActive) {
		t.Fatalf("expected ErrTimerAlreadyActive, got %v", err)
	}
	if store.createCalls != 0 {
		t.Fatalf("CreateActiveTimer must not be called when a timer is already active")
	}
}

func TestStartTimer_EmptyDescriptionRejected(t *testing.T) {
	store := &mockTimerStore{getErr: timer.ErrNoActiveTimer}
	svc := timer.NewService(store, &stubCategoryFinder{})

	_, err := svc.StartTimer(context.Background(), "u1", "   ", "cat-1")
	if !errors.Is(err, timer.ErrEmptyTimerDescription) {
		t.Fatalf("expected ErrEmptyTimerDescription, got %v", err)
	}
}

func TestStartTimer_InvalidCategoryRejected(t *testing.T) {
	store := &mockTimerStore{getErr: timer.ErrNoActiveTimer}
	svc := timer.NewService(store, &stubCategoryFinder{err: errors.New("missing")})

	_, err := svc.StartTimer(context.Background(), "u1", "Work session", "bad-cat")
	if !errors.Is(err, timer.ErrInvalidTimerCategory) {
		t.Fatalf("expected ErrInvalidTimerCategory, got %v", err)
	}
	if store.createCalls != 0 {
		t.Fatalf("must not create timer for invalid category")
	}
}

func TestStartTimer_CreatesWhenIdle(t *testing.T) {
	store := &mockTimerStore{getErr: timer.ErrNoActiveTimer}
	svc := timer.NewService(store, &stubCategoryFinder{})

	got, err := svc.StartTimer(context.Background(), "u1", "  Focus  ", "cat-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.createCalls != 1 {
		t.Fatalf("expected one create call")
	}
	if got.ElapsedSeconds != 0 {
		t.Fatalf("new timer elapsed must be 0, got %d", got.ElapsedSeconds)
	}
	if store.lastCreated.Description != "Focus" {
		t.Fatalf("description should be trimmed, got %q", store.lastCreated.Description)
	}
}

// --- Rule 6: stop delegates duration calculation to store transaction ---

func TestStopTimer_DelegatesToStore(t *testing.T) {
	expected := &models.Activity{
		ID:              "a1",
		DurationSeconds: 5025,
		Description:     "Focus",
	}
	store := &mockTimerStore{stopResult: expected}
	svc := timer.NewService(store, &stubCategoryFinder{})

	note := "done"
	got, err := svc.StopTimer(context.Background(), "u1", &note, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.DurationSeconds != 5025 {
		t.Fatalf("expected duration 5025, got %d", got.DurationSeconds)
	}
}

func TestDiscardTimer_CallsStore(t *testing.T) {
	store := &mockTimerStore{}
	svc := timer.NewService(store, &stubCategoryFinder{})

	if err := svc.DiscardTimer(context.Background(), "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.discardCalls != 1 {
		t.Fatalf("expected DeleteActiveTimer once, got %d", store.discardCalls)
	}
}

func TestGetActiveTimer_ClampsNegativeElapsed(t *testing.T) {
	// StartedAt in the future → Since is negative → clamp to 0
	store := &mockTimerStore{
		active: &models.ActiveTimer{
			ID:        "tm-1",
			UserID:    "u1",
			StartedAt: time.Now().UTC().Add(5 * time.Minute),
		},
	}
	svc := timer.NewService(store, &stubCategoryFinder{})

	got, err := svc.GetActiveTimer(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ElapsedSeconds != 0 {
		t.Fatalf("expected clamped elapsed 0, got %d", got.ElapsedSeconds)
	}
}
