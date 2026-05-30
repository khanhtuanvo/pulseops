<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import type { DocumentNode } from 'graphql';

import { apolloClient } from '@/graphql/client';
import {
  ACKNOWLEDGE_INCIDENT,
  CREATE_POSTMORTEM,
  INCIDENT_QUERY,
  INVESTIGATE_INCIDENT,
  RESOLVE_INCIDENT,
} from '@/graphql/operations';
import { useAuthStore } from '@/stores/auth';
import type { Incident } from '@/types';

const route = useRoute();
const authStore = useAuthStore();
const incident = ref<Incident | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);
const statusMessage = ref('');
const postmortemSummary = ref('');
const postmortemTimeline = ref('');
const postmortemActionItems = ref('');
const toast = ref<string | null>(null);

const canAct = computed(() => !authStore.isViewer && incident.value);
const actionItems = computed(() =>
  postmortemActionItems.value
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean),
);

async function loadIncident() {
  loading.value = true;
  error.value = null;
  try {
    const { data } = await apolloClient.query({
      query: INCIDENT_QUERY,
      variables: { id: route.params.id },
      fetchPolicy: 'network-only',
    });
    incident.value = data.incident;
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to load incident';
  } finally {
    loading.value = false;
  }
}

async function mutateIncident(mutation: DocumentNode, variables: Record<string, unknown>, field: string) {
  if (!window.confirm('Apply this incident update?')) {
    return;
  }
  const previous = incident.value;
  try {
    const { data } = await apolloClient.mutate({ mutation, variables });
    incident.value = data[field];
    toast.value = 'Incident updated';
  } catch (err) {
    incident.value = previous;
    error.value = err instanceof Error ? err.message : 'Update failed';
  }
}

async function createPostmortem() {
  if (!incident.value) return;
  await apolloClient.mutate({
    mutation: CREATE_POSTMORTEM,
    variables: {
      incidentId: incident.value.id,
      summary: postmortemSummary.value,
      timeline: postmortemTimeline.value,
      actionItems: actionItems.value,
    },
  });
  toast.value = 'Postmortem saved';
}

onMounted(loadIncident);
</script>

