package sheets

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
	driveapi "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	sheetsapi "google.golang.org/api/sheets/v4"
)

// withSpreadsheetsMock registers spreadsheet-related mock handlers on mux.
func withSpreadsheetsMock(mux *http.ServeMux) {
	// spreadsheets.create (POST) — matches root path
	mux.HandleFunc("/v4/spreadsheets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			title := "Untitled"
			if props, ok := body["properties"].(map[string]any); ok {
				if t, ok := props["title"].(string); ok {
					title = t
				}
			}
			resp := map[string]any{
				"spreadsheetId":  "new-ss-id",
				"spreadsheetUrl": "https://docs.google.com/spreadsheets/d/new-ss-id",
				"properties": map[string]any{
					"title":  title,
					"locale": "en_US",
				},
				"sheets": []map[string]any{
					{
						"properties": map[string]any{
							"sheetId": 0,
							"title":   "Sheet1",
							"index":   0,
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// spreadsheets.get for ss1
	mux.HandleFunc("/v4/spreadsheets/ss1", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"spreadsheetId":  "ss1",
			"spreadsheetUrl": "https://docs.google.com/spreadsheets/d/ss1",
			"properties": map[string]any{
				"title":  "Test Spreadsheet",
				"locale": "en_US",
			},
			"sheets": []map[string]any{
				{
					"properties": map[string]any{
						"sheetId": 0,
						"title":   "Sheet1",
						"index":   0,
					},
				},
				{
					"properties": map[string]any{
						"sheetId": 1,
						"title":   "Sheet2",
						"index":   1,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withValuesMock registers values-related mock handlers on mux.
func withValuesMock(mux *http.ServeMux) {
	// values.get (GET) and values.update (PUT)
	mux.HandleFunc("/v4/spreadsheets/ss1/values/Sheet1!A1:B2", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := map[string]any{
				"range":          "Sheet1!A1:B2",
				"majorDimension": "ROWS",
				"values": [][]any{
					{"Name", "Age"},
					{"Alice", 30},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		case http.MethodPut:
			resp := map[string]any{
				"spreadsheetId":  "ss1",
				"updatedRange":   "Sheet1!A1:B2",
				"updatedRows":    2,
				"updatedColumns": 2,
				"updatedCells":   4,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// values.append — the Sheets API uses POST with :append suffix
	mux.HandleFunc("/v4/spreadsheets/ss1/values/Sheet1!A1:B2:append", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"spreadsheetId": "ss1",
			"tableRange":    "Sheet1!A1:B2",
			"updates": map[string]any{
				"spreadsheetId":  "ss1",
				"updatedRange":   "Sheet1!A3:B4",
				"updatedRows":    2,
				"updatedColumns": 2,
				"updatedCells":   4,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// values.clear — POST with :clear suffix
	mux.HandleFunc("/v4/spreadsheets/ss1/values/Sheet1!A1:B2:clear", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"spreadsheetId": "ss1",
			"clearedRange":  "Sheet1!A1:B2",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// values.batchGet
	mux.HandleFunc("/v4/spreadsheets/ss1/values:batchGet", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"spreadsheetId": "ss1",
			"valueRanges": []map[string]any{
				{
					"range":          "Sheet1!A1:B2",
					"majorDimension": "ROWS",
					"values": [][]any{
						{"Name", "Age"},
						{"Alice", 30},
					},
				},
				{
					"range":          "Sheet2!A1:B2",
					"majorDimension": "ROWS",
					"values": [][]any{
						{"City", "Pop"},
						{"NYC", 8000000},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// values.batchUpdate
	mux.HandleFunc("/v4/spreadsheets/ss1/values:batchUpdate", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"spreadsheetId":       "ss1",
			"totalUpdatedRows":    4,
			"totalUpdatedColumns": 2,
			"totalUpdatedCells":   8,
			"totalUpdatedSheets":  2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withTabsMock registers tab-related mock handlers on mux.
func withTabsMock(mux *http.ServeMux) {
	// spreadsheets.batchUpdate for tab operations (POST)
	mux.HandleFunc("/v4/spreadsheets/ss1:batchUpdate", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)

		requests, _ := body["requests"].([]any)
		replies := make([]map[string]any, 0)

		for _, req := range requests {
			reqMap, _ := req.(map[string]any)
			if _, ok := reqMap["addSheet"]; ok {
				replies = append(replies, map[string]any{
					"addSheet": map[string]any{
						"properties": map[string]any{
							"sheetId": 42,
							"title":   "NewTab",
							"index":   2,
						},
					},
				})
			} else {
				replies = append(replies, map[string]any{})
			}
		}

		resp := map[string]any{
			"spreadsheetId": "ss1",
			"replies":       replies,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withDriveFilesMock registers Drive API mock handlers for spreadsheet listing/deletion.
func withDriveFilesMock(mux *http.ServeMux) {
	// The Google Drive API client sends requests to paths under the root.
	// When using option.WithEndpoint(server.URL+"/"), the client sends to /<path>.
	// The actual Drive API paths may be /drive/v3/files or just /files depending on version.
	// We register both common patterns.

	driveFilesHandler := func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a specific file request (has path beyond /files or /drive/v3/files)
		path := r.URL.Path

		// Handle files.delete for ss1
		if strings.HasSuffix(path, "/ss1") {
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// files.list
		resp := map[string]any{
			"files": []map[string]any{
				{
					"id":          "ss1",
					"name":        "Test Spreadsheet",
					"webViewLink": "https://docs.google.com/spreadsheets/d/ss1",
				},
				{
					"id":          "ss2",
					"name":        "Another Sheet",
					"webViewLink": "https://docs.google.com/spreadsheets/d/ss2",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	// When using option.WithEndpoint(server.URL+"/"), the client strips the
	// API prefix and sends to /files directly.
	mux.HandleFunc("/files", driveFilesHandler)
	mux.HandleFunc("/files/", driveFilesHandler)
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
	withSpreadsheetsMock(mux)
	withValuesMock(mux)
	withTabsMock(mux)
	withDriveFilesMock(mux)
	withTokenMock(mux)
	return httptest.NewServer(mux)
}

// newTestSheetsServiceFactory returns a SheetsServiceFactory backed by the test server.
func newTestSheetsServiceFactory(server *httptest.Server) SheetsServiceFactory {
	return func(ctx context.Context) (*sheetsapi.Service, error) {
		return sheetsapi.NewService(ctx,
			option.WithoutAuthentication(),
			option.WithEndpoint(server.URL+"/"),
			option.WithHTTPClient(server.Client()),
		)
	}
}

// newTestDriveServiceFactory returns a DriveServiceFactory backed by the test server.
func newTestDriveServiceFactory(server *httptest.Server) DriveServiceFactory {
	return func(ctx context.Context) (*driveapi.Service, error) {
		return driveapi.NewService(ctx,
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

// buildTestValuesCmd creates a `values` subcommand tree for use in tests.
func buildTestValuesCmd(factory SheetsServiceFactory) *cobra.Command {
	valuesCmd := &cobra.Command{Use: "values"}
	valuesCmd.AddCommand(newValuesGetCmd(factory))
	valuesCmd.AddCommand(newValuesUpdateCmd(factory))
	valuesCmd.AddCommand(newValuesAppendCmd(factory))
	valuesCmd.AddCommand(newValuesClearCmd(factory))
	valuesCmd.AddCommand(newValuesBatchGetCmd(factory))
	valuesCmd.AddCommand(newValuesBatchUpdateCmd(factory))
	return valuesCmd
}

// buildTestSpreadsheetsCmd creates a `spreadsheets` subcommand tree for use in tests.
func buildTestSpreadsheetsCmd(sheetsFactory SheetsServiceFactory, driveFactory DriveServiceFactory) *cobra.Command {
	spreadsheetsCmd := &cobra.Command{Use: "spreadsheets"}
	spreadsheetsCmd.AddCommand(newSpreadsheetsListCmd(driveFactory))
	spreadsheetsCmd.AddCommand(newSpreadsheetsGetCmd(sheetsFactory))
	spreadsheetsCmd.AddCommand(newSpreadsheetsCreateCmd(sheetsFactory))
	spreadsheetsCmd.AddCommand(newSpreadsheetsDeleteCmd(driveFactory))
	return spreadsheetsCmd
}

// buildTestTabsCmd creates a `tabs` subcommand tree for use in tests.
func buildTestTabsCmd(factory SheetsServiceFactory) *cobra.Command {
	tabsCmd := &cobra.Command{Use: "tabs"}
	tabsCmd.AddCommand(newTabsListCmd(factory))
	tabsCmd.AddCommand(newTabsCreateCmd(factory))
	tabsCmd.AddCommand(newTabsDeleteCmd(factory))
	tabsCmd.AddCommand(newTabsRenameCmd(factory))
	return tabsCmd
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
