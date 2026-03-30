package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewFlyClient_MissingToken(t *testing.T) {
	os.Unsetenv("FLY_API_TOKEN")

	_, err := NewFlyClient(context.Background())
	if err == nil {
		t.Fatal("expected error for missing FLY_API_TOKEN, got nil")
	}
}

func TestNewFlyClient_Valid(t *testing.T) {
	t.Setenv("FLY_API_TOKEN", "test-fly-token")

	client, err := NewFlyClient(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestFlyClient_InjectsAuthHeader(t *testing.T) {
	t.Setenv("FLY_API_TOKEN", "my-fly-secret")

	var gotHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client, err := NewFlyClient(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	resp.Body.Close()

	want := "Bearer my-fly-secret"
	if gotHeader != want {
		t.Errorf("Authorization header = %q, want %q", gotHeader, want)
	}
}

func TestFlyBaseURL_Default(t *testing.T) {
	os.Unsetenv("FLY_API_BASE_URL")
	got := FlyBaseURL()
	if got != "https://api.machines.dev" {
		t.Errorf("FlyBaseURL() = %q, want %q", got, "https://api.machines.dev")
	}
}

func TestFlyBaseURL_Override(t *testing.T) {
	t.Setenv("FLY_API_BASE_URL", "http://localhost:9999")
	got := FlyBaseURL()
	if got != "http://localhost:9999" {
		t.Errorf("FlyBaseURL() = %q, want %q", got, "http://localhost:9999")
	}
}
