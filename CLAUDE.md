# Emdash Agents — Internal Integration Platform

## Overview
Internal admin tool for managing AI agent integrations. Go CLI binary (`integrations`) that AI agents call inside Docker containers to interact with external services. Includes an admin-only Next.js portal for centralized credential management, and a Go orchestrator that deploys Claude Agent SDK containers to Kubernetes.

Mark owns all integrations centrally. Clients get specialized agents configured via the admin dashboard. Session-bound integrations (Instagram, LinkedIn, X, Canvas, Zillow) use Playwright browser automation for cookie capture — no manual cookie pasting or Chrome extensions.

## Quick Start
```bash
make build          # → bin/integrations
make test           # run tests with coverage
make lint           # go vet

# Orchestrator
make orchestrator       # → bin/orchestrator
make orchestrator-dev   # run with doppler (localhost:8080)
make kind-setup         # create local k8s cluster
make sync-templates     # sync agents/*/template.yaml → Supabase

# Docker images
make docker-agent-base    # build agent base image
make docker-export-creds  # build export-creds image

# Portal
make portal-install             # npm install
make portal-dev                 # npm run dev (localhost:3000)
make portal-build               # npm run build
make portal-lint                # npm run lint
make portal-playwright-install  # install Playwright chromium for session capture
```

## CLI Usage
Run `bin/integrations <provider> --help` for full command reference. Each provider has nested subcommands with `--json` output and `--dry-run` for write operations.

## Providers

| Provider | Auth Type | Package | Key Files | Coverage |
|----------|-----------|---------|-----------|----------|
| Gmail | Google OAuth (auto-refresh) | `internal/providers/gmail/` | gmail.go, messages.go, helpers.go | 93.2% |
| Google Sheets | Google OAuth (dual: Sheets + Drive) | `internal/providers/sheets/` | sheets.go, spreadsheets.go, values_read.go, values_write.go, tabs.go | 85.5% |
| Google Calendar | Google OAuth | `internal/providers/calendar/` | calendar.go, events.go, calendars.go, freebusy.go | 92.9% |
| Google Drive | Google OAuth | `internal/providers/drive/` | drive.go, files.go, permissions.go | 88.9% |
| Google Places | Scraper (no API key) | `internal/providers/places/` | places.go, scraper.go, search.go, lookup.go | 94.6% |
| GitHub | GitHub OAuth (auto-refresh) | `internal/providers/github/` | github.go, helpers.go + 13 resource files | 85.8% |
| Instagram | Cookie session (web + mobile APIs) | `internal/providers/instagram/` | instagram.go, client.go + 17 resource files | 85.0% |
| LinkedIn | Cookie session (Voyager API) | `internal/providers/linkedin/` | linkedin.go, client.go + 17 resource files | 86.5% |
| Framer | API key + Node.js bridge | `internal/providers/framer/` | framer.go, bridge_client.go + 15 resource files | 80.5% |
| Supabase | OAuth 2.1 with PKCE | `internal/providers/supabase/` | supabase.go, helpers.go + 16 resource files | 82.5% |
| X (Twitter) | Cookie session (GraphQL + v1.1) | `internal/providers/x/` | x.go, client.go + 17 resource files | 84.2% |
| iMessage | BlueBubbles REST API | `internal/providers/imessage/` | imessage.go, client.go + 13 resource files | 83.9% |
| Canvas LMS | Cookie session (`/api/v1/`) | `internal/providers/canvas/` | canvas.go, client.go + 29 resource files | 80.0% |
| Zillow | Cookie session (internal APIs) | `internal/providers/zillow/` | zillow.go, client.go + 12 resource files | 86.9% |

## Architecture
- `cmd/integrations/main.go` — entrypoint, registers providers
- `internal/cli/` — Cobra root command, output helpers (JSON/text)
- `internal/auth/` — Google OAuth + GitHub OAuth token management with auto-refresh
- `internal/providers/provider.go` — Provider interface

### Provider Pattern
Each provider follows a consistent pattern:
- **provider.go** — Provider struct, `RegisterCommands()` with nested Cobra subcommand groups
- **client.go** — HTTP client (session-based providers: CSRF rotation, rate limit detection)
- **helpers.go** — Shared types (summaries, details) and formatter functions
- **Resource files** — One file per API resource group (e.g., posts.go, comments.go)
- **mock_server_test.go** — `httptest.NewServer` mock, `captureStdout`, `newTestRootCmd`

### Testing Pattern
- All providers use `ServiceFactory`/`ClientFactory` for dependency injection
- Tests use `httptest.NewServer` to mock APIs via `newFullMockServer(t)`
- Orchestrator uses `sqlmock` + `fake.NewSimpleClientset()` for DB and K8s tests
- Coverage target: 80%+

### Command Aliases
Most resource subcommands have short aliases (e.g., `msg` for `messages`, `ev` for `events`, `pr` for `pulls`). Provider aliases: `ig` (Instagram), `li` (LinkedIn), `fr` (Framer), `sb` (Supabase), `twitter` (X), `imsg` (iMessage), `cvs` (Canvas), `zw` (Zillow).

