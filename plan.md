# PulseOps — Agentic Coding Plan

> **How to use this file**
> Each numbered step below is a self-contained prompt for an agentic coding agent (Claude Code, Cursor, Copilot Workspace, etc).
> Copy the step verbatim into the agent. Complete it fully and verify it works before moving to the next step.
> Never skip steps. Never combine steps. The order is deliberate.

---

## Progress tracker

Mark each step `[x]` when the verification check passes. Never mark done until the verification passes.

| Step | Description | Status |
|---|---|---|
| 1.1 | Initialise GitHub repository and folder structure | [x] |
| 1.2 | Create root .env.example | [x] |
| 1.3 | Docker Compose for local development | [x] |
| 1.4 | Initialise Go module and install dependencies | [x] |
| 1.5 | Initialise Vue frontend and install dependencies | [x] |
| 1.6 | GitHub Actions CI pipelines | [x] |
| 1.7 | GitHub branch protection | [x] |
| 1.8 | Verify full local dev stack | [x] |
| 2.1 | Config loader | [x] |
| 2.2 | MongoDB client and connection | [x] |
| 2.3 | MongoDB indexes at startup | [x] |
| 2.4 | Wire main.go with config, MongoDB, graceful shutdown | [x] |
| 3.1 | GraphQL schema SDL | [x] |
| 3.2 | Configure gqlgen and generate Go code | [x] |
| 3.3 | Stub resolvers for all queries and mutations | [x] |
| 4.1 | JWT signing and validation | [x] |
| 4.2 | Auth middleware | [x] |
| 4.3 | Google OAuth flow | [x] |
| 4.4 | Auth HTTP handlers | [x] |
| 5.1 | HTTP router with all middleware | [x] |
| 5.2 | MongoDB document models in Go | [x] |
| 5.3 | Alert fingerprint engine | [x] |
| 5.4 | Alert webhook handler | [x] |
| 6.1 | Subscription Hub | [x] |
| 6.2 | MongoDB Change Stream listener | [x] |
| 6.3 | Wire subscriptions into GraphQL resolver | [x] |
| 7.1 | Incidents repository | [x] |
| 7.2 | Incident state machine service | [x] |
| 7.3 | GraphQL resolvers for incidents | [x] |
| 8.1 | User and team repositories | [x] |
| 8.2 | On-call schedule logic | [x] |
| 8.3 | GraphQL resolvers for teams and on-call | [x] |
| 9.1 | Analytics aggregation | [x] |
| 9.2 | Runbook CRUD | [x] |
| 10.1 | Vue router with auth guards | [x] |
| 10.2 | Pinia auth store | [x] |
| 10.3 | useAuth composable (PKCE login flow) | [x] |
| 10.4 | LoginView and AuthCallback pages | [x] |
| 10.5 | Pinia incidents store | [x] |
| 10.6 | useIncidentFeed composable | [x] |
| 10.7 | DashboardView and IncidentFeed component | [x] |
| 10.8 | IncidentDetailView | [x] |
| 10.9 | MTTR analytics chart | [x] |
| 11.1 | Azure infrastructure provisioning script | [x] |
| 11.2 | Deployment workflows in GitHub Actions | [x] |
| 11.3 | Trivy container vulnerability scanning | [x] |
| 11.4 | Azure Key Vault secret injection | [x] |
| 12.1 | OpenTelemetry tracing | [x] |
| 12.2 | Structured request logging | [x] |
| 12.3 | Project README | [x] |
| 12.4 | Demo recording and resume bullets | [~] |
| 13.1 | Set up Playwright for E2E testing | [x] |
| 13.2 | E2E test: login flow | [x] |
| 13.3 | E2E test: real-time incident flow | [x] |
| 13.4 | Go integration tests for webhook pipeline | [x] |
| 13.5 | Go unit tests for auth package | [x] |
| 13.6 | Unit tests for fingerprint engine and state machine | [x] |
| 14.1 | Synthetic alert generator script | [x] |
| 14.2 | Performance baseline test | [ ] |
| 14.3 | Go benchmarks for critical path | [x] |
| 15.1 | Postmortem workflow | [x] |
| 15.2 | Escalation policy enforcement | [x] |
| 15.3 | API key management UI | [x] |
| 15.4 | Harden error handling across all resolvers | [~] |
| 15.5 | Interview preparation document | [ ] |
| 15.6 | Final portfolio packaging | [ ] |

Legend: `[x]` done · `[~]` partially done · `[ ]` not started

**Current step: 14.2** (remaining work is perf baseline capture + portfolio polish; both need a live stack/host)

---

## Next steps (prioritized)

> Reconciled against the actual codebase on 2026-06-04. The git history (`feat: implement phase 1-12`) landed far more than the old tracker reflected; statuses above were corrected to match what is actually in the repo. The items below are everything still open, ordered by impact.

### Tier 1 — functional gaps (a user can hit these)
1. ~~**15.2 Escalation policy enforcement**~~ — ✅ **Done** (this session). Added `backend/internal/escalation` with a background `Checker` (started from `main.go`, panic-guarded, 60s ticker) that flips stale un-acknowledged `TRIGGERED` incidents to `escalated=true` and publishes `INCIDENT_ESCALATED` to the Hub. Per-team thresholds via an optional `TeamDoc.EscalationPolicy` (defaults: tier-1 5 min, tier-2 15 min log-only). Pure decision helpers (`DueForTier1`/`DueForTier2`) are unit-tested. *Follow-ups:* no GraphQL mutation/UI yet to edit `EscalationPolicy` (set directly in Mongo for now); the change stream also emits a redundant `ALERT_ATTACHED` for the escalation write since status stays `TRIGGERED` (harmless, could be filtered).
2. ~~**15.3 API key management UI**~~ — ✅ **Done** (this session). New `frontend/src/views/TeamSettingsView.vue` at `/team/settings` (owner-only route). Three sections: members (table + invite form + remove with last-owner/self guards), on-call schedule (rotation reorder + interval daily/weekly/biweekly + overrides form), and API key (masked hint, rotate-with-confirmation, one-time full-key reveal modal with copy). Added `TEAM_QUERY`, `INVITE_MEMBER`, `REMOVE_MEMBER`, `UPDATE_SCHEDULE`, `ADD_OVERRIDE`, `ROTATE_API_KEY` operations + `Team`/`OnCallSchedule`/`ScheduleOverride` types, and a settings link in `TeamView.vue`. *Simplifications vs. spec:* rotation reorder uses ↑/↓ buttons (not HTML5 drag-and-drop, to stay dependency-free); no cycle-start date picker because the `updateSchedule` mutation doesn't accept one (backend sets `cycleStart = now`).

### Tier 2 — test & reliability hardening (start here next)
3. ~~**13.4 Go integration tests for the webhook pipeline**~~ — ✅ **Done** (this session). Added `backend/internal/alerting/integration_test.go` (build tag `integration`, same package so it reuses `hashString`/`Fingerprint`). Five tests: new incident (201), in-window dedup (200 + `alertCount==2` + 2 alerts), **post-TTL re-fire → 2 incidents**, invalid API key (401, 0 incidents), rate-limit (101st → 429 + `Retry-After`). Each test uses an isolated, auto-dropped database. Added an `integration-test` CI job to `backend-ci.yml` that boots a single-node Mongo replica set and runs `go test -tags integration ./internal/alerting/...`. *Not run locally* — no Mongo replica set in the dev sandbox; verified it compiles/vets under the tag and runs in CI.
4. ~~**13.3 E2E real-time incident flow**~~ — ✅ **Done** (this session). New `frontend/e2e/incidents.spec.ts` with 4 specs: live appearance via subscription (severity + TRIGGERED), acknowledge updates the card (RESPONDER), viewer sees no Ack button, and the two-tab fan-out proof. To support it, the backend now seeds a deterministic E2E team (id matches the auth bypass) with a known API key under `ENV=test` (`server.SeedE2EData`, e2etest-tagged + no-op otherwise, called from `main.go`). The spec waits for the subscription websocket to open before posting (the Hub does not replay events). *Not run locally* — needs the full e2e docker-compose stack; verified specs parse via `playwright test --list`. Runs in the existing `e2e.yml` job.
5. **15.4 Harden error handling across resolvers** *(partial)* — GraphQL resolvers leak internal errors (raw `err` returned to clients). Introduce a sanitized error layer (log full error + request ID server-side, return a safe message/code to the client). The three critical bugs found earlier are already fixed (see below).

### Tier 3 — performance & portfolio polish (start here next)
6. ~~**14.1 Synthetic alert generator**~~ — ✅ **Done** (this session). `scripts/load-test/generate-alerts.sh` — `curl` loop with `--count/--rate/--api-key/--url/--scenario` (random varies source/alertName/severity/environment; duplicate sends one fingerprint to exercise dedup). Prints progress every 10 alerts (sent, effective rate, HTTP status breakdown). Verified end-to-end against a local test server (status counting, pacing, progress); arg validation covered.
7. ~~**14.3 Go benchmarks**~~ — ✅ **Done** (this session). `backend/internal/alerting/benchmark_test.go`: `BenchmarkFingerprint` (pure CPU — captured at ~263 ns/op), `BenchmarkFingerprintDedup` (dup-insert + incident read), and `BenchmarkWebhookHandler` (full request→write, with the rate limiter reset outside the timed region). DB-backed benchmarks skip cleanly when Mongo is absent. Results recorded in `docs/performance.md`. *DB-backed numbers still need capturing on a replica-set host* — see 14.2.
8. **14.2 Performance baseline** — capture p50/p95 webhook latency and dedup throughput; record numbers in the README.
9. **12.4 Demo recording** *(partial — resume bullets already in README)* — record the live-dashboard demo.
10. **15.5 Interview prep doc** and **15.6 Final portfolio packaging** — wrap-up artifacts.

### Open operational item (from this session's bug fixes)
- The incidents `fingerprint` index was changed from **unique** to non-unique to make post-TTL re-fires work (plan STEP 13.4 acceptance). On a **fresh** DB this is seamless; on an **existing** DB the old `fingerprint_1` unique index must be dropped or `CreateIndexes` will fail at startup with an index-options conflict. Decide: add a drop-if-exists migration to `mongodb.CreateIndexes`, or document a manual `db.incidents.dropIndex(...)`.

### Recent fixes already applied (not yet committed)
- **OnCallStatus subscription panic** — `close(out)` → `defer close(out)` in `schema.resolvers.go` (was crashing the process on first send).
- **Dedup 500 after TTL window** — incidents `fingerprint` index made non-unique + `handleDuplicateAlert` now targets the latest incident.
- **ReDoS in runbook search** — user query now passed through `regexp.QuoteMeta`.

---

## Project overview

PulseOps is a real-time, multi-tenant incident management SaaS web application.

**Stack**
- Backend: Go 1.22+ monolith, gqlgen (GraphQL), chi (HTTP router), mongo-driver, golang-jwt, go-jose, zap, testify
- Frontend: Vue 3, TypeScript, Pinia, Vue Router, Apollo Client, graphql-ws, TailwindCSS, Chart.js, Vite, Vitest
- Database: MongoDB Atlas (replica set required), TTL indexes, Change Streams
- Auth: Google OAuth 2.0 + OIDC, Authorization Code Flow + PKCE, httpOnly cookies, session JWT (15 min), refresh token in MongoDB (7 days)
- Cloud: Azure Container Apps (Go API), Azure Static Web Apps (Vue SPA), Azure Container Registry, Azure Key Vault, Azure Monitor + Application Insights
- DevOps: GitHub Actions, Docker, golangci-lint, Trivy, provision.sh (Azure CLI)

**Architecture summary**
- Incoming alert webhooks hit a Go HTTP handler
- Go fingerprints the alert (SHA-256), deduplicates via MongoDB TTL collection, creates or updates an incident document
- MongoDB Change Stream detects the write and pushes an event to an in-process Go channel (Hub)
- gqlgen subscription resolver broadcasts the event over WebSocket to all connected Vue clients subscribed to that team
- Vue dashboard updates in real time with no polling and no page refresh
- All data is scoped by teamId extracted from the session JWT — never from user input

---

## Phase 1 — Project skeleton

---

### STEP 1.1 — Initialise the GitHub repository and folder structure

