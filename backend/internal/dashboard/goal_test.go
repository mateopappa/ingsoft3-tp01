package dashboard_test

import (
	"testing"

	"flow/internal/dashboard"
)

func TestGoalProgress(t *testing.T) {
	cases := []struct {
		name         string
		week, goal   int
		wantProgress int
	}{
		{"zero goal returns 0", 3600, 0, 0},
		{"negative goal returns 0", 3600, -1, 0},
		{"zero week returns 0", 0, 36000, 0},
		{"half way", 18000, 36000, 50},
		{"exact goal", 36000, 36000, 100},
		{"over goal clamps to 100", 50000, 36000, 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dashboard.GoalProgress(tc.week, tc.goal)
			if got != tc.wantProgress {
				t.Fatalf("GoalProgress(%d,%d)=%d want %d", tc.week, tc.goal, got, tc.wantProgress)
			}
		})
	}
}

func TestRemainingToGoal(t *testing.T) {
	if got := dashboard.RemainingToGoal(1000, 5000); got != 4000 {
		t.Fatalf("got %d", got)
	}
	if got := dashboard.RemainingToGoal(6000, 5000); got != 0 {
		t.Fatalf("over goal should return 0, got %d", got)
	}
	if got := dashboard.RemainingToGoal(100, 0); got != 0 {
		t.Fatalf("invalid goal should return 0, got %d", got)
	}
}

func TestClassifyGoalStatus(t *testing.T) {
	cases := []struct {
		pct  int
		want string
	}{
		{0, "not_started"},
		{-1, "not_started"},
		{25, "behind"},
		{50, "on_track"},
		{99, "on_track"},
		{100, "completed"},
		{150, "completed"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := dashboard.ClassifyGoalStatus(tc.pct); got != tc.want {
				t.Fatalf("ClassifyGoalStatus(%d)=%q want %q", tc.pct, got, tc.want)
			}
		})
	}
}
