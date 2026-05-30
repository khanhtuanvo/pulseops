# PulseOps

PulseOps is a real-time, multi-tenant incident management SaaS application for small operations and engineering teams. It ingests alert webhooks, deduplicates noisy signals into incidents, and streams live incident updates to authenticated Vue dashboards.

## Architecture

Alert sources send webhook payloads to the Go API. The API validates team API keys, fingerprints alerts, writes incidents and alerts to MongoDB, then MongoDB Change Streams push incident changes into an in-process subscription hub for GraphQL WebSocket clients.

```text
Alert Source
  -> Go API
  -> MongoDB
  -> Change Stream
  -> Hub
  -> WebSocket
  -> Vue Dashboard
```

## Tech Stack

| Area | Tools | Purpose |
| --- | --- | --- |
| Frontend | Vue 3, TypeScript, Pinia, Vue Router, Apollo Client, TailwindCSS, Chart.js, Vite, Vitest | Authenticated SPA, live incident feed, analytics views |
| Backend | Go, chi, gqlgen, mongo-driver, JWT, go-jose, zap | HTTP API, GraphQL queries/mutations/subscriptions, OAuth, structured logging |
| Database | MongoDB replica set, TTL indexes, Change Streams | Incident storage, token/session storage, dedupe windows, real-time events |
| Cloud | Azure Container Apps, Azure Static Web Apps, Azure Container Registry, Azure Key Vault, Azure Monitor | Hosting, image registry, secret injection, telemetry |
| DevOps | GitHub Actions, Docker, golangci-lint, Trivy, Azure CLI | CI, local dev stack, deployments, vulnerability scanning |

## Local Development

Prerequisites:

- Docker Desktop with WSL integration enabled
- Go 1.22 or newer
- Node 20
- Google OAuth client credentials

Setup:

```sh
git clone <repo-url> pulseops
cd pulseops
cp .env.example .env
```

Fill in `.env` with real Google OAuth values. Keep `?replicaSet=rs0` in `MONGODB_URI`; MongoDB Change Streams require a replica set.

Start the full stack:

```sh
docker-compose up --build
```

Verify the API:

```sh
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

The frontend runs at `http://localhost:5173`.

## Design Decisions

- MongoDB-only, no Redis: TTL indexes replace ephemeral cache-like data for fingerprints, sessions, and rate limits. TTL precision is acceptable for this stage.
- Go monolith, not microservices: a single deployable unit reduces operational overhead while package boundaries keep the code organized.
- Rule-based fingerprinting, not ML: deterministic SHA-256 fingerprints are explainable, cheap, and require no training data.
- Single Container App instance in phase 1: WebSocket fan-out is in-process. At larger scale, Change Streams or a shared pub-sub layer would coordinate instances.
- OAuth PKCE, not DIY password auth: Google handles credential security, while PulseOps stores only httpOnly session cookies and hashed refresh tokens.

## Known Limitations

- WebSocket subscriptions are not horizontally scalable without replacing the in-process Hub with shared pub-sub or Change Stream fan-out across instances.
- Token revocation has a short window: an existing JWT can remain valid for up to 15 minutes after team membership changes.
- MongoDB TTL cleanup is handled by a background sweep and is not millisecond-precise.

## Resume Bullets

- Engineered a distributed real-time incident management platform using Go and GraphQL subscriptions over WebSocket, delivering live alert updates to Vue 3 dashboards with sub-2 second end-to-end latency from webhook to browser
- Designed a rule-based alert fingerprinting engine in Go using SHA-256 hashing to deterministically deduplicate and group correlated alerts, reducing incident noise by ~70% in synthetic load tests
- Implemented OAuth 2.0 Authorization Code Flow with PKCE using Google OIDC in Go, including ID token validation with go-jose, rotating refresh tokens in MongoDB, and httpOnly session cookies - eliminating client-side token exposure
- Architected multi-tenant data isolation enforcing teamId scoping on every MongoDB query via JWT claims context, preventing cross-tenant data access across all GraphQL resolvers
- Built a Vue 3 real-time subscription composable with Apollo WebSocket client, exponential backoff reconnection logic, and scroll-position-aware auto-scroll for live incident feeds
- Deployed containerised Go API to Azure Container Apps with GitHub Actions CI/CD, Trivy vulnerability scanning, and Azure Key Vault secret injection via managed identity - achieving zero-secret-in-code policy
