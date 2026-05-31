<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';

import AnalyticsChart from '@/components/AnalyticsChart.vue';
import IncidentFeed from '@/components/IncidentFeed.vue';
import { useIncidentFeed } from '@/composables/useIncidentFeed';
import { useAuthStore } from '@/stores/auth';
import { useIncidentsStore } from '@/stores/incidents';
import type { IncidentStatus, Severity } from '@/types';

const authStore = useAuthStore();
const incidentsStore = useIncidentsStore();
const feedElement = ref<HTMLElement | null>(null);
const statusFilter = ref<IncidentStatus | ''>('');
const severityFilter = ref<Severity | ''>('');
const range = ref<'24h' | '7d' | '30d'>('24h');
const feed = useIncidentFeed(feedElement);

const rangeDates = computed(() => {
  const to = new Date();
  const from = new Date(to);
  if (range.value === '24h') from.setDate(to.getDate() - 1);
  if (range.value === '7d') from.setDate(to.getDate() - 7);
  if (range.value === '30d') from.setDate(to.getDate() - 30);
  return { from, to };
});

const filteredIncidents = computed(() =>
  incidentsStore.incidents.filter((incident) => {
    if (statusFilter.value && incident.status !== statusFilter.value) return false;
    if (severityFilter.value && incident.severity !== severityFilter.value) return false;
    return true;
  }),
);

const triggeredCount = computed(() => incidentsStore.incidents.filter((incident) => incident.status === 'TRIGGERED').length);

onMounted(async () => {
  if (authStore.user) {
    await incidentsStore.loadIncidents(authStore.user.teamId, { limit: 50 });
  }
});
</script>

<template>
  <main class="min-h-screen bg-zinc-950 text-zinc-50">
    <nav class="border-b border-zinc-800 bg-zinc-950/95">
      <div class="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
        <div>
          <p class="text-sm font-semibold text-cyan-300">
            PulseOps
          </p>
          <h1 class="text-xl font-semibold">
            Incident dashboard
          </h1>
        </div>
        <div class="flex items-center gap-4 text-sm">
          <span class="text-zinc-300">{{ authStore.user?.name }}</span>
          <button
            class="rounded-md border border-zinc-700 px-3 py-2 text-zinc-200 hover:bg-zinc-900"
            type="button"
            @click="authStore.logout"
          >
            Logout
          </button>
        </div>
      </div>
    </nav>

    <div class="mx-auto grid max-w-7xl gap-6 px-6 py-6">
      <section class="grid gap-4 md:grid-cols-3">
        <div class="rounded-md border border-zinc-800 bg-zinc-950 p-4">
          <p class="text-xs text-zinc-500">
            Connection
          </p>
          <p class="mt-2 flex items-center gap-2 text-sm">
            <span
              class="size-2 rounded-full"
              :class="feed.isConnected.value ? 'bg-emerald-400' : 'bg-red-400'"
            />
            {{ feed.isConnected.value ? 'Connected' : `Reconnecting (${feed.reconnectCount.value})` }}
          </p>
        </div>
        <div class="rounded-md border border-zinc-800 bg-zinc-950 p-4">
          <p class="text-xs text-zinc-500">
            Triggered
          </p>
          <p class="mt-2 text-2xl font-semibold">
            {{ triggeredCount }}
          </p>
        </div>
        <div class="rounded-md border border-zinc-800 bg-zinc-950 p-4">
          <p class="text-xs text-zinc-500">
            Current on-call
          </p>
          <p class="mt-2 text-sm text-zinc-200">
            {{ authStore.user?.name ?? 'Unknown' }}
          </p>
        </div>
      </section>

      <AnalyticsChart
        v-if="authStore.user"
        :team-id="authStore.user.teamId"
        :from="rangeDates.from"
        :to="rangeDates.to"
        @range-change="range = $event"
      />

      <section class="grid gap-4">
        <div class="flex flex-wrap gap-3">
          <select
            v-model="statusFilter"
            class="rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm"
          >
            <option value="">
              All statuses
            </option>
            <option value="TRIGGERED">
              Triggered
            </option>
            <option value="ACKNOWLEDGED">
              Acknowledged
            </option>
            <option value="INVESTIGATING">
              Investigating
            </option>
            <option value="RESOLVED">
              Resolved
            </option>
          </select>
          <select
            v-model="severityFilter"
            class="rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm"
          >
            <option value="">
              All severities
            </option>
            <option value="CRITICAL">
              Critical
            </option>
            <option value="HIGH">
              High
            </option>
            <option value="MEDIUM">
              Medium
            </option>
            <option value="LOW">
              Low
            </option>
          </select>
        </div>
        <div
          ref="feedElement"
          class="max-h-[58vh] overflow-y-auto pr-1"
        >
          <IncidentFeed :incidents="filteredIncidents" />
        </div>
      </section>
    </div>
  </main>
</template>
