package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewVercelClient_MissingToken(t *testing.T) {
	os.Unsetenv("VERCEL_TOKEN")

	_, err := NewVercelClient(context.Background())
	if err == nil {
		t.Fatal("expected error for missing VERCEL_TOKEN, got nil")
	}
}

func TestNewVercelClient_Valid(t *testing.T) {
	t.Setenv("VERCEL_TOKEN", "test-vercel-token")

	client, err := NewVercelClient(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestVercelClient_InjectsAuthHeader(t *testing.T) {
	t.Setenv("VERCEL_TOKEN", "my-secret-token")

	var gotHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client, err := NewVercelClient(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	resp.Body.Close()

	want := "Bearer my-secret-token"
	if gotHeader != want {
		t.Errorf("Authorization header = %q, want %q", gotHeader, want)
	}
}

func TestVercelBaseURL_Default(t *testing.T) {
	os.Unsetenv("VERCEL_API_BASE_URL")
	got := VercelBaseURL()
	if got != "https://api.vercel.com" {
		t.Errorf("VercelBaseURL() = %q, want %q", got, "https://api.vercel.com")
	}
}

func TestVercelBaseURL_Override(t *testing.T) {
	t.Setenv("VERCEL_API_BASE_URL", "http://localhost:9999")
	got := VercelBaseURL()
	if got != "http://localhost:9999" {
		t.Errorf("VercelBaseURL() = %q, want %q", got, "http://localhost:9999")
	}
}