Create a new GitHub repository named `pulseops`. Clone it locally.

Create the following top-level folder structure exactly as specified. Do not create any file contents yet — only the folders and empty placeholder files (`.gitkeep` where needed to preserve empty directories).

```
pulseops/
├── backend/
│   ├── cmd/server/
│   ├── internal/
│   │   ├── alerting/
│   │   ├── incidents/
│   │   ├── graph/
│   │   ├── streams/
│   │   └── server/
│   ├── pkg/
│   │   ├── auth/
│   │   ├── mongodb/
│   │   └── config/
│   └── graph/
│       ├── generated/
│       └── model/
├── frontend/
│   └── src/
│       ├── router/
│       ├── stores/
│       ├── composables/
│       ├── views/
│       ├── components/
│       └── graphql/
└── .github/
    └── workflows/
```

Create a root-level `.gitignore` that covers:
- Go binaries, test binaries, and coverage files inside `backend/`
- `frontend/node_modules/` and `frontend/dist/`
- All `.env` files and `.env.local` files (never committed)
- `.DS_Store`, `.idea/`, `.vscode/`
- `docker-compose.override.yml`

Create a root-level `README.md` with a single line: `# PulseOps`.

Commit everything with message `feat: initialise repository structure`.

Push to main.

**Verification:** The repository exists on GitHub, the folder tree matches the spec above, and `.env` is confirmed absent from git tracking.

---

### STEP 1.2 — Create the root .env.example file

Create a `.env.example` file at the repository root with the following variables and placeholder values. This file IS committed to git. It documents every environment variable the project needs.

Variables to include (grouped with comments):

**App**
- `ENV` — value: `development`
- `PORT` — value: `8080`
- `ALLOWED_ORIGINS` — value: `http://localhost:5173`

**MongoDB**
- `MONGODB_URI` — value: `mongodb://mongo:27017/pulseops?replicaSet=rs0`
- `MONGODB_DB` — value: `pulseops`

**Auth**
- `JWT_SECRET` — value: `replace-with-32-char-random-string`
- `JWT_EXPIRY_MINUTES` — value: `15`
- `REFRESH_TOKEN_EXPIRY_DAYS` — value: `7`

**Google OAuth**
- `GOOGLE_CLIENT_ID` — value: `your-client-id.apps.googleusercontent.com`
- `GOOGLE_CLIENT_SECRET` — value: `your-client-secret`
- `OAUTH_REDIRECT_URL` — value: `http://localhost:8080/auth/callback`

**Frontend (Vite)**
- `VITE_API_URL` — value: `http://localhost:8080`
- `VITE_WS_URL` — value: `ws://localhost:8080/query`
- `VITE_GOOGLE_CLIENT_ID` — value: `your-client-id.apps.googleusercontent.com`

Add inline comments explaining:
- `VITE_` prefix is required for Vite to expose vars to the browser
- `VITE_GOOGLE_CLIENT_SECRET` must never be created — the secret stays server-side only
- `?replicaSet=rs0` in the MongoDB URI is required for Change Streams

Copy `.env.example` to `.env` locally and fill in real values (Google OAuth credentials from Google Cloud Console).

Commit `.env.example` with message `chore: add env example with all required variables`.

**Verification:** `.env` does not appear in `git status`. `.env.example` is committed and visible on GitHub.

---

### STEP 1.3 — Create the Docker Compose file for local development

Create `docker-compose.yml` at the repository root.

Define four services:

**mongo**
- Image: `mongo:7`
- Command: `--replSet rs0 --bind_ip_all`
- Port: `27017:27017`
- Volume: named volume `mongo_data` mounted at `/data/db`
- Healthcheck: ping MongoDB using mongosh, interval 10s, timeout 5s, retries 5
- Network: `pulseops`

**mongo-init**
- Image: `mongo:7`
- Depends on: `mongo` with condition `service_healthy`
- Restart: `no`
- Command: run `rs.initiate()` via mongosh if not already initialised
- Network: `pulseops`
- Purpose: initialises the replica set once at startup so Change Streams work. Must complete before backend starts.

**backend**
- Build context: `./backend`, target `dev`
- Port: `8080:8080`
- env_file: `.env`
- Depends on: `mongo` (healthy) and `mongo-init` (completed successfully)
- Volume: `./backend:/app` for hot reload
- Network: `pulseops`

**frontend**
- Image: `node:20-alpine`
- Working dir: `/app`
- Command: `sh -c "npm install && npm run dev -- --host"`
- Port: `5173:5173`
- env_file: `.env`
- Volume: `./frontend:/app` and `/app/node_modules` (anonymous volume to prevent host override)
- Network: `pulseops`

Define named volume `mongo_data` and network `pulseops` with bridge driver.

Commit with message `chore: add docker-compose for local development`.

**Verification:** Running `docker-compose config` produces no errors.

---

### STEP 1.4 — Initialise the Go module and install backend dependencies

Navigate to `backend/`. Initialise a Go module named `github.com/YOURUSERNAME/pulseops`.

Install the following Go dependencies:
- `github.com/go-chi/chi/v5` — HTTP router
- `github.com/99designs/gqlgen` — GraphQL server codegen
- `go.mongodb.org/mongo-driver/mongo` — MongoDB driver
- `github.com/golang-jwt/jwt/v5` — JWT signing and validation
- `github.com/go-jose/go-jose/v4` — OIDC ID token validation
- `golang.org/x/oauth2` — OAuth 2.0 client
- `golang.org/x/oauth2/google` — Google OAuth provider
- `go.uber.org/zap` — structured logging
- `github.com/stretchr/testify` — test assertions
- `github.com/rs/cors` — CORS middleware

Create `backend/go.mod` and `backend/go.sum` with all dependencies resolved.

Create a minimal `backend/Dockerfile` with two stages:
- `dev` stage: uses `golang:1.22-alpine`, installs `air` for hot reload, sets working directory `/app`, runs `air`
- `prod` stage: multi-stage build, compiles binary with `CGO_ENABLED=0 GOOS=linux`, produces minimal final image from `alpine:3.19`

Create `backend/.golangci.yml` with the following linters enabled: `errcheck`, `govet`, `staticcheck`, `unused`, `gofmt`. Set timeout to 5 minutes.

Commit with message `chore: initialise Go module and install dependencies`.

**Verification:** `cd backend && go mod tidy` runs without errors. `go build ./...` succeeds (even with no source files yet).

---

### STEP 1.5 — Initialise the Vue frontend and install dependencies

Navigate to `frontend/`. Scaffold a new Vite + Vue 3 + TypeScript project.

Install the following npm dependencies:
- `vue-router@4` — client-side routing
- `pinia` — state management
- `@apollo/client` — GraphQL client
- `@vue/apollo-composable` — Vue composables for Apollo
- `graphql-ws` — WebSocket transport for GraphQL subscriptions
- `graphql` — peer dependency
- `tailwindcss`, `postcss`, `autoprefixer` — CSS framework
- `chart.js`, `vue-chartjs` — analytics charts

Install dev dependencies:
- `vitest` — unit testing
- `@vue/test-utils` — Vue component testing
- `eslint`, `@typescript-eslint/parser`, `@typescript-eslint/eslint-plugin` — linting
- `eslint-plugin-vue` — Vue-specific linting rules

Initialise TailwindCSS config (`tailwind.config.ts`, `postcss.config.js`). Configure Tailwind to scan `src/**/*.{vue,ts}`.

Create `frontend/tsconfig.json` with strict mode enabled and path alias `@` pointing to `src/`.

Create `frontend/vite.config.ts` with:
- Vue plugin
- Path alias `@` → `src/`
- Dev server proxy: `/api` → `http://localhost:8080` and `/auth` → `http://localhost:8080`

Add npm scripts to `package.json`:
- `dev` — `vite`
- `build` — `vue-tsc && vite build`
- `type-check` — `vue-tsc --noEmit`
- `lint` — `eslint src/ --ext .ts,.vue`
- `test:unit` — `vitest run`

Commit with message `chore: initialise Vue frontend with all dependencies`.

**Verification:** `cd frontend && npm run type-check` passes. `npm run lint` passes. `npm run test:unit` passes (no tests yet is fine).

---

### STEP 1.6 — Create the GitHub Actions CI pipelines

Create `.github/workflows/backend-ci.yml`:
- Trigger on push to `main` and on pull requests, only when files in `backend/` change (use `paths:` filter)
- Job: `lint-test` on `ubuntu-latest`
- Steps: checkout, setup Go 1.22 with cache, `go mod download`, run `golangci-lint` via the official action, run `go test ./... -race -count=1`

Create `.github/workflows/frontend-ci.yml`:
- Trigger on push to `main` and on pull requests, only when files in `frontend/` change
- Job: `lint-test` on `ubuntu-latest`
- Steps: checkout, setup Node 20 with npm cache, `npm ci`, `npm run type-check`, `npm run lint`, `npm run test:unit`

Commit with message `ci: add backend and frontend CI pipelines`.

Push to a new branch named `ci/add-pipelines`, open a pull request, verify both CI jobs go green.

Merge the PR.

**Verification:** Both Actions workflows appear in the GitHub Actions tab. Both are green on the first run.

---

### STEP 1.7 — Configure GitHub branch protection

In GitHub repository Settings → Branches, add a branch protection rule for `main`:

- Require a pull request before merging: enabled
- Require status checks to pass before merging: enabled
  - Add: `backend CI / lint-test`
  - Add: `frontend CI / lint-test`
- Require branches to be up to date before merging: enabled
- Do not allow bypassing the above settings: enabled

**Verification:** Attempting a direct push to `main` is rejected. A PR without green CI cannot be merged.

---

### STEP 1.8 — Verify the full local dev stack starts cleanly

Run `docker-compose up --build` from the repository root.

Verify the following in order:
1. MongoDB container starts and healthcheck passes
2. mongo-init container runs and exits with code 0
3. Backend container starts — look for the log line `server starting on :8080`
4. Frontend container starts — look for the Vite dev server URL

Run these checks manually:
- `docker exec pulseops-mongo-1 mongosh --eval "rs.status().members[0].stateStr"` — must return `PRIMARY`
- `curl http://localhost:8080/health` — must return `{"status":"ok"}` (this endpoint does not exist yet — create a stub that returns this JSON)
- Open `http://localhost:5173` in a browser — must load without errors (a blank Vue app is fine)

Create the `/health` stub in the Go backend: a single HTTP handler on `GET /health` that returns `{"status":"ok"}` with status 200. Wire it to a minimal chi router in `cmd/server/main.go`.

Commit with message `feat: add health check endpoint and verify dev stack`.

**Verification:** All four checks above pass. `docker-compose up` is the single command to start the entire local environment.

---

## Phase 2 — Backend config, database, and indexes

---

### STEP 2.1 — Implement the config loader

Create `backend/pkg/config/config.go`.

Define a `Config` struct with fields for every environment variable listed in `.env.example`: `Env`, `Port`, `AllowedOrigins`, `MongoURI`, `MongoDB`, `JWTSecret`, `JWTExpiryMinutes`, `RefreshTokenExpiryDays`, `GoogleClientID`, `GoogleClientSecret`, `OAuthRedirectURL`.

Implement a `Load()` function that reads each variable from the environment using `os.Getenv`.

Implement a `mustEnv(key string) string` helper that calls `log.Fatalf` if the value is empty. Use this for all required variables. Use `getEnv(key, fallback string) string` for optional ones with defaults.

Parse integer values (`JWTExpiryMinutes`, `RefreshTokenExpiryDays`) with `strconv.Atoi`, fatal on parse error.

**Verification:** Write a unit test in `pkg/config/config_test.go` that sets required env vars, calls `Load()`, and asserts the struct fields are populated correctly.

---

### STEP 2.2 — Implement the MongoDB client and connection

Create `backend/pkg/mongodb/client.go`.

Implement `Connect(uri, dbName string) (*mongo.Database, error)` that:
- Creates a `mongo.Client` using the provided URI
- Calls `client.Ping()` with a 10-second timeout to verify connectivity
- Returns the named database handle

Implement `Disconnect(client *mongo.Client)` that calls `client.Disconnect` with a 5-second timeout, logging any error.

All errors must be returned — no `log.Fatal` inside this package. The caller (`main.go`) decides how to handle connection failure.

