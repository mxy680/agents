package portal

import (
	"encoding/hex"
	"fmt"
	"os"
)

// Config holds all portal configuration loaded from environment variables.
type Config struct {
	DatabaseURL     string
	EncryptionKey   []byte
	SessionSecret   string
	GoogleClientID  string
	GoogleClientSecret string
	GitHubClientID  string
	GitHubClientSecret string
	Port            string
	BaseURL         string
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	cfg := &Config{}

	// Required fields
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	encryptionKeyHex := os.Getenv("PORTAL_ENCRYPTION_KEY")
	if encryptionKeyHex == "" {
		return nil, fmt.Errorf("PORTAL_ENCRYPTION_KEY is required")
	}
	key, err := hex.DecodeString(encryptionKeyHex)
	if err != nil {
		return nil, fmt.Errorf("PORTAL_ENCRYPTION_KEY must be a valid hex string: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("PORTAL_ENCRYPTION_KEY must be a 32-byte hex string (64 hex chars), got %d bytes", len(key))
	}
	cfg.EncryptionKey = key

	cfg.SessionSecret = os.Getenv("PORTAL_SESSION_SECRET")
	if cfg.SessionSecret == "" {
		return nil, fmt.Errorf("PORTAL_SESSION_SECRET is required")
	}

	cfg.GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	if cfg.GoogleClientID == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}

	cfg.GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	if cfg.GoogleClientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_SECRET is required")
	}

	// Optional fields
	cfg.GitHubClientID = os.Getenv("GITHUB_CLIENT_ID")
	cfg.GitHubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	cfg.Port = os.Getenv("PORT")
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	cfg.BaseURL = os.Getenv("PORTAL_BASE_URL")
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:" + cfg.Port
	}

	return cfg, nil
}