## Web Portal (Next.js 15 + Supabase) — Admin Only

- `portal/` — Next.js 15 App Router, TypeScript, Tailwind CSS
- Auth: Supabase Auth with Google OAuth, admin-only access enforced in `proxy.ts`
- Database: Supabase PostgreSQL with RLS policies
- Integration tokens: AES-256-GCM encrypted, stored in `user_integrations` table
- Session-bound auth: Playwright browser automation captures cookies (Instagram, LinkedIn, X, Canvas, Zillow)
- OAuth providers: Google, GitHub, Supabase use standard connect flows
- Manual config: Framer (API key), iMessage/BlueBubbles (server URL + password)

### Key Portal Paths
```
portal/
  app/integrations/page.tsx          # Integration dashboard with provider cards
  app/admin/clients/page.tsx         # Client management CRUD
  app/api/integrations/playwright/   # Session capture API (connect, status)
  app/api/integrations/{provider}/   # OAuth connect/callback per provider
  lib/playwright/session-capture.ts  # Core: launch browser, poll cookies
  lib/playwright/providers/          # Per-provider cookie configs
  lib/crypto.ts                      # AES-256-GCM (Go-compatible wire format)
  lib/credentials.ts                 # resolveAdminCredentials() — single-tenant
  proxy.ts                           # Admin-only middleware (ADMIN_EMAILS check)
```

### Token Bridge (Go)
```
internal/tokenbridge/
  crypto.go       # AES-256-GCM decrypt (Go side)
  bridge.go       # ExportEnvForUser() → reads integrations table, decrypts to env map
```

### Cross-Language Encryption Wire Format
`base64(nonce [12 bytes] || ciphertext || auth_tag [16 bytes])`
Shared key: `ENCRYPTION_MASTER_KEY` (64 hex chars = 32 bytes)

## Orchestrator

```
Portal (Next.js)  ──HTTP──►  Orchestrator (Go :8080)  ──K8s API──►  Agent Pods
                                    │                                    │
                                    ├── Supabase (templates, instances)  │
                                    └── tokenbridge (decrypt creds)──────┘
                                                                   (init container)
```

### Orchestrator REST API
All endpoints under `/api/v1/`, auth via Supabase JWT.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/templates` | List active templates |
| GET | `/templates/{id}` | Template details |
| POST | `/agents/deploy` | Deploy agent |
| GET | `/agents` | List user's instances |
| GET | `/agents/{id}` | Instance status |
| GET | `/agents/{id}/logs` | Stream logs (SSE) |
| POST | `/agents/{id}/stop` | Stop running agent |
| DELETE | `/agents/{id}` | Delete stopped instance |

### Key Orchestrator Paths
```
cmd/orchestrator/main.go              # HTTP server entrypoint
cmd/sync-templates/main.go            # Template sync CLI
internal/orchestrator/
  store.go, k8s.go, pod_spec.go       # DB, K8s, pod builder
  credentials.go                       # Credential resolution (reuses tokenbridge)
  server.go, handlers.go              # chi router + REST handlers
  reconciler.go                        # Background pod status sync
```

### Agent Templates
```
agents/email-assistant/
  template.yaml      # name, description, required_integrations, docker_image
  role.md, CLAUDE.md, entrypoint.py, requirements.txt
```

## Environment Variables
```
# Google (Gmail, Sheets, Calendar, Drive)
GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_ACCESS_TOKEN, GOOGLE_REFRESH_TOKEN

# Google Places — GOOGLE_MAPS_SCRAPER_BIN (optional, falls back to PATH)

# Instagram — INSTAGRAM_SESSION_ID, INSTAGRAM_CSRF_TOKEN, INSTAGRAM_DS_USER_ID
#   Optional: INSTAGRAM_MID, INSTAGRAM_IG_DID, INSTAGRAM_USER_AGENT

# LinkedIn — LINKEDIN_LI_AT, LINKEDIN_JSESSIONID
# Framer — FRAMER_API_KEY, FRAMER_PROJECT_URL
# GitHub — GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, GITHUB_ACCESS_TOKEN, GITHUB_REFRESH_TOKEN
# iMessage — BLUEBUBBLES_URL, BLUEBUBBLES_PASSWORD
# Supabase — SUPABASE_INTEGRATION_CLIENT_ID, SUPABASE_INTEGRATION_CLIENT_SECRET, SUPABASE_ACCESS_TOKEN, SUPABASE_REFRESH_TOKEN
# X (Twitter) — X_AUTH_TOKEN, X_CSRF_TOKEN
# Canvas LMS — CANVAS_BASE_URL, CANVAS_SESSION_COOKIE, CANVAS_CSRF_TOKEN
# Zillow — ZILLOW_COOKIES, ZILLOW_PROXY_URL (optional)

# Orchestrator
SUPABASE_DB_URL, ENCRYPTION_MASTER_KEY, SUPABASE_JWT_SECRET
PORT (default: 8080), KUBE_NAMESPACE (default: agents)
AGENT_BASE_IMAGE, EXPORT_CREDS_IMAGE, ANTHROPIC_API_KEY_SECRET

# Portal — ORCHESTRATOR_URL (required for chat API)
```
