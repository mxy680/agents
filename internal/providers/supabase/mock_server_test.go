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

// withBranchesMock registers branch-related mock handlers on mux.
func withBranchesMock(mux *http.ServeMux) {
	branchData := map[string]any{
		"id":         "branch-001",
		"name":       "feat-login",
		"git_branch": "feat/login",
		"is_default": false,
		"status":     "ACTIVE_HEALTHY",
		"created_at": "2026-01-15T10:30:00Z",
	}

	// GET /v1/projects/test-ref/branches — list branches
	// POST /v1/projects/test-ref/branches — create branch
	// DELETE /v1/projects/test-ref/branches — disable branching
	mux.HandleFunc("/v1/projects/test-ref/branches", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode([]map[string]any{branchData})
		case http.MethodPost:
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":         "branch-001",
				"name":       "feat-login",
				"git_branch": body["git_branch"],
				"is_default": false,
				"status":     "ACTIVE_HEALTHY",
				"created_at": "2026-01-15T10:30:00Z",
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET /v1/branches/branch-001 — get branch
	// PATCH /v1/branches/branch-001 — update branch
	// DELETE /v1/branches/branch-001 — delete branch
	mux.HandleFunc("/v1/branches/branch-001", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(branchData)
		case http.MethodPatch:
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":         "branch-001",
				"name":       "feat-login",
				"git_branch": "feat/login",
				"is_default": false,
				"status":     "ACTIVE_HEALTHY",
				"created_at": "2026-01-15T10:30:00Z",
			}
			if gb, ok := body["git_branch"].(string); ok && gb != "" {
				resp["git_branch"] = gb
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// POST /v1/branches/branch-001/push — push
	mux.HandleFunc("/v1/branches/branch-001/push", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "pushed"})
	})

	// POST /v1/branches/branch-001/merge — merge
	mux.HandleFunc("/v1/branches/branch-001/merge", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "merged"})
	})

	// POST /v1/branches/branch-001/reset — reset
	mux.HandleFunc("/v1/branches/branch-001/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "reset"})
	})

	// GET /v1/branches/branch-001/diff — diff
	mux.HandleFunc("/v1/branches/branch-001/diff", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"diff": "--- main\n+++ feat-login\n@@ -1,0 +1,5 @@\n+CREATE TABLE users (id uuid);"})
	})
}

// withKeysMock registers API key-related mock handlers on mux.
func withKeysMock(mux *http.ServeMux) {
	keyData := map[string]any{
		"id":      "key-001",
		"name":    "anon",
		"api_key": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test-anon-key",
		"type":    "anon",
	}

	// GET /v1/projects/test-ref/api-keys — list keys
	// POST /v1/projects/test-ref/api-keys — create key
	mux.HandleFunc("/v1/projects/test-ref/api-keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode([]map[string]any{keyData})
		case http.MethodPost:
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			name, _ := body["name"].(string)
			keyType, _ := body["type"].(string)
			resp := map[string]any{
				"id":      "key-001",
				"name":    name,
				"api_key": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test-anon-key",
				"type":    keyType,
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET /v1/projects/test-ref/api-keys/key-001 — get key
	// PATCH /v1/projects/test-ref/api-keys/key-001 — update key
	// DELETE /v1/projects/test-ref/api-keys/key-001 — delete key
	mux.HandleFunc("/v1/projects/test-ref/api-keys/key-001", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(keyData)
		case http.MethodPatch:
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":      "key-001",
				"name":    "anon",
				"api_key": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test-anon-key",
				"type":    "anon",
			}
			if name, ok := body["name"].(string); ok && name != "" {
				resp["name"] = name
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// withSecretsMock registers Edge Function secret mock handlers on mux.
func withSecretsMock(mux *http.ServeMux) {
	secretsData := []map[string]any{
		{"name": "MY_SECRET", "value": "super-secret-value"},
		{"name": "DB_URL", "value": "postgresql://user:pass@host:5432/db"},
	}

	// GET /v1/projects/test-ref/secrets — list secrets
	// POST /v1/projects/test-ref/secrets — create/upsert secrets
	// DELETE /v1/projects/test-ref/secrets — delete secrets
	mux.HandleFunc("/v1/projects/test-ref/secrets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(secretsData)
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"status": "created"})
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// withAuthMock registers Auth configuration mock handlers on mux.
func withAuthMock(mux *http.ServeMux) {
	authConfig := map[string]any{
		"site_url":              "http://localhost:3000",
		"jwt_exp":               3600,
		"enable_signup":         true,
		"mailer_autoconfirm":    false,
		"sms_autoconfirm":       false,
		"external_email_enabled": true,
	}

	signingKeysData := []map[string]any{
		{"id": "sk-001", "status": "active", "created_at": "2026-01-01T00:00:00Z"},
	}

	tpaData := []map[string]any{
		{"id": "tpa-001", "type": "google", "created_at": "2026-01-01T00:00:00Z"},
	}

	// GET /v1/projects/test-ref/config/auth — get auth config
	// PATCH /v1/projects/test-ref/config/auth — update auth config
	mux.HandleFunc("/v1/projects/test-ref/config/auth", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(authConfig)
		case http.MethodPatch:
			json.NewEncoder(w).Encode(authConfig)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET /v1/projects/test-ref/config/auth/signing-keys — list signing keys
	// POST /v1/projects/test-ref/config/auth/signing-keys — create signing key
	mux.HandleFunc("/v1/projects/test-ref/config/auth/signing-keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(signingKeysData)
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(signingKeysData[0])
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET /v1/projects/test-ref/config/auth/signing-keys/sk-001 — get signing key
	// PATCH /v1/projects/test-ref/config/auth/signing-keys/sk-001 — update signing key
	// DELETE /v1/projects/test-ref/config/auth/signing-keys/sk-001 — delete signing key
	mux.HandleFunc("/v1/projects/test-ref/config/auth/signing-keys/sk-001", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(signingKeysData[0])
		case http.MethodPatch:
			json.NewEncoder(w).Encode(signingKeysData[0])
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET /v1/projects/test-ref/config/auth/third-party-auth — list TPA
	// POST /v1/projects/test-ref/config/auth/third-party-auth — create TPA
	mux.HandleFunc("/v1/projects/test-ref/config/auth/third-party-auth", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(tpaData)
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(tpaData[0])
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET /v1/projects/test-ref/config/auth/third-party-auth/tpa-001 — get TPA
	// DELETE /v1/projects/test-ref/config/auth/third-party-auth/tpa-001 — delete TPA
	mux.HandleFunc("/v1/projects/test-ref/config/auth/third-party-auth/tpa-001", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(tpaData[0])
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// newFullMockServer creates a test HTTP server with all Supabase Management API routes mocked.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	withProjectsMock(mux)
	withOrgsMock(mux)
	withBranchesMock(mux)
	withKeysMock(mux)
	withSecretsMock(mux)
	withAuthMock(mux)

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
