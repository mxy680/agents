package gcp

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

// withProjectsMock registers handlers for the Resource Manager projects API.
func withProjectsMock(mux *http.ServeMux) {
	// GET /v3/projects — list projects
	// POST /v3/projects — create project (returns completed Operation)
	mux.HandleFunc("/v3/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			// Return a completed operation with the project embedded.
			op := map[string]any{
				"name": "operations/create-proj-op",
				"done": true,
				"response": map[string]any{
					"projectId":   body["projectId"],
					"displayName": body["displayName"],
					"state":       "ACTIVE",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(op)
			return
		}
		resp := map[string]any{
			"projects": []map[string]any{
				{
					"name":        "projects/my-project-1",
					"projectId":   "my-project-1",
					"displayName": "My First Project",
					"state":       "ACTIVE",
					"createTime":  "2023-01-01T00:00:00Z",
				},
				{
					"name":        "projects/my-project-2",
					"projectId":   "my-project-2",
					"displayName": "My Second Project",
					"state":       "ACTIVE",
					"createTime":  "2023-02-01T00:00:00Z",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /v3/projects/my-project-1 — get project details
	// DELETE /v3/projects/my-project-1 — delete project
	mux.HandleFunc("/v3/projects/my-project-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			resp := map[string]any{
				"name": "operations/delete-proj-op",
				"done": true,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"name":        "projects/my-project-1",
			"projectId":   "my-project-1",
			"displayName": "My First Project",
			"state":       "ACTIVE",
			"parent":      "organizations/123456",
			"createTime":  "2023-01-01T00:00:00Z",
			"updateTime":  "2023-06-01T00:00:00Z",
			"etag":        "etag-abc123",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withServicesMock registers handlers for the Service Usage API.
func withServicesMock(mux *http.ServeMux) {
	// GET /v1/projects/my-project-1/services?filter=... — list services
	mux.HandleFunc("/v1/projects/my-project-1/services", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"services": []map[string]any{
				{
					"name":  "projects/my-project-1/services/iap.googleapis.com",
					"state": "ENABLED",
					"config": map[string]any{
						"title": "Cloud Identity-Aware Proxy API",
					},
				},
				{
					"name":  "projects/my-project-1/services/iam.googleapis.com",
					"state": "ENABLED",
					"config": map[string]any{
						"title": "Identity and Access Management (IAM) API",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /v1/projects/my-project-1/services/iap.googleapis.com:enable
	mux.HandleFunc("/v1/projects/my-project-1/services/iap.googleapis.com:enable", func(w http.ResponseWriter, r *http.Request) {
		op := map[string]any{
			"name": "operations/enable-svc-op",
			"done": true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(op)
	})

	// POST /v1/projects/my-project-1/services/iap.googleapis.com:disable
	mux.HandleFunc("/v1/projects/my-project-1/services/iap.googleapis.com:disable", func(w http.ResponseWriter, r *http.Request) {
		op := map[string]any{
			"name": "operations/disable-svc-op",
			"done": true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(op)
	})
}

// withOAuthMock registers handlers for the IAM OAuth clients API.
func withOAuthMock(mux *http.ServeMux) {
	// GET /v1/projects/my-project-1/locations/global/oauthClients — list clients
	// POST /v1/projects/my-project-1/locations/global/oauthClients?oauthClientId=... — create client
	mux.HandleFunc("/v1/projects/my-project-1/locations/global/oauthClients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			clientID := r.URL.Query().Get("oauthClientId")
			resp := map[string]any{
				"name":                "projects/my-project-1/locations/global/oauthClients/" + clientID,
				"clientId":            clientID + "@clientid.example.com",
				"displayName":         body["displayName"],
				"allowedRedirectUris": body["allowedRedirectUris"],
				"disabled":            false,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"oauthClients": []map[string]any{
				{
					"name":                "projects/my-project-1/locations/global/oauthClients/my-app-client",
					"clientId":            "my-app-client@clientid.example.com",
					"displayName":         "My App Client",
					"allowedRedirectUris": []any{"http://localhost:3000/callback"},
					"disabled":            false,
				},
				{
					"name":                "projects/my-project-1/locations/global/oauthClients/another-client",
					"clientId":            "another-client@clientid.example.com",
					"displayName":         "Another Client",
					"allowedRedirectUris": []any{"https://app.example.com/callback"},
					"disabled":            false,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// PATCH /v1/projects/my-project-1/locations/global/oauthClients/my-app-client — update
	// DELETE /v1/projects/my-project-1/locations/global/oauthClients/my-app-client — delete
	mux.HandleFunc("/v1/projects/my-project-1/locations/global/oauthClients/my-app-client", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		if r.Method == http.MethodPatch {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"name":                "projects/my-project-1/locations/global/oauthClients/my-app-client",
				"clientId":            "my-app-client@clientid.example.com",
				"displayName":         "My App Client",
				"allowedRedirectUris": body["allowedRedirectUris"],
				"disabled":            false,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// POST /v1/projects/my-project-1/locations/global/oauthClients/my-app-client/credentials — create credential
	// GET  /v1/projects/my-project-1/locations/global/oauthClients/my-app-client/credentials — list credentials
	mux.HandleFunc("/v1/projects/my-project-1/locations/global/oauthClients/my-app-client/credentials", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			resp := map[string]any{
				"name":         "projects/my-project-1/locations/global/oauthClients/my-app-client/credentials/cred-1",
				"clientId":     "my-app-client@clientid.example.com",
				"clientSecret": "super-secret-value",
				"disabled":     false,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"oauthClientCredentials": []map[string]any{
				{
					"name":     "projects/my-project-1/locations/global/oauthClients/my-app-client/credentials/cred-1",
					"clientId": "my-app-client@clientid.example.com",
					"disabled": false,
				},
				{
					"name":     "projects/my-project-1/locations/global/oauthClients/my-app-client/credentials/cred-2",
					"clientId": "my-app-client@clientid.example.com",
					"disabled": true,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withBrandsMock registers handlers for the IAP brands API.
func withBrandsMock(mux *http.ServeMux) {
	// GET /v1/projects/my-project-1/brands — list brands
	// POST /v1/projects/my-project-1/brands — create brand
	mux.HandleFunc("/v1/projects/my-project-1/brands", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"name":             "projects/my-project-1/brands/123456789",
				"applicationTitle": body["applicationTitle"],
				"supportEmail":     body["supportEmail"],
				"orgInternalOnly":  true,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"brands": []map[string]any{
				{
					"name":             "projects/my-project-1/brands/123456789",
					"applicationTitle": "My Application",
					"supportEmail":     "support@example.com",
					"orgInternalOnly":  true,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /v1/projects/my-project-1/brands/123456789 — get brand
	mux.HandleFunc("/v1/projects/my-project-1/brands/123456789", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"name":             "projects/my-project-1/brands/123456789",
			"applicationTitle": "My Application",
			"supportEmail":     "support@example.com",
			"orgInternalOnly":  true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withIAMMock registers handlers for the IAM service accounts API.
func withIAMMock(mux *http.ServeMux) {
	// GET  /v1/projects/my-project-1/serviceAccounts — list service accounts
	// POST /v1/projects/my-project-1/serviceAccounts — create service account
	mux.HandleFunc("/v1/projects/my-project-1/serviceAccounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			accountID, _ := body["accountId"].(string)
			sa, _ := body["serviceAccount"].(map[string]any)
			displayName := ""
			if sa != nil {
				displayName, _ = sa["displayName"].(string)
			}
			resp := map[string]any{
				"name":        "projects/my-project-1/serviceAccounts/" + accountID + "@my-project-1.iam.gserviceaccount.com",
				"email":       accountID + "@my-project-1.iam.gserviceaccount.com",
				"displayName": displayName,
				"disabled":    false,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"accounts": []map[string]any{
				{
					"name":        "projects/my-project-1/serviceAccounts/svc-acct-1@my-project-1.iam.gserviceaccount.com",
					"email":       "svc-acct-1@my-project-1.iam.gserviceaccount.com",
					"displayName": "Service Account One",
					"disabled":    false,
				},
				{
					"name":        "projects/my-project-1/serviceAccounts/svc-acct-2@my-project-1.iam.gserviceaccount.com",
					"email":       "svc-acct-2@my-project-1.iam.gserviceaccount.com",
					"displayName": "Service Account Two",
					"disabled":    false,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /v1/projects/my-project-1/serviceAccounts/svc-acct-1@.../keys — create key
	mux.HandleFunc("/v1/projects/my-project-1/serviceAccounts/svc-acct-1@my-project-1.iam.gserviceaccount.com/keys", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"name":           "projects/my-project-1/serviceAccounts/svc-acct-1@my-project-1.iam.gserviceaccount.com/keys/key123",
			"privateKeyData": "base64-encoded-json-key-data",
			"keyType":        "USER_MANAGED",
			"validAfterTime": "2023-01-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	// DELETE /v1/projects/my-project-1/serviceAccounts/svc-acct-1@... — delete service account
	mux.HandleFunc("/v1/projects/my-project-1/serviceAccounts/svc-acct-1@my-project-1.iam.gserviceaccount.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

// newFullMockServer creates an httptest.Server with handlers for all GCP API endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withProjectsMock(mux)
	withServicesMock(mux)
	withOAuthMock(mux)
	withBrandsMock(mux)
	withIAMMock(mux)
	return httptest.NewServer(mux)
}

// newTestClientFactory returns a ClientFactory that creates a *Client pointed at the
// test server for all API surfaces (resource manager, service usage, IAM, IAP).
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return &Client{
			http:               server.Client(),
			projectID:          "my-project-1",
			resourceManagerURL: server.URL + "/v3",
			serviceUsageURL:    server.URL + "/v1",
			iamURL:             server.URL + "/v1",
			iapURL:             server.URL + "/v1",
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
