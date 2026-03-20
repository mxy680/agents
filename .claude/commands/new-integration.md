# New Integration Provider

Create a complete integration provider for: $ARGUMENTS

## Pre-Implementation Checklist

Before writing any code, research the target API:
1. Find the official API documentation and Go SDK (if one exists)
2. Identify the authentication method (OAuth2, API key, cookie-based, etc.)
3. List ALL resource groups the API exposes — do NOT omit any as "lower value"
4. For each resource group, list ALL CRUD operations available
5. Present the full command tree and WAIT for user confirmation before coding

## Implementation Order (7 Steps)

Execute these steps sequentially. Commit after each step.

---

### Step 1: Auth Module — `internal/auth/<provider>.go`

Create the authentication factory following the existing pattern.

**For OAuth2 providers** (like Google, GitHub):
```go
package auth

// <Provider>EnvConfig holds env var names for OAuth credentials.
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

// OR for Google APIs that have a generated Go SDK:
func New<Provider>Service(ctx context.Context) (*<api>.Service, error) { ... }
```

**For cookie/session-based providers** (like Instagram):
- Create `internal/auth/<provider>.go` that reads session cookies from env vars
- Build an HTTP client with cookie injection via a custom `http.RoundTripper`

**Key rules:**
- Use `readEnv()` for required vars, `os.Getenv()` for optional ones
- Wrap with `tokenNotifySource` if the provider supports token refresh
- Write `internal/auth/<provider>_test.go` with tests for missing env vars and client creation

---

### Step 2: Provider Scaffold — `internal/providers/<name>/`

Create the provider package with these files:

#### `<name>.go` — Provider struct + RegisterCommands
```go
package <name>

type ServiceFactory func(ctx context.Context) (*api.Service, error)
// OR for raw HTTP APIs:
type ClientFactory func(ctx context.Context) (*http.Client, error)

type Provider struct {
    ServiceFactory ServiceFactory  // or ClientFactory
}

func New() *Provider {
    return &Provider{
        ServiceFactory: auth.New<Provider>Service,
    }
}

func (p *Provider) Name() string { return "<name>" }

func (p *Provider) RegisterCommands(parent *cobra.Command) {
    rootCmd := &cobra.Command{
        Use:   "<name>",
        Short: "Interact with <Provider>",
        Long:  "<description>",
    }

    // Add resource subcommands here (Step 3)

    parent.AddCommand(rootCmd)
}
```

#### `helpers.go` — Shared types and utilities
Define JSON-serializable summary/detail types for each resource:
```go
type <Resource>Summary struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    // ... fields visible in list output
}

type <Resource>Detail struct {
    // ... full fields for get output
}
```

Plus shared utility functions:
- `truncate(s string, max int) string`
- `confirmDestructive(cmd *cobra.Command) error` — checks `--confirm` flag
- `dryRunResult(cmd *cobra.Command, desc string, data any) error` — handles `--dry-run`
- Resource-specific formatters and parsers

#### `helpers_test.go` — Unit tests for all helper functions

#### `<name>_test.go` — Provider-level tests
```go
func TestProviderNew(t *testing.T) { ... }
func TestProviderName(t *testing.T) { ... }
func TestProviderRegisterCommands(t *testing.T) { ... }
```

---

### Step 3: Resource Commands

Create one file per resource group. Each file contains all commands for that resource.

**Command function pattern:**
```go
func new<Resource>ListCmd(factory ServiceFactory) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List <resources>",
        RunE: makeRun<Resource>List(factory),
    }
    cmd.Flags().Int("limit", 25, "Maximum results")
    cmd.Flags().String("page-token", "", "Pagination token")
    cmd.Flags().Bool("json", false, "JSON output")  // only if not inherited
    return cmd
}

func makeRun<Resource>List(factory ServiceFactory) func(*cobra.Command, []string) error {
    return func(cmd *cobra.Command, args []string) error {
        ctx := cmd.Context()
        svc, err := factory(ctx)
        if err != nil {
            return fmt.Errorf("create service: %w", err)
        }

        // API call
        // Format output (JSON via cli.WriteJSON or text table)
        return nil
    }
}
```

**Flag conventions (MUST follow):**
- `--json` — JSON output mode (inherited from root if using `PersistentFlags`)
- `--dry-run` — preview destructive/write operations without executing
- `--confirm` — skip interactive confirmation for destructive operations
- `--limit` — max results (default 25)
- `--page-token` — pagination cursor
- `--query` or `--q` — search/filter string
- Use `--<resource>-id` for resource identifiers (e.g., `--file-id`, `--repo`)

**Alias conventions:**
- Resource groups get singular aliases: `messages` → `msg`, `files` → `file`, `f`
- Provider root can get short alias: `instagram` → `ig`

