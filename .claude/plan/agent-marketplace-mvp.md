# Implementation Plan: Agent Marketplace MVP

## Overview

A marketplace platform where an admin creates AI agents with specific roles/jobs, and users purchase those agents to automate tasks. Each agent is a Claude Code instance with skills loaded, running in its own Dockerized workspace in the cloud, using a custom CLI for integrations instead of MCP servers or browser automation.

## Task Type
- [x] Backend
- [x] Frontend
- [x] Fullstack (→ Parallel)

---

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────┐
│                    Frontend (Next.js)                │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │ Storefront│  │  Admin   │  │  User Dashboard   │  │
│  │ (Browse)  │  │  Panel   │  │ (My Agents/Logs)  │  │
│  └──────────┘  └──────────┘  └───────────────────┘  │
└─────────────────────┬───────────────────────────────┘
                      │ REST + WebSocket
┌─────────────────────▼───────────────────────────────┐
│                  API Server (FastAPI)                │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │  Auth    │  │ Agents   │  │  Billing/Stripe   │  │
│  │  Module  │  │ CRUD     │  │  Module           │  │
│  └──────────┘  └──────────┘  └───────────────────┘  │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │ Scheduler│  │ Execution│  │  Integration      │  │
│  │ (cron)   │  │ Engine   │  │  Registry         │  │
│  └──────────┘  └──────────┘  └───────────────────┘  │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│              Agent Orchestration Layer               │
│  ┌──────────────────────────────────────────────┐   │
│  │  Docker Manager (per-user agent containers)  │   │
│  │  ┌────────────────────────────────────────┐  │   │
│  │  │  Agent Container (Claude Code + Skills)│  │   │
│  │  │  ┌──────────┐  ┌───────────────────┐   │  │   │
│  │  │  │ Skills   │  │  Integration CLI  │   │  │   │
│  │  │  │ (.md)    │  │  (custom binary)  │   │  │   │
│  │  │  └──────────┘  └───────────────────┘   │  │   │
│  │  └────────────────────────────────────────┘  │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                   Data Layer                        │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │ Postgres │  │  Redis   │  │  S3 (logs/assets) │  │
│  │ (primary)│  │ (queue)  │  │                   │  │
│  └──────────┘  └──────────┘  └───────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Key Design Decisions

1. **FastAPI backend** — async-native, great for managing concurrent agent executions
2. **Next.js frontend** — SSR for storefront SEO, client-side for dashboards
3. **Docker per agent instance** — isolation, reproducibility, resource limits
4. **Custom Integration CLI** — lightweight binary agents call instead of MCP/Playwright
5. **Redis for job queue** — schedule and dispatch agent runs
6. **Postgres** — relational data (users, agents, subscriptions, execution logs)
7. **Stripe** — subscription billing for agent purchases

---

## Tech Stack

| Layer | Technology | Rationale |
|-------|------------|-----------|
| Frontend | Next.js 15 + TypeScript + Tailwind | SSR for storefront, great DX |
| Backend API | FastAPI (Python 3.12) | Async, fast, typed |
| Database | PostgreSQL 16 + SQLAlchemy | Relational data, migrations |
| Cache/Queue | Redis + Bull (or ARQ for Python) | Job scheduling, caching |
| Auth | Clerk or NextAuth + JWT | Quick auth MVP |
| Payments | Stripe Subscriptions | Recurring billing |
| Container Orchestration | Docker API (docker-py) | Agent isolation |
| Agent Runtime | Claude Code CLI + custom skills | Core product |
| Integration CLI | Go binary | Fast, single binary, easy to distribute |
| Cloud | Hetzner or AWS EC2 | Cost-effective for Docker hosts |
| Logs/Storage | S3-compatible (MinIO for dev) | Agent output storage |

---

## Data Model

### Core Tables

