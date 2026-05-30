import { computed, ref } from 'vue';
import { defineStore } from 'pinia';

import type { User } from '@/types';

const apiBaseUrl = () => import.meta.env.VITE_API_URL ?? '';

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null);
  const checked = ref(false);

  const isAuthenticated = computed(() => user.value !== null);
  const isOwner = computed(() => user.value?.role === 'OWNER');
  const isResponder = computed(() => user.value?.role === 'RESPONDER');
  const isViewer = computed(() => user.value?.role === 'VIEWER');

  async function fetchMe() {
    try {
      const response = await fetch(`${apiBaseUrl()}/auth/me`, { credentials: 'include' });
      if (!response.ok) {
        user.value = null;
        return;
      }
      user.value = (await response.json()) as User;
    } finally {
      checked.value = true;
    }
  }

  async function logout() {
    await fetch(`${apiBaseUrl()}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    });
    user.value = null;
    checked.value = false;
    const { default: router } = await import('@/router');
    await router.push('/login');
  }

  return {
    user,
    checked,
    isAuthenticated,
    isOwner,
    isResponder,
    isViewer,
    fetchMe,
    logout,
  };
});
