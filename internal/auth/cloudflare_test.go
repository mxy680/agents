package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewCloudflareClient_MissingToken(t *testing.T) {
	os.Unsetenv("CLOUDFLARE_API_TOKEN")

	_, err := NewCloudflareClient(context.Background())
	if err == nil {
		t.Fatal("expected error for missing CLOUDFLARE_API_TOKEN, got nil")
	}
}

func TestNewCloudflareClient_Valid(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_TOKEN", "test-cf-token")

	client, err := NewCloudflareClient(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestCloudflareClient_InjectsAuthHeader(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_TOKEN", "my-cf-token")

	var gotHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client, err := NewCloudflareClient(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	resp.Body.Close()

	want := "Bearer my-cf-token"
	if gotHeader != want {
		t.Errorf("Authorization header = %q, want %q", gotHeader, want)
	}
}

func TestCloudflareBaseURL_Default(t *testing.T) {
	os.Unsetenv("CLOUDFLARE_API_BASE_URL")
	got := CloudflareBaseURL()
	if got != "https://api.cloudflare.com/client/v4" {
		t.Errorf("CloudflareBaseURL() = %q, want %q", got, "https://api.cloudflare.com/client/v4")
	}
}

func TestCloudflareBaseURL_Override(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_BASE_URL", "http://localhost:9999")
	got := CloudflareBaseURL()
	if got != "http://localhost:9999" {
		t.Errorf("CloudflareBaseURL() = %q, want %q", got, "http://localhost:9999")
	}
}