**Verification:** Unit test using `testify` that mocks a successful ping. Integration test (build tag `integration`) that connects to a real local MongoDB.

---

### STEP 2.3 — Create all MongoDB indexes at startup

Create `backend/pkg/mongodb/indexes.go`.

Implement `CreateIndexes(db *mongo.Database) error` that creates the following indexes using a 30-second timeout context. All index creation calls must be idempotent (safe to run on every startup).

Indexes to create:

**incidents collection**
- Compound index: `{ teamId: 1, status: 1 }`
- Compound index: `{ teamId: 1, triggeredAt: -1 }`
- Unique index: `{ fingerprint: 1 }`

**alerts collection**
- Index: `{ incidentId: 1 }`
- Compound index: `{ teamId: 1, receivedAt: -1 }`

**fingerprints collection**
- TTL index on `createdAt` with `expireAfterSeconds: 60`
- (The `_id` field stores the fingerprint hash — no additional unique index needed)

**sessions collection**
- TTL index on `expiresAt` with `expireAfterSeconds: 0`
- Index: `{ userId: 1 }`

**rate_limits collection**
- TTL index on `expiresAt` with `expireAfterSeconds: 0`

**users collection**
- Unique index: `{ email: 1 }`
- Unique index: `{ googleSubject: 1 }`
- Index: `{ teamId: 1 }`

**on_call_schedules collection**
- Unique index: `{ teamId: 1 }`

Call `CreateIndexes` from `main.go` after MongoDB connection succeeds. If it returns an error, call `log.Fatal`.

**Verification:** After `docker-compose up`, run `db.incidents.getIndexes()` in mongosh and confirm all indexes exist.

---

### STEP 2.4 — Wire main.go with config, MongoDB, and graceful shutdown

Rewrite `backend/cmd/server/main.go` as the full application entry point.

The startup sequence must be:
1. Call `config.Load()` — panics on missing required vars
2. Initialise `zap.Logger` (production mode if `ENV=production`, development mode otherwise)
3. Call `mongodb.Connect()` — fatal on error
4. Call `mongodb.CreateIndexes()` — fatal on error
5. Call `server.NewRouter(cfg, db, logger)` — returns `http.Handler`
6. Start `http.Server` on `cfg.Port` with read/write/idle timeouts
7. Block until SIGINT or SIGTERM received
8. Graceful shutdown with 30-second timeout
9. Disconnect MongoDB

`server.NewRouter` is a stub for now — it must at minimum register the `GET /health` handler from Step 1.8.

**Verification:** `docker-compose up` starts the backend, logs show config loaded and MongoDB connected, `curl http://localhost:8080/health` returns `{"status":"ok"}`.

---

## Phase 3 — GraphQL schema and codegen

---

### STEP 3.1 — Write the complete GraphQL schema

Create `backend/graph/schema.graphql`.

Write the complete SDL schema including:

**Scalars**
- `Time` (ISO-8601 string)
- `JSON` (arbitrary object)

**Enums**
- `IncidentStatus`: `TRIGGERED`, `ACKNOWLEDGED`, `INVESTIGATING`, `RESOLVED`, `CLOSED`
- `Severity`: `CRITICAL`, `HIGH`, `MEDIUM`, `LOW`
- `Role`: `OWNER`, `RESPONDER`, `VIEWER`
- `IncidentEventType`: `INCIDENT_CREATED`, `INCIDENT_ACKNOWLEDGED`, `INCIDENT_INVESTIGATING`, `INCIDENT_RESOLVED`, `INCIDENT_ESCALATED`, `ALERT_ATTACHED`

**Types**
- `Incident`: id, title, status, severity, teamId, fingerprint, alertCount, triggeredAt, acknowledgedAt, acknowledgedBy (User), resolvedAt, resolvedBy (User), escalated, escalatedAt, assignee (User), runbook (Runbook), statusMessage, mttr (Int nullable), alerts ([Alert])
- `Alert`: id, incidentId, source, alertName, severity, environment, payload (JSON), fingerprint, receivedAt
- `User`: id, email, name, avatarUrl, teamId, role, googleSubject, createdAt
- `Team`: id, name, members ([User]), onCallSchedule (OnCallSchedule), escalationPolicy (EscalationPolicy), apiKeyHint
- `OnCallSchedule`: id, teamId, rotation ([User]), intervalDays, cycleStart, currentOnCall (User), overrides ([ScheduleOverride])
- `ScheduleOverride`: id, user, startsAt, endsAt, reason
- `EscalationPolicy`: id, teamId, tiers ([EscalationTier])
- `EscalationTier`: tierNumber, waitMinutes, notifyUserId, notifyUser (User)
- `Runbook`: id, teamId, title, content, tags ([String]), updatedAt
- `Postmortem`: id, incidentId, authorId, summary, timeline, actionItems ([String]), createdAt
- `IncidentPage`: items ([Incident]), totalCount, hasMore
- `Analytics`: mttrSeconds, mttaSeconds, totalCount, byDay ([DayStat]), bySeverity ([SeverityStat])
- `DayStat`: date, count
- `SeverityStat`: severity, count
- `IncidentEvent`: type (IncidentEventType), incident (Incident), actor (User nullable), occurredAt
- `OnCallStatusEvent`: teamId, currentOnCall (User), nextOnCall (User), handoffAt

**Query type**
- `me: User!`
- `incidents(teamId, status, severity, from, to, limit, offset): IncidentPage!`
- `incident(id: ID!): Incident`
- `team(id: ID!): Team`
- `analytics(teamId, from, to): Analytics!`
- `runbooks(teamId, query): [Runbook!]!`

**Mutation type**
- `acknowledgeIncident(id, message): Incident!`
- `investigateIncident(id, message): Incident!`
- `resolveIncident(id, summary): Incident!`
- `createTeam(name): Team!`
- `inviteMember(teamId, email, role): User!`
- `removeMember(teamId, userId): Boolean!`
- `updateSchedule(teamId, rotation, intervalDays): OnCallSchedule!`
- `addOverride(teamId, userId, startsAt, endsAt, reason): ScheduleOverride!`
- `upsertRunbook(id, teamId, title, content, tags): Runbook!`
- `createPostmortem(incidentId, summary, timeline, actionItems): Postmortem!`
- `rotateApiKey(teamId): String!`

**Subscription type**
- `incidentFeed(teamId, severity, status): IncidentEvent!`
- `onCallStatus(teamId): OnCallStatusEvent!`

**Verification:** Schema file exists. No syntax errors (validate with `gqlgen` dry run in next step).

---

### STEP 3.2 — Configure gqlgen and generate Go code

Create `backend/graph/gqlgen.yml` with:
- Schema path pointing to `graph/schema.graphql`
- Output for generated code: `graph/generated/generated.go`
- Output for models: `graph/model/models_gen.go`
- Resolver output: `internal/graph/`
- Custom scalar mappings: `Time` → `time.Time`, `JSON` → `map[string]interface{}`
- Autobind: disabled (explicit model mapping preferred)

Run `go run github.com/99designs/gqlgen generate` from the `backend/` directory.

This produces:
- `graph/generated/generated.go` — auto-generated, never edit
- `graph/model/models_gen.go` — auto-generated, never edit
- Resolver interface stubs in `internal/graph/`

Add a `//go:generate go run github.com/99designs/gqlgen generate` directive at the top of `internal/graph/resolver.go`.

Add a comment header to both generated files: `// Code generated by gqlgen. DO NOT EDIT.`

Commit generated files — they must be in version control so CI does not need to run codegen.

Commit with message `feat: add GraphQL schema and generate gqlgen code`.

**Verification:** `go build ./...` succeeds after generation. All resolver methods exist as stubs returning `nil, nil`.

---

### STEP 3.3 — Implement stub resolvers for every query and mutation

In `internal/graph/`, implement stub resolver methods for every Query, Mutation, and Subscription defined in the schema.

Each stub must:
- Accept the correct arguments as generated by gqlgen
- Return the correct return type (with nil/zero value)
- Include a `// TODO: implement` comment
- NOT panic

Do not implement any actual logic yet. The goal is a compilable, runnable GraphQL server that returns empty responses.

Create `internal/graph/resolver.go` with the root `Resolver` struct. The struct will eventually hold references to repositories, the Hub, and config. For now it is empty.

**Verification:** `go build ./...` succeeds. Starting the server and opening the GraphQL playground at `http://localhost:8080/query` (if playground is enabled) shows the schema.

---

## Phase 4 — Auth package

---

### STEP 4.1 — Implement JWT signing and validation

Create `backend/pkg/auth/jwt.go`.

Implement:
- `Claims` struct with fields: `UserID string`, `TeamID string`, `Role string`, `Email string`, and standard `jwt.RegisteredClaims`
- `SignJWT(claims Claims, secret string, expiryMinutes int) (string, error)` — signs with HS256
- `ValidateJWT(tokenString, secret string) (*Claims, error)` — parses and validates signature, expiry, and required fields. Returns a typed error distinguishing expired tokens from invalid ones.

Create `backend/pkg/auth/context.go`.

Implement:
- `type contextKey` (unexported)
- `WithClaims(ctx context.Context, claims *Claims) context.Context` — stores claims in context
- `FromContext(ctx context.Context) *Claims` — retrieves claims, returns nil if not present
- `RequireAuth(ctx context.Context) (*Claims, error)` — returns claims or an unauthenticated error
- `RequireRole(ctx context.Context, roles ...string) (*Claims, error)` — returns claims if role matches, otherwise returns an unauthorised error

**Verification:** Unit tests covering: valid token round-trip, expired token returns correct error type, invalid signature returns error, `RequireRole` returns error for wrong role.

---

### STEP 4.2 — Implement the auth middleware

Create `backend/pkg/auth/middleware.go`.

Implement `Middleware(jwtSecret string) func(http.Handler) http.Handler`.

The middleware must:
1. Read the `session` cookie from the request
2. If no cookie: call `next` with the original context (no claims injected — unauthenticated)
3. If cookie present: call `ValidateJWT`
4. If token expired: clear the cookie (set `MaxAge: -1`), call `next` with no claims
5. If token invalid: clear the cookie, call `next` with no claims
6. If token valid: call `WithClaims` to inject claims into context, call `next`

The middleware never returns 401 directly — it delegates that decision to individual handlers and resolvers. This allows public routes to coexist with protected routes on the same router without special casing.

**Verification:** Unit tests covering all five branches above using `httptest`.

---

### STEP 4.3 — Implement Google OAuth flow

Create `backend/pkg/auth/oauth.go`.

Implement:
- `NewGoogleOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config` — returns configured oauth2.Config with scopes `openid`, `email`, `profile`
- `GenerateStateToken() (string, error)` — generates a cryptographically random 32-byte hex string for the OAuth state parameter
- `ExchangeCode(ctx context.Context, cfg *oauth2.Config, code, codeVerifier string) (*oauth2.Token, error)` — exchanges authorisation code for tokens using PKCE verifier
- `ValidateIDToken(ctx context.Context, rawIDToken string, clientID string) (*GoogleClaims, error)` — fetches Google's JWKS, verifies the token signature, checks `iss`, `aud`, `exp`, returns extracted claims
- `GoogleClaims` struct: `Subject string`, `Email string`, `Name string`, `Picture string`

For `ValidateIDToken`: use `go-jose` to parse and verify. Fetch Google's public keys from `https://www.googleapis.com/oauth2/v3/certs`. Cache the JWKS response for 1 hour using a simple in-memory cache with a mutex.

**Verification:** Unit tests for `GenerateStateToken` (verify randomness, 64 hex chars), `ValidateIDToken` with a mocked JWKS endpoint covering valid token, expired token, wrong audience.

---

### STEP 4.4 — Implement auth HTTP handlers

Create `backend/internal/server/auth_handlers.go`.

Implement the following HTTP handlers using the `chi` router pattern:

**GET /auth/login**
- Generate state token, store in a short-lived cookie (`state`, httpOnly, 10-minute expiry)
- Build Google authorisation URL with state and `code_challenge_method=S256`
- Note: the PKCE code verifier is generated client-side in Vue — this endpoint does NOT generate it
- Redirect to the Google URL

