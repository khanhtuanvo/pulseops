<script setup lang="ts">
import { useAuthStore } from '@/stores/auth';

const authStore = useAuthStore();
</script>

<template>
  <main class="min-h-screen bg-zinc-950 px-6 py-6 text-zinc-50">
    <section class="mx-auto max-w-5xl rounded-md border border-zinc-800 p-5">
      <RouterLink
        class="text-sm text-cyan-300"
        to="/dashboard"
      >
        Back to dashboard
      </RouterLink>
      <h1 class="mt-4 text-2xl font-semibold">
        Team
      </h1>
      <p class="mt-2 text-sm text-zinc-400">
        {{ authStore.user?.teamId }}
      </p>
      <RouterLink
        v-if="authStore.isOwner"
        class="mt-4 inline-block rounded-md bg-cyan-400 px-3 py-2 text-sm font-semibold text-zinc-950"
        to="/team/settings"
      >
        Team settings
      </RouterLink>
    </section>
  </main>
</template>
