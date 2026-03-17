package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewGitHubOAuthConfig(t *testing.T) {
	t.Run("missing client ID", func(t *testing.T) {
		t.Setenv("GITHUB_CLIENT_ID", "")
		t.Setenv("GITHUB_CLIENT_SECRET", "test-secret")
		_, err := newGitHubOAuthConfig()
		if err == nil {
			t.Fatal("expected error for missing client ID")
		}
	})

	t.Run("missing client secret", func(t *testing.T) {
		t.Setenv("GITHUB_CLIENT_ID", "test-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "")
		_, err := newGitHubOAuthConfig()
		if err == nil {
			t.Fatal("expected error for missing client secret")
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Setenv("GITHUB_CLIENT_ID", "test-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "test-secret")
		cfg, err := newGitHubOAuthConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ClientID != "test-id" {
			t.Errorf("expected client ID test-id, got %s", cfg.ClientID)
		}
		if cfg.Endpoint.TokenURL != "https://github.com/login/oauth/access_token" {
			t.Errorf("unexpected token URL: %s", cfg.Endpoint.TokenURL)
		}
	})
}

func TestNewGitHubToken(t *testing.T) {
	t.Run("missing access token", func(t *testing.T) {
		t.Setenv("GITHUB_ACCESS_TOKEN", "")
		t.Setenv("GITHUB_REFRESH_TOKEN", "test-refresh")
		_, err := newGitHubToken()
		if err == nil {
			t.Fatal("expected error for missing access token")
		}
	})

	t.Run("missing refresh token", func(t *testing.T) {
		t.Setenv("GITHUB_ACCESS_TOKEN", "test-access")
		t.Setenv("GITHUB_REFRESH_TOKEN", "")
		_, err := newGitHubToken()
		if err == nil {
			t.Fatal("expected error for missing refresh token")
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Setenv("GITHUB_ACCESS_TOKEN", "test-access")
		t.Setenv("GITHUB_REFRESH_TOKEN", "test-refresh")
		tok, err := newGitHubToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tok.AccessToken != "test-access" {
			t.Errorf("expected access token test-access, got %s", tok.AccessToken)
		}
	})
}

func TestGitHubBaseURL(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("GITHUB_API_BASE_URL", "")
		if u := GitHubBaseURL(); u != "https://api.github.com" {
			t.Errorf("expected default URL, got %s", u)
		}
	})

	t.Run("custom", func(t *testing.T) {
		t.Setenv("GITHUB_API_BASE_URL", "https://github.example.com/api/v3")
		if u := GitHubBaseURL(); u != "https://github.example.com/api/v3" {
			t.Errorf("expected custom URL, got %s", u)
		}
	})
}

func TestGitHubHeaderTransport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Errorf("Accept header = %q, want application/vnd.github+json", got)
		}
		if got := r.Header.Get("X-GitHub-Api-Version"); got != "2022-11-28" {
			t.Errorf("X-GitHub-Api-Version = %q, want 2022-11-28", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &githubHeaderTransport{base: http.DefaultTransport},
	}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestNewGitHubClient(t *testing.T) {
	t.Run("missing credentials", func(t *testing.T) {
		t.Setenv("GITHUB_CLIENT_ID", "")
		t.Setenv("GITHUB_CLIENT_SECRET", "")
		t.Setenv("GITHUB_ACCESS_TOKEN", "")
		t.Setenv("GITHUB_REFRESH_TOKEN", "")
		_, err := NewGitHubClient(context.Background())
		if err == nil {
			t.Fatal("expected error for missing credentials")
		}
	})

	t.Run("success", func(t *testing.T) {
		// Clear any previously set env vars that might interfere
		for _, key := range []string{"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "GITHUB_ACCESS_TOKEN", "GITHUB_REFRESH_TOKEN"} {
			os.Unsetenv(key)
		}
		t.Setenv("GITHUB_CLIENT_ID", "test-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "test-secret")
		t.Setenv("GITHUB_ACCESS_TOKEN", "test-access")
		t.Setenv("GITHUB_REFRESH_TOKEN", "test-refresh")
		client, err := NewGitHubClient(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})
}
