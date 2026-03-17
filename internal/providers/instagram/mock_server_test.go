package instagram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// withProfileMock registers all profile-related mock handlers on mux.
func withProfileMock(mux *http.ServeMux) {
	// GET /api/v1/users/web_profile_info/?username=X
	mux.HandleFunc("/api/v1/users/web_profile_info/", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, `{"status":"fail","message":"missing username"}`, http.StatusBadRequest)
			return
		}
		resp := map[string]any{
			"data": map[string]any{
				"user": map[string]any{
					"pk":                     "42544748138",
					"username":               username,
					"full_name":              "Test User",
					"is_private":             false,
					"is_verified":            false,
					"biography":              "Test bio",
					"external_url":           "https://example.com",
					"profile_pic_url":        "https://example.com/pic.jpg",
					"is_business_account":    false,
					"is_professional_account": false,
					"category_name":          "",
					"edge_followed_by":       100,
					"edge_follow":            50,
				},
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/users/{id}/info/
	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		// path: /api/v1/users/{id}/info/
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) < 2 || parts[1] != "info/" {
			http.NotFound(w, r)
			return
		}
		userID := parts[0]
		resp := map[string]any{
			"user": map[string]any{
				"pk":                     userID,
				"username":               "testuser",
				"full_name":              "Test User",
				"is_private":             false,
				"is_verified":            false,
				"biography":              "Bio from user info",
				"external_url":           "",
				"follower_count":         int64(200),
				"following_count":        int64(80),
				"media_count":            int64(15),
				"total_clips_count":      int64(3),
				"is_business":            false,
				"account_type":           1,
				"profile_pic_url":        "https://example.com/pic2.jpg",
				"has_profile_pic":        true,
				"is_professional_account": false,
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/accounts/edit/web_form_data/
	mux.HandleFunc("/api/v1/accounts/edit/web_form_data/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"form_data": map[string]any{
				"first_name":        "Test",
				"last_name":         "User",
				"email":             "testuser@example.com",
				"username":          "testuser",
				"phone_number":      "+15551234567",
				"gender":            1,
				"biography":         "My bio",
				"external_url":      "https://example.com",
				"is_email_confirmed": true,
				"is_phone_confirmed": false,
				"business_account":  false,
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// newFullMockServer creates an httptest server with all Instagram mock handlers.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withProfileMock(mux)
	return httptest.NewServer(mux)
}

// newTestClientFactory returns a ClientFactory that creates an Instagram Client
// backed by the given httptest server, bypassing real auth entirely.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		session := &auth.InstagramSession{
			SessionID: "test-session-id",
			CSRFToken: "test-csrf-token",
			DSUserID:  "42544748138",
			UserAgent: "test-agent/1.0",
		}
		return newClientWithBase(session, server.Client(), server.URL), nil
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

// buildTestProfileCmd creates a `profile` subcommand tree for use in tests.
func buildTestProfileCmd(factory ClientFactory) *cobra.Command {
	profileCmd := &cobra.Command{Use: "profile", Aliases: []string{"prof"}}
	profileCmd.AddCommand(newProfileGetCmd(factory))
	profileCmd.AddCommand(newProfileEditFormCmd(factory))
	return profileCmd
}

// runCmd is a test helper that executes a cobra command tree with args and returns stdout.
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
	// Silence usage on error so test output is clean
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
