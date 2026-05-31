<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { RouterLink } from 'vue-router';

import { useAuthStore } from '@/stores/auth';
import { useIncidentsStore } from '@/stores/incidents';
import type { Incident } from '@/types';

const props = defineProps<{ incident: Incident }>();
const authStore = useAuthStore();
const incidentsStore = useIncidentsStore();
const now = ref(Date.now());
let timer: number | undefined;

const severityClass = computed(() => ({
  CRITICAL: 'bg-red-500/15 text-red-200 ring-red-500/30',
  HIGH: 'bg-orange-500/15 text-orange-200 ring-orange-500/30',
  MEDIUM: 'bg-amber-500/15 text-amber-100 ring-amber-500/30',
  LOW: 'bg-emerald-500/15 text-emerald-100 ring-emerald-500/30',
}[props.incident.severity]));

const age = computed(() => {
  const seconds = Math.max(0, Math.floor((now.value - new Date(props.incident.triggeredAt).getTime()) / 1000));
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m`;
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
});

onMounted(() => {
  timer = window.setInterval(() => {
    now.value = Date.now();
  }, 1000);
});

onUnmounted(() => window.clearInterval(timer));
</script>

<template>
  <article class="rounded-md border border-zinc-800 bg-zinc-950 p-4 transition hover:border-zinc-700">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <div class="flex flex-wrap items-center gap-2">
          <span
            class="rounded-full px-2 py-1 text-xs font-semibold ring-1"
            :class="severityClass"
          >{{ incident.severity }}</span>
          <span class="rounded-full bg-zinc-800 px-2 py-1 text-xs text-zinc-300">{{ incident.status }}</span>
          <span class="text-xs text-zinc-500">{{ age }} ago</span>
        </div>
        <RouterLink
          class="mt-3 block text-base font-semibold text-zinc-50 hover:text-cyan-200"
          :to="`/incidents/${incident.id}`"
        >
          {{ incident.title }}
        </RouterLink>
        <p class="mt-1 text-sm text-zinc-400">
          {{ incident.alertCount }} alert{{ incident.alertCount === 1 ? '' : 's' }}
        </p>
      </div>
      <div
        v-if="!authStore.isViewer"
        class="flex shrink-0 gap-2"
      >
        <button
          v-if="incident.status === 'TRIGGERED'"
          class="rounded-md bg-cyan-400 px-3 py-2 text-xs font-semibold text-zinc-950 hover:bg-cyan-300"
          type="button"
          @click="incidentsStore.acknowledgeIncident(incident.id)"
        >
          Ack
        </button>
        <button
          v-if="incident.status === 'ACKNOWLEDGED'"
          class="rounded-md bg-emerald-400 px-3 py-2 text-xs font-semibold text-zinc-950 hover:bg-emerald-300"
          type="button"
          @click="incidentsStore.resolveIncident(incident.id, 'Resolved from dashboard')"
        >
          Resolve
        </button>
      </div>
    </div>
  </article>
</template>
