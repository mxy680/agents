# Integrating the CLI with Claude Code

This guide explains how to give Claude Code (or any Claude-based agent) access to the `integrations` CLI so it can interact with Gmail, Google Calendar, Google Drive, Google Sheets, GitHub, Instagram, LinkedIn, Framer, Google Places, and Supabase on your behalf.

## How it works

The `integrations` CLI is a stateless binary. Every invocation reads credentials from environment variables, calls an API, and prints the result. There is no login flow, no config file, no daemon. If the right env vars are set and the binary is in PATH, any process — including Claude Code — can use it.

## Setup

### 1. Build the binary

```bash
cd /path/to/agents
make build    # produces bin/integrations
```

Copy or symlink it somewhere on your PATH:

```bash
cp bin/integrations /usr/local/bin/integrations
# or
ln -s "$(pwd)/bin/integrations" /usr/local/bin/integrations
```

Verify it works:

```bash
integrations --help
```

### 2. Set credentials

The CLI reads credentials from environment variables. You have three options for providing them:

#### Option A: Doppler (recommended if you already use it)

If your project uses Doppler for secrets management, wrap your Claude Code session:

```bash
doppler run -- claude
```

Every shell command Claude Code runs will inherit the Doppler-injected env vars automatically.

#### Option B: Export manually

Export the env vars for the providers you need before starting Claude Code:

```bash
# Google (Gmail, Calendar, Drive, Sheets)
export GOOGLE_CLIENT_ID=...
export GOOGLE_CLIENT_SECRET=...
export GOOGLE_ACCESS_TOKEN=...
export GOOGLE_REFRESH_TOKEN=...

# GitHub
export GITHUB_ACCESS_TOKEN=...

# Instagram (cookie-based)
export INSTAGRAM_SESSION_ID=...
export INSTAGRAM_CSRF_TOKEN=...
export INSTAGRAM_DS_USER_ID=...

# LinkedIn (cookie-based)
export LINKEDIN_LI_AT=...
export LINKEDIN_JSESSIONID=...

# Framer
export FRAMER_API_KEY=...
export FRAMER_PROJECT_URL=...

# Supabase
export SUPABASE_ACCESS_TOKEN=...

# Google Places (no credentials needed — just needs the scraper binary in PATH)
```

Then start Claude Code in the same shell.

#### Option C: Token bridge (for portal users)

If you have tokens stored in the portal's Supabase database, you can use the token bridge to export them:

```bash
# The export-creds binary decrypts tokens from the DB and prints them as env vars
eval $(export-creds --user-id=YOUR_USER_ID)
claude
```

### 3. Tell Claude Code about the CLI

Add the CLI reference to your project's `CLAUDE.md` so Claude Code knows the commands are available. The project's existing `CLAUDE.md` already documents all commands, so if you're working in this repo, Claude Code already knows about it.

For other projects, copy the relevant command reference sections into that project's `CLAUDE.md`, or add a pointer:

```markdown
## Tools Available

You have access to the `integrations` CLI binary for interacting with external services.
Run `integrations --help` and `integrations <provider> --help` to discover available commands.
Always use the `--json` flag for machine-readable output.
```

## Usage patterns

### Basic: read and respond

```bash
# List unread emails
integrations gmail messages list --query=is:unread --since=24h --json

# Read a specific email
integrations gmail messages get --id=MESSAGE_ID --json

# Check today's calendar
integrations calendar events list \
  --time-min=$(date -u +%Y-%m-%dT00:00:00Z) \
  --time-max=$(date -u +%Y-%m-%dT23:59:59Z) \
  --single-events --order-by=startTime --json
```

### Write operations (use with care)

```bash
# Send an email (always confirm with user first)
integrations gmail messages send \
  --to=someone@example.com \
  --subject="Meeting follow-up" \
  --body="Thanks for the call today." \
  --json

# Create a calendar event
integrations calendar events create \
  --summary="Team standup" \
  --start=2026-03-17T10:00:00-04:00 \
  --end=2026-03-17T10:30:00-04:00 \
  --json

# Create a GitHub issue
integrations github issues create \
  --owner=myorg --repo=myrepo \
  --title="Bug: login fails on mobile" \
  --body="Steps to reproduce..." \
  --json
```

### Chaining commands

Claude Code can chain commands to build workflows:

```bash
# Find an email and reply to it
MSG_ID=$(integrations gmail messages list --query="from:boss subject:urgent" --limit=1 --json | jq -r '.messages[0].id')
integrations gmail messages get --id=$MSG_ID --json
# ... Claude drafts a reply based on the content ...
integrations gmail messages send --to=boss@company.com --subject="Re: urgent" --body="..." --reply-to=$MSG_ID --json
```

### The `--json` flag

Always use `--json` for programmatic access. Without it, the CLI outputs human-readable tables. With it, you get structured JSON that Claude Code can parse with `jq` or read directly.

### The `--dry-run` flag

Write operations (send, create, update, delete) support `--dry-run` which shows what would happen without actually doing it. Use this for previewing actions before committing.

## Security considerations

- **Never commit credentials.** The CLI reads from env vars, not files. Keep it that way.
- **Use `--dry-run` first.** For any destructive or externally-visible operation, preview before executing.
- **Confirm before sending.** Claude Code should always show you drafts of emails, posts, and comments before sending them. This is enforced by convention in the agent role files, but worth reinforcing in your own `CLAUDE.md`.
- **Token refresh.** Google OAuth tokens expire after ~1 hour. If you have `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and `GOOGLE_REFRESH_TOKEN` set, the CLI auto-refreshes the access token. For Instagram and LinkedIn (cookie-based), tokens last longer but will eventually expire and need manual rotation.
- **Scope of access.** The CLI has the same access as the tokens you provide. Review your OAuth scopes and cookie permissions before granting access.

## Comparison with the orchestrator flow

| | Orchestrator (production) | Claude Code (local) |
|---|---|---|
| Credentials | Encrypted in Supabase DB, decrypted by init container | Env vars in your shell |
| Runtime | K8s pod with Docker | Your local machine |
| Agent SDK | Anthropic Agent SDK (Python) | Claude Code (native) |
| Binary | Baked into Docker image | Installed in PATH |
| Token refresh | Token bridge handles it | CLI auto-refreshes (Google/GitHub) |

The CLI binary is identical in both cases — the only difference is how credentials get into the environment.
