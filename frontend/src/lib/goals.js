/**
 * Weekly goal helpers for the dashboard widget.
 * TP5 demo: shipped without tests first so the coverage gate can fire.
 */

export function goalProgress(weekSeconds, goalSeconds) {
  if (!goalSeconds || goalSeconds <= 0) return 0;
  if (!weekSeconds || weekSeconds <= 0) return 0;
  const pct = Math.floor((weekSeconds / goalSeconds) * 100);
  if (pct > 100) return 100;
  if (pct < 0) return 0;
  return pct;
}

export function remainingToGoal(weekSeconds, goalSeconds) {
  if (!goalSeconds || goalSeconds <= 0) return 0;
  const left = goalSeconds - (weekSeconds || 0);
  return left < 0 ? 0 : left;
}

export function classifyGoalStatus(progressPercent) {
  if (progressPercent <= 0) return 'not_started';
  if (progressPercent < 50) return 'behind';
  if (progressPercent < 100) return 'on_track';
  return 'completed';
}

export function formatGoalSummary(weekSeconds, goalSeconds) {
  const progress = goalProgress(weekSeconds, goalSeconds);
  const remaining = remainingToGoal(weekSeconds, goalSeconds);
  const status = classifyGoalStatus(progress);
  return {
    progress,
    remaining,
    status,
    label: `${progress}% · ${status}`,
  };
}

export function isGoalReached(weekSeconds, goalSeconds) {
  if (!goalSeconds || goalSeconds <= 0) return false;
  return (weekSeconds || 0) >= goalSeconds;
}

export function nextMilestone(progressPercent) {
  if (progressPercent < 25) return 25;
  if (progressPercent < 50) return 50;
  if (progressPercent < 75) return 75;
  if (progressPercent < 100) return 100;
  return null;
}