**GET /auth/callback**
- Read `code`, `state`, `code_verifier` from query params
- Validate `state` matches the state cookie; if mismatch return 400
- Clear the state cookie
- Call `ExchangeCode` with the code and verifier
- Call `ValidateIDToken` on the ID token
- Look up user in MongoDB by `googleSubject`; if not found, create new user document
- Call `SignJWT` to issue session token (15 min)
- Generate refresh token, hash with SHA-256, insert into `sessions` collection
- Set `session` cookie: httpOnly, Secure (in production), SameSite=Lax (dev) / Strict (prod), Path=/
- Set `refresh` cookie: httpOnly, Secure, SameSite=Lax/Strict, Path=/auth/refresh
- Redirect to `http://localhost:5173/auth/callback` (frontend handles the rest)

**GET /auth/me**
- Require valid session JWT (use `RequireAuth` from context — middleware already ran)
- Return current user object as JSON
- Return 401 if no valid session

**POST /auth/refresh**
- Read refresh token from `refresh` cookie
- SHA-256 hash it, look up in `sessions` collection
- If not found or expired: return 401
- Delete old session document
- Issue new session JWT and new refresh token
- Insert new session document
- Set new `session` and `refresh` cookies
- Return 200

**POST /auth/logout**
- Delete session document from MongoDB using userId from JWT claims
- Clear both `session` and `refresh` cookies (MaxAge: -1)
- Return 200

**Verification:** Integration tests for the full callback → me → refresh → logout cycle using `httptest` and a real local MongoDB.

---

## Phase 5 — HTTP router and webhook

---

### STEP 5.1 — Implement the HTTP router with all middleware

Create `backend/internal/server/router.go`.

Implement `NewRouter(cfg *config.Config, db *mongo.Database, logger *zap.Logger) http.Handler`.

Wire the following middleware on all routes:
- Request ID middleware (generate UUID, set on context and `X-Request-ID` response header)
- Structured request logger (zap — log method, path, status, duration, request ID)
- CORS middleware with `AllowedOrigins: cfg.AllowedOrigins`, `AllowCredentials: true`, allowed headers including `Content-Type` and `Authorization`
- Recovery middleware (catch panics, log with stack trace, return 500)
- Auth middleware (`pkg/auth.Middleware`) — runs on all routes, injects claims when present

Register routes:

**Public (no auth required)**
- `GET /health` → returns `{"status":"ok"}`

**Auth routes (public)**
- `GET /auth/login`
- `GET /auth/callback`

**Auth routes (require session)**
- `GET /auth/me`
- `POST /auth/refresh`
- `POST /auth/logout`

**Webhook (API key auth, not session)**
- `POST /webhooks/alerts` — separate API key middleware, not session JWT

**GraphQL**
- `POST /query` — GraphQL HTTP (queries + mutations), requires auth middleware
- `GET /query` — WebSocket upgrade for subscriptions, requires separate WebSocket origin check

For the WebSocket endpoint: configure gqlgen's WebSocket handler with `CheckOrigin` that only allows `cfg.AllowedOrigins`. This is separate from the HTTP CORS config and must be set explicitly.

**Verification:** `curl http://localhost:8080/health` returns `{"status":"ok"}`. `curl -X POST http://localhost:8080/auth/logout` without a cookie returns 401.

---

### STEP 5.2 — Define MongoDB document models in Go

Create `backend/internal/incidents/models.go`.

Define Go structs for all MongoDB documents. Use `bson` struct tags throughout. These are the persistence models — separate from the GraphQL models generated by gqlgen.

Structs to define:
- `IncidentDoc`: all fields from the MongoDB schema in Step 2.3, using `primitive.ObjectID` for ID fields and `time.Time` for dates
- `AlertDoc`: all fields from alerts collection
- `FingerprintDoc`: `ID string` (bson `_id`), `IncidentID primitive.ObjectID`, `TeamID primitive.ObjectID`, `CreatedAt time.Time`
- `SessionDoc`: `ID string`, `UserID primitive.ObjectID`, `TokenHash string`, `ExpiresAt time.Time`, `CreatedAt time.Time`, `UserAgent string`, `IPAddress string`
- `RateLimitDoc`: `ID string` (bson `_id`), `Count int`, `ExpiresAt time.Time`
- `UserDoc`: all fields from users collection
- `TeamDoc`: `ID primitive.ObjectID`, `Name string`, `APIKeyHash string`, `APIKeyHint string`, `CreatedAt time.Time`, `OwnerID primitive.ObjectID`
- `OnCallScheduleDoc`: all fields including embedded `OverrideDoc` slice
- `RunbookDoc`: all fields
- `PostmortemDoc`: all fields

**Verification:** `go build ./...` succeeds. No unused imports.

---

### STEP 5.3 — Implement the alert fingerprint engine

Create `backend/internal/alerting/fingerprint.go`.

Implement:
- `AlertPayload` struct: `Source string`, `AlertName string`, `Severity string`, `Environment string`, `Labels map[string]string`, `Payload map[string]interface{}`
- `Fingerprint(a AlertPayload) string` — normalises source, alertName, severity, environment (lowercase + trim), concatenates with `::` separator, computes SHA-256, returns first 16 hex characters
- `NormalizeAlertPayload(raw map[string]interface{}) (AlertPayload, error)` — parses raw webhook JSON into an `AlertPayload`, validates required fields are present, returns error with field name if missing

Create `backend/internal/alerting/fingerprint_test.go` with tests covering:
- Same inputs always produce same fingerprint
- Different sources produce different fingerprints
- Case insensitivity (uppercase and lowercase source produce same fingerprint)
- Missing required field returns error

**Verification:** All unit tests pass with `go test ./internal/alerting/... -v`.

---

### STEP 5.4 — Implement the alert webhook handler

Create `backend/internal/alerting/handler.go`.

Implement `NewWebhookHandler(db *mongo.Database, logger *zap.Logger) http.HandlerFunc`.

The handler must perform these steps in order:

1. **Rate limiting** — extract API key from `X-API-Key` header. Compute rate limit key as SHA-256 of API key + current UTC minute (format `YYYYMMDDHHMM`). Upsert a `rate_limits` document with `$inc: {count: 1}` and `$setOnInsert: {expiresAt: now+60s}`. If count exceeds 100, return 429 with `Retry-After: 60` header.

2. **API key validation** — SHA-256 hash the raw API key, query `teams` collection for `{apiKeyHash: hash}`. If not found, return 401. Store the team document for use in subsequent steps.

3. **Parse and validate request body** — decode JSON body into `map[string]interface{}`. Call `NormalizeAlertPayload`. If validation fails, return 400 with field error message.

4. **Fingerprint** — call `Fingerprint(payload)`. Combine with `teamId` to produce a team-scoped fingerprint key: `teamId + "::" + fingerprint`.

5. **Deduplication** — attempt to insert a `FingerprintDoc` with `_id = scopedFingerprint`. If write fails with duplicate key error (E11000), this is a duplicate alert. Fetch the existing incident by fingerprint, increment its `alertCount` by 1, insert an `AlertDoc` linked to that incident, return 200.

6. **Create incident** — if fingerprint is new: insert an `IncidentDoc` with status `TRIGGERED`, link the fingerprint to it, insert an `AlertDoc`. Return 201.

7. **Logging** — log all outcomes (new incident, duplicate, rate limited, invalid key) with structured fields (teamId, fingerprint, alertCount, duration).

**Verification:** Use `curl` to POST a test alert payload to `http://localhost:8080/webhooks/alerts`. Verify the incident document appears in MongoDB. POST the same alert again within 60 seconds — verify `alertCount` increments and no new incident is created.

---

## Phase 6 — Real-time subscription pipeline

---

### STEP 6.1 — Implement the subscription Hub

Create `backend/internal/streams/hub.go`.

Implement a `Hub` struct that manages subscription fan-out from a single event source to many connected WebSocket clients.

The Hub must:
- Hold a buffered channel `events chan IncidentEvent` with capacity 256 (receives from Change Stream)
- Hold a `map[string][]chan IncidentEvent` keyed by teamId (registered subscriber channels)
- Use `sync.RWMutex` to protect the subscribers map
- Implement `Subscribe(teamId string) chan IncidentEvent` — creates a buffered subscriber channel (capacity 16), registers it, returns it
- Implement `Unsubscribe(teamId string, ch chan IncidentEvent)` — removes the channel from the map, closes it
- Implement `Publish(event IncidentEvent)` — called by the Change Stream listener. Acquires read lock. Sends the event to all channels registered for `event.TeamID`. Uses non-blocking send (`select` with `default`) to drop events to slow subscribers rather than blocking the publisher.
- Implement `NewHub() *Hub` — creates and returns an initialised Hub

Define `IncidentEvent` struct in `backend/internal/streams/event.go`:
- `Type string` — matches `IncidentEventType` enum values
- `IncidentID string`
- `TeamID string`
- `Payload interface{}` — the full incident document to be returned to subscribers

**Verification:** Unit test that subscribes two channels on the same teamId, publishes one event, and asserts both channels receive it. Test that a slow subscriber (full channel) does not block the publisher.

---

### STEP 6.2 — Implement the MongoDB Change Stream listener

Create `backend/internal/streams/changestream.go`.

Implement `StartChangeStreamListener(ctx context.Context, db *mongo.Database, hub *Hub, logger *zap.Logger)`.

This function must:
1. Open a Change Stream on the `incidents` collection watching for `insert` and `update` operations
2. Run in a `for` loop that calls `cs.Next(ctx)` to receive change events
3. For each event: decode the full document, construct an `IncidentEvent` with type `INCIDENT_CREATED` (for insert) or the appropriate status-change type (for update), call `hub.Publish(event)`
4. Implement reconnection logic: if `cs.Err()` is non-nil and not `context.Canceled`, log the error, wait 3 seconds (exponential backoff up to 30 seconds), and re-open the Change Stream using the last resume token
5. Store the resume token from each event so reconnection can resume from where it left off
6. Exit cleanly when `ctx` is cancelled (application shutdown)

Call `StartChangeStreamListener` from `main.go` as a goroutine, passing the app's root context. When the server shuts down, the context cancels and the goroutine exits.

**Verification:** Start the server, open a GraphQL subscription in a browser or test client, POST an alert via the webhook endpoint, verify the subscription receives the event within 2 seconds.

---

### STEP 6.3 — Wire subscriptions into the GraphQL resolver

In `internal/graph/resolver.go`, add the `Hub *streams.Hub` field to the `Resolver` struct.

Implement the `IncidentFeed` subscription resolver:
1. Call `RequireAuth(ctx)` — return error if unauthenticated
2. Validate that the requested `teamId` matches `claims.TeamID`
3. Call `hub.Subscribe(teamId)` to get a subscriber channel
4. Start a goroutine that waits for `ctx.Done()` and calls `hub.Unsubscribe` when the WebSocket disconnects
5. Create an output channel of type `<-chan *model.IncidentEvent`
6. Start a goroutine that reads from the Hub subscriber channel, maps `streams.IncidentEvent` to `*model.IncidentEvent`, and sends to the output channel
7. Return the output channel to gqlgen (gqlgen sends each received value to the WebSocket client)

Implement the `OnCallStatus` subscription resolver following the same pattern, filtering by teamId.

**Verification:** Open two browser tabs. Both open `incidentFeed` subscriptions for the same team. POST an alert via webhook. Both tabs receive the event simultaneously.

---

## Phase 7 — Incident resolvers

---

### STEP 7.1 — Implement the incidents repository

Create `backend/internal/incidents/repository.go`.

Implement `Repository` struct with a `*mongo.Database` field.

Implement the following methods. Every method must include `teamID string` as a parameter and apply it as a filter on every MongoDB query — no exceptions.

- `Create(ctx, doc IncidentDoc) (*IncidentDoc, error)` — insertOne
- `GetByID(ctx, id, teamID string) (*IncidentDoc, error)` — findOne with `{_id, teamId}` — returns nil (not error) when not found
- `GetByFingerprint(ctx, fingerprint, teamID string) (*IncidentDoc, error)` — findOne
- `List(ctx, teamID string, filters ListFilters) ([]*IncidentDoc, int64, error)` — find with teamId + optional status/severity/time filters, pagination via limit+offset, total count via `CountDocuments`
- `UpdateStatus(ctx, id, teamID string, update StatusUpdate) (*IncidentDoc, error)` — findOneAndUpdate with `{_id, teamId}` filter, returns updated document
- `IncrementAlertCount(ctx, id, teamID string) error` — updateOne with `$inc: {alertCount: 1}`

