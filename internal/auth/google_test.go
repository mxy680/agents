package auth

import (
	"os"
	"testing"
)

func setEnvVars(t *testing.T) {
	t.Helper()
	t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	t.Setenv("GMAIL_ACCESS_TOKEN", "test-access-token")
	t.Setenv("GMAIL_REFRESH_TOKEN", "test-refresh-token")
}

func TestNewOAuthConfig_Success(t *testing.T) {
	setEnvVars(t)

	config, err := NewOAuthConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.ClientID != "test-client-id" {
		t.Errorf("expected ClientID=test-client-id, got %s", config.ClientID)
	}
	if config.ClientSecret != "test-client-secret" {
		t.Errorf("expected ClientSecret=test-client-secret, got %s", config.ClientSecret)
	}
}

func TestNewOAuthConfig_MissingClientID(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "secret")

	_, err := NewOAuthConfig()
	if err == nil {
		t.Fatal("expected error for missing GOOGLE_CLIENT_ID")
	}
}

func TestNewOAuthConfig_MissingClientSecret(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")

	_, err := NewOAuthConfig()
	if err == nil {
		t.Fatal("expected error for missing GOOGLE_CLIENT_SECRET")
	}
}

func TestNewToken_Success(t *testing.T) {
	setEnvVars(t)

	token, err := NewToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "test-access-token" {
		t.Errorf("expected AccessToken=test-access-token, got %s", token.AccessToken)
	}
	if token.RefreshToken != "test-refresh-token" {
		t.Errorf("expected RefreshToken=test-refresh-token, got %s", token.RefreshToken)
	}
}

func TestNewToken_MissingAccessToken(t *testing.T) {
	t.Setenv("GMAIL_ACCESS_TOKEN", "")
	t.Setenv("GMAIL_REFRESH_TOKEN", "refresh")

	_, err := NewToken()
	if err == nil {
		t.Fatal("expected error for missing GMAIL_ACCESS_TOKEN")
	}
}

func TestNewToken_MissingRefreshToken(t *testing.T) {
	t.Setenv("GMAIL_ACCESS_TOKEN", "access")
	t.Setenv("GMAIL_REFRESH_TOKEN", "")

	_, err := NewToken()
	if err == nil {
		t.Fatal("expected error for missing GMAIL_REFRESH_TOKEN")
	}
}

func TestReadEnv_Unset(t *testing.T) {
	os.Unsetenv("NONEXISTENT_VAR")
	_, err := readEnv("NONEXISTENT_VAR")
	if err == nil {
		t.Fatal("expected error for unset env var")
	}
}

func TestNewGmailService_MissingCredentials(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("GMAIL_ACCESS_TOKEN", "")
	t.Setenv("GMAIL_REFRESH_TOKEN", "")

	_, err := NewGmailService(t.Context())
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
}
