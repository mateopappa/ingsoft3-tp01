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

export function suggestBreakMinutes(durationSeconds) {
  if (!durationSeconds || durationSeconds < 25 * 60) return 0;
  if (durationSeconds < 50 * 60) return 5;
  if (durationSeconds < 90 * 60) return 10;
  return 15;
}

export function isLongSession(durationSeconds) {
  return (durationSeconds || 0) >= 90 * 60;
}

export function normalizeNote(note) {
  const trimmed = (note ?? '').trim();
  if (!trimmed) return '';
  if (trimmed.length <= 280) return trimmed;
  return trimmed.slice(0, 280);
}

export function mergeFavoritePreference(current, incoming) {
  return Boolean(current) || Boolean(incoming);
}

export function buildFocusInsight(durationSeconds, favorite, note) {
  const score = estimateFocusScore(durationSeconds, favorite, note);
  return {
    score,
    band: classifyFocusBand(score),
    label: formatFocusLabel(durationSeconds, favorite, note),
    breakMinutes: suggestBreakMinutes(durationSeconds),
    longSession: isLongSession(durationSeconds),
    note: normalizeNote(note),
  };
}
