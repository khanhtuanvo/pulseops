<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { RouterLink, useRoute } from 'vue-router';

import { useAuth } from '@/composables/useAuth';

const route = useRoute();
const { handleCallback } = useAuth();
const error = ref<string | null>(null);
const loading = ref(true);

onMounted(async () => {
  const code = route.query.code;
  const state = route.query.state;
  if (typeof code !== 'string' || typeof state !== 'string') {
    error.value = 'Invalid callback - missing code or state';
    loading.value = false;
    return;
  }

  try {
    await handleCallback(code, state);
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Authentication failed';
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <main class="grid min-h-screen place-items-center bg-zinc-950 px-6 text-zinc-50">
    <section class="w-full max-w-sm text-center">
      <div
        v-if="loading"
        class="mx-auto size-10 animate-spin rounded-full border-2 border-cyan-300 border-t-transparent"
      />
      <p
        v-if="loading"
        class="mt-4 text-sm text-zinc-300"
      >
        Completing sign in...
      </p>
      <template v-else-if="error">
        <h1 class="text-2xl font-semibold">
          Could not sign in
        </h1>
        <p class="mt-3 text-sm text-red-200">
          {{ error }}
        </p>
        <RouterLink
          class="mt-6 inline-flex rounded-md bg-white px-4 py-2 text-sm font-semibold text-zinc-950"
          to="/login"
        >
          Try again
        </RouterLink>
      </template>
    </section>
  </main>
</template>
