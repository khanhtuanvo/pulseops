<script setup lang="ts">
import { useRoute } from 'vue-router';

import { useAuth } from '@/composables/useAuth';

const route = useRoute();
const { login } = useAuth();
</script>

<template>
  <main class="grid min-h-screen place-items-center bg-zinc-950 px-6 text-zinc-50">
    <section class="w-full max-w-sm">
      <p class="text-sm font-semibold text-cyan-300">
        PulseOps
      </p>
      <h1 class="mt-3 text-3xl font-semibold">
        Sign in
      </h1>
      <p
        v-if="route.query.error"
        class="mt-4 rounded-md border border-red-500/40 bg-red-500/10 px-3 py-2 text-sm text-red-200"
      >
        {{ route.query.error }}
      </p>
      <button
        class="mt-8 flex w-full items-center justify-center gap-3 rounded-md bg-white px-4 py-3 text-sm font-semibold text-zinc-950 shadow-sm hover:bg-zinc-100"
        type="button"
        @click="login"
      >
        <span class="grid size-5 place-items-center rounded-full bg-blue-600 text-xs text-white">G</span>
        Sign in with Google
      </button>
    </section>
  </main>
</template>
