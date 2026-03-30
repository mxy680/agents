package fly

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

// --- REST mock handlers ---

func withAppsMock(mux *http.ServeMux) {
	// GET /v1/apps — list apps
	mux.HandleFunc("/v1/apps", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":       "app_created1",
				"name":     body["app_name"],
				"status":   "pending",
				"org_slug": body["org_slug"],
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"apps": []map[string]any{
				{
					"id":       "app_abc1",
					"name":     "my-app",
					"status":   "running",
					"org_slug": "my-org",
				},
				{
					"id":       "app_def2",
					"name":     "my-other-app",
					"status":   "suspended",
					"org_slug": "my-org",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET/DELETE /v1/apps/my-app
	mux.HandleFunc("/v1/apps/my-app", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		resp := map[string]any{
			"id":       "app_abc1",
			"name":     "my-app",
			"status":   "running",
			"org_slug": "my-org",
			"hostname": "my-app.fly.dev",
			"app_url":  "https://my-app.fly.dev",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// DELETE /v1/apps/my-app?force=true
	mux.HandleFunc("/v1/apps/my-app-force", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
}

func withMachinesMock(mux *http.ServeMux) {
	// GET /v1/apps/my-app/machines — list
	// POST /v1/apps/my-app/machines — create
	mux.HandleFunc("/v1/apps/my-app/machines", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			config, _ := body["config"].(map[string]any)
			image := ""
			if config != nil {
				image, _ = config["image"].(string)
			}
			resp := map[string]any{
				"id":     "mach_created1",
				"name":   "my-machine",
				"state":  "started",
				"region": "iad",
				"config": map[string]any{"image": image},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{
				"id":     "mach_abc1",
				"name":   "machine-one",
				"state":  "started",
				"region": "iad",
				"config": map[string]any{"image": "registry.fly.io/my-app:latest"},
			},
			{
				"id":     "mach_def2",
				"name":   "machine-two",
				"state":  "stopped",
				"region": "lhr",
				"config": map[string]any{"image": "registry.fly.io/my-app:v2"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET/POST/DELETE /v1/apps/my-app/machines/mach_abc1
	mux.HandleFunc("/v1/apps/my-app/machines/mach_abc1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if r.Method == http.MethodPost {
			// update
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			config, _ := body["config"].(map[string]any)
			image := "registry.fly.io/my-app:latest"
			if config != nil {
				if img, ok := config["image"].(string); ok && img != "" {
					image = img
				}
			}
			resp := map[string]any{
				"id":     "mach_abc1",
				"name":   "machine-one",
				"state":  "started",
				"region": "iad",
				"config": map[string]any{"image": image},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		// GET
		resp := map[string]any{
			"id":          "mach_abc1",
			"name":        "machine-one",
			"state":       "started",
			"region":      "iad",
			"instance_id": "inst_abc1",
			"private_ip":  "fdaa::1",
			"created_at":  "2024-01-01T00:00:00Z",
			"updated_at":  "2024-01-02T00:00:00Z",
			"config":      map[string]any{"image": "registry.fly.io/my-app:latest"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /v1/apps/my-app/machines/mach_abc1/start
	mux.HandleFunc("/v1/apps/my-app/machines/mach_abc1/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "started"})
	})

	// POST /v1/apps/my-app/machines/mach_abc1/stop
	mux.HandleFunc("/v1/apps/my-app/machines/mach_abc1/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
	})

	// GET /v1/apps/my-app/machines/mach_abc1/wait
	mux.HandleFunc("/v1/apps/my-app/machines/mach_abc1/wait", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}

func withVolumesMock(mux *http.ServeMux) {
	// GET/POST /v1/apps/my-app/volumes
	mux.HandleFunc("/v1/apps/my-app/volumes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			size := 1
			if s, ok := body["size_gb"].(float64); ok {
				size = int(s)
			}
			resp := map[string]any{
				"id":      "vol_created1",
				"name":    body["name"],
				"state":   "created",
				"region":  body["region"],
				"size_gb": size,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{
				"id":      "vol_abc1",
				"name":    "my-volume",
				"state":   "created",
				"region":  "iad",
				"size_gb": 10,
			},
			{
				"id":      "vol_def2",
				"name":    "my-other-volume",
				"state":   "created",
				"region":  "lhr",
				"size_gb": 20,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET/DELETE /v1/apps/my-app/volumes/vol_abc1
	mux.HandleFunc("/v1/apps/my-app/volumes/vol_abc1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		resp := map[string]any{
			"id":                  "vol_abc1",
			"name":                "my-volume",
			"state":               "created",
			"region":              "iad",
			"size_gb":             10,
			"encrypted":           true,
			"attached_machine_id": "mach_abc1",
			"created_at":          "2024-01-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// PUT /v1/apps/my-app/volumes/vol_abc1/extend
	mux.HandleFunc("/v1/apps/my-app/volumes/vol_abc1/extend", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		size := 10
		if s, ok := body["size_gb"].(float64); ok {
			size = int(s)
		}
		resp := map[string]any{
			"id":      "vol_abc1",
			"name":    "my-volume",
			"state":   "created",
			"region":  "iad",
			"size_gb": size,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /v1/apps/my-app/volumes/vol_abc1/snapshots
	mux.HandleFunc("/v1/apps/my-app/volumes/vol_abc1/snapshots", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{
				"id":         "snap_abc1",
				"status":     "complete",
				"size":       1073741824,
				"created_at": "2024-01-01T00:00:00Z",
				"digest":     "sha256:abc123",
			},
			{
				"id":         "snap_def2",
				"status":     "complete",
				"size":       2147483648,
				"created_at": "2024-01-02T00:00:00Z",
				"digest":     "sha256:def456",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withCertsMock(mux *http.ServeMux) {
	// GET /v1/apps/my-app/certificates
	mux.HandleFunc("/v1/apps/my-app/certificates", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{
				"hostname":      "example.com",
				"client_status": "Ready",
				"issued":        true,
			},
			{
				"hostname":      "www.example.com",
				"client_status": "Pending",
				"issued":        false,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /v1/apps/my-app/certificates/acme
	mux.HandleFunc("/v1/apps/my-app/certificates/acme", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		resp := map[string]any{
			"hostname":              body["hostname"],
			"client_status":        "Pending",
			"issued":               false,
			"acme_dns_configured":  false,
			"acme_alpn_configured": false,
			"created_at":           "2024-01-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	// GET/DELETE /v1/apps/my-app/certificates/example.com
	mux.HandleFunc("/v1/apps/my-app/certificates/example.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		// GET
		resp := map[string]any{
			"hostname":              "example.com",
			"client_status":        "Ready",
			"issued":               true,
			"acme_dns_configured":  true,
			"acme_alpn_configured": true,
			"dns_validation_target": "_acme-challenge.example.com",
			"created_at":           "2024-01-01T00:00:00Z",
			"expires_at":           "2025-01-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /v1/apps/my-app/certificates/example.com/check
	mux.HandleFunc("/v1/apps/my-app/certificates/example.com/check", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"hostname":              "example.com",
			"client_status":        "Ready",
			"issued":               true,
			"acme_dns_configured":  true,
			"acme_alpn_configured": true,
			"dns_validation_target": "_acme-challenge.example.com",
			"created_at":           "2024-01-01T00:00:00Z",
			"expires_at":           "2025-01-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withGraphQLMock(mux *http.ServeMux) {
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &req)

		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(req.Query, "platform"):
			// regions query
			resp := map[string]any{
				"data": map[string]any{
					"platform": map[string]any{
						"regions": []map[string]any{
							{"code": "iad", "name": "Ashburn, Virginia (US)"},
							{"code": "lhr", "name": "London, United Kingdom"},
							{"code": "nrt", "name": "Tokyo, Japan"},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)

		case strings.Contains(req.Query, "unsetSecrets"):
			// secrets unset mutation — must appear before "setSecrets" check since "unsetSecrets" contains "setSecrets"
			resp := map[string]any{
				"data": map[string]any{
					"unsetSecrets": map[string]any{
						"app": map[string]any{
							"secrets": []map[string]any{
								{"name": "DATABASE_URL", "digest": "def456digest", "createdAt": "2024-01-01T00:00:00Z"},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)

		case strings.Contains(req.Query, "setSecrets"):
			// secrets set mutation
			resp := map[string]any{
				"data": map[string]any{
					"setSecrets": map[string]any{
						"app": map[string]any{
							"secrets": []map[string]any{
								{"name": "API_KEY", "digest": "abc123digest", "createdAt": "2024-01-01T00:00:00Z"},
								{"name": "DATABASE_URL", "digest": "def456digest", "createdAt": "2024-01-01T00:00:00Z"},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)

		default:
			// secrets list query (app.secrets)
			resp := map[string]any{
				"data": map[string]any{
					"app": map[string]any{
						"secrets": []map[string]any{
							{"name": "API_KEY", "digest": "abc123digest", "createdAt": "2024-01-01T00:00:00Z"},
							{"name": "DATABASE_URL", "digest": "def456digest", "createdAt": "2024-01-01T00:00:00Z"},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	})
}

// newFullMockServer creates an httptest.Server with handlers for all Fly.io API endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withAppsMock(mux)
	withMachinesMock(mux)
	withVolumesMock(mux)
	withCertsMock(mux)
	withGraphQLMock(mux)
	return httptest.NewServer(mux)
}

// newTestClientFactory returns a ClientFactory that creates a *Client pointed at the test server.
// Both baseURL and graphqlURL are set to the same test server since the mock handles both REST and GraphQL.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return &Client{
			http:       server.Client(),
			baseURL:    server.URL,
			graphqlURL: server.URL + "/graphql",
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
