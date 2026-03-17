package drive

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
	api "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// withFilesMock registers all file-related mock handlers on mux.
func withFilesMock(mux *http.ServeMux) {
	// files.create (upload) — multipart upload goes to /upload/drive/v3/files
	mux.HandleFunc("/upload/drive/v3/files", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":           "file-uploaded1",
			"name":         "uploaded.txt",
			"mimeType":     "text/plain",
			"size":         "42",
			"modifiedTime": "2026-03-16T10:00:00Z",
			"createdTime":  "2026-03-16T10:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// files.list
	mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		// files.list (GET)
		resp := map[string]any{
			"files": []map[string]any{
				{
					"id":           "file1",
					"name":         "Project Plan.docx",
					"mimeType":     "application/vnd.google-apps.document",
					"size":         "0",
					"modifiedTime": "2026-03-15T10:00:00Z",
					"parents":      []string{"root"},
				},
				{
					"id":           "file2",
					"name":         "Budget.xlsx",
					"mimeType":     "application/vnd.google-apps.spreadsheet",
					"size":         "0",
					"modifiedTime": "2026-03-14T10:00:00Z",
					"parents":      []string{"folder1"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// files.get, files.update, files.delete for file1
	mux.HandleFunc("/files/file1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodPatch {
			// files.update (trash, untrash, move)
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":           "file1",
				"name":         "Project Plan.docx",
				"mimeType":     "application/vnd.google-apps.document",
				"modifiedTime": "2026-03-16T12:00:00Z",
				"parents":      []string{"root"},
			}
			if trashed, ok := body["trashed"]; ok {
				resp["trashed"] = trashed
			}
			// If move: check for addParents query param
			if addP := r.URL.Query().Get("addParents"); addP != "" {
				resp["parents"] = []string{addP}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		// files.get (GET)
		// Check for alt=media (download)
		if r.URL.Query().Get("alt") == "media" {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte("file-content-here"))
			return
		}
		resp := map[string]any{
			"id":             "file1",
			"name":           "Project Plan.docx",
			"mimeType":       "application/vnd.google-apps.document",
			"size":           "1024",
			"modifiedTime":   "2026-03-15T10:00:00Z",
			"createdTime":    "2026-03-01T10:00:00Z",
			"parents":        []string{"root"},
			"description":    "Q1 project plan",
			"webViewLink":    "https://docs.google.com/document/d/file1",
			"webContentLink": "https://drive.google.com/uc?id=file1",
			"shared":         true,
			"owners": []map[string]any{
				{"emailAddress": "alice@example.com", "displayName": "Alice"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// files.copy for file1
	mux.HandleFunc("/files/file1/copy", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":           "file1-copy",
			"name":         "Project Plan (copy).docx",
			"mimeType":     "application/vnd.google-apps.document",
			"modifiedTime": "2026-03-16T10:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// files.export for file1 (Google Workspace export)
	mux.HandleFunc("/files/file1/export", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("exported-pdf-content"))
	})
}

// withPermissionsMock registers all permission-related mock handlers on mux.
func withPermissionsMock(mux *http.ServeMux) {
	// permissions.list and permissions.create for file1
	mux.HandleFunc("/files/file1/permissions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// permissions.create
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":           "perm-new1",
				"role":         body["role"],
				"type":         body["type"],
				"emailAddress": body["emailAddress"],
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		// permissions.list (GET)
		resp := map[string]any{
			"permissions": []map[string]any{
				{
					"id":           "perm1",
					"role":         "owner",
					"type":         "user",
					"emailAddress": "alice@example.com",
					"displayName":  "Alice",
				},
				{
					"id":           "perm2",
					"role":         "reader",
					"type":         "user",
					"emailAddress": "bob@example.com",
					"displayName":  "Bob",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// permissions.get and permissions.delete for perm1
	mux.HandleFunc("/files/file1/permissions/perm1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// GET
		resp := map[string]any{
			"id":           "perm1",
			"role":         "owner",
			"type":         "user",
			"emailAddress": "alice@example.com",
			"displayName":  "Alice",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// newFullMockServer creates an httptest server with all Drive mock handlers.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withFilesMock(mux)
	withPermissionsMock(mux)
	return httptest.NewServer(mux)
}

// newTestServiceFactory returns a ServiceFactory that creates a *drive.Service
// backed by the given httptest server, bypassing OAuth entirely.
func newTestServiceFactory(server *httptest.Server) ServiceFactory {
	return func(ctx context.Context) (*api.Service, error) {
		return api.NewService(ctx,
			option.WithoutAuthentication(),
			option.WithEndpoint(server.URL+"/"),
			option.WithHTTPClient(server.Client()),
		)
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

// buildTestFilesCmd creates a `files` subcommand tree for use in tests.
func buildTestFilesCmd(factory ServiceFactory) *cobra.Command {
	filesCmd := &cobra.Command{Use: "files", Aliases: []string{"file", "f"}}
	filesCmd.AddCommand(newFilesListCmd(factory))
	filesCmd.AddCommand(newFilesGetCmd(factory))
	filesCmd.AddCommand(newFilesDownloadCmd(factory))
	filesCmd.AddCommand(newFilesUploadCmd(factory))
	filesCmd.AddCommand(newFilesCopyCmd(factory))
	filesCmd.AddCommand(newFilesMoveCmd(factory))
	filesCmd.AddCommand(newFilesTrashCmd(factory))
	filesCmd.AddCommand(newFilesUntrashCmd(factory))
	filesCmd.AddCommand(newFilesDeleteCmd(factory))
	return filesCmd
}

// buildTestPermissionsCmd creates a `permissions` subcommand tree for use in tests.
func buildTestPermissionsCmd(factory ServiceFactory) *cobra.Command {
	permCmd := &cobra.Command{Use: "permissions", Aliases: []string{"permission", "perm"}}
	permCmd.AddCommand(newPermissionsListCmd(factory))
	permCmd.AddCommand(newPermissionsGetCmd(factory))
	permCmd.AddCommand(newPermissionsCreateCmd(factory))
	permCmd.AddCommand(newPermissionsDeleteCmd(factory))
	return permCmd
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
