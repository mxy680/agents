package orchestrator

// Config holds all configuration for the orchestrator service.
type Config struct {
	// HTTP server
	Port int

	// Database
	DatabaseURL         string
	EncryptionMasterKey string

	// Google OAuth (for token refresh)
	GoogleClientID     string
	GoogleClientSecret string

	// Runtime selection: "docker" or "k8s" (default: "k8s")
	Runtime string

	// Docker runtime
	ClaudeOAuthToken string // CLAUDE_CODE_OAUTH_TOKEN passed directly to containers

	// K8s runtime
	KubeNamespace          string
	AgentBaseImage         string
	ExportCredsImage       string
	ClaudeSessionSecretRef string // K8s Secret name for CLAUDE_CODE_OAUTH_TOKEN

	// Supabase JWT
	SupabaseJWTSecret string

	// Static API key auth (alternative to JWT for internal admin tool)
	APIKey      string // If set, Bearer <APIKey> is accepted as auth
	AdminUserID string // User ID to use when authenticating with API key

	// Agent templates directory (on host filesystem)
	AgentsDir string // e.g. "/opt/agents/agents" — mounted read-only into containers

	// CORS
	AllowedOrigin string // If empty, defaults to "*" (development only)
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Port:                   8080,
		Runtime:                "k8s",
		KubeNamespace:          "agents",
		AgentBaseImage:         "ghcr.io/emdash-projects/agent-base:dev",
		ExportCredsImage:       "ghcr.io/emdash-projects/export-creds:dev",
		ClaudeSessionSecretRef: "claude-session-token",
	}
}