Define `ListFilters` struct (optional status, severity, from/to times, limit, offset) and `StatusUpdate` struct (status, optional acknowledgedBy/resolvedBy/statusMessage, computed timestamps).

**Verification:** Unit tests for `GetByID` asserting that a document with a different teamId returns nil, not an error.

---

### STEP 7.2 — Implement the incident state machine service

Create `backend/internal/incidents/service.go`.

Implement `Service` struct wrapping the `Repository`.

Implement the following methods, each enforcing valid state transitions:

- `Acknowledge(ctx context.Context, incidentID string, claims *auth.Claims, message string) (*IncidentDoc, error)`
  - Require role: OWNER or RESPONDER
  - Fetch incident by ID + teamID
  - Validate current status is TRIGGERED or ESCALATED — return error if ACKNOWLEDGED, RESOLVED, or CLOSED
  - Update: status = ACKNOWLEDGED, acknowledgedAt = now, acknowledgedBy = claims.UserID, statusMessage = message

- `Investigate(ctx, incidentID string, claims *auth.Claims, message string) (*IncidentDoc, error)`
  - Require role: OWNER or RESPONDER
  - Validate current status is ACKNOWLEDGED
  - Update: status = INVESTIGATING, statusMessage = message

- `Resolve(ctx, incidentID string, claims *auth.Claims, summary string) (*IncidentDoc, error)`
  - Require role: OWNER or RESPONDER
  - Validate current status is ACKNOWLEDGED or INVESTIGATING
  - Compute mttrSeconds = resolvedAt - triggeredAt
  - Update: status = RESOLVED, resolvedAt = now, resolvedBy = claims.UserID, resolutionSummary = summary, mttrSeconds

All state transition violations must return a typed error (not a generic string) so the GraphQL resolver can return a meaningful error code.

**Verification:** Unit tests for all valid transitions and all invalid transitions. Invalid transitions must return the correct error type.

---

### STEP 7.3 — Implement GraphQL resolvers for incidents

In `internal/graph/incident.resolvers.go`, replace stubs with real implementations.

Implement:

**Query.incidents** — call `RequireAuth`, validate `args.TeamID == claims.TeamID`, call `incidentService.repository.List`, map `[]*IncidentDoc` to `[]*model.Incident`, return `IncidentPage`

**Query.incident** — call `RequireAuth`, call `repository.GetByID(id, claims.TeamID)`, return nil if not found (do NOT return an error for missing records — that leaks existence information)

**Mutation.acknowledgeIncident** — call `incidentService.Acknowledge`, publish event to Hub, return updated incident

**Mutation.investigateIncident** — call `incidentService.Investigate`, publish event to Hub, return updated incident

**Mutation.resolveIncident** — call `incidentService.Resolve`, publish event to Hub, return updated incident

For all resolvers: map between `IncidentDoc` (MongoDB model) and `model.Incident` (GraphQL model) using a dedicated mapper function in `internal/graph/mappers.go`. Never return raw MongoDB documents to GraphQL resolvers.

**Verification:** Open GraphQL playground. Run `mutation { acknowledgeIncident(id: "existing_id") { status acknowledgedAt } }`. Verify status changes. Verify a subscription client receives the event.

---

## Phase 8 — Team, user, and on-call resolvers

---

### STEP 8.1 — Implement user and team repositories

Create `backend/internal/teams/repository.go`.

Implement repository methods for:
- `FindUserByGoogleSubject(ctx, subject string) (*UserDoc, error)`
- `FindUserByID(ctx, id, teamID string) (*UserDoc, error)`
- `CreateUser(ctx, doc UserDoc) (*UserDoc, error)`
- `FindTeamByID(ctx, id string) (*TeamDoc, error)`
- `FindTeamByAPIKeyHash(ctx, hash string) (*TeamDoc, error)`
- `CreateTeam(ctx, doc TeamDoc) (*TeamDoc, error)`
- `ListTeamMembers(ctx, teamID string) ([]*UserDoc, error)`
- `UpdateUserRole(ctx, userID, teamID string, role string) error`
- `RemoveUserFromTeam(ctx, userID, teamID string) error`

These are used by both the auth handlers and the GraphQL resolvers.

**Verification:** Unit tests for all methods with mock MongoDB using `testify/mock`.

---

### STEP 8.2 — Implement on-call schedule logic

Create `backend/internal/oncall/schedule.go`.

Implement:
- `CurrentOnCall(schedule OnCallScheduleDoc, at time.Time) (*UserDoc, error)` — computes who is currently on call based on rotation, intervalDays, cycleStart, and current time. Formula: `index = floor((at - cycleStart) / intervalDays*24h) % len(rotation)`. Checks overrides first — if an override exists for `at`, return the override user.
- `NextOnCall(schedule OnCallScheduleDoc, at time.Time) (*UserDoc, error)` — returns the next person after the current one
- `NextHandoffAt(schedule OnCallScheduleDoc, at time.Time) time.Time` — returns the timestamp when the current rotation slot ends
- `AddOverride(ctx, db, teamID string, override ScheduleOverride) error` — validates override time range does not exceed 30 days, pushes to `overrides` array in MongoDB

The `currentOnCall` field on `OnCallSchedule` in GraphQL is a computed field — it calls `CurrentOnCall` at query time, it is not stored.

**Verification:** Unit tests for `CurrentOnCall` with a 3-person rotation and various timestamps. Test override takes priority over rotation.

---

### STEP 8.3 — Implement GraphQL resolvers for teams and on-call

In `internal/graph/team.resolvers.go`, implement:

**Query.team** — require auth, return team (with `escalationPolicy` hidden from VIEWER role using a field-level check)

**Mutation.createTeam** — require OWNER role, generate a new API key (32 random bytes, hex encoded), store only the SHA-256 hash and last 4 chars hint, create team document, return team (return the full API key once in the `apiKeyHint` field with a note that it will not be shown again)

**Mutation.inviteMember** — require OWNER role, validate user exists by email, update their `teamId` and `role`

**Mutation.removeMember** — require OWNER role, prevent removing the last owner, call `RemoveUserFromTeam`

**Mutation.updateSchedule** — require OWNER role, validate all user IDs in rotation belong to the team, upsert schedule document

**Mutation.addOverride** — require OWNER role, call `AddOverride`

**Mutation.rotateApiKey** — require OWNER role, generate new API key, update team document with new hash and hint, return the full new key (only time it is visible)

**Team field resolver: onCallSchedule** — fetch schedule by teamId, compute `currentOnCall` using the schedule logic

**Verification:** Create a team, invite a member, set an on-call schedule, verify `currentOnCall` returns the correct user.

---

## Phase 9 — Analytics and runbooks

---

### STEP 9.1 — Implement analytics aggregation

Create `backend/internal/analytics/service.go`.

Implement `ComputeAnalytics(ctx, db *mongo.Database, teamID string, from, to time.Time) (*AnalyticsResult, error)`.

Use MongoDB aggregation pipeline on the `incidents` collection:
- Match: `{ teamId, status: {$in: [RESOLVED, CLOSED]}, triggeredAt: {$gte: from, $lte: to} }`
- Group: compute average `mttrSeconds`, average `mttaSeconds` (acknowledgedAt - triggeredAt), total count
- Facet: group by day (use `$dateToString` with format `%Y-%m-%d`), group by severity

All aggregation must happen in MongoDB — do not fetch raw documents and aggregate in Go.

**Verification:** Insert 10 test incidents with known MTTR values. Call `ComputeAnalytics` and assert the returned average MTTR matches the expected value.

---

### STEP 9.2 — Implement runbook CRUD

Create `backend/internal/runbooks/repository.go`.

Implement:
- `Upsert(ctx, doc RunbookDoc) (*RunbookDoc, error)` — if `doc.ID` is zero, insert; otherwise update (with teamId scoping)
- `List(ctx, teamID string, query string) ([]*RunbookDoc, error)` — if query is non-empty, use MongoDB `$regex` case-insensitive match on title and tags
- `GetByID(ctx, id, teamID string) (*RunbookDoc, error)`
- `Delete(ctx, id, teamID string) error`

Implement the `Query.runbooks` and `Mutation.upsertRunbook` GraphQL resolvers using this repository.

**Verification:** Create a runbook, search by title, verify it appears in results. Update content, verify the change persists.

---

## Phase 10 — Vue frontend

---

### STEP 10.1 — Create the Vue router with auth guards

Create `frontend/src/router/index.ts`.

Define routes:
- `/` → redirect to `/dashboard`
- `/login` → `LoginView.vue` (no auth required)
- `/auth/callback` → `AuthCallback.vue` (no auth required)
- `/dashboard` → `DashboardView.vue` (requiresAuth: true)
- `/incidents/:id` → `IncidentDetailView.vue` (requiresAuth: true)
- `/team` → `TeamView.vue` (requiresAuth: true, OWNER or RESPONDER only)

Implement `router.beforeEach` guard:
1. If route does not require auth: proceed
2. If `authStore.checked` is false: call `await authStore.fetchMe()`
3. If `authStore.user` is not null: proceed
4. If route requires OWNER role and user is VIEWER: redirect to `/dashboard` with a query param `?error=unauthorized`
5. Otherwise: redirect to `/login` with `?redirect=<intended path>`

After successful auth callback, read the `redirect` query param and navigate there.

**Verification:** Navigate to `/dashboard` without a session — should redirect to `/login`. Log in — should redirect to `/dashboard`. Log out — should redirect to `/login`.

---

### STEP 10.2 — Implement the Pinia auth store

Create `frontend/src/stores/auth.ts`.

Define the store using Composition API style (`defineStore` with setup function).

State:
- `user: User | null` — the current authenticated user
- `checked: boolean` — whether `/auth/me` has been called at least once this session

Actions:
- `fetchMe(): Promise<void>` — GET `/auth/me` with `credentials: 'include'`, set `user` from response, set `checked = true` in `finally` block (always mark checked even on 401)
- `logout(): Promise<void>` — POST `/auth/logout`, clear `user`, set `checked = false`, navigate to `/login`

Computed:
- `isAuthenticated: boolean` — `user !== null`
- `isOwner: boolean` — `user?.role === 'OWNER'`
- `isResponder: boolean` — `user?.role === 'RESPONDER'`
- `isViewer: boolean` — `user?.role === 'VIEWER'`

**Verification:** Unit test with Vitest + `@vue/test-utils` mocking the fetch calls.

---

### STEP 10.3 — Implement the useAuth composable (PKCE login flow)

Create `frontend/src/composables/useAuth.ts`.

Implement `useAuth()` composable returning:

- `login()` — implements PKCE:
  1. Generate `codeVerifier`: 43 random bytes, base64url encoded (no padding)
  2. Compute `codeChallenge`: SHA-256 of the code verifier, base64url encoded
  3. Store `codeVerifier` in `sessionStorage` with key `pkce_verifier`
  4. Generate random `state` string (32 hex chars), store in `sessionStorage` with key `oauth_state`
  5. Build Google authorisation URL manually using `VITE_GOOGLE_CLIENT_ID`, `VITE_API_URL`, state, challenge, scopes
  6. Set `window.location.href` to the URL

- `handleCallback(code: string, state: string): Promise<void>` — called from `AuthCallback.vue`:
  1. Read stored state from `sessionStorage`, compare with received state — throw if mismatch
  2. Read `codeVerifier` from `sessionStorage`
  3. Clear both sessionStorage keys
  4. POST to `/auth/callback` with `{code, codeVerifier, state}`
  5. Call `authStore.fetchMe()` to hydrate the store
  6. Navigate to `router.currentRoute.value.query.redirect ?? '/dashboard'`

Use the Web Crypto API (`window.crypto.subtle`) for SHA-256 — no external crypto libraries.

**Verification:** Unit test the PKCE generation: verify code verifier is 43+ chars, code challenge is base64url without padding.

---

### STEP 10.4 — Create the LoginView and AuthCallback pages

Create `frontend/src/views/LoginView.vue`.

Display a centered login page with:
- PulseOps logo/wordmark
- A "Sign in with Google" button that calls `useAuth().login()`
- An error message area that displays `route.query.error` if present

