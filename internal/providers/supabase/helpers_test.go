package supabase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello w…"},
		{"ab", 5, "ab"},
		{"", 5, ""},
		{"abcdef", 6, "abcdef"},
		{"abcdefg", 6, "abcde…"},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abcdefgh", "abcd****efgh"},
		{"short", "****"},
		{"1234567", "****"},
		{"eyJhbGciOiJIUzI1NiJ9.abc.xyz_end", "eyJh****_end"},
		{"", "****"},
	}
	for _, tt := range tests {
		got := maskKey(tt.input)
		if got != tt.want {
			t.Errorf("maskKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDoSupabaseSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":"abc","name":"test"}]`)
	}))
	t.Cleanup(server.Close)

	t.Setenv("SUPABASE_API_BASE_URL", server.URL)

	data, err := doSupabase(server.Client(), http.MethodGet, "/projects", nil)
	if err != nil {
		t.Fatalf("doSupabase returned error: %v", err)
	}

	var projects []map[string]any
	if err := json.Unmarshal(data, &projects); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(projects) != 1 || projects[0]["id"] != "abc" {
		t.Errorf("unexpected projects: %v", projects)
	}
}

func TestDoSupabaseErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"Project not found"}`)
	}))
	t.Cleanup(server.Close)

	t.Setenv("SUPABASE_API_BASE_URL", server.URL)

	_, err := doSupabase(server.Client(), http.MethodGet, "/projects/nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "API error 404") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "API error 404")
	}
}

func TestDoSupabaseBadURL(t *testing.T) {
	t.Setenv("SUPABASE_API_BASE_URL", "://invalid-url")

	_, err := doSupabase(&http.Client{}, http.MethodGet, "/projects", nil)
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}