```sql
-- Users (customers)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    stripe_customer_id VARCHAR(255),
    role VARCHAR(20) DEFAULT 'user', -- 'user' | 'admin'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Agent Templates (admin-created blueprints)
CREATE TABLE agent_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100), -- 'productivity', 'communication', 'dev-tools'
    icon_url VARCHAR(500),
    skills JSONB NOT NULL DEFAULT '[]',       -- skill file paths/content
    system_prompt TEXT NOT NULL,               -- agent role/personality
    integrations JSONB NOT NULL DEFAULT '[]',  -- required integrations
    schedule_type VARCHAR(50),                 -- 'cron', 'on-demand', 'event-driven'
    default_schedule VARCHAR(100),             -- e.g., '0 7 * * *' for daily 7am
    price_monthly_cents INTEGER NOT NULL,
    is_published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User Agent Instances (purchased/active agents)
CREATE TABLE agent_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    template_id UUID REFERENCES agent_templates(id),
    status VARCHAR(20) DEFAULT 'provisioning', -- 'provisioning'|'active'|'paused'|'error'
    container_id VARCHAR(255),                  -- Docker container ID
    config JSONB DEFAULT '{}',                  -- user-specific config overrides
    schedule VARCHAR(100),                      -- user's cron schedule
    stripe_subscription_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Integration Credentials (user's connected services)
CREATE TABLE integration_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    integration_type VARCHAR(100) NOT NULL,     -- 'gmail', 'slack', 'notion'
    credentials_encrypted BYTEA NOT NULL,       -- encrypted OAuth tokens / API keys
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, integration_type)
);

-- Execution Logs
CREATE TABLE execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_instance_id UUID REFERENCES agent_instances(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'running',       -- 'running'|'success'|'error'
    output_summary TEXT,
    output_url VARCHAR(500),                    -- S3 link to full output
    tokens_used INTEGER,
    cost_cents INTEGER,
    error_message TEXT
);
```

---

## Integration CLI Design

The Integration CLI is a Go binary that agents invoke instead of MCP/Playwright:

```
# Example usage by an agent
integrations gmail list-unread --limit=50 --since="24h"
integrations gmail send --to="user@example.com" --subject="Daily Digest" --body-file=/tmp/digest.md
integrations slack post --channel="#general" --message="Good morning!"
integrations notion query --database="Tasks" --filter='{"status":"In Progress"}'
integrations calendar events --date="today"
```

### CLI Architecture

```
integrations/
├── cmd/
│   └── integrations/
│       └── main.go              # Entry point
├── internal/
│   ├── cli/
│   │   └── root.go              # Cobra root command
│   ├── providers/
│   │   ├── provider.go          # Provider interface
│   │   ├── gmail/
│   │   │   ├── gmail.go         # Gmail provider
│   │   │   ├── list.go          # list-unread command
│   │   │   └── send.go          # send command
│   │   ├── slack/
│   │   │   └── slack.go
│   │   ├── notion/
│   │   │   └── notion.go
│   │   └── calendar/
│   │       └── calendar.go
│   └── auth/
│       └── credentials.go       # Read creds from mounted secrets
├── go.mod
└── go.sum
```

Key principles:
- **Single binary** compiled into each agent container
- **Stdout/stderr** for output — agents parse CLI output naturally
- **JSON output mode** (`--json`) for structured data
- **Credentials mounted** as env vars or files in the container (never hardcoded)
- **Rate limiting** built into each provider
- **Dry-run mode** for testing

### MVP Integrations (Phase 1)

1. **Gmail** — list, read, send, search
2. **Google Calendar** — list events, create events
3. **Slack** — post message, read channels

---

## Agent Container Design

Each agent instance gets a Docker container:

```dockerfile
FROM ubuntu:24.04

# Install Claude Code CLI
RUN curl -fsSL https://claude.ai/install.sh | sh

# Install Integration CLI
COPY integrations /usr/local/bin/integrations

# Create workspace
RUN mkdir -p /workspace /skills /config

# Copy agent skills
COPY skills/ /skills/

# Copy CLAUDE.md with agent system prompt + instructions
COPY CLAUDE.md /workspace/CLAUDE.md

# Entry point: run claude with the agent's task
ENTRYPOINT ["/usr/local/bin/claude", "--workspace", "/workspace"]
```