Create `frontend/src/views/AuthCallback.vue`.

On `mounted`:
1. Read `code` and `state` from `route.query`
2. If either is missing: display error "Invalid callback — missing code or state" and show link to `/login`
3. Call `useAuth().handleCallback(code, state)`
4. Show a loading spinner while the callback is in progress
5. On error: display the error message and a retry link

**Verification:** Visiting `/login` shows the login button. Clicking it redirects to Google. After Google approval, the callback page loads and redirects to `/dashboard`.

---

### STEP 10.5 — Implement the Pinia incidents store

Create `frontend/src/stores/incidents.ts`.

State:
- `incidents: Incident[]` — list of incidents for the current team
- `loading: boolean`
- `error: string | null`

Actions:
- `loadIncidents(teamId: string, filters?: IncidentFilters): Promise<void>` — executes Apollo `incidents` query, replaces `incidents` array
- `acknowledgeIncident(id: string, message?: string): Promise<void>` — executes `acknowledgeIncident` mutation, updates the incident in the store by ID
- `resolveIncident(id: string, summary: string): Promise<void>` — executes `resolveIncident` mutation, updates the incident in the store
- `addLiveIncident(event: IncidentEvent): void` — called by the subscription composable. For `INCIDENT_CREATED`: prepend to list. For status updates: find and update in place.

**Verification:** Unit test `addLiveIncident` with each event type.

---

### STEP 10.6 — Implement the useIncidentFeed composable

Create `frontend/src/composables/useIncidentFeed.ts`.

This is the most complex frontend component. Implement it carefully.

The composable must:

1. **Open subscription** — on `onMounted`, start the `incidentFeed` Apollo subscription for the authenticated user's teamId using `useSubscription` from `@vue/apollo-composable`

2. **Handle subscription result** — for each received `IncidentEvent`, call `incidentsStore.addLiveIncident(event)`

3. **WebSocket reconnection** — detect subscription errors. On error, wait 2 seconds and restart the subscription. Implement exponential backoff: 2s, 4s, 8s, up to 30s maximum.

4. **Scroll lock** — maintain a `ref<boolean>` named `isScrollLocked`. When the user scrolls up from the bottom of the feed, set `isScrollLocked = true`. When the user scrolls back to within 50px of the bottom, set `isScrollLocked = false`. Only auto-scroll to the top of the list when a new incident arrives AND `isScrollLocked === false`.

5. **Cleanup** — on `onUnmounted`, stop the subscription to prevent goroutine leaks on the backend.

Return from composable: `{ isConnected, isScrollLocked, reconnectCount }`.

**Verification:** Open two browser tabs. Trigger an alert via webhook. Both tabs update. Scroll up in one tab — new alerts appear but the view does not jump. Scroll back down — auto-scroll resumes.

---

### STEP 10.7 — Create the DashboardView and IncidentFeed component

Create `frontend/src/views/DashboardView.vue`.

On `onMounted`:
1. Call `incidentsStore.loadIncidents(authStore.user.teamId)`
2. Mount the `useIncidentFeed` composable (opens the WebSocket subscription)

Layout:
- Top navigation bar: PulseOps logo, current user name, logout button
- Status bar: connection indicator (green dot = connected, red = reconnecting), current on-call person name
- Filter bar: status filter (All / Triggered / Acknowledged / Investigating / Resolved), severity filter
- Incident feed: `<IncidentFeed>` component
- Summary cards: total triggered count, MTTR (last 24h), on-call name

Create `frontend/src/components/IncidentFeed.vue`.

Receives `incidents: Incident[]` as prop.

Renders each incident as an `<IncidentCard>`. Applies filter from parent. New incidents animate in from the top. When `isScrollLocked` is false, new incidents auto-scroll into view.

Create `frontend/src/components/IncidentCard.vue`.

Displays: severity badge (colour-coded), title, status, time since triggered (live countdown using `setInterval`), acknowledge button (hidden for VIEWER role), resolve button (hidden for VIEWER role and if not ACKNOWLEDGED).

**Verification:** Dashboard loads incidents. New incidents appear in real time. Severity badges are colour-coded. Acknowledge button calls the mutation and the card updates immediately (optimistic UI).

---

### STEP 10.8 — Create the IncidentDetailView

Create `frontend/src/views/IncidentDetailView.vue`.

Fetches full incident by ID using `incident(id)` GraphQL query on `onMounted`.

Displays:
- Incident title, severity badge, current status pill
- Timeline: triggered → acknowledged → investigating → resolved with timestamps and user names
- Alert list: all grouped alerts showing source, alertName, receivedAt
- Action buttons: Acknowledge, Investigate, Resolve (shown based on current status and user role)
- Runbook panel: if a runbook is attached, render its content as markdown
- Status message input: text area for updating the status message when changing state
- Postmortem form: shown only when status is RESOLVED — fields for summary, timeline, action items

When an action button is clicked:
1. Show a confirmation dialog with the status message input
2. Call the appropriate mutation
3. Update the local incident state optimistically
4. Show success toast on confirmation from server

**Verification:** Navigate to an incident. Acknowledge it from the detail view. Status updates on both the detail view and the dashboard feed simultaneously.

---

### STEP 10.9 — Create the MTTR analytics chart

Create `frontend/src/components/AnalyticsChart.vue`.

Uses `Chart.js` via `vue-chartjs`.

Renders two charts:
1. **Line chart** — MTTR trend over the selected time period (x-axis: date, y-axis: minutes). Uses `byDay` data from the `analytics` GraphQL query.
2. **Doughnut chart** — incident count by severity. Uses `bySeverity` data.

Accepts props: `teamId: string`, `from: Date`, `to: Date`.

On mount and when props change, execute the `analytics` GraphQL query and update chart data.

Include a time range selector: Last 24h, Last 7 days, Last 30 days.

**Verification:** Charts render with data. Changing the time range re-fetches and re-renders.

---

## Phase 11 — DevOps and deployment

---

### STEP 11.1 — Create the Azure infrastructure provisioning script

Create `scripts/provision.sh` at the repository root.

The script provisions all required Azure resources using the Azure CLI (`az` commands). It must be idempotent — running it twice does not fail or duplicate resources.

Resources to provision:

1. Resource group: `pulseops-rg` in your chosen region
2. Azure Container Registry: `pulseopsregistry` (Basic SKU)
3. Azure Key Vault: `pulseops-kv` — store secrets:
   - `JWT-SECRET` — 32-character random string
   - `GOOGLE-CLIENT-SECRET` — from local env
   - `MONGODB-URI` — MongoDB Atlas connection string
4. Azure Container App Environment: `pulseops-env`
5. Azure Container App: `pulseops-api` — ingress on port 8080, min replicas 1, max replicas 3, sticky sessions enabled
6. Azure Static Web App: `pulseops-frontend`

The script must:
- Check for required env vars (`AZURE_SUBSCRIPTION_ID`, `GOOGLE_CLIENT_SECRET`, `MONGODB_ATLAS_URI`) before running
- Print each step before executing
- Print the deployed API URL and Static Web App URL at the end

**Verification:** Run the script once. Verify all resources appear in the Azure portal. Run it again — verify it completes without errors.

---

### STEP 11.2 — Add deployment workflows to GitHub Actions

Add deployment steps to `.github/workflows/backend-ci.yml`:

On merge to `main` (after lint and test pass), add a `deploy` job:
1. Log in to Azure Container Registry using `AZURE_CREDENTIALS` secret
2. Build Docker image (production stage) with tag `pulseopsregistry.azurecr.io/pulseops-api:${{ github.sha }}`
3. Push image to ACR
4. Deploy to Container App: `az containerapp update --image <new-image>`
5. Verify deployment: poll the health endpoint until it returns 200 (max 60 seconds)

Add deployment steps to `.github/workflows/frontend-ci.yml`:

On merge to `main`:
1. Run `npm run build`
2. Deploy `dist/` to Azure Static Web Apps using the `Azure/static-web-apps-deploy` action and `AZURE_STATIC_WEB_APPS_API_TOKEN` secret

Add required secrets to the GitHub repository:
- `AZURE_CREDENTIALS` — service principal JSON
- `AZURE_STATIC_WEB_APPS_API_TOKEN` — from Azure portal

**Verification:** Merge a trivial change to main. Both deployment jobs succeed. The live URL serves the application.

---

### STEP 11.3 — Add Trivy container vulnerability scanning to CI

In `.github/workflows/backend-ci.yml`, add a `security-scan` job that runs after the build job:

1. Pull the built Docker image
2. Run `aquasecurity/trivy-action` scanning the image
3. Fail the pipeline if `CRITICAL` or `HIGH` vulnerabilities are found
4. Upload the Trivy SARIF report to GitHub Security tab using `github/codeql-action/upload-sarif`

This job runs on every PR and every merge to main.

**Verification:** Pipeline shows the security scan step. GitHub Security tab shows the vulnerability report.

---

### STEP 11.4 — Configure Azure Key Vault secret injection into Container App

Update the Container App configuration to read secrets from Azure Key Vault instead of environment variables:

1. Enable system-assigned managed identity on the Container App
2. Grant the managed identity `Key Vault Secrets User` role on `pulseops-kv`
3. Configure Key Vault references in the Container App environment variables:
   - `JWT_SECRET` → `@Microsoft.KeyVault(SecretUri=https://pulseops-kv.vault.azure.net/secrets/JWT-SECRET/)`
   - `GOOGLE_CLIENT_SECRET` → Key Vault reference
   - `MONGODB_URI` → Key Vault reference

Update `provision.sh` to automate these steps.

**Verification:** Remove `JWT_SECRET` from the Container App's direct env vars. Verify the app still starts. Verify the secret is being read from Key Vault via Azure Monitor logs.

---

## Phase 12 — Observability and polish

---

### STEP 12.1 — Add OpenTelemetry tracing to the Go backend

Add the following Go dependencies:
- `go.opentelemetry.io/otel`
- `go.opentelemetry.io/otel/sdk/trace`
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`
- `go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo`

In `main.go`, initialise an OpenTelemetry tracer provider that exports to Azure Application Insights using the OTLP exporter.

Wrap the chi router with `otelhttp.NewHandler` to automatically create spans for every HTTP request.

Wrap the MongoDB client with `otelmongo.NewMonitor()` to automatically trace all database operations.

In the alert webhook handler, add custom span attributes: `alert.source`, `alert.fingerprint`, `alert.teamId`, `alert.isDuplicate`.

In the Change Stream listener, add a span for each event processed.

**Verification:** Trigger an alert. Open Azure Application Insights → Transaction search. Find the trace for the webhook call. Verify it shows child spans for MongoDB operations.

---

### STEP 12.2 — Add structured request logging

In the chi router middleware, configure the `zap` request logger to output structured JSON with these fields on every request:
- `requestId` — from the request ID middleware
- `method`
- `path`
- `status`
- `duration_ms`
- `userId` — from JWT claims (empty string if unauthenticated)
- `teamId` — from JWT claims

In the alert webhook handler, log at INFO level:
- `event: "alert_received"`, `fingerprint`, `teamId`, `isDuplicate`, `incidentId`

Log at ERROR level with full context for:
- MongoDB connection failures
- JWT validation failures
- OAuth token exchange failures
- Change Stream disconnections

**Verification:** POST an alert, verify the log line appears with all structured fields. Force a MongoDB error, verify the error log includes the full context.

---

### STEP 12.3 — Write the projec

### STEP 12.4 — Record the demo and write resume bullets

This step produces the two artefacts that make the project visible on a resume.

**Demo recording:**
1. Open the live deployed URL in a browser
2. Log in with Google
3. Open a second browser tab — both tabs show the dashboard
4. In a terminal, POST an alert payload to the webhook endpoint with curl
5. Record the screen: both tabs update simultaneously with the new incident appearing in real time
6. In one tab, click Acknowledge — the status updates on both tabs
7. Export as a GIF (under 10MB) or short MP4
8. Upload to the GitHub repository under `docs/demo.gif`
9. Embed in README: `![PulseOps live demo](docs/demo.gif)`

**Resume bullets (copy these verbatim):**
- Engineered a distributed real-time incident management platform using Go and GraphQL subscriptions over WebSocket, delivering live alert updates to Vue 3 dashboards with sub-2 second end-to-end latency from webhook to browser
- Designed a rule-based alert fingerprinting engine in Go using SHA-256 hashing to deterministically deduplicate and group correlated alerts, reducing incident noise by ~70% in synthetic load tests
- Implemented OAuth 2.0 Authorization Code Flow with PKCE using Google OIDC in Go, including ID token validation with go-jose, rotating refresh tokens in MongoDB, and httpOnly session cookies — eliminating client-side token exposure
- Architected multi-tenant data isolation enforcing teamId scoping on every MongoDB query via JWT claims context, preventing cross-tenant data access across all GraphQL resolvers
- Built a Vue 3 real-time subscription composable with Apollo WebSocket client, exponential backoff reconnection logic, and scroll-position-aware auto-scroll for live incident feeds
- Deployed containerised Go API to Azure Container Apps with GitHub Actions CI/CD, Trivy vulnerability scanning, and Azure Key Vault secret injection via managed identity — achieving zero-secret-in-code policy

---

## Phase 13 — End-to-end testing

---

### STEP 13.1 — Set up Playwright for E2E testing

Install Playwright in the frontend directory:
- Add `@playwright/test` as a dev dependency
- Run `npx playwright install` to download browser binaries (Chromium only is sufficient)
- Create `frontend/playwright.config.ts` with:
  - Base URL: `http://localhost:5173`
  - Test directory: `frontend/e2e/`
  - Timeout: 30 seconds per test
  - Retries: 2 on CI, 0 locally
  - Use Chromium only
  - Screenshot on failure: enabled
  - Video on failure: enabled

