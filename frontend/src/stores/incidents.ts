import { ref } from 'vue';
import { defineStore } from 'pinia';

import { apolloClient } from '@/graphql/client';
import {
  ACKNOWLEDGE_INCIDENT,
  INCIDENTS_QUERY,
  RESOLVE_INCIDENT,
} from '@/graphql/operations';
import type { Incident, IncidentEvent, IncidentFilters } from '@/types';

export const useIncidentsStore = defineStore('incidents', () => {
  const incidents = ref<Incident[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  function replaceIncident(incident: Incident) {
    const index = incidents.value.findIndex((item) => item.id === incident.id);
    if (index >= 0) {
      incidents.value[index] = incident;
    } else {
      incidents.value.unshift(incident);
    }
  }

  async function loadIncidents(teamId: string, filters: IncidentFilters = {}) {
    loading.value = true;
    error.value = null;
    try {
      const { data } = await apolloClient.query({
        query: INCIDENTS_QUERY,
        variables: { teamId, ...filters },
        fetchPolicy: 'network-only',
      });
      incidents.value = data.incidents.items;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Unable to load incidents';
    } finally {
      loading.value = false;
    }
  }

  async function acknowledgeIncident(id: string, message?: string) {
    const { data } = await apolloClient.mutate({
      mutation: ACKNOWLEDGE_INCIDENT,
      variables: { id, message },
    });
    replaceIncident(data.acknowledgeIncident);
  }

  async function resolveIncident(id: string, summary: string) {
    const { data } = await apolloClient.mutate({
      mutation: RESOLVE_INCIDENT,
      variables: { id, summary },
    });
    replaceIncident(data.resolveIncident);
  }

  function addLiveIncident(event: IncidentEvent) {
    if (event.type === 'INCIDENT_CREATED') {
      incidents.value = [
        event.incident,
        ...incidents.value.filter((incident) => incident.id !== event.incident.id),
      ];
      return;
    }
    replaceIncident(event.incident);
  }

  return {
    incidents,
    loading,
    error,
    loadIncidents,
    acknowledgeIncident,
    resolveIncident,
    addLiveIncident,
  };
});
