import { describe, it, expect } from 'vitest';
import {
  goalProgress,
  remainingToGoal,
  classifyGoalStatus,
  formatGoalSummary,
  isGoalReached,
  nextMilestone,
} from './goals.js';

describe('goalProgress', () => {
  it.each([
    [3600, 0, 0],
    [0, 36000, 0],
    [18000, 36000, 50],
    [36000, 36000, 100],
    [50000, 36000, 100],
  ])('goalProgress(%i, %i) → %i', (week, goal, expected) => {
    expect(goalProgress(week, goal)).toBe(expected);
  });
});

describe('remainingToGoal', () => {
  it('returns seconds left', () => {
    expect(remainingToGoal(1000, 5000)).toBe(4000);
  });

  it('clamps at zero when over goal', () => {
    expect(remainingToGoal(6000, 5000)).toBe(0);
  });

  it('rejects invalid goal (error / edge case)', () => {
    expect(remainingToGoal(100, 0)).toBe(0);
  });
});

describe('classifyGoalStatus', () => {
  it.each([
    [0, 'not_started'],
    [25, 'behind'],
    [50, 'on_track'],
    [100, 'completed'],
  ])('classifyGoalStatus(%i) → %s', (pct, status) => {
    expect(classifyGoalStatus(pct)).toBe(status);
  });
});

describe('formatGoalSummary / isGoalReached / nextMilestone', () => {
  it('builds a summary object', () => {
    const summary = formatGoalSummary(18000, 36000);
    expect(summary.progress).toBe(50);
    expect(summary.remaining).toBe(18000);
    expect(summary.status).toBe('on_track');
    expect(summary.label).toContain('50%');
  });

  it('detects when the goal is reached', () => {
    expect(isGoalReached(36000, 36000)).toBe(true);
    expect(isGoalReached(1000, 36000)).toBe(false);
    expect(isGoalReached(1000, 0)).toBe(false);
  });

  it.each([
    [10, 25],
    [30, 50],
    [60, 75],
    [80, 100],
    [100, null],
  ])('nextMilestone(%i) → %s', (pct, next) => {
    expect(nextMilestone(pct)).toBe(next);
  });
});
