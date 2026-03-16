package gmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// withMessagesMock registers all message-related mock handlers on mux.
func withMessagesMock(mux *http.ServeMux) {
	// messages.list (GET) and messages.insert (POST)
	mux.HandleFunc("/gmail/v1/users/me/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// messages.insert
			resp := map[string]string{
				"id":       "inserted1",
				"threadId": "thread-inserted1",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"messages": []map[string]string{
				{"id": "msg1", "threadId": "thread1"},
				{"id": "msg2", "threadId": "thread2"},
			},
			"resultSizeEstimate": 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// messages.get msg1 (also handles DELETE for messages.delete)
	mux.HandleFunc("/gmail/v1/users/me/messages/msg1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		msg := map[string]any{
			"id":       "msg1",
			"snippet":  "Hello world",
			"threadId": "thread1",
			"payload": map[string]any{
				"headers": []map[string]string{
					{"name": "From", "value": "alice@example.com"},
					{"name": "To", "value": "bob@example.com"},
					{"name": "Subject", "value": "Test Email"},
					{"name": "Date", "value": "Mon, 16 Mar 2026 10:00:00 -0500"},
					{"name": "Message-ID", "value": "<abc123@example.com>"},
				},
				"mimeType": "text/plain",
				"body": map[string]string{
					"data": "SGVsbG8gV29ybGQ=", // base64url of "Hello World"
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	})

	// messages.get msg2
	mux.HandleFunc("/gmail/v1/users/me/messages/msg2", func(w http.ResponseWriter, r *http.Request) {
		msg := map[string]any{
			"id":       "msg2",
			"snippet":  "Second email",
			"threadId": "thread2",
			"payload": map[string]any{
				"headers": []map[string]string{
					{"name": "From", "value": "charlie@example.com"},
					{"name": "To", "value": "bob@example.com"},
					{"name": "Subject", "value": "Another Test"},
					{"name": "Date", "value": "Mon, 16 Mar 2026 11:00:00 -0500"},
				},
				"mimeType": "text/plain",
				"body": map[string]string{
					"data": "U2Vjb25kIGJvZHk=", // base64url of "Second body"
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	})

	// messages.send
	mux.HandleFunc("/gmail/v1/users/me/messages/send", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"id":       "sent1",
			"threadId": "thread-sent1",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// messages.batchModify
	mux.HandleFunc("/gmail/v1/users/me/messages/batchModify", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// messages.batchDelete
	mux.HandleFunc("/gmail/v1/users/me/messages/batchDelete", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// messages.import
	mux.HandleFunc("/gmail/v1/users/me/messages/import", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"id":       "imported1",
			"threadId": "thread-imported1",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// messages.trash, messages.untrash, messages.modify, messages.delete for msg1
	mux.HandleFunc("/gmail/v1/users/me/messages/msg1/trash", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"id": "msg1"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/gmail/v1/users/me/messages/msg1/untrash", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"id": "msg1"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/gmail/v1/users/me/messages/msg1/modify", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":       "msg1",
			"labelIds": []string{"INBOX", "STARRED"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
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
	withMessagesMock(mux)
	withTokenMock(mux)
	return httptest.NewServer(mux)
}

// newTestServiceFactory returns a ServiceFactory that creates a *gmail.Service
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

	buf := make([]byte, 65536)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// newTestRootCmd creates a root command with global flags for testing.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}
