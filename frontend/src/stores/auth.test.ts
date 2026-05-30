import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

import { useAuthStore } from './auth';

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.restoreAllMocks();
  });

  it('hydrates the current user from /auth/me', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          id: 'user-1',
          email: 'owner@example.com',
          name: 'Owner',
          teamId: 'team-1',
          role: 'OWNER',
          googleSubject: 'subject-1',
          createdAt: new Date().toISOString(),
        }),
      }),
    );

    const store = useAuthStore();
    await store.fetchMe();

    expect(store.checked).toBe(true);
    expect(store.user?.email).toBe('owner@example.com');
    expect(store.isOwner).toBe(true);
  });

  it('marks auth checked and clears user on 401', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false }));

    const store = useAuthStore();
    await store.fetchMe();

    expect(store.checked).toBe(true);
    expect(store.user).toBeNull();
  });
});
