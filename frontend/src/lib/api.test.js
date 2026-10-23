import { describe, it, expect, vi } from 'vitest';
import { createApiClient } from './api.js';

describe('createApiClient (mock fetch)', () => {
  it('returns JSON body on success', async () => {
    // Arrange — mock replaces the real network dependency
    const fetchMock = vi.fn().mockResolvedValue({
      status: 200,
      ok: true,
      json: async () => ({ id: 'a1', description: 'Coding' }),
    });
    const api = createApiClient(fetchMock);

    // Act
    const data = await api.request('/api/activities');

    // Assert
    expect(data.description).toBe('Coding');
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/activities',
      expect.objectContaining({
        headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
      }),
    );
  });

  it('returns null on 204 No Content', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ status: 204, ok: true, json: async () => ({}) });
    const api = createApiClient(fetchMock);

    await expect(api.request('/api/activities/1', { method: 'DELETE' })).resolves.toBeNull();
  });

  it('throws with server error message when response is not ok', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      status: 409,
      ok: false,
      json: async () => ({ error: 'a timer is already active', code: 'TIMER_ACTIVE' }),
    });
    const api = createApiClient(fetchMock);

    await expect(api.request('/api/timer/start', { method: 'POST' })).rejects.toMatchObject({
      message: 'a timer is already active',
      status: 409,
      code: 'TIMER_ACTIVE',
    });
  });
});
