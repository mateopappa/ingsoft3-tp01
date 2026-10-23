import { describe, it, expect } from 'vitest';
import { validateAuthInput, validateActivityForm } from './validate.js';

describe('validateAuthInput', () => {
  it.each([
    ['user@example.com', 'password1', true],
    ['a@b.co', '12345678', true],
  ])('accepts valid credentials (%s)', (email, password, ok) => {
    expect(validateAuthInput(email, password).ok).toBe(ok);
  });

  // Error cases
  it('rejects empty email', () => {
    const result = validateAuthInput('  ', 'password1');
    expect(result.ok).toBe(false);
    expect(result.error).toMatch(/email/i);
  });

  it('rejects email without @', () => {
    const result = validateAuthInput('not-an-email', 'password1');
    expect(result.ok).toBe(false);
    expect(result.error).toMatch(/@/);
  });

  it('rejects password shorter than 8 characters', () => {
    const result = validateAuthInput('user@example.com', '1234567');
    expect(result.ok).toBe(false);
    expect(result.error).toMatch(/8 characters/i);
  });
});

describe('validateActivityForm', () => {
  it('accepts a complete valid form', () => {
    const result = validateActivityForm({
      description: 'Coding',
      categoryId: 'cat-1',
      activityDate: '2026-08-19',
      hours: 1,
      minutes: 30,
    });
    expect(result.ok).toBe(true);
    expect(result.durationSeconds).toBe(5400);
  });

  it('rejects empty description (error case)', () => {
    const result = validateActivityForm({
      description: '   ',
      categoryId: 'cat-1',
      activityDate: '2026-08-19',
      hours: 1,
      minutes: 0,
    });
    expect(result.ok).toBe(false);
    expect(result.error).toMatch(/description/i);
  });

  it('rejects zero duration (error case)', () => {
    const result = validateActivityForm({
      description: 'Idle',
      categoryId: 'cat-1',
      activityDate: '2026-08-19',
      hours: 0,
      minutes: 0,
    });
    expect(result.ok).toBe(false);
    expect(result.error).toMatch(/duration/i);
  });

  it('rejects missing category', () => {
    const result = validateActivityForm({
      description: 'Coding',
      categoryId: '',
      activityDate: '2026-08-19',
      hours: 1,
      minutes: 0,
    });
    expect(result.ok).toBe(false);
    expect(result.error).toMatch(/category/i);
  });
});
