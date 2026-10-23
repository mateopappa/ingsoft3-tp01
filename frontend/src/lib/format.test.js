import { describe, it, expect } from 'vitest';
import { formatSeconds, pad } from './format.js';

describe('formatSeconds', () => {
  it.each([
    [0, '0m'],
    [null, '0m'],
    [undefined, '0m'],
    [-5, '0m'],
    [45, '0m'],
    [60, '1m'],
    [3599, '59m'],
    [3600, '1h 00m'],
    [5022, '1h 23m'],
    [20520, '5h 42m'],
  ])('formatSeconds(%s) → %s', (seconds, expected) => {
    // Arrange + Act
    const result = formatSeconds(seconds);
    // Assert
    expect(result).toBe(expected);
  });
});

describe('pad', () => {
  it('pads single digit with leading zero', () => {
    expect(pad(5)).toBe('05');
    expect(pad(12)).toBe('12');
  });
});