<template>
  <main class="min-h-screen bg-zinc-950 px-6 py-6 text-zinc-50">
    <section class="mx-auto grid max-w-5xl gap-6">
      <RouterLink class="text-sm text-cyan-300 hover:text-cyan-200" to="/dashboard">Back to dashboard</RouterLink>
      <p v-if="loading" class="text-sm text-zinc-400">Loading incident...</p>
      <p v-else-if="error" class="rounded-md border border-red-500/40 bg-red-500/10 p-3 text-sm text-red-200">{{ error }}</p>

      <template v-if="incident">
        <header class="rounded-md border border-zinc-800 bg-zinc-950 p-5">
          <div class="flex flex-wrap items-center gap-2">
            <span class="rounded-full bg-red-500/15 px-2 py-1 text-xs font-semibold text-red-200">{{ incident.severity }}</span>
            <span class="rounded-full bg-zinc-800 px-2 py-1 text-xs text-zinc-300">{{ incident.status }}</span>
          </div>
          <h1 class="mt-4 text-2xl font-semibold">{{ incident.title }}</h1>
          <p class="mt-2 text-sm text-zinc-400">{{ incident.fingerprint }}</p>
        </header>

        <section class="grid gap-4 md:grid-cols-4">
          <div class="rounded-md border border-zinc-800 p-4">
            <p class="text-xs text-zinc-500">Triggered</p>
            <p class="mt-2 text-sm">{{ new Date(incident.triggeredAt).toLocaleString() }}</p>
          </div>
          <div class="rounded-md border border-zinc-800 p-4">
            <p class="text-xs text-zinc-500">Acknowledged</p>
            <p class="mt-2 text-sm">{{ incident.acknowledgedAt ? new Date(incident.acknowledgedAt).toLocaleString() : 'Pending' }}</p>
          </div>
          <div class="rounded-md border border-zinc-800 p-4">
            <p class="text-xs text-zinc-500">Investigating</p>
            <p class="mt-2 text-sm">{{ incident.status === 'INVESTIGATING' ? 'Active' : 'Not active' }}</p>
          </div>
          <div class="rounded-md border border-zinc-800 p-4">
            <p class="text-xs text-zinc-500">Resolved</p>
            <p class="mt-2 text-sm">{{ incident.resolvedAt ? new Date(incident.resolvedAt).toLocaleString() : 'Pending' }}</p>
          </div>
        </section>

        <section v-if="canAct" class="rounded-md border border-zinc-800 p-4">
          <label class="text-sm font-medium" for="status-message">Status message</label>
          <textarea id="status-message" v-model="statusMessage" class="mt-2 min-h-24 w-full rounded-md border border-zinc-800 bg-zinc-950 p-3 text-sm" />
          <div class="mt-3 flex flex-wrap gap-2">
            <button
              v-if="incident.status === 'TRIGGERED'"
              class="rounded-md bg-cyan-400 px-3 py-2 text-sm font-semibold text-zinc-950"
              type="button"
              @click="mutateIncident(ACKNOWLEDGE_INCIDENT, { id: incident.id, message: statusMessage }, 'acknowledgeIncident')"
            >
              Acknowledge
            </button>
            <button
              v-if="incident.status === 'ACKNOWLEDGED'"
              class="rounded-md bg-amber-300 px-3 py-2 text-sm font-semibold text-zinc-950"
              type="button"
              @click="mutateIncident(INVESTIGATE_INCIDENT, { id: incident.id, message: statusMessage }, 'investigateIncident')"
            >
              Investigate
            </button>
            <button
              v-if="incident.status === 'ACKNOWLEDGED' || incident.status === 'INVESTIGATING'"
              class="rounded-md bg-emerald-400 px-3 py-2 text-sm font-semibold text-zinc-950"
              type="button"
              @click="mutateIncident(RESOLVE_INCIDENT, { id: incident.id, summary: statusMessage }, 'resolveIncident')"
            >
              Resolve
            </button>
          </div>
        </section>

        <section class="rounded-md border border-zinc-800 p-4">
          <h2 class="text-sm font-semibold">Alerts</h2>
          <div class="mt-3 grid gap-2">
            <article v-for="alert in incident.alerts" :key="alert.id" class="rounded-md bg-zinc-900 p-3 text-sm">
              <p class="font-medium">{{ alert.source }} / {{ alert.alertName }}</p>
              <p class="text-zinc-500">{{ new Date(alert.receivedAt).toLocaleString() }}</p>
            </article>
            <p v-if="incident.alerts.length === 0" class="text-sm text-zinc-500">No alert payloads are attached yet.</p>
          </div>
        </section>

        <section v-if="incident.runbook" class="rounded-md border border-zinc-800 p-4">
          <h2 class="text-sm font-semibold">{{ incident.runbook.title }}</h2>
          <pre class="mt-3 whitespace-pre-wrap text-sm text-zinc-300">{{ incident.runbook.content }}</pre>
        </section>

        <section v-if="incident.status === 'RESOLVED'" class="rounded-md border border-zinc-800 p-4">
          <h2 class="text-sm font-semibold">Postmortem</h2>
          <input v-model="postmortemSummary" class="mt-3 w-full rounded-md border border-zinc-800 bg-zinc-950 p-3 text-sm" placeholder="Summary" />
          <textarea v-model="postmortemTimeline" class="mt-3 min-h-24 w-full rounded-md border border-zinc-800 bg-zinc-950 p-3 text-sm" placeholder="Timeline" />
          <textarea v-model="postmortemActionItems" class="mt-3 min-h-24 w-full rounded-md border border-zinc-800 bg-zinc-950 p-3 text-sm" placeholder="Action items, one per line" />
          <button class="mt-3 rounded-md bg-cyan-400 px-3 py-2 text-sm font-semibold text-zinc-950" type="button" @click="createPostmortem">
            Save postmortem
          </button>
        </section>
      </template>
      <p v-if="toast" class="fixed bottom-6 right-6 rounded-md bg-emerald-400 px-4 py-2 text-sm font-semibold text-zinc-950">{{ toast }}</p>
    </section>
  </main>
</template>
