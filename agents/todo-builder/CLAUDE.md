# Todo Builder Agent — Tool Documentation

## CRITICAL RULES

1. **DO NOT prefix commands with `doppler run --`** — credentials are already in your environment
2. **Always read `known-issues.md` FIRST** before starting any work. This file contains issues from prior runs — avoid repeating them.
3. **Always WRITE to `known-issues.md`** when you encounter an error or unexpected behavior. Append a line: `- [YYYY-MM-DD] [description] → [solution]`. This persists across sessions.
4. **Always create a Linear issue BEFORE starting work** — in the `todo-app` Linear project (NOT the Engagent team)
5. **Always commit to GitHub AFTER completing work**
6. **Always deploy to Vercel AFTER committing**
7. **Always verify the deployment is healthy** — curl the URL and check for HTTP 200
8. **Close the Linear issue with a summary when done**
9. **The GitHub org is `engagentdev`**, the repo name is `todo-app`
10. **NEVER use existing projects** — create NEW projects on every platform (GitHub, Supabase, Vercel, Linear). The todo app gets its OWN Supabase project, its OWN Linear team/project.
11. **For file operations, use `integrations github repos contents create/update`**
12. **NEVER modify, query, or touch the `engagent` Supabase project or `Engagent` Linear team** — those are the production platform. Create separate resources for the todo app.
13. **Create a Linear team called `Todo App`** (if it doesn't exist) and use that for all issues — NOT the `Engagent` (ENG) team.

## Authentication
All credentials are pre-configured via environment variables. Run commands directly — no `doppler run` needed.

---

## Tool 1: GitHub CLI — Code Management

### Repositories
```bash
# List repos
integrations github repos list --json

# Get a specific repo
integrations github repos get --repo=todo-app --owner=engagentdev --json

# Create a new repo
integrations github repos create --name=todo-app --description="Autonomous todo app built by Engagent" --private=false --json
```

### File Contents (read/write files in the repo)
```bash
# List files at a path
integrations github repos contents list --repo=todo-app --owner=engagentdev --path=/ --json
integrations github repos contents list --repo=todo-app --owner=engagentdev --path=src/app --json

# Get a file (also returns the current SHA needed for updates)
integrations github repos contents get --repo=todo-app --owner=engagentdev --path=package.json --json

# Create a new file (content must be base64-encoded)
integrations github repos contents create \
  --repo=todo-app \
  --owner=engagentdev \
  --path=src/app/page.tsx \
  --message="feat: add home page" \
  --content="$(echo -n 'file content here' | base64)" \
  --json

# Update an existing file (requires the current file SHA)
integrations github repos contents update \
  --repo=todo-app \
  --owner=engagentdev \
  --path=src/app/page.tsx \
  --message="feat: update home page" \
  --content="$(echo -n 'new file content' | base64)" \
  --sha=<current-sha-from-get> \
  --json
```

### Commits
```bash
# List recent commits
integrations github repos commits list --repo=todo-app --owner=engagentdev --limit=10 --json
```

**Note on base64 encoding:** For multi-line file content, write to a temp file first, then encode:
```bash
cat > /tmp/myfile.tsx << 'EOF'
// file content here
EOF
CONTENT=$(base64 < /tmp/myfile.tsx)
integrations github repos contents create --repo=todo-app --owner=engagentdev --path=src/app/page.tsx --message="feat: add page" --content="$CONTENT" --json
```

---

## Tool 2: Supabase CLI — Database

### Projects
```bash
integrations supabase projects list --json
```

For table operations (creating tables, querying data), use `curl` to interact with the Supabase REST API directly:

```bash
# Create a table via SQL (using the Supabase SQL API)
curl -s -X POST "${SUPABASE_URL}/rest/v1/rpc/exec_sql" \
  -H "apikey: ${SUPABASE_SERVICE_ROLE_KEY}" \
  -H "Authorization: Bearer ${SUPABASE_SERVICE_ROLE_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"query": "CREATE TABLE IF NOT EXISTS todos (id uuid DEFAULT gen_random_uuid() PRIMARY KEY, title text NOT NULL, completed boolean DEFAULT false, created_at timestamptz DEFAULT now())"}'

# Query a table
curl -s "${SUPABASE_URL}/rest/v1/todos?select=*" \
  -H "apikey: ${SUPABASE_SERVICE_ROLE_KEY}" \
  -H "Authorization: Bearer ${SUPABASE_SERVICE_ROLE_KEY}"

# Insert a row
curl -s -X POST "${SUPABASE_URL}/rest/v1/todos" \
  -H "apikey: ${SUPABASE_SERVICE_ROLE_KEY}" \
  -H "Authorization: Bearer ${SUPABASE_SERVICE_ROLE_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test todo"}'
```

The environment variables `SUPABASE_URL` (or `NEXT_PUBLIC_SUPABASE_URL`) and `SUPABASE_SERVICE_ROLE_KEY` are available in your environment.

---

## Tool 3: Vercel CLI — Deployment

### Projects
```bash
# List projects
integrations vercel projects list --json

# Get a specific project
integrations vercel projects get --name=todo-app --json

# Create a new project
integrations vercel projects create --name=todo-app --framework=nextjs --json
```

### Deployments
```bash
# List deployments for a project
integrations vercel deployments list --project=todo-app --limit=5 --json

# Get a specific deployment
integrations vercel deployments get --id=<deployment-id> --json

# Trigger a new deployment
integrations vercel deployments create --project=todo-app --json
```

### Environment Variables
```bash
# List env vars for a project
integrations vercel env list --project=todo-app --json

# Set an env var
integrations vercel env set --project=todo-app --key=NEXT_PUBLIC_SUPABASE_URL --value="$NEXT_PUBLIC_SUPABASE_URL" --target=production --json
integrations vercel env set --project=todo-app --key=NEXT_PUBLIC_SUPABASE_ANON_KEY --value="$NEXT_PUBLIC_SUPABASE_ANON_KEY" --target=production --json
```

### Domains
```bash
# List domains
integrations vercel domains list --json

# Add a domain to a project
integrations vercel domains add --project=todo-app --domain=todo-app.vercel.app --json
```

---

## Tool 4: Linear CLI — Issue Tracking

### Teams
```bash
# List all teams
integrations linear teams list --json
```

### Users
```bash
# Get the current user (useful for assigning issues)
integrations linear users me --json
```

### Issues
```bash
# List issues
integrations linear issues list --json
integrations linear issues list --team=ENG --json
integrations linear issues list --team=ENG --status=Done --json

# Get a specific issue
integrations linear issues get --id=<issue-id> --json

# Create a new issue
integrations linear issues create \
  --title="Add authentication to todo app" \
  --team=ENG \
  --description="Implement Supabase Auth with email/password login and signup" \
  --priority=2 \
  --json

# Update an issue (e.g., close it when done)
integrations linear issues update --id=<issue-id> --status=Done --json
integrations linear issues update --id=<issue-id> --status="In Progress" --json
```

**Priority levels:** 1=Urgent, 2=High, 3=Medium, 4=Low

---

## Workflow

The standard workflow for every evolution cycle:

1. **Check state** — does the GitHub repo exist? What's been built?
2. **Create Linear issue** — document what you're about to build
3. **Implement** — write/update files on GitHub via the contents API
4. **Verify deployment** — Vercel auto-deploys on push; wait ~30s then check status
5. **Curl the URL** — confirm HTTP 200 before closing the issue
6. **Close issue** — update Linear issue to Done with a summary comment
