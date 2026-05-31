<script setup lang="ts">
import IncidentCard from '@/components/IncidentCard.vue';
import type { Incident } from '@/types';

defineProps<{ incidents: Incident[] }>();
</script>

<template>
  <div class="space-y-3">
    <TransitionGroup name="incident">
      <IncidentCard
        v-for="incident in incidents"
        :key="incident.id"
        :incident="incident"
      />
    </TransitionGroup>
    <p
      v-if="incidents.length === 0"
      class="rounded-md border border-dashed border-zinc-800 p-8 text-center text-sm text-zinc-500"
    >
      No incidents match the current filters.
    </p>
  </div>
</template>

<style scoped>
.incident-enter-active,
.incident-leave-active {
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.incident-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.incident-leave-to {
  opacity: 0;
}
</style>