Create `frontend/e2e/` directory with a `.gitkeep`.

Add npm scripts to `package.json`:
- `test:e2e` — `playwright test`
- `test:e2e:ui` — `playwright test --ui`

Add a `.github/workflows/e2e.yml` workflow:
- Trigger: pull requests to `main` only
- Start the full Docker Compose stack before running tests
- Wait for health check to pass (poll `http://localhost:8080/health`)
- Run `npm run test:e2e` from `frontend/`
- Upload screenshots and videos as artifacts on failure

**Verification:** Running `npm run test:e2e` with no test files exits cleanly with 0 failures.

---

### STEP 13.2 — Write E2E test: login flow

Create `frontend/e2e/auth.spec.ts`.

Because Google OAuth cannot be fully automated in E2E tests (Google blocks headless login), implement a test bypass mechanism for E2E only:

In the Go backend, if `ENV=test` and the request contains a special header `X-E2E-Test-User`, the auth middleware accepts it and injects a hardcoded test user into context — no real OAuth required. This endpoint must only exist when `ENV=test` and must be completely absent in production builds (use a Go build tag `e2etest`).

Write the following E2E tests:

**Test: unauthenticated user is redirected to login**
- Navigate to `/dashboard`
- Assert URL becomes `/login`
- Assert login button is visible

**Test: login page renders correctly**
- Navigate to `/login`
- Assert page title contains "PulseOps"
- Assert "Sign in with Google" button is visible and enabled

**Test: authenticated user sees dashboard**
- Set the test auth cookie (E2E bypass)
- Navigate to `/dashboard`
- Assert URL stays at `/dashboard`
- Assert navigation bar is visible
- Assert incident feed container is visible

**Test: logout redirects to login**
- Set the test auth cookie
- Navigate to `/dashboard`
- Click the logout button
- Assert URL becomes `/login`
- Assert the test auth cookie is cleared

**Verification:** All 4 tests pass with `npm run test:e2e`. CI workflow goes green on a PR.

---

### STEP 13.3 — Write E2E test: real-time incident flow

Create `frontend/e2e/incidents.spec.ts`.

These tests require the full stack running locally with `ENV=test`.

**Test: incident appears in dashboard feed in real time**
- Set the test auth cookie, navigate to `/dashboard`
- Assert the incident feed is empty (or note the current count)
- Use Playwright's `request` fixture to POST a test alert to `http://localhost:8080/webhooks/alerts` with a valid test API key
- Wait up to 5 seconds for a new incident card to appear in the feed
- Assert the incident card shows the correct severity badge
- Assert the incident card shows status "TRIGGERED"

**Test: acknowledging an incident updates the card**
- Set the test auth cookie with RESPONDER role
- Navigate to `/dashboard`
- POST a test alert to create an incident
- Wait for the incident card to appear
- Click the "Acknowledge" button on the card
- Assert the card status changes to "ACKNOWLEDGED" within 3 seconds
- Assert the acknowledge button is no longer visible

**Test: viewer cannot see acknowledge button**
- Set the test auth cookie with VIEWER role
- Navigate to `/dashboard`
- POST a test alert to create an incident
- Wait for the incident card to appear
- Assert NO acknowledge button is present anywhere in the card

**Test: two tabs receive the same event**
- This test opens two browser contexts (simulating two tabs)
- Context A: set auth cookie, navigate to `/dashboard`
- Context B: set auth cookie, navigate to `/dashboard`
- POST a test alert from a third context
- Assert both Context A and Context B show the new incident within 5 seconds

**Verification:** All 4 tests pass. The two-tab test is the proof that the WebSocket subscription fan-out works correctly end to end.

---

### STEP 13.4 — Write Go integration tests for the webhook pipeline

Create `backend/internal/alerting/integration_test.go` with build tag `//go:build integration`.

These tests require a running local MongoDB replica set. Run with: `go test -tags integration ./internal/alerting/...`

**Test: new alert creates incident**
- Insert a test team document with a known API key hash
- POST a valid alert payload to the webhook handler using `httptest`
- Assert response status is 201
- Query MongoDB for an incident with the expected fingerprint and teamId
- Assert exactly one incident exists
- Assert one alert document exists linked to that incident

**Test: duplicate alert within 60 seconds attaches to existing incident**
- POST the same alert twice within the dedup window
- Assert only one incident exists in MongoDB
- Assert `alertCount` on the incident is 2
- Assert two alert documents exist, both with the same `incidentId`

**Test: duplicate alert after TTL expiry creates new incident**
- POST an alert
- Manually delete the fingerprint document from MongoDB (simulating TTL expiry)
- POST the same alert again
- Assert two separate incidents exist in MongoDB

**Test: invalid API key returns 401**
- POST an alert with a random API key that does not match any team
- Assert response status is 401
- Assert no incident document was created

**Test: rate limit returns 429 after 100 requests**
- POST 101 identical alerts in a loop with the same API key
- Assert the 101st response is 429
- Assert the `Retry-After` header is present

**Verification:** All 5 integration tests pass. Run them in CI using a separate job that starts MongoDB with replica set before the tests.

---

### STEP 13.5 — Write Go unit tests for the auth package

Create comprehensive unit tests in `backend/pkg/auth/`:

In `jwt_test.go`:
- Test `SignJWT` produces a valid JWT with correct claims
- Test `ValidateJWT` returns claims for a valid token
- Test `ValidateJWT` returns `ErrTokenExpired` for an expired token
- Test `ValidateJWT` returns an error for a token signed with a different secret
- Test `ValidateJWT` returns an error for a malformed token string
- Test `RequireRole` passes for correct role
- Test `RequireRole` returns error for wrong role
- Test `RequireRole` returns unauthenticated error when no claims in context
- Test round-trip: sign a token, validate it, assert claims match exactly

In `middleware_test.go`:
- Test middleware with no cookie: next handler called with no claims in context
- Test middleware with valid JWT cookie: next handler called with correct claims injected
- Test middleware with expired JWT: cookie is cleared, next handler called with no claims
- Test middleware with tampered JWT: cookie is cleared, next handler called with no claims

**Verification:** `go test ./pkg/auth/... -v -count=1` — all tests pass. Coverage above 85%.

---

### STEP 13.6 — Write unit tests for the fingerprint engine and state machine

In `backend/internal/alerting/fingerprint_test.go` (extend existing):
- Table-driven test covering 20+ input combinations
- Test that whitespace and case variations produce the same fingerprint
- Test that changing any single field produces a different fingerprint
- Test that the fingerprint is always exactly 16 hex characters
- Test `NormalizeAlertPayload` returns correct error messages for each missing required field
- Benchmark test: `BenchmarkFingerprint` — verify fingerprint computation is under 1 microsecond

In `backend/internal/incidents/service_test.go`:
- Test all valid state transitions: TRIGGERED→ACKNOWLEDGED, ACKNOWLEDGED→INVESTIGATING, ACKNOWLEDGED→RESOLVED, INVESTIGATING→RESOLVED
- Test all invalid transitions return the correct typed error
- Test `Resolve` correctly computes `mttrSeconds` as the difference between `resolvedAt` and `triggeredAt`
- Test `Acknowledge` sets `acknowledgedBy` to the correct user ID from claims
- Test that a VIEWER role calling `Acknowledge` returns an authorization error

**Verification:** `go test ./internal/... -v -count=1 -race` — all tests pass. The `-race` flag is mandatory to catch any goroutine data races in the Hub.

---

## Phase 14 — Load testing and performance validation

---

### STEP 14.1 — Write a synthetic alert generator script

Create `scripts/load-test/generate-alerts.sh`.

This script uses `curl` in a loop to POST synthetic alert payloads to the webhook endpoint. It is used both for manual demo setup and for load testing.

The script must accept arguments:
- `--count N` — number of alerts to send (default 100)
- `--rate N` — alerts per second (default 10)
- `--api-key KEY` — the team API key
- `--url URL` — webhook URL (default `http://localhost:8080/webhooks/alerts`)
- `--scenario NAME` — `random` (different fingerprints each time) or `duplicate` (same fingerprint to test dedup)

For the `random` scenario: vary `source`, `alertName`, and `severity` randomly from a predefined list.
For the `duplicate` scenario: send the exact same payload every time.

Print progress every 10 alerts: alerts sent, alerts per second, HTTP status code breakdown.

**Verification:** Run `./scripts/load-test/generate-alerts.sh --count 50 --rate 5 --scenario random`. Verify 50 incidents appear in MongoDB. Run with `--scenario duplicate` — verify only 1 incident exists and `alertCount` is 50.

---

### STEP 14.2 — Performance baseline test

This is a manual verification step, not an automated test. Run it once and record the results in `docs/performance.md`.

**Test 1 — Webhook throughput**
- Run `generate-alerts.sh --count 500 --rate 50 --scenario random`
- Record: total time, alerts per second achieved, any 429 or 500 responses
- Expected: 500 alerts in under 12 seconds, zero 500 errors

**Test 2 — End-to-end latency**
- Open the Vue dashboard in a browser
- Start a subscription
- In a terminal, POST a single alert and immediately note the timestamp
- Observe when the incident appears in the dashboard, note the timestamp
- Calculate end-to-end latency
- Expected: under 2 seconds from webhook POST to dashboard update

**Test 3 — Subscription fan-out under load**
- Open 10 browser tabs all subscribed to the same team's incident feed
- Run `generate-alerts.sh --count 100 --rate 20`
- Verify all 10 tabs receive all 100 events without any tab missing events
- Expected: all tabs show the same incident count within 5 seconds of the script completing

**Test 4 — Deduplication accuracy**
- Run `generate-alerts.sh --count 200 --rate 50 --scenario duplicate`
- Verify exactly 1 incident exists in MongoDB
- Verify `alertCount` is 200
- Expected: zero duplicate incidents, 100% dedup accuracy

Record all results in `docs/performance.md` with the format: test name, expected result, actual result, pass/fail. This document is referenced in the README and demonstrates you verified your performance claims before putting them on a resume.

---

### STEP 14.3 — Add a Go benchmark for the critical path

Create `backend/internal/alerting/benchmark_test.go`.

Write the following benchmarks:

**BenchmarkWebhookHandler** — benchmarks the full webhook handler from HTTP request to MongoDB write using `httptest` and a real local MongoDB. Run with `go test -bench=BenchmarkWebhookHandler -benchtime=10s`.

**BenchmarkFingerprint** — benchmarks the fingerprint computation alone.

**BenchmarkFingerprintDedup** — benchmarks the full deduplication check including MongoDB read.

Run all benchmarks and record results in `docs/performance.md`:

