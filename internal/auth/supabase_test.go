package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewSupabaseOAuthConfig(t *testing.T) {
	t.Run("missing client ID", func(t *testing.T) {
		t.Setenv("SUPABASE_INTEGRATION_CLIENT_ID", "")
		t.Setenv("SUPABASE_INTEGRATION_CLIENT_SECRET", "test-secret")
		_, err := newSupabaseOAuthConfig()
		if err == nil {
			t.Fatal("expected error for missing client ID")
		}
	})

	t.Run("missing client secret", func(t *testing.T) {
		t.Setenv("SUPABASE_INTEGRATION_CLIENT_ID", "test-id")
		t.Setenv("SUPABASE_INTEGRATION_CLIENT_SECRET", "")
		_, err := newSupabaseOAuthConfig()
		if err == nil {
			t.Fatal("expected error for missing client secret")
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Setenv("SUPABASE_INTEGRATION_CLIENT_ID", "test-id")
		t.Setenv("SUPABASE_INTEGRATION_CLIENT_SECRET", "test-secret")
		cfg, err := newSupabaseOAuthConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ClientID != "test-id" {
			t.Errorf("expected client ID test-id, got %s", cfg.ClientID)
		}
		if cfg.Endpoint.TokenURL != "https://api.supabase.com/v1/oauth/token" {
			t.Errorf("unexpected token URL: %s", cfg.Endpoint.TokenURL)
		}
	})
}

func TestNewSupabaseToken(t *testing.T) {
	t.Run("missing access token", func(t *testing.T) {
		t.Setenv("SUPABASE_ACCESS_TOKEN", "")
		t.Setenv("SUPABASE_REFRESH_TOKEN", "test-refresh")
		_, err := newSupabaseToken()
		if err == nil {
			t.Fatal("expected error for missing access token")
		}
	})

	t.Run("optional refresh token", func(t *testing.T) {
		t.Setenv("SUPABASE_ACCESS_TOKEN", "test-access")
		t.Setenv("SUPABASE_REFRESH_TOKEN", "")
		tok, err := newSupabaseToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tok.AccessToken != "test-access" {
			t.Errorf("expected access token test-access, got %s", tok.AccessToken)
		}
		if tok.RefreshToken != "" {
			t.Errorf("expected empty refresh token, got %s", tok.RefreshToken)
		}
	})

	t.Run("success with refresh token", func(t *testing.T) {
		t.Setenv("SUPABASE_ACCESS_TOKEN", "test-access")
		t.Setenv("SUPABASE_REFRESH_TOKEN", "test-refresh")
		tok, err := newSupabaseToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tok.AccessToken != "test-access" {
			t.Errorf("expected access token test-access, got %s", tok.AccessToken)
		}
		if tok.RefreshToken != "test-refresh" {
			t.Errorf("expected refresh token test-refresh, got %s", tok.RefreshToken)
		}
	})
}

func TestSupabaseBaseURL(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("SUPABASE_API_BASE_URL", "")
		if u := SupabaseBaseURL(); u != "https://api.supabase.com" {
			t.Errorf("expected default URL, got %s", u)
		}
	})

	t.Run("custom", func(t *testing.T) {
		t.Setenv("SUPABASE_API_BASE_URL", "https://custom.supabase.example.com")
		if u := SupabaseBaseURL(); u != "https://custom.supabase.example.com" {
			t.Errorf("expected custom URL, got %s", u)
		}
	})
}

func TestSupabaseHeaderTransport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept header = %q, want application/json", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &supabaseHeaderTransport{base: http.DefaultTransport},
	}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestNewSupabaseClient(t *testing.T) {
	t.Run("missing credentials", func(t *testing.T) {
		t.Setenv("SUPABASE_INTEGRATION_CLIENT_ID", "")
		t.Setenv("SUPABASE_INTEGRATION_CLIENT_SECRET", "")
		t.Setenv("SUPABASE_ACCESS_TOKEN", "")
		t.Setenv("SUPABASE_REFRESH_TOKEN", "")
		_, err := NewSupabaseClient(context.Background())
		if err == nil {
			t.Fatal("expected error for missing credentials")
		}
	})

	t.Run("success", func(t *testing.T) {
		for _, key := range []string{"SUPABASE_INTEGRATION_CLIENT_ID", "SUPABASE_INTEGRATION_CLIENT_SECRET", "SUPABASE_ACCESS_TOKEN", "SUPABASE_REFRESH_TOKEN"} {
			os.Unsetenv(key)
		}
		t.Setenv("SUPABASE_INTEGRATION_CLIENT_ID", "test-id")
		t.Setenv("SUPABASE_INTEGRATION_CLIENT_SECRET", "test-secret")
		t.Setenv("SUPABASE_ACCESS_TOKEN", "test-access")
		t.Setenv("SUPABASE_REFRESH_TOKEN", "test-refresh")
		client, err := NewSupabaseClient(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})
}
