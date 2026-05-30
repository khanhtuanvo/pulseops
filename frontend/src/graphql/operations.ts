import { gql } from '@apollo/client/core';

export const USER_FIELDS = gql`
  fragment UserFields on User {
    id
    email
    name
    avatarUrl
    teamId
    role
    googleSubject
    createdAt
  }
`;

export const INCIDENT_FIELDS = gql`
  fragment IncidentFields on Incident {
    id
    title
    status
    severity
    teamId
    fingerprint
    alertCount
    triggeredAt
    acknowledgedAt
    resolvedAt
    escalated
    escalatedAt
    statusMessage
    mttr
    runbook {
      id
      teamId
      title
      content
      tags
      updatedAt
    }
    alerts {
      id
      incidentId
      source
      alertName
      severity
      environment
      payload
      fingerprint
      receivedAt
    }
  }
`;

export const INCIDENTS_QUERY = gql`
  ${INCIDENT_FIELDS}
  query Incidents($teamId: ID, $status: IncidentStatus, $severity: Severity, $limit: Int, $offset: Int) {
    incidents(teamId: $teamId, status: $status, severity: $severity, limit: $limit, offset: $offset) {
      items {
        ...IncidentFields
      }
      totalCount
      hasMore
    }
  }
`;

export const INCIDENT_QUERY = gql`
  ${INCIDENT_FIELDS}
  query Incident($id: ID!) {
    incident(id: $id) {
      ...IncidentFields
    }
  }
`;

export const ACKNOWLEDGE_INCIDENT = gql`
  ${INCIDENT_FIELDS}
  mutation AcknowledgeIncident($id: ID!, $message: String) {
    acknowledgeIncident(id: $id, message: $message) {
      ...IncidentFields
    }
  }
`;

export const INVESTIGATE_INCIDENT = gql`
  ${INCIDENT_FIELDS}
  mutation InvestigateIncident($id: ID!, $message: String) {
    investigateIncident(id: $id, message: $message) {
      ...IncidentFields
    }
  }
`;

export const RESOLVE_INCIDENT = gql`
  ${INCIDENT_FIELDS}
  mutation ResolveIncident($id: ID!, $summary: String) {
    resolveIncident(id: $id, summary: $summary) {
      ...IncidentFields
    }
  }
`;

export const INCIDENT_FEED = gql`
  ${INCIDENT_FIELDS}
  subscription IncidentFeed($teamId: ID!) {
    incidentFeed(teamId: $teamId) {
      type
      occurredAt
      incident {
        ...IncidentFields
      }
    }
  }
`;

export const ANALYTICS_QUERY = gql`
  query Analytics($teamId: ID, $from: Time, $to: Time) {
    analytics(teamId: $teamId, from: $from, to: $to) {
      mttrSeconds
      mttaSeconds
      totalCount
      byDay {
        date
        count
      }
      bySeverity {
        severity
        count
      }
    }
  }
`;

export const CREATE_POSTMORTEM = gql`
  mutation CreatePostmortem($incidentId: ID!, $summary: String!, $timeline: String!, $actionItems: [String!]!) {
    createPostmortem(incidentId: $incidentId, summary: $summary, timeline: $timeline, actionItems: $actionItems) {
      id
      incidentId
      authorId
      summary
      timeline
      actionItems
      createdAt
    }
  }
`;
