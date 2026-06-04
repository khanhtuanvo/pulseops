export type Role = 'OWNER' | 'RESPONDER' | 'VIEWER';
export type IncidentStatus =
  | 'TRIGGERED'
  | 'ACKNOWLEDGED'
  | 'INVESTIGATING'
  | 'RESOLVED'
  | 'CLOSED';
export type Severity = 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW';
export type IncidentEventType =
  | 'INCIDENT_CREATED'
  | 'INCIDENT_ACKNOWLEDGED'
  | 'INCIDENT_INVESTIGATING'
  | 'INCIDENT_RESOLVED'
  | 'INCIDENT_ESCALATED'
  | 'ALERT_ATTACHED';

export interface User {
  id: string;
  email: string;
  name: string;
  avatarUrl?: string | null;
  teamId: string;
  role: Role;
  googleSubject: string;
  createdAt: string;
}

export interface Alert {
  id: string;
  incidentId: string;
  source: string;
  alertName: string;
  severity: Severity;
  environment: string;
  payload: Record<string, unknown>;
  fingerprint: string;
  receivedAt: string;
}

export interface Runbook {
  id: string;
  teamId: string;
  title: string;
  content: string;
  tags: string[];
  updatedAt: string;
}

export interface Incident {
  id: string;
  title: string;
  status: IncidentStatus;
  severity: Severity;
  teamId: string;
  fingerprint: string;
  alertCount: number;
  triggeredAt: string;
  acknowledgedAt?: string | null;
  acknowledgedBy?: User | null;
  resolvedAt?: string | null;
  resolvedBy?: User | null;
  escalated: boolean;
  escalatedAt?: string | null;
  assignee?: User | null;
  runbook?: Runbook | null;
  statusMessage?: string | null;
  mttr?: number | null;
  alerts: Alert[];
}

export interface IncidentEvent {
  type: IncidentEventType;
  incident: Incident;
  actor?: User | null;
  occurredAt: string;
}

export interface IncidentFilters {
  status?: IncidentStatus;
  severity?: Severity;
  limit?: number;
  offset?: number;
}

export interface ScheduleOverride {
  id: string;
  user: Pick<User, 'id' | 'name'>;
  startsAt: string;
  endsAt: string;
  reason: string;
}

export interface OnCallSchedule {
  id: string;
  intervalDays: number;
  cycleStart: string;
  rotation: Array<Pick<User, 'id' | 'name' | 'email' | 'role'>>;
  currentOnCall: Pick<User, 'id' | 'name'>;
  overrides: ScheduleOverride[];
}

export interface Team {
  id: string;
  name: string;
  apiKeyHint?: string | null;
  members: User[];
  onCallSchedule?: OnCallSchedule | null;
}

export interface Analytics {
  mttrSeconds: number;
  mttaSeconds: number;
  totalCount: number;
  byDay: Array<{ date: string; count: number }>;
  bySeverity: Array<{ severity: Severity; count: number }>;
}
