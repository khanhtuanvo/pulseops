<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { Doughnut, Line } from 'vue-chartjs';
import {
  ArcElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
} from 'chart.js';

import { apolloClient } from '@/graphql/client';
import { ANALYTICS_QUERY } from '@/graphql/operations';
import type { Analytics } from '@/types';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Tooltip, Legend);

const props = defineProps<{ teamId: string; from: Date; to: Date }>();
const emit = defineEmits<{ rangeChange: [range: '24h' | '7d' | '30d'] }>();
const analytics = ref<Analytics | null>(null);
const selectedRange = ref<'24h' | '7d' | '30d'>('24h');

async function loadAnalytics() {
  const { data } = await apolloClient.query({
    query: ANALYTICS_QUERY,
    variables: {
      teamId: props.teamId,
      from: props.from.toISOString(),
      to: props.to.toISOString(),
    },
    fetchPolicy: 'network-only',
  });
  analytics.value = data.analytics;
}

const lineData = computed(() => ({
  labels: analytics.value?.byDay.map((stat) => new Date(stat.date).toLocaleDateString()) ?? [],
  datasets: [
    {
      label: 'Incidents',
      data: analytics.value?.byDay.map((stat) => stat.count) ?? [],
      borderColor: '#22d3ee',
      backgroundColor: 'rgba(34, 211, 238, 0.18)',
      tension: 0.3,
    },
  ],
}));

const doughnutData = computed(() => ({
  labels: analytics.value?.bySeverity.map((stat) => stat.severity) ?? [],
  datasets: [
    {
      data: analytics.value?.bySeverity.map((stat) => stat.count) ?? [],
      backgroundColor: ['#ef4444', '#f97316', '#f59e0b', '#10b981'],
    },
  ],
}));

function setRange(range: '24h' | '7d' | '30d') {
  selectedRange.value = range;
  emit('rangeChange', range);
}

watch(() => [props.teamId, props.from, props.to], loadAnalytics, { deep: true });
onMounted(loadAnalytics);
</script>

<template>
  <section class="grid gap-4 xl:grid-cols-[1.4fr_1fr]">
    <div class="rounded-md border border-zinc-800 bg-zinc-950 p-4">
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-sm font-semibold text-zinc-100">Incident trend</h2>
          <p class="text-xs text-zinc-500">MTTR {{ Math.round((analytics?.mttrSeconds ?? 0) / 60) }}m</p>
        </div>
        <div class="flex rounded-md border border-zinc-800 p-1 text-xs">
          <button
            v-for="range in ['24h', '7d', '30d']"
            :key="range"
            class="rounded px-2 py-1"
            :class="selectedRange === range ? 'bg-cyan-400 text-zinc-950' : 'text-zinc-400'"
            type="button"
            @click="setRange(range as '24h' | '7d' | '30d')"
          >
            {{ range }}
          </button>
        </div>
      </div>
      <Line :data="lineData" :options="{ responsive: true, maintainAspectRatio: false }" class="h-64" />
    </div>
    <div class="rounded-md border border-zinc-800 bg-zinc-950 p-4">
      <h2 class="mb-4 text-sm font-semibold text-zinc-100">Severity mix</h2>
      <Doughnut :data="doughnutData" :options="{ responsive: true, maintainAspectRatio: false }" class="h-64" />
    </div>
  </section>
</template>
