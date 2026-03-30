package vercel

import (
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

func withProjectsMock(mux *http.ServeMux) {
	// POST /v11/projects — create project
	mux.HandleFunc("/v11/projects", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		name, _ := body["name"].(string)
		resp := map[string]any{
			"id":          "prj_created1",
			"name":        name,
			"framework":   "nextjs",
			"nodeVersion": "18.x",
			"createdAt":   1700000000000,
			"updatedAt":   1700000000000,
			"accountId":   "acct_123",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v10/projects", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"projects": []map[string]any{
				{
					"id":          "prj_abc123",
					"name":        "my-nextjs-app",
					"framework":   "nextjs",
					"nodeVersion": "18.x",
					"createdAt":   1700000000000,
					"updatedAt":   1700001000000,
				},
				{
					"id":          "prj_def456",
					"name":        "my-vite-app",
					"framework":   "vite",
					"nodeVersion": "20.x",
					"createdAt":   1700100000000,
					"updatedAt":   1700200000000,
				},
			},
			"pagination": map[string]any{"next": 0},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v9/projects/my-nextjs-app", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":          "prj_abc123",
				"name":        "my-nextjs-app",
				"framework":   "nextjs",
				"nodeVersion": "18.x",
				"createdAt":   1700000000000,
				"updatedAt":   1700002000000,
				"accountId":   "acct_123",
			}
			if n, ok := body["name"].(string); ok && n != "" {
				resp["name"] = n
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resp := map[string]any{
			"id":              "prj_abc123",
			"name":            "my-nextjs-app",
			"framework":       "nextjs",
			"nodeVersion":     "18.x",
			"rootDirectory":   ".",
			"buildCommand":    "npm run build",
			"outputDirectory": ".next",
			"installCommand":  "npm install",
			"devCommand":      "npm run dev",
			"createdAt":       1700000000000,
			"updatedAt":       1700001000000,
			"accountId":       "acct_123",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withDeploymentsMock(mux *http.ServeMux) {
	mux.HandleFunc("/v6/deployments", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"deployments": []map[string]any{
				{
					"id":        "dpl_abc123",
					"url":       "my-app-abc123.vercel.app",
					"name":      "my-nextjs-app",
					"state":     "READY",
					"type":      "LAMBDAS",
					"target":    "production",
					"source":    "cli",
					"createdAt": 1700000000000,
				},
				{
					"id":        "dpl_def456",
					"url":       "my-app-def456.vercel.app",
					"name":      "my-nextjs-app",
					"state":     "BUILDING",
					"type":      "LAMBDAS",
					"target":    "preview",
					"source":    "git",
					"createdAt": 1700100000000,
				},
			},
			"pagination": map[string]any{"next": 0},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v13/deployments/dpl_abc123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resp := map[string]any{
			"id":         "dpl_abc123",
			"url":        "my-app-abc123.vercel.app",
			"name":       "my-nextjs-app",
			"state":      "READY",
			"readyState": "READY",
			"type":       "LAMBDAS",
			"target":     "production",
			"source":     "cli",
			"creator":    map[string]any{"username": "alice"},
			"meta": map[string]any{
				"githubCommitRef": "main",
				"githubCommitSha": "abc123def456",
			},
			"createdAt":  1700000000000,
			"buildingAt": 1700000100000,
			"ready":      1700000500000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v13/deployments", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		resp := map[string]any{
			"id":        "dpl_new789",
			"url":       "my-app-new789.vercel.app",
			"name":      body["name"],
			"state":     "INITIALIZING",
			"target":    body["target"],
			"source":    "cli",
			"createdAt": 1700200000000,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v12/deployments/dpl_abc123/cancel", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":    "dpl_abc123",
			"state": "CANCELED",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withDomainsMock(mux *http.ServeMux) {
	mux.HandleFunc("/v5/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":        "dom_added1",
				"name":      body["name"],
				"verified":  false,
				"createdAt": 1700000000000,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"domains": []map[string]any{
				{
					"id":        "dom_abc1",
					"name":      "example.com",
					"verified":  true,
					"createdAt": 1700000000000,
					"expiresAt": 1731536000000,
				},
				{
					"id":        "dom_def2",
					"name":      "example.org",
					"verified":  false,
					"createdAt": 1700100000000,
				},
			},
			"pagination": map[string]any{"next": 0},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v5/domains/example.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resp := map[string]any{
			"domain": map[string]any{
				"id":                  "dom_abc1",
				"name":                "example.com",
				"verified":            true,
				"serviceType":         "external",
				"nameservers":         []any{"ns1.vercel.com", "ns2.vercel.com"},
				"intendedNameservers": []any{"ns1.vercel.com", "ns2.vercel.com"},
				"createdAt":           1700000000000,
				"expiresAt":           1731536000000,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v5/domains/example.com/verify", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"verified": true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withEnvMock(mux *http.ServeMux) {
	mux.HandleFunc("/v10/projects/prj_abc123/env", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":        "env_created1",
				"key":       body["key"],
				"value":     body["value"],
				"type":      "plain",
				"target":    body["target"],
				"createdAt": 1700000000000,
				"updatedAt": 1700000000000,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"envs": []map[string]any{
				{
					"id":        "env_abc1",
					"key":       "API_KEY",
					"value":     "",
					"type":      "encrypted",
					"target":    []any{"production"},
					"createdAt": 1700000000000,
					"updatedAt": 1700000000000,
				},
				{
					"id":        "env_def2",
					"key":       "DATABASE_URL",
					"value":     "",
					"type":      "plain",
					"target":    []any{"production", "preview"},
					"createdAt": 1700100000000,
					"updatedAt": 1700100000000,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v10/projects/prj_abc123/env/env_abc1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resp := map[string]any{
			"id":        "env_abc1",
			"key":       "API_KEY",
			"value":     "secret-value",
			"type":      "encrypted",
			"target":    []any{"production"},
			"createdAt": 1700000000000,
			"updatedAt": 1700000000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withDNSMock(mux *http.ServeMux) {
	mux.HandleFunc("/v4/domains/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"records": []map[string]any{
				{
					"id":        "rec_abc1",
					"name":      "www",
					"type":      "CNAME",
					"value":     "cname.vercel-dns.com",
					"ttl":       3600,
					"createdAt": 1700000000000,
				},
				{
					"id":        "rec_def2",
					"name":      "@",
					"type":      "A",
					"value":     "76.76.21.21",
					"ttl":       3600,
					"createdAt": 1700100000000,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v2/domains/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		resp := map[string]any{
			"id":        "rec_new1",
			"name":      body["name"],
			"type":      body["type"],
			"value":     body["value"],
			"ttl":       3600,
			"createdAt": 1700200000000,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v2/domains/example.com/records/rec_abc1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

func withCertsMock(mux *http.ServeMux) {
	mux.HandleFunc("/v4/certs", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"certs": []map[string]any{
				{
					"id":         "cert_abc1",
					"cns":        []any{"example.com", "www.example.com"},
					"expiration": 1731536000000,
					"createdAt":  1700000000000,
					"autoRenew":  true,
				},
				{
					"id":         "cert_def2",
					"cns":        []any{"example.org"},
					"expiration": 1731536000000,
					"createdAt":  1700100000000,
					"autoRenew":  false,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v4/certs/cert_abc1", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":         "cert_abc1",
			"cns":        []any{"example.com", "www.example.com"},
			"expiration": 1731536000000,
			"createdAt":  1700000000000,
			"autoRenew":  true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withTeamsMock(mux *http.ServeMux) {
	mux.HandleFunc("/v2/teams", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"teams": []map[string]any{
				{
					"id":        "team_abc1",
					"slug":      "acme-corp",
					"name":      "Acme Corp",
					"createdAt": 1700000000000,
				},
				{
					"id":        "team_def2",
					"slug":      "startup-xyz",
					"name":      "Startup XYZ",
					"createdAt": 1700100000000,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v2/teams/team_abc1", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":        "team_abc1",
			"slug":      "acme-corp",
			"name":      "Acme Corp",
			"createdAt": 1700000000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v2/teams/team_abc1/members", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"members": []map[string]any{
				{
					"uid":      "usr_alice1",
					"username": "alice",
					"email":    "alice@example.com",
					"role":     "OWNER",
					"joinedAt": 1700000000000,
				},
				{
					"uid":      "usr_bob2",
					"username": "bob",
					"email":    "bob@example.com",
					"role":     "MEMBER",
					"joinedAt": 1700100000000,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withAliasesMock(mux *http.ServeMux) {
	mux.HandleFunc("/v2/deployments/dpl_abc123/aliases", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"uid":          "ali_new1",
				"alias":        body["alias"],
				"deploymentId": "dpl_abc123",
				"createdAt":    1700200000000,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"aliases": []map[string]any{
				{
					"uid":          "ali_abc1",
					"alias":        "my-app.vercel.app",
					"deploymentId": "dpl_abc123",
					"createdAt":    1700000000000,
				},
				{
					"uid":          "ali_def2",
					"alias":        "my-app-custom.example.com",
					"deploymentId": "dpl_abc123",
					"createdAt":    1700100000000,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withLogsMock(mux *http.ServeMux) {
	mux.HandleFunc("/v2/deployments/dpl_abc123/events", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{
				"id":           "evt_abc1",
				"text":         "Build started",
				"type":         "stdout",
				"source":       "build",
				"deploymentId": "dpl_abc123",
				"date":         1700000100000,
			},
			{
				"id":           "evt_def2",
				"text":         "Build completed successfully",
				"type":         "stdout",
				"source":       "build",
				"deploymentId": "dpl_abc123",
				"date":         1700000500000,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withWebhooksMock(mux *http.ServeMux) {
	mux.HandleFunc("/v1/webhooks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":        "hook_new1",
				"url":       body["url"],
				"events":    body["events"],
				"createdAt": 1700200000000,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{
				"id":        "hook_abc1",
				"url":       "https://hooks.example.com/vercel",
				"events":    []any{"deployment.created", "deployment.ready"},
				"createdAt": 1700000000000,
			},
			{
				"id":        "hook_def2",
				"url":       "https://hooks.example.com/vercel2",
				"events":    []any{"deployment.error"},
				"createdAt": 1700100000000,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v1/webhooks/hook_abc1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

// newFullMockServer creates an httptest.Server with handlers for all Vercel API endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withProjectsMock(mux)
	withDeploymentsMock(mux)
	withDomainsMock(mux)
	withEnvMock(mux)
	withDNSMock(mux)
	withCertsMock(mux)
	withTeamsMock(mux)
	withAliasesMock(mux)
	withLogsMock(mux)
	withWebhooksMock(mux)
	return httptest.NewServer(mux)
}

// newTestClientFactory returns a ClientFactory that creates a *Client pointed at the test server.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return &Client{
			http:    server.Client(),
			baseURL: server.URL,
			teamID:  "",
		}, nil
	}
}

// captureStdout runs f with os.Stdout redirected to a pipe and returns the output.
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

	out, _ := io.ReadAll(r)
	return string(out)
}

// newTestRootCmd creates a root command with global flags for testing.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}

// buildTestCmd creates a subcommand tree for a resource group for use in tests.
func buildTestCmd(name string, cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{Use: name}
	for _, c := range cmds {
		cmd.AddCommand(c)
	}
	return cmd
}

// runCmd executes a cobra command tree with args and returns stdout.
func runCmd(t *testing.T, root *cobra.Command, args ...string) string {
	t.Helper()
	return captureStdout(t, func() {
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}
	})
}

// runCmdErr executes a cobra command tree and returns any error (does not fatal).
func runCmdErr(t *testing.T, root *cobra.Command, args ...string) error {
	t.Helper()
	root.SetArgs(args)
	root.SilenceUsage = true
	root.SilenceErrors = true
	return root.Execute()
}

// mustContain asserts that output contains substr.
func mustContain(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, output)
	}
}
