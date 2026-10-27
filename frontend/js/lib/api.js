/**
 * Injectable API client — pass a fetch double in unit tests instead of hitting the network.
 */

/**
 * @param {typeof fetch} [fetchImpl]
 */
export function createApiClient(fetchImpl = globalThis.fetch) {
  return {
    /**
     * @param {string} endpoint
     * @param {RequestInit} [options]
     */
    async request(endpoint, options = {}) {
      const config = {
        credentials: 'same-origin',
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...(options.headers || {}),
        },
      };

      const response = await fetchImpl(endpoint, config);
      if (response.status === 204) {
        return null;
      }

      const data = await response.json();
      if (!response.ok) {
        const error = new Error(data.error || 'An error occurred');
        error.status = response.status;
        error.code = data.code;
        throw error;
      }
      return data;
    },
  };
}
