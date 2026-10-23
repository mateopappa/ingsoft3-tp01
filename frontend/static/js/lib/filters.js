/**
 * Builds query-string params for the activities list from UI filter state.
 */

/**
 * @param {{ search?: string, categoryId?: string, fromDate?: string, toDate?: string, favoriteOnly?: boolean }} filters
 * @returns {URLSearchParams}
 */
export function buildActivityQueryParams(filters = {}) {
  const params = new URLSearchParams();
  const search = (filters.search ?? '').trim();
  if (search) params.set('search', search);
  if (filters.categoryId) params.set('category_id', filters.categoryId);
  if (filters.fromDate) params.set('from_date', filters.fromDate);
  if (filters.toDate) params.set('to_date', filters.toDate);
  if (filters.favoriteOnly) params.set('favorite', 'true');
  return params;
}
