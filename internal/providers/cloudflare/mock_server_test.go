package cloudflare

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

// cfEnvelope wraps a result value in the Cloudflare API envelope.
func cfEnvelope(result any) map[string]any {
	return map[string]any{
		"success":  true,
		"result":   result,
		"errors":   []any{},
		"messages": []any{},
	}
}

// cfEnvelopeNull returns a Cloudflare envelope with a null result (for deletes).
func cfEnvelopeNull() map[string]any {
	return map[string]any{
		"success":  true,
		"result":   nil,
		"errors":   []any{},
		"messages": []any{},
	}
}

// writeJSON writes a JSON-encoded value to the response writer.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// --- Mock handler sections ---

const testAccountID = "acct_test123"
const testZoneID = "zone_abc123"

func withZonesMock(mux *http.ServeMux) {
	// GET /zones — list
	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		result := []map[string]any{
			{
				"id":     testZoneID,
				"name":   "example.com",
				"status": "active",
				"plan":   map[string]any{"name": "Free Website"},
				"type":   "full",
				"paused": false,
				"name_servers":          []any{"ns1.cloudflare.com", "ns2.cloudflare.com"},
				"original_name_servers": []any{"ns1.example.com"},
			},
			{
				"id":     "zone_def456",
				"name":   "example.org",
				"status": "pending",
				"plan":   map[string]any{"name": "Pro"},
				"type":   "full",
				"paused": true,
			},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// GET /zones/{id} — get
	mux.HandleFunc("/zones/"+testZoneID, func(w http.ResponseWriter, r *http.Request) {
		result := map[string]any{
			"id":     testZoneID,
			"name":   "example.com",
			"status": "active",
			"plan":   map[string]any{"name": "Free Website"},
			"type":   "full",
			"paused": false,
			"name_servers":          []any{"ns1.cloudflare.com", "ns2.cloudflare.com"},
			"original_name_servers": []any{"ns1.example.com"},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// POST /zones/{id}/purge_cache
	mux.HandleFunc("/zones/"+testZoneID+"/purge_cache", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, cfEnvelope(map[string]any{"id": testZoneID}))
	})
}

func withDNSMockCF(mux *http.ServeMux) {
	// GET /zones/{id}/dns_records — list
	mux.HandleFunc("/zones/"+testZoneID+"/dns_records", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			result := map[string]any{
				"id":      "rec_new1",
				"type":    body["type"],
				"name":    body["name"],
				"content": body["content"],
				"proxied": body["proxied"],
				"ttl":     body["ttl"],
			}
			writeJSON(w, cfEnvelope(result))
			return
		}
		result := []map[string]any{
			{
				"id":      "rec_abc1",
				"type":    "A",
				"name":    "example.com",
				"content": "1.2.3.4",
				"proxied": true,
				"ttl":     1,
			},
			{
				"id":      "rec_def2",
				"type":    "CNAME",
				"name":    "www.example.com",
				"content": "example.com",
				"proxied": false,
				"ttl":     3600,
			},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// GET/PUT/DELETE /zones/{id}/dns_records/{recID}
	mux.HandleFunc("/zones/"+testZoneID+"/dns_records/rec_abc1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			writeJSON(w, cfEnvelope(map[string]any{"id": "rec_abc1"}))
			return
		}
		if r.Method == http.MethodPut {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			result := map[string]any{
				"id":      "rec_abc1",
				"type":    "A",
				"name":    "example.com",
				"content": "5.6.7.8",
				"proxied": false,
				"ttl":     3600,
			}
			if c, ok := body["content"].(string); ok && c != "" {
				result["content"] = c
			}
			writeJSON(w, cfEnvelope(result))
			return
		}
		result := map[string]any{
			"id":      "rec_abc1",
			"type":    "A",
			"name":    "example.com",
			"content": "1.2.3.4",
			"proxied": true,
			"ttl":     1,
		}
		writeJSON(w, cfEnvelope(result))
	})
}

func withWorkersMock(mux *http.ServeMux) {
	// GET /accounts/{id}/workers/scripts — list
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts", func(w http.ResponseWriter, r *http.Request) {
		result := []map[string]any{
			{
				"id":          "my-worker",
				"etag":        "etag_abc1",
				"created_on":  "2024-01-01T00:00:00Z",
				"modified_on": "2024-06-01T00:00:00Z",
			},
			{
				"id":          "another-worker",
				"etag":        "etag_def2",
				"created_on":  "2024-02-01T00:00:00Z",
				"modified_on": "2024-07-01T00:00:00Z",
			},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// GET/PUT/DELETE /accounts/{id}/workers/scripts/{name}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/my-worker", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			writeJSON(w, cfEnvelopeNull())
			return
		}
		if r.Method == http.MethodPut {
			result := map[string]any{
				"id":          "my-worker",
				"etag":        "etag_new1",
				"created_on":  "2024-01-01T00:00:00Z",
				"modified_on": "2024-08-01T00:00:00Z",
			}
			writeJSON(w, cfEnvelope(result))
			return
		}
		result := map[string]any{
			"id":          "my-worker",
			"etag":        "etag_abc1",
			"created_on":  "2024-01-01T00:00:00Z",
			"modified_on": "2024-06-01T00:00:00Z",
		}
		writeJSON(w, cfEnvelope(result))
	})
}

