package gcpconsole

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

// --- Mock data ---

var mockClient1 = map[string]any{
	"clientId":      "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com",
	"projectNumber": "123456789012",
	"brandId":       "123456789012",
	"displayName":   "My Web App",
	"type":          "WEB",
	"authType":      "SHARED_SECRET",
	"redirectUris":  []any{"https://example.com/callback", "https://example.com/auth"},
	"creationTime":  "2024-01-15T10:00:00Z",
	"updateTime":    "2024-06-01T12:00:00Z",
	"clientSecrets": []any{
		map[string]any{
			"clientSecret": "GOCSPX-secret1abc",
			"createTime":   "2024-01-15T10:00:00Z",
			"state":        "ENABLED",
			"id":           "secret-id-1",
		},
	},
}

var mockClient2 = map[string]any{
	"clientId":      "123456789012-zyxwvutsrqponmlkjihgfedcba654321.apps.googleusercontent.com",
	"projectNumber": "123456789012",
	"brandId":       "123456789012",
	"displayName":   "Mobile Client",
	"type":          "WEB",
	"authType":      "SHARED_SECRET",
	"redirectUris":  []any{"https://mobile.example.com/callback"},
	"creationTime":  "2024-02-20T08:30:00Z",
	"updateTime":    "2024-05-10T09:00:00Z",
}

// withOAuthMock registers all OAuth client mock endpoints on mux.
func withOAuthMock(mux *http.ServeMux) {
	// GET /v1/clients?projectNumber=X — list clients
	mux.HandleFunc("/v1/clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)

			displayName, _ := body["displayName"].(string)
			projectNumber, _ := body["projectNumber"].(string)
			var uris []string
			if rawURIs, ok := body["redirectUris"].([]any); ok {
				for _, u := range rawURIs {
					if s, ok := u.(string); ok {
						uris = append(uris, s)
					}
				}
			}

			resp := map[string]any{
				"clientId":      "123456789012-newclientid.apps.googleusercontent.com",
				"projectNumber": projectNumber,
				"brandId":       projectNumber,
				"displayName":   displayName,
				"type":          "WEB",
				"authType":      "SHARED_SECRET",
				"redirectUris":  uris,
				"creationTime":  "2024-10-01T00:00:00Z",
				"updateTime":    "2024-10-01T00:00:00Z",
				"clientSecrets": []any{
					map[string]any{
						"clientSecret": "GOCSPX-newsecretvalue",
						"createTime":   "2024-10-01T00:00:00Z",
						"state":        "ENABLED",
						"id":           "new-secret-id",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
			return
		}

		// GET — list
		resp := map[string]any{
			"clients": []any{mockClient1, mockClient2},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	clientID1 := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"

	// GET/PUT/DELETE /v1/clients/{clientId}
	mux.HandleFunc("/v1/clients/"+clientID1, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockClient1)

		case http.MethodPut:
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)

			// Echo back with updated fields
			resp := map[string]any{
				"clientId":      clientID1,
				"projectNumber": "123456789012",
				"brandId":       "123456789012",
				"displayName":   "My Web App",
				"type":          "WEB",
				"authType":      "SHARED_SECRET",
				"redirectUris":  body["redirectUris"],
				"creationTime":  "2024-01-15T10:00:00Z",
				"updateTime":    "2024-10-01T00:00:00Z",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		}
	})
}

// newFullMockServer creates an httptest.Server with handlers for all GCP Console API endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withOAuthMock(mux)
	return httptest.NewServer(mux)
}

// newTestClientFactory returns a ClientFactory that creates a *Client pointed at the test server.
// The base URL is set to serverURL + "/v1" so that all paths like "/clients/..." resolve correctly.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(_ context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL+"/v1"), nil
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

// buildTestCmd creates a parent command with subcommands, for use in tests.
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
