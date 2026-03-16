# Agent Marketplace - Integration CLI

## Overview
Go CLI binary (`integrations`) that AI agents call inside Docker containers to interact with external services. Currently supports Gmail.

## Quick Start
```bash
make build          # → bin/integrations
make test           # run tests with coverage
make lint           # go vet
```

## Commands
```
integrations gmail list-unread [--limit=N] [--since=Xh] [--json]
integrations gmail read --id=MESSAGE_ID [--json]
integrations gmail send --to=EMAIL --subject=SUBJECT [--body=TEXT | --body-file=PATH] [--cc=EMAIL] [--reply-to=MSG_ID] [--dry-run]
integrations gmail search --query=QUERY [--limit=N] [--json]
```

## Architecture
- `cmd/integrations/main.go` — entrypoint, registers providers
- `internal/cli/` — Cobra root command, output helpers (JSON/text)
- `internal/auth/` — Google OAuth token management with auto-refresh
- `internal/providers/gmail/` — Gmail provider with injectable ServiceFactory
- `internal/providers/provider.go` — Provider interface

## Testing
- Gmail commands use `ServiceFactory` for dependency injection
- Tests use `httptest.NewServer` to mock the Gmail API
- Coverage target: 80%+ (currently 87.8%)

## Environment Variables
```
GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GMAIL_ACCESS_TOKEN, GMAIL_REFRESH_TOKEN
```