func withPagesMock(mux *http.ServeMux) {
	// GET /accounts/{id}/pages/projects — list
	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects", func(w http.ResponseWriter, r *http.Request) {
		result := []map[string]any{
			{
				"id":                "pages_abc1",
				"name":              "my-pages-app",
				"subdomain":         "my-pages-app.pages.dev",
				"production_branch": "main",
				"created_on":        "2024-01-01T00:00:00Z",
			},
			{
				"id":                "pages_def2",
				"name":              "another-pages-app",
				"subdomain":         "another-pages-app.pages.dev",
				"production_branch": "master",
				"created_on":        "2024-02-01T00:00:00Z",
			},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// GET /accounts/{id}/pages/projects/my-pages-app
	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/my-pages-app", func(w http.ResponseWriter, r *http.Request) {
		result := map[string]any{
			"id":                "pages_abc1",
			"name":              "my-pages-app",
			"subdomain":         "my-pages-app.pages.dev",
			"production_branch": "main",
			"created_on":        "2024-01-01T00:00:00Z",
		}
		writeJSON(w, cfEnvelope(result))
	})

	// GET /accounts/{id}/pages/projects/my-pages-app/deployments — list
	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/my-pages-app/deployments", func(w http.ResponseWriter, r *http.Request) {
		result := []map[string]any{
			{
				"id":          "deploy_abc1",
				"url":         "https://abc1.my-pages-app.pages.dev",
				"environment": "production",
				"latest_stage": map[string]any{"name": "deploy"},
				"created_on":  "2024-06-01T00:00:00Z",
			},
			{
				"id":          "deploy_def2",
				"url":         "https://def2.my-pages-app.pages.dev",
				"environment": "preview",
				"latest_stage": map[string]any{"name": "build"},
				"created_on":  "2024-07-01T00:00:00Z",
			},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// GET /accounts/{id}/pages/projects/my-pages-app/deployments/deploy_abc1
	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/my-pages-app/deployments/deploy_abc1", func(w http.ResponseWriter, r *http.Request) {
		result := map[string]any{
			"id":          "deploy_abc1",
			"url":         "https://abc1.my-pages-app.pages.dev",
			"environment": "production",
			"latest_stage": map[string]any{"name": "deploy"},
			"created_on":  "2024-06-01T00:00:00Z",
		}
		writeJSON(w, cfEnvelope(result))
	})
}

func withR2Mock(mux *http.ServeMux) {
	// GET /accounts/{id}/r2/buckets — list
	mux.HandleFunc("/accounts/"+testAccountID+"/r2/buckets", func(w http.ResponseWriter, r *http.Request) {
		result := map[string]any{
			"buckets": []map[string]any{
				{
					"name":          "my-bucket",
					"creation_date": "2024-01-01T00:00:00Z",
				},
				{
					"name":          "another-bucket",
					"creation_date": "2024-02-01T00:00:00Z",
				},
			},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// PUT /accounts/{id}/r2/buckets/new-bucket — create
	mux.HandleFunc("/accounts/"+testAccountID+"/r2/buckets/new-bucket", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			writeJSON(w, cfEnvelopeNull())
			return
		}
		if r.Method == http.MethodDelete {
			writeJSON(w, cfEnvelopeNull())
			return
		}
	})

	// DELETE /accounts/{id}/r2/buckets/my-bucket — delete
	mux.HandleFunc("/accounts/"+testAccountID+"/r2/buckets/my-bucket", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			writeJSON(w, cfEnvelopeNull())
			return
		}
		if r.Method == http.MethodPut {
			writeJSON(w, cfEnvelopeNull())
			return
		}
	})
}

