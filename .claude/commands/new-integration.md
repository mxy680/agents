# New Integration Provider

Create a complete integration provider for: $ARGUMENTS

## Pre-Implementation Research

Before writing any code, research the target API:
1. Find the official API documentation and Go SDK (if one exists)
2. Identify the authentication method (OAuth2, API key, cookie-based, etc.)
3. List ALL resource groups the API exposes — do NOT omit any as "lower value"
4. For each resource group, list ALL CRUD operations available
5. Present the full command tree and WAIT for user confirmation before coding

---

## Phase 1: Connect Authentication

Goal: Get credentials flowing from the portal to the CLI binary.

### 1a. Portal OAuth Routes

**For OAuth2 providers**, create three route files under `portal/app/api/integrations/<provider>/`:

#### `connect/route.ts` — Initiate OAuth flow
```typescript
export async function GET(request: NextRequest) {
    // 1. Verify user is authenticated via Supabase
    // 2. Read label from query params (default: "<Provider> Account")
    // 3. Create HMAC-signed state token via createOAuthState(userId, label)
    // 4. Build OAuth authorize URL with scopes
    // 5. Redirect to provider's OAuth endpoint
}
```

#### `callback/route.ts` — Handle OAuth callback
```typescript
export async function GET(request: NextRequest) {
    // 1. Extract code, state, error from query params
    // 2. Verify state token via verifyOAuthState(state)
    // 3. Verify authenticated user matches state userId
    // 4. Exchange code for tokens (POST to provider's token endpoint)
    // 5. Encrypt token JSON via encrypt() from lib/crypto.ts
    // 6. Upsert into user_integrations table (provider, label, credentials as \\x hex)
    // 7. Redirect to /integrations
}
```

#### `disconnect/route.ts` — Remove integration
```typescript
export async function POST(request: NextRequest) {
    // 1. Verify user is authenticated
    // 2. Delete from user_integrations where provider=<name> and user_id=userId
    // 3. Return success JSON
}
```

**For non-OAuth providers** (like Instagram's cookie-based auth):
Create `save/route.ts` instead of connect/callback — accepts credentials via POST form.

### 1b. Portal Credential Mapping

#### `portal/lib/credentials.ts`
Add a case to `credentialsToEnv()`:
```typescript
case "<provider>":
    if (credJson.access_token) env.<PROVIDER>_ACCESS_TOKEN = credJson.access_token
    if (credJson.refresh_token) env.<PROVIDER>_REFRESH_TOKEN = credJson.refresh_token
    // Include client_id/secret from process.env if needed for token refresh
    break
```

### 1c. Go Auth Module — `internal/auth/<provider>.go`

**For OAuth2 providers** (like Google, GitHub):
```go
package auth

var <Provider>EnvConfig = struct {
    ClientID     string
    ClientSecret string
    AccessToken  string
    RefreshToken string
}{
    ClientID:     "<PROVIDER>_CLIENT_ID",
    ClientSecret: "<PROVIDER>_CLIENT_SECRET",
    AccessToken:  "<PROVIDER>_ACCESS_TOKEN",
    RefreshToken: "<PROVIDER>_REFRESH_TOKEN",
}

// New<Provider>Client creates an authenticated HTTP client.
// Reuse newAuthenticatedClient() if it's a Google API.
// For non-Google OAuth2, follow the GitHub pattern (custom endpoint + header transport).
func New<Provider>Client(ctx context.Context) (*http.Client, error) { ... }

// OR for Google APIs with a generated Go SDK:
func New<Provider>Service(ctx context.Context) (*<api>.Service, error) { ... }
```

**Key rules:**
- Use `readEnv()` for required vars, `os.Getenv()` for optional ones
- Wrap with `tokenNotifySource` if the provider supports token refresh
- Write `internal/auth/<provider>_test.go` with tests for missing env vars

### 1d. Token Bridge — `internal/tokenbridge/bridge.go`

Add a case to `processIntegration()`:
```go
case "<provider>":
    mapCredentials(creds, env, map[string]string{
        "access_token":  "<PROVIDER>_ACCESS_TOKEN",
        "refresh_token": "<PROVIDER>_REFRESH_TOKEN",
    })
```

### 1e. Verify Auth End-to-End

1. Run portal locally: `make portal-dev` (uses `doppler run`)
2. Navigate to http://localhost:3000/integrations
3. Click connect for the new provider, complete OAuth
4. Verify credentials appear in `user_integrations` table
5. Build CLI: `make build`
6. Test that `doppler run -- bin/integrations <provider> --help` shows the command tree

**Commit after Phase 1 is verified.**

---

## Phase 2: Build CLI Tools (Incremental + E2E)

Goal: Implement each resource group one at a time, verifying against the real API with markshteyn1@gmail.com's connected credentials.

### 2a. Provider Scaffold

#### `internal/providers/<name>/<name>.go`
```go
package <name>

type ServiceFactory func(ctx context.Context) (*api.Service, error)
// OR: type ClientFactory func(ctx context.Context) (*http.Client, error)

type Provider struct {
    ServiceFactory ServiceFactory
}

func New() *Provider {
    return &Provider{ServiceFactory: auth.New<Provider>Service}
}

func (p *Provider) Name() string { return "<name>" }

func (p *Provider) RegisterCommands(parent *cobra.Command) {
    rootCmd := &cobra.Command{
        Use:   "<name>",
        Short: "Interact with <Provider>",
    }
    // Resource subcommands added incrementally below
    parent.AddCommand(rootCmd)
}
```

#### `internal/providers/<name>/helpers.go`
Define JSON-serializable summary/detail types for each resource:
```go
type <Resource>Summary struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

Plus shared utilities: `truncate`, `confirmDestructive`, `dryRunResult`, formatters.

#### `internal/providers/<name>/mock_server_test.go`
```go
func newFullMockServer(t *testing.T) *httptest.Server { ... }
func newTestServiceFactory(server *httptest.Server) ServiceFactory { ... }
func captureStdout(t *testing.T, f func()) string { ... }
func newTestRootCmd() *cobra.Command { ... }
```

#### `cmd/integrations/main.go`
Register the provider:
```go
<name>Provider := <name>.New()
<name>Provider.RegisterCommands(cli.RootCmd())
```

**Commit the scaffold.**

### 2b. Implement Resource Groups One-by-One

For EACH resource group, repeat this cycle:

1. **Create command file** — `<resource>.go` with all CRUD operations
2. **Create mock handlers** — add `with<Resource>Mock(mux)` to `mock_server_test.go`
3. **Create unit tests** — `<resource>_test.go` covering text, JSON, dry-run, error cases
4. **Run unit tests** — `make test` (must pass with 80%+ coverage)
5. **Run e2e test** — Build and test against real API:
   ```bash
   make build
   doppler run -- bin/integrations <provider> <resource> list --json
   doppler run -- bin/integrations <provider> <resource> get --id=<real-id> --json
   ```
6. **Commit** after each resource group passes both unit and e2e

**Command function pattern:**
```go
func new<Resource>ListCmd(factory ServiceFactory) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List <resources>",
        RunE:  makeRun<Resource>List(factory),
    }
    cmd.Flags().Int("limit", 25, "Maximum results")
    cmd.Flags().String("page-token", "", "Pagination token")
    return cmd
}

