import { describe, it, expect } from 'vitest';
import { buildActivityQueryParams } from './filters.js';

describe('buildActivityQueryParams', () => {
  it.each([
    [{}, ''],
    [{ search: '  go  ' }, 'search=go'],
    [{ categoryId: 'cat-1' }, 'category_id=cat-1'],
    [{ favoriteOnly: true }, 'favorite=true'],
    [{ favoriteOnly: false }, ''],
  ])('maps %j → %s', (filters, expected) => {
    expect(buildActivityQueryParams(filters).toString()).toBe(expected);
  });

  it('combines multiple filters', () => {
    const params = buildActivityQueryParams({
      search: 'test',
      categoryId: 'c1',
      fromDate: '2026-01-01',
      toDate: '2026-01-31',
      favoriteOnly: true,
    });
    expect(params.get('search')).toBe('test');
    expect(params.get('category_id')).toBe('c1');
    expect(params.get('from_date')).toBe('2026-01-01');
    expect(params.get('to_date')).toBe('2026-01-31');
    expect(params.get('favorite')).toBe('true');
  });
});