func withKVMock(mux *http.ServeMux) {
	// GET /accounts/{id}/storage/kv/namespaces — list
	mux.HandleFunc("/accounts/"+testAccountID+"/storage/kv/namespaces", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			result := map[string]any{
				"id":    "kv_new1",
				"title": body["title"],
			}
			writeJSON(w, cfEnvelope(result))
			return
		}
		result := []map[string]any{
			{
				"id":    "kv_abc1",
				"title": "MY_KV_NS",
			},
			{
				"id":    "kv_def2",
				"title": "ANOTHER_KV_NS",
			},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// GET /accounts/{id}/storage/kv/namespaces/kv_abc1/keys — list keys
	mux.HandleFunc("/accounts/"+testAccountID+"/storage/kv/namespaces/kv_abc1/keys", func(w http.ResponseWriter, r *http.Request) {
		result := []map[string]any{
			{"name": "key-one"},
			{"name": "key-two", "expiration": float64(9999999999)},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// GET/PUT/DELETE /accounts/{id}/storage/kv/namespaces/kv_abc1/values/my-key
	mux.HandleFunc("/accounts/"+testAccountID+"/storage/kv/namespaces/kv_abc1/values/my-key", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// KV get returns raw value — we still wrap in envelope because client.do unwraps it
			// Actually the KV value endpoint returns the raw value (not JSON envelope).
			// The client.do tries to parse as JSON envelope; if it fails and status is 2xx,
			// it returns the raw bytes. So we return plain text here.
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("hello-world"))
			return
		}
		if r.Method == http.MethodPut {
			writeJSON(w, cfEnvelope(nil))
			return
		}
		if r.Method == http.MethodDelete {
			writeJSON(w, cfEnvelope(nil))
			return
		}
	})
}

func withFirewallMock(mux *http.ServeMux) {
	// GET/POST /zones/{id}/firewall/rules
	mux.HandleFunc("/zones/"+testZoneID+"/firewall/rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			result := []map[string]any{
				{
					"id":          "rule_new1",
					"action":      "block",
					"description": "block bad actors",
					"paused":      false,
				},
			}
			writeJSON(w, cfEnvelope(result))
			return
		}
		result := []map[string]any{
			{
				"id":          "rule_abc1",
				"action":      "block",
				"description": "Block known bots",
				"paused":      false,
			},
			{
				"id":          "rule_def2",
				"action":      "challenge",
				"description": "Challenge suspicious IPs",
				"paused":      true,
			},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// DELETE /zones/{id}/firewall/rules/{ruleID}
	mux.HandleFunc("/zones/"+testZoneID+"/firewall/rules/rule_abc1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			writeJSON(w, cfEnvelope(map[string]any{"id": "rule_abc1"}))
			return
		}
	})
}

func withCertsMockCF(mux *http.ServeMux) {
	// GET /zones/{id}/ssl/certificate_packs — list
	mux.HandleFunc("/zones/"+testZoneID+"/ssl/certificate_packs", func(w http.ResponseWriter, r *http.Request) {
		result := []map[string]any{
			{
				"id":     "cert_abc1",
				"type":   "advanced",
				"hosts":  []any{"example.com", "*.example.com"},
				"status": "active",
			},
			{
				"id":     "cert_def2",
				"type":   "universal",
				"hosts":  []any{"example.com"},
				"status": "initializing",
			},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// GET /zones/{id}/ssl/certificate_packs/{certID}
	mux.HandleFunc("/zones/"+testZoneID+"/ssl/certificate_packs/cert_abc1", func(w http.ResponseWriter, r *http.Request) {
		result := map[string]any{
			"id":     "cert_abc1",
			"type":   "advanced",
			"hosts":  []any{"example.com", "*.example.com"},
			"status": "active",
		}
		writeJSON(w, cfEnvelope(result))
	})
}

func withAccountsMock(mux *http.ServeMux) {
	// GET /accounts — list
	mux.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		result := []map[string]any{
			{
				"id":   testAccountID,
				"name": "My Cloudflare Account",
				"type": "standard",
			},
			{
				"id":   "acct_other456",
				"name": "Another Account",
				"type": "enterprise",
			},
		}
		writeJSON(w, cfEnvelope(result))
	})

	// GET /accounts/{id}
	mux.HandleFunc("/accounts/"+testAccountID, func(w http.ResponseWriter, r *http.Request) {
		result := map[string]any{
			"id":   testAccountID,
			"name": "My Cloudflare Account",
			"type": "standard",
		}
		writeJSON(w, cfEnvelope(result))
	})
}

func withIPsMock(mux *http.ServeMux) {
	// GET /ips
	mux.HandleFunc("/ips", func(w http.ResponseWriter, r *http.Request) {
		result := map[string]any{
			"ipv4_cidrs": []any{"103.21.244.0/22", "103.22.200.0/22"},
			"ipv6_cidrs": []any{"2400:cb00::/32", "2606:4700::/32"},
		}
		writeJSON(w, cfEnvelope(result))
	})
}

// newFullMockServer creates an httptest.Server with handlers for all Cloudflare API endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withZonesMock(mux)
	withDNSMockCF(mux)
	withWorkersMock(mux)
	withPagesMock(mux)
	withR2Mock(mux)
	withKVMock(mux)
	withFirewallMock(mux)
	withCertsMockCF(mux)
	withAccountsMock(mux)
	withIPsMock(mux)
	return httptest.NewServer(mux)
}

// newTestClientFactory returns a ClientFactory that creates a *Client pointed at the test server,
// with the test account ID pre-configured.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return &Client{
			http:      server.Client(),
			baseURL:   server.URL,
			accountID: testAccountID,
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

// buildTestCmd creates a subcommand tree for a resource group.
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
