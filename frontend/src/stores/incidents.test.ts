import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

import { useIncidentsStore } from './incidents';
import type { Incident, IncidentEvent } from '@/types';

function incident(id: string, status: Incident['status']): Incident {
  return {
    id,
    title: `Incident ${id}`,
    status,
    severity: 'HIGH',
    teamId: 'team-1',
    fingerprint: `fp-${id}`,
    alertCount: 1,
    triggeredAt: new Date().toISOString(),
    escalated: false,
    alerts: [],
  };
}

function event(type: IncidentEvent['type'], item: Incident): IncidentEvent {
  return {
    type,
    incident: item,
    occurredAt: new Date().toISOString(),
  };
}

describe('incidents store live updates', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('prepends created incidents', () => {
    const store = useIncidentsStore();
    store.incidents = [incident('old', 'TRIGGERED')];

    store.addLiveIncident(event('INCIDENT_CREATED', incident('new', 'TRIGGERED')));

    expect(store.incidents.map((item) => item.id)).toEqual(['new', 'old']);
  });

  it('updates existing incidents for status events', () => {
    const store = useIncidentsStore();
    store.incidents = [incident('a', 'TRIGGERED')];

    store.addLiveIncident(event('INCIDENT_ACKNOWLEDGED', incident('a', 'ACKNOWLEDGED')));

    expect(store.incidents).toHaveLength(1);
    expect(store.incidents[0].status).toBe('ACKNOWLEDGED');
  });
});
