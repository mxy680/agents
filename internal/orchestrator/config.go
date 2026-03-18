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

	// K8s
	KubeNamespace         string
	AgentBaseImage        string
	ExportCredsImage      string
	ClaudeSessionSecretRef string // K8s Secret name for CLAUDE_CODE_OAUTH_TOKEN

	// Supabase JWT
	SupabaseJWTSecret string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Port:                   8080,
		KubeNamespace:          "agents",
		AgentBaseImage:         "ghcr.io/emdash-projects/agent-base:dev",
		ExportCredsImage:       "ghcr.io/emdash-projects/export-creds:dev",
		ClaudeSessionSecretRef: "claude-session-token",
	}
}
