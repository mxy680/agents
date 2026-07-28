# ENGAGENT

Production infrastructure for deploying and operating AI agents.

ENGAGENT combines a Go orchestration layer, agent services, a client portal, desktop tooling, and Supabase-backed workflows. It is built for real operational use: connecting agents to external systems, controlling execution, and giving operators a clear place to review and manage work.

## What it includes

- Agent orchestration and scheduled workflows
- Encrypted third-party integrations and OAuth flows
- Containerized execution and deployment tooling
- Admin and client-facing control surfaces
- Human approval boundaries for sensitive actions
- Shared infrastructure for reusable, vertical-specific agents

## Architecture

- **Go** for orchestration and core services
- **TypeScript / Next.js** for web control surfaces
- **Python** for agent and model-backed workflows
- **Postgres / Supabase** for data, authentication, and operational state
- **Docker / Kubernetes** for isolated execution and deployment

## Repository map

- `agents/` — agent definitions and workflows
- `internal/` — orchestration and platform services
- `client-portal/` — client-facing web application
- `desktop/` — desktop tooling
- `cmd/` — service entry points
- `docs/` — architecture and integration documentation

## Security model

Secrets stay outside source control. Integrations use environment-provided credentials or OAuth, sensitive actions can require explicit approval, and agent workloads are designed to run in isolated environments.

## Status

Actively developed by [ENGAGENT LLC](https://engagent.dev). This public repository shows the platform architecture and selected tooling; deployments require environment-specific configuration.
