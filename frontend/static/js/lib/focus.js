/**
 * Focus-score helpers for the activity insight widget.
 * TP5: intentionally untested — PR left open and red for the defense.
 */

export function estimateFocusScore(durationSeconds, favorite, note) {
  if (!durationSeconds || durationSeconds <= 0) return 0;
  let score = Math.floor(durationSeconds / 60);
  if (favorite) score += 10;
  if (note && note.length > 0) score += 5;
  if (score > 100) return 100;
  if (score < 0) return 0;
  return score;
}

export function classifyFocusBand(score) {
  if (score <= 0) return 'idle';
  if (score < 30) return 'light';
  if (score < 70) return 'deep';
  return 'flow';
}

export function formatFocusLabel(durationSeconds, favorite, note) {
  const score = estimateFocusScore(durationSeconds, favorite, note);
  const band = classifyFocusBand(score);
  return `${band} (${score})`;
}

export function isDeepFocus(durationSeconds, favorite, note) {
  return estimateFocusScore(durationSeconds, favorite, note) >= 30;
}