**Output rules:**
- Text mode: human-readable table/list (use `fmt.Fprintf` to stdout)
- JSON mode: use `cli.WriteJSON(cmd, data)` for consistent envelope
- Destructive ops: require `--confirm` flag OR `--dry-run`, print what would happen

---

### Step 4: Mock Server + Tests

#### `mock_server_test.go`
```go
package <name>

func with<Resource>Mock(mux *http.ServeMux) {
    // Register handlers for each API endpoint
    // Return realistic JSON responses matching the real API
}

func newFullMockServer(t *testing.T) *httptest.Server {
    t.Helper()
    mux := http.NewServeMux()
    with<Resource1>Mock(mux)
    with<Resource2>Mock(mux)
    // ... all resources
    return httptest.NewServer(mux)
}

func newTestServiceFactory(server *httptest.Server) ServiceFactory {
    // Return factory that creates service pointing at test server
}

func captureStdout(t *testing.T, f func()) string {
    // Capture stdout for assertion (copy from existing provider)
}

func newTestRootCmd() *cobra.Command {
    root := &cobra.Command{Use: "integrations"}
    root.PersistentFlags().Bool("json", false, "")
    root.PersistentFlags().Bool("dry-run", false, "")
    return root
}
```

#### `<resource>_test.go` — One test file per resource command file
Test every command (list, get, create, update, delete, etc.):
- Text output mode
- JSON output mode (`--json`)
- Dry-run mode (`--dry-run`)
- Error cases (missing required flags)
- Edge cases (empty results, pagination)

**Coverage target: 80%+**

---

### Step 5: Register Provider

#### `cmd/integrations/main.go`
Add import and registration:
```go
import "<provider>provider" "github.com/emdash-projects/agents/internal/providers/<name>"

// In main():
<name>Provider := <name>provider.New()  // or <name>.New() if no conflict
<name>Provider.RegisterCommands(cli.RootCmd())
```

Run `make build && make test` to verify everything compiles and passes.

---

### Step 6: Token Bridge + Portal Credentials

#### `internal/tokenbridge/bridge.go`
Add a case to `processIntegration()`:
```go
case "<provider>":
    mapCredentials(creds, env, map[string]string{
        "access_token":  "<PROVIDER>_ACCESS_TOKEN",
        "refresh_token": "<PROVIDER>_REFRESH_TOKEN",
        // ... map all credential fields to env vars
    })
```

#### `portal/lib/credentials.ts`
Add a case to `credentialsToEnv()`:
```typescript
case "<provider>":
    if (credJson.access_token) env.<PROVIDER>_ACCESS_TOKEN = credJson.access_token
    if (credJson.refresh_token) env.<PROVIDER>_REFRESH_TOKEN = credJson.refresh_token
    // ... match the Go-side mapping
    break
```

---

### Step 7: Portal OAuth Routes (if OAuth-based)

Create three route files under `portal/app/api/integrations/<provider>/`:

#### `connect/route.ts` — Initiate OAuth flow
```typescript
export async function GET(request: NextRequest) {
    // 1. Verify user is authenticated via Supabase
    // 2. Read label from query params
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
    // 5. Encrypt token JSON via encrypt()
    // 6. Upsert into user_integrations table
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

---

## Post-Implementation Checklist

After all steps are complete:
- [ ] `make build` succeeds
- [ ] `make test` passes with 80%+ coverage for the new provider
- [ ] `make lint` passes (go vet)
- [ ] All commands work in text and JSON modes
- [ ] Destructive commands require `--confirm` or `--dry-run`
- [ ] Token bridge maps credentials correctly (Go + TypeScript match)
- [ ] Update `CLAUDE.md` with:
  - Full command reference for the new provider
  - Package layout section
  - Environment variables section
  - Updated coverage numbers in Testing section

## Environment Variables to Document

```
# <Provider>
<PROVIDER>_CLIENT_ID, <PROVIDER>_CLIENT_SECRET   (if OAuth)
<PROVIDER>_ACCESS_TOKEN, <PROVIDER>_REFRESH_TOKEN (if OAuth)
<PROVIDER>_API_KEY                                 (if API key)
<PROVIDER>_<CREDENTIAL>                            (if cookie/session)
```

## Key Files to Touch

| File | Change |
|------|--------|
| `internal/auth/<provider>.go` | Auth factory |
| `internal/auth/<provider>_test.go` | Auth tests |
| `internal/providers/<name>/*.go` | Provider + commands |
| `internal/providers/<name>/*_test.go` | Tests |
| `cmd/integrations/main.go` | Register provider |
| `internal/tokenbridge/bridge.go` | Credential mapping |
| `portal/lib/credentials.ts` | TS credential mapping |
| `portal/app/api/integrations/<name>/` | OAuth routes |
| `CLAUDE.md` | Documentation |
