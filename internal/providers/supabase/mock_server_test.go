package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	// GET /v1/projects/available-regions?organization_slug=...
	mux.HandleFunc("/v1/projects/available-regions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"all": map[string]any{
				"specific": []map[string]any{
					{"code": "us-east-1", "name": "East US (North Virginia)", "type": "specific", "provider": "AWS"},
					{"code": "us-west-2", "name": "West US (Oregon)", "type": "specific", "provider": "AWS"},
					{"code": "eu-central-1", "name": "Central EU (Frankfurt)", "type": "specific", "provider": "AWS"},
				},
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

// withDatabaseMock registers database-related mock handlers on mux.
func withDatabaseMock(mux *http.ServeMux) {
	// GET /v1/projects/test-ref/database/migrations — list migrations
	mux.HandleFunc("/v1/projects/test-ref/database/migrations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := []map[string]any{
			{
				"version":    "20260101000000",
				"name":       "create_users_table",
				"statements": []string{"CREATE TABLE users (id uuid PRIMARY KEY)"},
			},
			{
				"version":    "20260102000000",
				"name":       "add_email_column",
				"statements": []string{"ALTER TABLE users ADD COLUMN email text"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// GET /v1/projects/test-ref/types/typescript — get type definitions
	mux.HandleFunc("/v1/projects/test-ref/types/typescript", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "export type Database = { public: { Tables: { users: { Row: { id: string } } } } }")
	})

	// GET /v1/projects/test-ref/ssl-enforcement — get SSL enforcement
	// PUT /v1/projects/test-ref/ssl-enforcement — update SSL enforcement
	mux.HandleFunc("/v1/projects/test-ref/ssl-enforcement", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			enforced, _ := body["enforced"].(bool)
			json.NewEncoder(w).Encode(map[string]any{"enforced": enforced})
			return
		}
		// GET
		json.NewEncoder(w).Encode(map[string]any{"enforced": true})
	})

	// GET /v1/projects/test-ref/jit-access — get JIT access
	// PUT /v1/projects/test-ref/jit-access — update JIT access
	mux.HandleFunc("/v1/projects/test-ref/jit-access", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			json.NewEncoder(w).Encode(map[string]any{"enabled": true, "config": body})
			return
		}
		// GET
		json.NewEncoder(w).Encode(map[string]any{"enabled": false, "allowed_roles": []string{"authenticated"}})
	})
}

// withNetworkMock registers network-related mock handlers on mux.
func withNetworkMock(mux *http.ServeMux) {
	// GET /v1/projects/test-ref/network-restrictions — get restrictions
	// PATCH /v1/projects/test-ref/network-restrictions — update restrictions
	mux.HandleFunc("/v1/projects/test-ref/network-restrictions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			json.NewEncoder(w).Encode(body)
			return
		}
		// GET
		json.NewEncoder(w).Encode(map[string]any{
			"allowed_cidrs":        []string{"0.0.0.0/0"},
			"allowed_cidrs_config": []string{},
		})
	})

	// POST /v1/projects/test-ref/network-restrictions/apply — apply restrictions
	mux.HandleFunc("/v1/projects/test-ref/network-restrictions/apply", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "applied"})
	})

	// GET /v1/projects/test-ref/network-bans/retrieve — list bans
	mux.HandleFunc("/v1/projects/test-ref/network-bans/retrieve", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"banned_ipv4_addresses": []string{"1.2.3.4", "5.6.7.8"},
		})
	})

	// DELETE /v1/projects/test-ref/network-bans — remove bans
	mux.HandleFunc("/v1/projects/test-ref/network-bans", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
}

// withDomainsMock registers domain-related mock handlers on mux.
func withDomainsMock(mux *http.ServeMux) {
	// GET /v1/projects/test-ref/custom-hostname — get custom hostname
	// DELETE /v1/projects/test-ref/custom-hostname — delete custom hostname
	mux.HandleFunc("/v1/projects/test-ref/custom-hostname", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			return
		}
		// GET
		json.NewEncoder(w).Encode(map[string]any{
			"custom_hostname": "app.example.com",
			"status":          "Active",
		})
	})

	// POST /v1/projects/test-ref/custom-hostname/initialize — initialize custom hostname
	mux.HandleFunc("/v1/projects/test-ref/custom-hostname/initialize", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		hostname, _ := body["custom_hostname"].(string)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"custom_hostname": hostname,
			"status":          "Pending",
		})
	})

	// POST /v1/projects/test-ref/custom-hostname/reverify — reverify custom hostname
	mux.HandleFunc("/v1/projects/test-ref/custom-hostname/reverify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "Pending"})
	})

	// POST /v1/projects/test-ref/custom-hostname/activate — activate custom hostname
	mux.HandleFunc("/v1/projects/test-ref/custom-hostname/activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "Active"})
	})

	// GET /v1/projects/test-ref/vanity-subdomain — get vanity subdomain
	// DELETE /v1/projects/test-ref/vanity-subdomain — delete vanity subdomain
	mux.HandleFunc("/v1/projects/test-ref/vanity-subdomain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			return
		}
		// GET
		json.NewEncoder(w).Encode(map[string]any{
			"vanity_subdomain": "my-app",
		})
	})

	// POST /v1/projects/test-ref/vanity-subdomain/check-availability — check availability
	mux.HandleFunc("/v1/projects/test-ref/vanity-subdomain/check-availability", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		subdomain, _ := body["vanity_subdomain"].(string)
		available := subdomain != "taken-subdomain"
		json.NewEncoder(w).Encode(map[string]any{"available": available})
	})

	// POST /v1/projects/test-ref/vanity-subdomain/activate — activate vanity subdomain
	mux.HandleFunc("/v1/projects/test-ref/vanity-subdomain/activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		subdomain, _ := body["vanity_subdomain"].(string)
		json.NewEncoder(w).Encode(map[string]any{"vanity_subdomain": subdomain, "status": "active"})
	})
}

