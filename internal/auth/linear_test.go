package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewLinearClient_MissingAPIKey(t *testing.T) {
	os.Unsetenv("LINEAR_API_KEY")

	_, err := NewLinearClient(context.Background())
	if err == nil {
		t.Fatal("expected error for missing LINEAR_API_KEY, got nil")
	}
}

func TestNewLinearClient_Valid(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "lin_api_test_key")

	client, err := NewLinearClient(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestLinearClient_InjectsAuthHeader(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "lin_api_my_secret")

	var gotHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client, err := NewLinearClient(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	resp.Body.Close()

	want := "Bearer lin_api_my_secret"
	if gotHeader != want {
		t.Errorf("Authorization header = %q, want %q", gotHeader, want)
	}
}

func TestLinearBaseURL_Default(t *testing.T) {
	os.Unsetenv("LINEAR_API_BASE_URL")
	got := LinearBaseURL()
	if got != "https://api.linear.app/graphql" {
		t.Errorf("LinearBaseURL() = %q, want %q", got, "https://api.linear.app/graphql")
	}
}

func TestLinearBaseURL_Override(t *testing.T) {
	t.Setenv("LINEAR_API_BASE_URL", "http://localhost:9999/graphql")
	got := LinearBaseURL()
	if got != "http://localhost:9999/graphql" {
		t.Errorf("LinearBaseURL() = %q, want %q", got, "http://localhost:9999/graphql")
	}
}