func makeRun<Resource>List(factory ServiceFactory) func(*cobra.Command, []string) error {
    return func(cmd *cobra.Command, args []string) error {
        ctx := cmd.Context()
        svc, err := factory(ctx)
        if err != nil {
            return fmt.Errorf("create service: %w", err)
        }
        // API call → format output (JSON via cli.WriteJSON or text table)
        return nil
    }
}
```

**Flag conventions (MUST follow):**
- `--json` — JSON output (inherited from root PersistentFlags)
- `--dry-run` — preview write operations without executing
- `--confirm` — skip confirmation for destructive operations
- `--limit` — max results (default 25)
- `--page-token` — pagination cursor
- `--query` — search/filter string
- `--<resource>-id` for identifiers (e.g., `--file-id`, `--repo`)

**Alias conventions:**
- Resource groups get singular aliases: `messages` → `msg`, `files` → `file`, `f`
- Provider root can get short alias: `instagram` → `ig`

**Output rules:**
- Text mode: human-readable table/list via `fmt.Fprintf` to stdout
- JSON mode: `cli.WriteJSON(cmd, data)` for consistent envelope
- Destructive ops: require `--confirm` or `--dry-run`

**Coverage target: 80%+ per resource file.**

---

## Phase 3: Specialized Agent + Marketplace

Goal: Create a Claude Agent SDK agent that uses only this integration, test it, and add it to the marketplace.

### 3a. Agent Template

Create `agents/<agent-name>/`:

#### `template.yaml`
```yaml
name: <agent-name>
description: <what the agent does>
required_integrations:
  - <provider>
docker_image: agent-base:latest
```

#### `role.md`
Agent persona — who it is, how it should behave, what it's good at.

#### `CLAUDE.md`
Claude-specific instructions for tool usage, output format, safety constraints.

#### `entrypoint.py`
```python
import anthropic
from claude_agent_sdk import Agent

# Initialize with only this integration's tools
agent = Agent(
    model="claude-sonnet-4-6-20250514",
    tools=[...],  # CLI commands available via integrations binary
)
```

#### `requirements.txt`
```
anthropic
claude-agent-sdk
```

### 3b. Sync Template to Supabase

```bash
make sync-templates
```

### 3c. Test Agent Locally

1. Deploy locally via orchestrator:
   ```bash
   make orchestrator-dev
   # POST /api/v1/agents/deploy with template_id
   ```
2. Test prompts against the agent — verify it uses the integration correctly
3. Check logs: `GET /api/v1/agents/{id}/logs`

### 3d. Add to Marketplace

1. Verify template is visible in portal: `GET /api/v1/templates`
2. Test deploy from portal UI
3. Verify agent runs with user's connected credentials

**Commit after Phase 3 is verified.**

---

## Post-Implementation Checklist

- [ ] `make build` succeeds
- [ ] `make test` passes with 80%+ coverage for the new provider
- [ ] `make lint` passes (go vet)
- [ ] All commands verified against real API (markshteyn1@gmail.com)
- [ ] Token bridge maps credentials correctly (Go + TypeScript sides match)
- [ ] Agent template synced and deployable
- [ ] Agent tested with real prompts
- [ ] Update `CLAUDE.md` with:
  - Full command reference for the new provider
  - Package layout section
  - Environment variables section
  - Updated coverage numbers in Testing section

## Key Files to Touch

| File | Phase | Change |
|------|-------|--------|
| `portal/app/api/integrations/<name>/` | 1 | OAuth routes |
| `portal/lib/credentials.ts` | 1 | TS credential mapping |
| `internal/auth/<provider>.go` | 1 | Auth factory |
| `internal/tokenbridge/bridge.go` | 1 | Credential mapping |
| `internal/providers/<name>/*.go` | 2 | Provider + commands |
| `internal/providers/<name>/*_test.go` | 2 | Unit tests |
| `cmd/integrations/main.go` | 2 | Register provider |
| `agents/<agent-name>/` | 3 | Agent template |
| `CLAUDE.md` | all | Documentation |
