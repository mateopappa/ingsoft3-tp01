/**
 * Client-side validation rules mirroring the backend business rules.
 */

/**
 * @returns {{ ok: true } | { ok: false, error: string }}
 */
export function validateAuthInput(email, password) {
  const trimmed = (email ?? '').trim();
  if (!trimmed) {
    return { ok: false, error: 'Email is required' };
  }
  if (!trimmed.includes('@')) {
    return { ok: false, error: 'Email must contain @' };
  }
  if (!password || password.length < 8) {
    return { ok: false, error: 'Password must be at least 8 characters long' };
  }
  return { ok: true };
}

/**
 * @returns {{ ok: true, durationSeconds: number } | { ok: false, error: string }}
 */
export function validateActivityForm({ description, categoryId, activityDate, hours, minutes }) {
  const desc = (description ?? '').trim();
  if (!desc) {
    return { ok: false, error: 'Description is required' };
  }
  if (!categoryId) {
    return { ok: false, error: 'Category is required' };
  }
  if (!activityDate) {
    return { ok: false, error: 'Activity date is required' };
  }
  const h = Number(hours) || 0;
  const m = Number(minutes) || 0;
  const durationSeconds = h * 3600 + m * 60;
  if (durationSeconds <= 0) {
    return { ok: false, error: 'Duration must be greater than zero' };
  }
  return { ok: true, durationSeconds };
}
