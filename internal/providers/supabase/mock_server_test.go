package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// --- Mock server setup ---

// withProjectsMock registers project-related mock handlers on mux.
func withProjectsMock(mux *http.ServeMux) {
	// GET /v1/projects — list projects
	// POST /v1/projects — create project
	mux.HandleFunc("/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			name, _ := body["name"].(string)
			resp := map[string]any{
				"id":              "newprojectref",
				"name":            name,
				"organization_id": "org-uuid-1234",
				"region":          "us-east-1",
				"status":          "ACTIVE_HEALTHY",
				"created_at":      "2026-03-16T00:00:00Z",
				"db_host":         "db.newprojectref.supabase.co",
				"db_version":      "15.1.0.117",
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{
				"id":              "abcdefghijkl",
				"name":            "my-app",
				"organization_id": "org-uuid-1234",
				"region":          "us-east-1",
				"status":          "ACTIVE_HEALTHY",
				"created_at":      "2025-06-01T00:00:00Z",
				"db_host":         "db.abcdefghijkl.supabase.co",
				"db_version":      "15.1.0.117",
			},
			{
				"id":              "mnopqrstuvwx",
				"name":            "my-staging",
				"organization_id": "org-uuid-1234",
				"region":          "us-west-2",
				"status":          "ACTIVE_HEALTHY",
				"created_at":      "2025-08-15T00:00:00Z",
				"db_host":         "db.mnopqrstuvwx.supabase.co",
				"db_version":      "15.1.0.117",
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// GET /v1/projects/test-ref — single project detail
	// PATCH /v1/projects/test-ref — update project
	// DELETE /v1/projects/test-ref — delete project
	mux.HandleFunc("/v1/projects/test-ref", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPatch {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":              "test-ref",
				"name":            "updated-project",
				"organization_id": "org-uuid-1234",
				"region":          "us-east-1",
				"status":          "ACTIVE_HEALTHY",
				"created_at":      "2025-06-01T00:00:00Z",
				"db_host":         "db.test-ref.supabase.co",
				"db_version":      "15.1.0.117",
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		// GET
		resp := map[string]any{
			"id":              "test-ref",
			"name":            "my-app",
			"organization_id": "org-uuid-1234",
			"region":          "us-east-1",
			"status":          "ACTIVE_HEALTHY",
			"created_at":      "2025-06-01T00:00:00Z",
			"db_host":         "db.test-ref.supabase.co",
			"db_version":      "15.1.0.117",
		}
		json.NewEncoder(w).Encode(resp)
	})

	// GET /v1/projects/available-regions
	mux.HandleFunc("/v1/projects/available-regions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"regions": []map[string]any{
				{"key": "us-east-1", "displayName": "US East (N. Virginia)"},
				{"key": "us-west-2", "displayName": "US West (Oregon)"},
				{"key": "eu-central-1", "displayName": "EU Central (Frankfurt)"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// POST /v1/projects/test-ref/pause
	mux.HandleFunc("/v1/projects/test-ref/pause", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "PAUSED"})
	})

	// POST /v1/projects/test-ref/restore
	mux.HandleFunc("/v1/projects/test-ref/restore", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "ACTIVE_HEALTHY"})
	})

	// GET /v1/projects/test-ref/health
	mux.HandleFunc("/v1/projects/test-ref/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := []map[string]any{
			{"name": "database", "status": "HEALTHY", "error": ""},
			{"name": "auth", "status": "HEALTHY", "error": ""},
			{"name": "storage", "status": "HEALTHY", "error": ""},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// withOrgsMock registers organization-related mock handlers on mux.
func withOrgsMock(mux *http.ServeMux) {
	// GET /v1/organizations — list orgs
	// POST /v1/organizations — create org
	mux.HandleFunc("/v1/organizations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			name, _ := body["name"].(string)
			resp := map[string]any{
				"id":   "org-new-5678",
				"name": name,
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{"id": "org-uuid-1234", "name": "Acme Corp"},
			{"id": "org-uuid-5678", "name": "My Personal Org"},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// newFullMockServer creates a test HTTP server with all Supabase Management API routes mocked.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	withProjectsMock(mux)
	withOrgsMock(mux)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// newTestClientFactory returns a ClientFactory that creates an HTTP client
// pointing at the given mock server, using the SUPABASE_API_BASE_URL env var.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*http.Client, error) {
		return server.Client(), nil
	}
}

// captureStdout captures anything written to os.Stdout during f's execution.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// newTestRootCmd creates a minimal root command with persistent flags required
// by provider subcommands (--json, --dry-run).
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "Output as JSON")
	root.PersistentFlags().Bool("dry-run", false, "Print what would happen without executing")
	return root
}

// setEnv sets an environment variable and registers a cleanup to restore it.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	old, hadOld := os.LookupEnv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if hadOld {
			os.Setenv(key, old)
		} else {
			os.Unsetenv(key)
		}
	})
}

// mustContain fails the test if s does not contain substr.
func mustContain(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, s)
	}
}