Container lifecycle:
1. **Provisioning**: Build image with agent's skills, create container
2. **Execution**: Start container, run Claude Code with task prompt
3. **Capture**: Stream stdout/stderr to execution logs
4. **Cleanup**: Stop container after task completes (keep for on-demand agents)

---

## Implementation Steps

### Phase 1: Foundation (Week 1-2)

#### Step 1: Project Setup
- Initialize monorepo structure
- Set up FastAPI backend with project skeleton
- Set up Next.js frontend with Tailwind
- Configure Docker Compose for local dev (Postgres, Redis)
- **Deliverable**: Running dev environment with health check endpoints

#### Step 2: Database & Auth
- Define SQLAlchemy models matching data model above
- Set up Alembic migrations
- Implement Clerk auth (or simple JWT auth for MVP)
- Admin vs user role middleware
- **Deliverable**: Auth flow working, DB seeded with admin user

#### Step 3: Agent Templates CRUD (Admin)
- Admin API: create, update, publish, unpublish agent templates
- Admin dashboard UI: form to create agent templates
- Storefront: list published agents with details
- **Deliverable**: Admin can create agents, users can browse

### Phase 2: Core Engine (Week 3-4)

#### Step 4: Integration CLI (Go)
- Scaffold Go CLI with Cobra
- Implement Gmail provider (OAuth2 + API)
- Implement Google Calendar provider
- Implement Slack provider (Bot token)
- Add `--json` output mode
- **Deliverable**: CLI binary that can send/read emails, list calendar, post to Slack

#### Step 5: Agent Container Runtime
- Build base Docker image with Claude Code + Integration CLI
- Implement container provisioning in FastAPI (docker-py)
- Skill injection: copy skill .md files into container
- Credential mounting: inject user's integration creds as env vars
- Agent execution: invoke Claude Code with task prompt, capture output
- **Deliverable**: Can spin up an agent container, run a task, get output

#### Step 6: Scheduler
- Implement cron-based scheduler (APScheduler or Celery Beat)
- Wire to agent execution engine
- Support: daily, hourly, custom cron expressions
- **Deliverable**: Agents run on schedule automatically

### Phase 3: Marketplace & Billing (Week 5-6)

#### Step 7: Stripe Integration
- Set up Stripe products/prices matching agent templates
- Implement subscription checkout flow
- Webhook handlers: subscription created, cancelled, payment failed
- **Deliverable**: Users can subscribe to agents with real payments

#### Step 8: Agent Instance Management
- Purchase flow: checkout → provision container → start schedule
- User dashboard: view my agents, pause/resume, view logs
- Execution log viewer: see what the agent did, output summary
- **Deliverable**: Full purchase-to-running pipeline

#### Step 9: Integration OAuth Flow
- OAuth consent screens for Gmail, Google Calendar
- Slack app installation flow
- Credential storage (encrypted in DB)
- User settings page to connect/disconnect integrations
- **Deliverable**: Users can connect their accounts

### Phase 4: Polish & Launch (Week 7-8)

#### Step 10: First Agent — Email Daily Digest
- Create "Email Manager" agent template
- Skills: summarize emails, categorize priority, draft responses
- Schedule: daily at 7am user's timezone
- Output: formatted digest sent via email/Slack
- **Deliverable**: Working end-to-end demo agent

#### Step 11: Monitoring & Error Handling
- Agent health checks
- Execution failure alerts (to admin)
- Cost tracking per execution
- Container resource limits
- **Deliverable**: Production-ready error handling

#### Step 12: Landing Page & Deployment
- Marketing landing page on storefront
- Deploy to cloud (Hetzner/AWS)
- Set up CI/CD
- **Deliverable**: Live product

---

## Project Structure