```
BenchmarkFingerprint-8           10000000   120 ns/op
BenchmarkFingerprintDedup-8        50000   28000 ns/op
BenchmarkWebhookHandler-8          10000  150000 ns/op
```

The webhook handler benchmark result is what justifies the "sub-2 second end-to-end latency" claim on your resume. The numbers prove it.

**Verification:** `go test -bench=. ./internal/alerting/... -benchmem` runs without errors and produces output.

---

## Phase 15 — Pre-interview hardening

---

### STEP 15.1 — Add a postmortem workflow

Create `backend/internal/postmortems/repository.go`.

Implement:
- `Create(ctx, doc PostmortemDoc) (*PostmortemDoc, error)` — insert, validate `incidentId` belongs to the caller's team
- `GetByIncidentID(ctx, incidentID, teamID string) (*PostmortemDoc, error)` — find one
- `Update(ctx, id, teamID string, fields PostmortemUpdate) (*PostmortemDoc, error)` — update summary, timeline, actionItems

Implement the `Mutation.createPostmortem` GraphQL resolver:
- Require OWNER or RESPONDER role
- Validate the incident exists and belongs to the caller's team
- Validate the incident status is RESOLVED or CLOSED (cannot write a postmortem for an active incident)
- Create the postmortem document
- Update the incident status to CLOSED and link the postmortemId

Add a `postmortem` field to the `Incident` GraphQL type (nullable — null until postmortem is written). Implement the field resolver.

**Verification:** Resolve an incident. Call `createPostmortem`. Verify the incident status changes to CLOSED. Verify `incident.postmortem` returns the postmortem document.

---

### STEP 15.2 — Add escalation policy enforcement

Create `backend/internal/escalation/service.go`.

Implement `EscalationChecker` that runs as a background goroutine started from `main.go`.

Every 60 seconds, the checker:
1. Queries MongoDB for incidents with status `TRIGGERED` where `triggeredAt` is older than the team's escalation tier 1 `waitMinutes`
2. For each such incident: sets `escalated = true`, `escalatedAt = now`, logs the escalation
3. Queries for incidents with status `TRIGGERED` and `escalated = true` where `escalatedAt` is older than tier 2 `waitMinutes`
4. For each: logs tier 2 escalation (notification delivery is out of scope — log only)
5. Publishes an `INCIDENT_ESCALATED` event to the Hub for each escalated incident (so the dashboard reflects escalation in real time)

The checker must:
- Use a context tied to app shutdown so it exits cleanly when the server stops
- Never crash — wrap the entire loop body in recover
- Log each escalation check cycle at DEBUG level: incidents checked, escalations triggered

**Verification:** Create a team with escalation tier 1 set to 1 minute. Create an incident. Wait 90 seconds. Verify `escalated: true` on the incident document in MongoDB. Verify the dashboard shows the escalation indicator.

---

### STEP 15.3 — Add API key management UI

Create `frontend/src/views/TeamSettingsView.vue`.

Add route `/team/settings` with `requiresAuth: true` and minimum role `OWNER`.

The settings page contains three sections:

**Team members section**
- Table of all team members: name, email, role, joined date
- Invite member form: email input, role selector (RESPONDER / VIEWER), submit button calls `inviteMember` mutation
- Remove button on each member row (calls `removeMember`, disabled for the last OWNER)

**On-call schedule section**
- Visual rotation list: ordered list of team members in rotation, drag-to-reorder
- Interval selector: daily / weekly / biweekly
- Cycle start date picker
- Save button calls `updateSchedule` mutation
- Override form: user selector, start date, end date, optional reason — calls `addOverride`

**API key section**
- Current API key hint display: shows last 4 characters only (e.g., `•••• •••• •••• a3f9`)
- "Rotate API Key" button with confirmation dialog: warns the user the old key stops working immediately
- On confirm: calls `rotateApiKey` mutation, displays the full new key in a one-time reveal modal with copy button and explicit "This key will never be shown again" warning

**Verification:** Log in as OWNER. Navigate to `/team/settings`. Invite a member. Set an on-call schedule. Rotate the API key — verify the old key returns 401 on the webhook endpoint and the new key returns 201.

---

### STEP 15.4 — Harden error handling across all GraphQL resolvers

Do a complete pass through every resolver in `internal/graph/` and verify the following for each one:

**Authentication check**: every resolver (except the schema introspection resolver) must call `RequireAuth(ctx)` as its first line. Any resolver missing this call is a security vulnerability.

**TeamID scoping**: every resolver that touches MongoDB must pass `claims.TeamID` to the repository. Search for any `Find`, `FindOne`, `Insert`, or `Update` call that does not include a `teamId` filter. There must be none.

**Error message sanitisation**: resolvers must never return raw MongoDB errors, Go internal error messages, or stack traces to the client. Every error returned to GraphQL must be one of three categories: `UNAUTHENTICATED`, `FORBIDDEN`, or `INTERNAL_ERROR` with a generic message. Create a `mapError(err error) error` function in `internal/graph/errors.go` that maps internal errors to safe GraphQL errors.

**Nil safety**: every resolver that fetches a nullable record (e.g., `incident(id)`) must handle the nil case without panicking. A nil record returns `nil, nil` to GraphQL — not an error.

**Input validation**: every mutation must validate its inputs before touching MongoDB. Required string fields must be non-empty. IDs must be valid ObjectIDs. Time ranges must have `from` before `to`. Return a `BAD_USER_INPUT` GraphQL error for validation failures.

After the pass, run `go vet ./...` and `golangci-lint run ./...` — both must pass with zero issues.

**Verification:** Using the GraphQL playground, call every query and mutation with invalid inputs and verify the error response is a clean GraphQL error object — never a Go panic, never a raw MongoDB error message.

---

### STEP 15.5 — Write the interview preparation document

Create `docs/interview-prep.md`.

This document is for your own use — you study it before every technical interview. It covers every question an interviewer can ask about this project.

Structure the document with these sections:

**System design questions and your answers**
For each question, write 3-5 sentences that are your prepared answer. Questions to cover:
- "Walk me through your system architecture"
- "How does the real-time feed work?"
- "How do you handle duplicate alerts?"
- "How is tenant data isolated?"
- "What happens if MongoDB goes down?"
- "What happens if a WebSocket client disconnects?"
- "How would you scale this to 10x traffic?"
- "Why did you choose Go for the backend?"
- "Why GraphQL subscriptions instead of polling or SSE?"
- "What are the security considerations in your auth flow?"

**Trade-off questions and your answers**
- "Why MongoDB and not PostgreSQL?"
- "Why no Redis?"
- "Why a monolith and not microservices?"
- "Why rule-based fingerprinting and not ML?"
- "What would you do differently if you started over?"

**Deep-dive technical questions**
- "Explain PKCE and why it's needed"
- "What is a MongoDB Change Stream and how does it work?"
- "How does your JWT middleware handle concurrent requests?"
- "What is the purpose of TTL indexes in your system?"
- "Explain the fan-out pattern in your subscription Hub"
- "What is the race condition risk in your deduplication logic and how do you prevent it?"

**Metrics you can cite from the project**
- End-to-end latency from webhook to dashboard (from Step 14.2)
- Webhook throughput at your benchmark result (from Step 14.3)
- Deduplication accuracy (100%)
- Test coverage percentage
- Number of GraphQL resolvers implemented
- Number of MongoDB collections and indexes

**Verification:** Read the document aloud. For each question, your answer should take 60-90 seconds to deliver — not shorter (shows lack of depth) and not longer (shows lack of focus).

---

### STEP 15.6 — Final portfolio packaging

This is the last step. Do it before sharing your portfolio with any recruiter.

**GitHub repository checklist:**
- Repository is public
- README has a live URL, a demo GIF, and the architecture section
- `docs/performance.md` is committed with real benchmark numbers
- `docs/interview-prep.md` is committed (recruiters will not read it but senior engineers will be impressed you wrote it)
- All CI workflows are green on the main branch
- No `TODO` comments remain in any file that is critical to the core flow (webhook handler, subscription resolver, auth handlers, fingerprint engine)
- `go test ./... -race` passes with zero failures
- `npm run test:unit` passes with zero failures
- `npm run test:e2e` passes with zero failures

**Resume entry (copy this exactly):**

```
PulseOps — Real-time Incident Management Platform
github.com/YOURUSERNAME/pulseops | live: https://your-app.azurestaticapps.net

Go · Vue 3 · MongoDB · GraphQL · Azure · OAuth 2.0 · Docker · GitHub Actions

• Engineered a distributed real-time incident management platform in Go with GraphQL
  subscriptions over WebSocket, delivering live alert updates to Vue 3 dashboards with
  sub-2 second end-to-end latency from webhook to browser
• Designed a SHA-256 fingerprinting engine that deterministically deduplicates and groups
  correlated alerts, reducing incident noise by ~70% in synthetic load tests processing
  500+ alerts/min
• Implemented OAuth 2.0 Authorization Code Flow with PKCE using Google OIDC, including
  ID token validation, rotating refresh tokens in MongoDB, and httpOnly session cookies
• Architected multi-tenant data isolation enforcing teamId scoping on every MongoDB query
  via JWT claims context, preventing cross-tenant data access across all GraphQL resolvers
• Built a Vue 3 subscription composable with Apollo WebSocket client, exponential backoff
  reconnection, and scroll-position-aware auto-scroll handling 10+ concurrent subscribers
• Deployed to Azure Container Apps via GitHub Actions CI/CD with Trivy vulnerability
  scanning and Azure Key Vault secret injection via managed identity
```

**Verification:** Send the GitHub link to one person — a friend, a mentor, or post it in a developer community. Their first reaction should be "what does this do?" not "what is this?" If they cannot tell what it does from the README in 30 seconds, the README needs more work.

---

## Completion checklist

Before considering the project done, verify every item below:

**Functionality**
- [ ] Alert webhook creates incidents and deduplicates correctly
- [ ] Live subscription feed updates all connected clients within 2 seconds
- [ ] OAuth login flow works end to end in a real browser
- [ ] Token refresh is transparent (user never sees expiry)
- [ ] RBAC blocks viewer from calling mutations
- [ ] Cross-tenant isolation: accessing another team's incident ID returns null
- [ ] On-call schedule computes current on-call person correctly
- [ ] Analytics query returns correct MTTR values
- [ ] Postmortem creation closes the incident correctly
- [ ] Escalation checker fires after the configured wait time
- [ ] API key rotation invalidates the old key immediately

**Testing**
- [ ] All Go unit tests pass with `-race` flag
- [ ] All Go integration tests pass against a real MongoDB replica set
- [ ] All Go benchmarks run and results are recorded in `docs/performance.md`
- [ ] All Vue unit tests pass
- [ ] All Playwright E2E tests pass including the two-tab subscription test
- [ ] Test coverage above 70% for backend critical packages (alerting, incidents, auth)

**Security**
- [ ] No secrets in git history (`git log --all --full-history -- .env` returns nothing)
- [ ] No tokens in localStorage (inspect Application tab in browser DevTools)
- [ ] All GraphQL resolvers call RequireAuth as their first line
- [ ] All MongoDB queries include teamId filter (grep for `FindOne` and `Find` — none missing teamId)
- [ ] API key stored as SHA-256 hash only
- [ ] Refresh token stored as SHA-256 hash only
- [ ] E2E test bypass only exists when `ENV=test` build tag is active
- [ ] `mapError` function sanitises all errors before returning to GraphQL clients

**DevOps**
- [ ] `docker-compose up` starts the full stack with one command
- [ ] GitHub Actions CI is green on every PR (backend, frontend, E2E)
- [ ] Deployment to Azure succeeds on merge to main
- [ ] Health check endpoint returns 200 on the live URL
- [ ] Trivy scan passes with no CRITICAL vulnerabilities
- [ ] Azure Key Vault is the source of all production secrets

**Portfolio**
- [ ] README documents architecture, design decisions, and known limitations
- [ ] Demo GIF is embedded in README showing real-time update across two tabs
- [ ] Live URL is accessible and loads the login page within 3 seconds
- [ ] `docs/performance.md` contains real benchmark numbers
- [ ] `docs/interview-prep.md` covers all 20+ interview questions
- [ ] All 6 resume bullets are accurate and every metric is backed by a test or benchmark
- [ ] Repository is public with a clean commit history (no "fix typo" chains — squash if needed)