// withRestMock registers PostgREST config mock handlers on mux.
func withRestMock(mux *http.ServeMux) {
	mux.HandleFunc("/v1/projects/test-ref/postgrest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		config := map[string]any{
			"db_schema":            "public",
			"db_extra_search_path": "public, extensions",
			"max_rows":             1000,
			"db_pool":              10,
		}
		if r.Method == http.MethodPatch {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			for k, v := range body {
				config[k] = v
			}
		}
		json.NewEncoder(w).Encode(config)
	})
}

// withAnalyticsMock registers analytics endpoint mock handlers on mux.
func withAnalyticsMock(mux *http.ServeMux) {
	analyticsResponse := map[string]any{
		"result": []map[string]any{
			{"timestamp": "2026-03-16T00:00:00Z", "count": 42},
		},
	}

	for _, endpoint := range []string{
		"logs.all",
		"usage.api-counts",
		"usage.api-requests-count",
		"functions.combined-stats",
	} {
		ep := endpoint // capture loop variable
		mux.HandleFunc("/v1/projects/test-ref/analytics/endpoints/"+ep, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(analyticsResponse)
		})
	}
}

// withAdvisorsMock registers advisor mock handlers on mux.
func withAdvisorsMock(mux *http.ServeMux) {
	advisorData := []map[string]any{
		{
			"title":       "Missing index on foreign key",
			"description": "Table 'orders' has a foreign key without an index",
			"severity":    "WARN",
		},
	}

	for _, advisorType := range []string{"performance", "security"} {
		at := advisorType
		mux.HandleFunc("/v1/projects/test-ref/advisors/"+at, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(advisorData)
		})
	}
}

// withBillingMock registers billing addon mock handlers on mux.
func withBillingMock(mux *http.ServeMux) {
	addonsData := []map[string]any{
		{"variant": "compute_2x", "name": "2x Compute", "status": "active"},
	}

	// GET /v1/projects/test-ref/billing/addons — list addons
	// PATCH /v1/projects/test-ref/billing/addons — apply addon
	mux.HandleFunc("/v1/projects/test-ref/billing/addons", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			json.NewEncoder(w).Encode(map[string]string{"status": "applied"})
			return
		}
		json.NewEncoder(w).Encode(addonsData)
	})

	// DELETE /v1/projects/test-ref/billing/addons/compute_2x — remove addon
	mux.HandleFunc("/v1/projects/test-ref/billing/addons/compute_2x", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
}

// withSnippetsMock registers SQL snippet mock handlers on mux.
func withSnippetsMock(mux *http.ServeMux) {
	snippetData := map[string]any{
		"id":          "snippet-001",
		"name":        "Get recent users",
		"description": "Returns users created in the last 30 days",
		"content":     "SELECT * FROM users WHERE created_at > NOW() - INTERVAL '30 days'",
	}

	// GET /v1/snippets — list snippets
	mux.HandleFunc("/v1/snippets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{snippetData})
	})

	// GET /v1/snippets/snippet-001 — get snippet
	mux.HandleFunc("/v1/snippets/snippet-001", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snippetData)
	})
}

// withActionsMock registers CI/CD action run mock handlers on mux.
func withActionsMock(mux *http.ServeMux) {
	actionData := map[string]any{
		"id":     "run-abc123",
		"status": "completed",
		"type":   "deploy",
	}

	// GET /v1/projects/test-ref/actions — list actions
	mux.HandleFunc("/v1/projects/test-ref/actions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{actionData})
	})

	// GET /v1/projects/test-ref/actions/run-abc123 — get action
	mux.HandleFunc("/v1/projects/test-ref/actions/run-abc123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(actionData)
	})

	// GET /v1/projects/test-ref/actions/run-abc123/logs — action logs
	mux.HandleFunc("/v1/projects/test-ref/actions/run-abc123/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"timestamp": "2026-03-16T00:00:00Z", "message": "Deploy started"},
			{"timestamp": "2026-03-16T00:01:00Z", "message": "Deploy completed"},
		})
	})

	// PATCH /v1/projects/test-ref/actions/run-abc123/status — update status
	mux.HandleFunc("/v1/projects/test-ref/actions/run-abc123/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		status, _ := body["status"].(string)
		json.NewEncoder(w).Encode(map[string]string{"id": "run-abc123", "status": status})
	})
}

// withEncryptionMock registers pgsodium encryption mock handlers on mux.
func withEncryptionMock(mux *http.ServeMux) {
	encryptionData := map[string]any{
		"root_key": "some-root-key-id",
		"enabled":  true,
	}

	mux.HandleFunc("/v1/projects/test-ref/pgsodium", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			for k, v := range body {
				encryptionData[k] = v
			}
		}
		json.NewEncoder(w).Encode(encryptionData)
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
	withDatabaseMock(mux)
	withNetworkMock(mux)
	withDomainsMock(mux)
	withRestMock(mux)
	withAnalyticsMock(mux)
	withAdvisorsMock(mux)
	withBillingMock(mux)
	withSnippetsMock(mux)
	withActionsMock(mux)
	withEncryptionMock(mux)

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
