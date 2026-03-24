package docs

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
	docsapi "google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// withDocsMock registers Google Docs API mock handlers on mux.
func withDocsMock(mux *http.ServeMux) {
	// documents.create (POST /v1/documents)
	mux.HandleFunc("/v1/documents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		title := "Untitled"
		if t, ok := body["title"].(string); ok && t != "" {
			title = t
		}
		resp := map[string]any{
			"documentId": "new-doc-id",
			"title":      title,
			"body": map[string]any{
				"content": []map[string]any{
					{
						"endIndex": 1,
						"paragraph": map[string]any{
							"elements": []map[string]any{
								{
									"textRun": map[string]any{
										"content": "\n",
									},
								},
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// documents.get (GET /v1/documents/doc1) and documents.batchUpdate (POST /v1/documents/doc1:batchUpdate)
	mux.HandleFunc("/v1/documents/doc1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp := testDocResponse("doc1", "Test Document", "Hello world\n")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// documents.batchUpdate (POST /v1/documents/doc1:batchUpdate)
	mux.HandleFunc("/v1/documents/doc1:batchUpdate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)

		requests, _ := body["requests"].([]any)
		replies := make([]map[string]any, len(requests))
		for i := range requests {
			replies[i] = map[string]any{}
		}

		resp := map[string]any{
			"documentId": "doc1",
			"replies":    replies,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// testDocResponse builds a mock Document API response.
func testDocResponse(id, title, bodyText string) map[string]any {
	return map[string]any{
		"documentId": id,
		"title":      title,
		"body": map[string]any{
			"content": []map[string]any{
				{
					"startIndex": 0,
					"endIndex":   int64(len(bodyText) + 1),
					"paragraph": map[string]any{
						"elements": []map[string]any{
							{
								"startIndex": 0,
								"endIndex":   int64(len(bodyText)),
								"textRun": map[string]any{
									"content": bodyText,
								},
							},
						},
					},
				},
				{
					"startIndex": int64(len(bodyText) + 1),
					"endIndex":   int64(len(bodyText) + 2),
					"paragraph": map[string]any{
						"elements": []map[string]any{
							{
								"textRun": map[string]any{
									"content": "\n",
								},
							},
						},
					},
				},
			},
		},
	}
}

// withTokenMock registers a mock OAuth token endpoint on mux.
func withTokenMock(mux *http.ServeMux) {
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"access_token":  "new-access-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "new-refresh-token",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// newFullMockServer creates an httptest.Server with all mock handlers registered.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withDocsMock(mux)
	withTokenMock(mux)
	return httptest.NewServer(mux)
}

// newTestDocsServiceFactory returns a DocsServiceFactory backed by the test server.
func newTestDocsServiceFactory(server *httptest.Server) DocsServiceFactory {
	return func(ctx context.Context) (*docsapi.Service, error) {
		return docsapi.NewService(ctx,
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

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured stdout: %v", err)
	}
	return string(data)
}

// newTestRootCmd creates a root command with global flags for testing.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}

// buildTestDocumentsCmd creates a `documents` subcommand tree for use in tests.
func buildTestDocumentsCmd(factory DocsServiceFactory) *cobra.Command {
	documentsCmd := &cobra.Command{Use: "documents"}
	documentsCmd.AddCommand(newDocumentsCreateCmd(factory))
	documentsCmd.AddCommand(newDocumentsGetCmd(factory))
	documentsCmd.AddCommand(newDocumentsAppendCmd(factory))
	documentsCmd.AddCommand(newDocumentsBatchUpdateCmd(factory))
	return documentsCmd
}

// runCmd is a test helper that executes a command with the given args and returns output or error.
func runCmd(t *testing.T, root *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var cmdErr error
	out := captureStdout(t, func() {
		root.SetArgs(args)
		cmdErr = root.Execute()
	})
	return strings.TrimSpace(out), cmdErr
}