```
agents/
├── apps/
│   ├── api/                          # FastAPI backend
│   │   ├── app/
│   │   │   ├── main.py
│   │   │   ├── config.py
│   │   │   ├── models/               # SQLAlchemy models
│   │   │   │   ├── user.py
│   │   │   │   ├── agent_template.py
│   │   │   │   ├── agent_instance.py
│   │   │   │   ├── integration.py
│   │   │   │   └── execution_log.py
│   │   │   ├── routers/              # API endpoints
│   │   │   │   ├── auth.py
│   │   │   │   ├── templates.py
│   │   │   │   ├── instances.py
│   │   │   │   ├── integrations.py
│   │   │   │   ├── billing.py
│   │   │   │   └── executions.py
│   │   │   ├── services/             # Business logic
│   │   │   │   ├── agent_engine.py   # Container lifecycle
│   │   │   │   ├── scheduler.py      # Cron management
│   │   │   │   ├── billing.py        # Stripe logic
│   │   │   │   └── crypto.py         # Credential encryption
│   │   │   └── middleware/
│   │   │       ├── auth.py
│   │   │       └── admin.py
│   │   ├── alembic/                  # DB migrations
│   │   ├── tests/
│   │   ├── pyproject.toml
│   │   └── Dockerfile
│   │
│   └── web/                          # Next.js frontend
│       ├── src/
│       │   ├── app/
│       │   │   ├── (storefront)/     # Public pages
│       │   │   │   ├── page.tsx      # Landing/browse
│       │   │   │   └── agents/[slug]/page.tsx
│       │   │   ├── (dashboard)/      # User dashboard
│       │   │   │   ├── my-agents/
│       │   │   │   ├── integrations/
│       │   │   │   └── logs/
│       │   │   └── (admin)/          # Admin panel
│       │   │       ├── templates/
│       │   │       └── analytics/
│       │   ├── components/
│       │   └── lib/
│       ├── package.json
│       └── Dockerfile
│
├── packages/
│   └── integrations-cli/             # Go CLI
│       ├── cmd/integrations/main.go
│       ├── internal/
│       ├── go.mod
│       └── Dockerfile
│
├── agent-images/                     # Agent Docker images
│   ├── base/Dockerfile               # Base image with Claude Code + CLI
│   └── templates/                    # Per-template skill bundles
│       └── email-manager/
│           ├── CLAUDE.md
│           └── skills/
│
├── docker-compose.yml                # Local dev
├── docker-compose.prod.yml
└── README.md
```

---

## Risks and Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Claude Code CLI pricing at scale | HIGH | Track tokens per execution, set cost caps per agent run, alert on spikes |
| Container resource exhaustion | HIGH | CPU/memory limits per container, max concurrent containers |
| Integration OAuth token expiry | MEDIUM | Auto-refresh tokens, alert user on auth failures |
| Agent hallucination / bad actions | HIGH | Sandboxed containers, CLI-only integrations (no raw API), dry-run mode for new agents |
| Credential security | CRITICAL | Encrypt at rest (Fernet/AES), never log credentials, rotate encryption keys |
| Stripe webhook reliability | MEDIUM | Idempotent handlers, webhook signature verification, dead-letter queue |
| Docker image build time | MEDIUM | Base image caching, pre-built layers, image registry |
| User timezone handling | LOW | Store timezone per user, schedule in UTC, convert for display |

---

## MVP Scope (What's IN vs OUT)

### IN (MVP)
- Admin creates agent templates with skills + system prompt
- Users browse and purchase agents (Stripe subscription)
- Users connect Gmail, Google Calendar, Slack via OAuth
- Agents run on cron schedule in Docker containers
- Integration CLI with 3 providers (Gmail, Calendar, Slack)
- Execution logs viewable in dashboard
- One pre-built agent: "Email Daily Digest"

### OUT (Post-MVP)
- Agent marketplace with third-party creators
- Custom agent builder for users
- Real-time agent chat/interaction
- Agent-to-agent communication
- Usage-based pricing
- Mobile app
- Advanced analytics dashboard
- Custom integration CLI plugins

---

## SESSION_ID (for /ccg:execute use)
- CODEX_SESSION: N/A (codeagent-wrapper not available)
- GEMINI_SESSION: N/A (codeagent-wrapper not available)
