import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useSubscription } from '@vue/apollo-composable';

import { INCIDENT_FEED } from '@/graphql/operations';
import { useAuthStore } from '@/stores/auth';
import { useIncidentsStore } from '@/stores/incidents';
import type { IncidentEvent } from '@/types';

export function useIncidentFeed(feedElement?: { value: HTMLElement | null }) {
  const authStore = useAuthStore();
  const incidentsStore = useIncidentsStore();
  const isConnected = ref(false);
  const isScrollLocked = ref(false);
  const reconnectCount = ref(0);
  const enabled = ref(false);
  const backoffMs = ref(2000);
  let reconnectTimer: number | undefined;

  const variables = computed(() => ({ teamId: authStore.user?.teamId ?? '' }));
  const subscription = useSubscription(INCIDENT_FEED, variables, () => ({
    enabled: enabled.value && Boolean(authStore.user?.teamId),
    fetchPolicy: 'no-cache',
  }));

  function atFeedBottom(element: HTMLElement) {
    return element.scrollHeight - element.scrollTop - element.clientHeight < 50;
  }

  function handleScroll() {
    const element = feedElement?.value;
    if (!element) {
      return;
    }
    isScrollLocked.value = !atFeedBottom(element);
  }

  function scrollToLatest() {
    const element = feedElement?.value;
    if (element && !isScrollLocked.value) {
      element.scrollTo({ top: 0, behavior: 'smooth' });
    }
  }

  function scheduleReconnect() {
    window.clearTimeout(reconnectTimer);
    isConnected.value = false;
    reconnectTimer = window.setTimeout(() => {
      reconnectCount.value += 1;
      enabled.value = false;
      window.setTimeout(() => {
        enabled.value = true;
      }, 0);
      backoffMs.value = Math.min(backoffMs.value * 2, 30000);
    }, backoffMs.value);
  }

  watch(subscription.result, (result) => {
    const event = result?.incidentFeed as IncidentEvent | undefined;
    if (!event) {
      return;
    }
    isConnected.value = true;
    backoffMs.value = 2000;
    incidentsStore.addLiveIncident(event);
    scrollToLatest();
  });

  watch(subscription.error, (err) => {
    if (err) {
      scheduleReconnect();
    }
  });

  onMounted(() => {
    enabled.value = true;
    feedElement?.value?.addEventListener('scroll', handleScroll, { passive: true });
  });

  onUnmounted(() => {
    enabled.value = false;
    subscription.stop?.();
    window.clearTimeout(reconnectTimer);
    feedElement?.value?.removeEventListener('scroll', handleScroll);
  });

  return { isConnected, isScrollLocked, reconnectCount };
}
