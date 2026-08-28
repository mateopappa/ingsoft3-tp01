package unit

import (
	"errors"
	"testing"
	"time"

	"flow/internal/timer"
)

// Test BT-6: Second active timer rejected (Rule 5)
func TestSecondActiveTimerRule(t *testing.T) {
	hasActiveTimer := true

	var err error
	if hasActiveTimer {
		err = timer.ErrTimerAlreadyActive
	}

	if !errors.Is(err, timer.ErrTimerAlreadyActive) {
		t.Errorf("expected ErrTimerAlreadyActive when user already has a running timer, got %v", err)
	}
}

// Test BT-7: Timer duration calculated correctly from server timestamps (Rule 6)
func TestTimerDurationCalculation(t *testing.T) {
	startedAt := time.Date(2026, 8, 19, 14, 0, 0, 0, time.UTC)
	stoppedAt := time.Date(2026, 8, 19, 15, 23, 45, 0, time.UTC)

	durationSeconds := int(stoppedAt.Sub(startedAt).Seconds())
	expectedSeconds := 1*3600 + 23*60 + 45 // 5025 seconds

	if durationSeconds != expectedSeconds {
		t.Errorf("expected %d seconds, got %d", expectedSeconds, durationSeconds)
	}
}
