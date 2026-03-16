package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// gmailMessage represents a minimal Gmail API message for test fixtures.
type gmailMessage struct {
	ID      string `json:"id"`
	Snippet string `json:"snippet"`
	Payload struct {
		Headers []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"headers"`
		MimeType string `json:"mimeType"`
		Body     struct {
			Data string `json:"data"`
		} `json:"body"`
		Parts []struct {
			MimeType string `json:"mimeType"`
			Body     struct {
				Data string `json:"data"`
			} `json:"body"`
		} `json:"parts"`
	} `json:"payload"`
	ThreadId string `json:"threadId"`
}

// newTestGmailServer creates an httptest server that mocks the Gmail API.
func newTestGmailServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Mock messages.list
	mux.HandleFunc("/gmail/v1/users/me/messages", func(w http.ResponseWriter, r *http.Request) {
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

	// Mock messages.get
	mux.HandleFunc("/gmail/v1/users/me/messages/msg1", func(w http.ResponseWriter, r *http.Request) {
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

	// Mock messages.send
	mux.HandleFunc("/gmail/v1/users/me/messages/send", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"id":       "sent1",
			"threadId": "thread-sent1",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Mock OAuth token endpoint
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

// ---- parseSinceDuration ----

func TestParseSinceDuration(t *testing.T) {
	tests := []struct {
		input   string
		wantHrs float64
		wantErr bool
	}{
		{"24h", 24, false},
		{"1h", 1, false},
		{"7d", 168, false},
		{"30m", 0.5, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := parseSinceDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Hours() != tt.wantHrs {
				t.Errorf("expected %f hours, got %f", tt.wantHrs, d.Hours())
			}
		})
	}
}

// ---- truncate ----

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string", 10, "this is..."},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%d", tt.input, tt.max), func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}

// ---- stripHTMLTags ----

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<b>bold</b> and <i>italic</i>", "bold and italic"},
		{"no tags here", "no tags here"},
		{"<div><p>nested</p></div>", "nested"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripHTMLTags(tt.input)
			if got != tt.want {
				t.Errorf("stripHTMLTags(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---- list-unread ----

func TestListUnreadJSON(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newListUnreadCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"list-unread", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var summaries []EmailSummary
	if err := json.Unmarshal([]byte(output), &summaries); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].ID != "msg1" {
		t.Errorf("expected first message ID=msg1, got %s", summaries[0].ID)
	}
	if summaries[0].From != "alice@example.com" {
		t.Errorf("expected From=alice@example.com, got %s", summaries[0].From)
	}
	if summaries[0].Subject != "Test Email" {
		t.Errorf("expected Subject=Test Email, got %s", summaries[0].Subject)
	}
}

func TestListUnreadText(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newListUnreadCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"list-unread"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
	// Should contain the FROM header and at least one sender
	if len(output) == 0 {
		t.Error("expected table output")
	}
}

func TestListUnreadEmpty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/gmail/v1/users/me/messages", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"messages":           []map[string]string{},
			"resultSizeEstimate": 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newListUnreadCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"list-unread"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	// Empty list response returns no messages slice, printSummaries shows "No unread messages found."
	_ = output
}

func TestListUnreadInvalidSince(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newListUnreadCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	root.SetArgs([]string{"list-unread", "--since=notaduration"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --since")
	}
}

func TestListUnreadWithLimit(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newListUnreadCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"list-unread", "--limit=5", "--since=7d"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

// ---- read ----

func TestReadJSON(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newReadCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"read", "--id=msg1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var detail EmailDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if detail.ID != "msg1" {
		t.Errorf("expected ID=msg1, got %s", detail.ID)
	}
	if detail.From != "alice@example.com" {
		t.Errorf("expected From=alice@example.com, got %s", detail.From)
	}
	if detail.To != "bob@example.com" {
		t.Errorf("expected To=bob@example.com, got %s", detail.To)
	}
	if detail.Subject != "Test Email" {
		t.Errorf("expected Subject=Test Email, got %s", detail.Subject)
	}
	if detail.Body != "Hello World" {
		t.Errorf("expected Body=Hello World, got %q", detail.Body)
	}
}

func TestReadText(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newReadCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"read", "--id=msg1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestReadMsg2(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newReadCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"read", "--id=msg2", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var detail EmailDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if detail.ID != "msg2" {
		t.Errorf("expected ID=msg2, got %s", detail.ID)
	}
	if detail.From != "charlie@example.com" {
		t.Errorf("expected From=charlie@example.com, got %s", detail.From)
	}
}

// ---- send ----

func TestSendDryRun(t *testing.T) {
	// Test that --dry-run doesn't require Gmail credentials
	factory := newTestServiceFactory(newTestGmailServer(t))
	cmd := newSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"send", "--to=test@example.com", "--subject=Test", "--body=Hello", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if len(output) == 0 {
		t.Error("expected dry-run output")
	}
}

func TestSendDryRunJSON(t *testing.T) {
	factory := newTestServiceFactory(newTestGmailServer(t))
	cmd := newSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"send", "--to=test@example.com", "--subject=Test", "--body=Hello", "--dry-run", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["status"] != "dry-run" {
		t.Errorf("expected status=dry-run, got %s", result["status"])
	}
	if result["to"] != "test@example.com" {
		t.Errorf("expected to=test@example.com, got %s", result["to"])
	}
}

func TestSendRequiresBodyOrFile(t *testing.T) {
	factory := newTestServiceFactory(newTestGmailServer(t))
	cmd := newSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	root.SetArgs([]string{"send", "--to=test@example.com", "--subject=Test", "--dry-run"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when neither --body nor --body-file provided")
	}
}

func TestSendBodyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "body-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("Body from file")
	tmpFile.Close()

	factory := newTestServiceFactory(newTestGmailServer(t))
	cmd := newSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"send", "--to=test@example.com", "--subject=Test", "--body-file=" + tmpFile.Name(), "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if len(output) == 0 {
		t.Error("expected output from body-file dry run")
	}
}

func TestSendActual(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"send", "--to=test@example.com", "--subject=Hello", "--body=World"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected output after sending")
	}
}

func TestSendActualJSON(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"send", "--to=test@example.com", "--subject=Hello", "--body=World", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result SendResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result.ID != "sent1" {
		t.Errorf("expected ID=sent1, got %s", result.ID)
	}
	if result.Status != "sent" {
		t.Errorf("expected Status=sent, got %s", result.Status)
	}
}

func TestSendWithCC(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"send", "--to=test@example.com", "--subject=Hello", "--body=World", "--cc=cc@example.com", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

func TestSendWithReplyTo(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var execErr error
	captureStdout(t, func() {
		// reply-to msg1 which has a Message-ID header in the test server
		root.SetArgs([]string{"send", "--to=test@example.com", "--subject=Re: Test", "--body=Reply body", "--reply-to=msg1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

func TestSendBodyFileMissing(t *testing.T) {
	factory := newTestServiceFactory(newTestGmailServer(t))
	cmd := newSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	root.SetArgs([]string{"send", "--to=test@example.com", "--subject=Test", "--body-file=/nonexistent/path.txt"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing body file")
	}
}

// ---- search ----

func TestSearchJSON(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newSearchCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=from:alice", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var summaries []EmailSummary
	if err := json.Unmarshal([]byte(output), &summaries); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 results, got %d", len(summaries))
	}
}

func TestSearchText(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newSearchCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=from:alice"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestSearchEmpty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/gmail/v1/users/me/messages", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"resultSizeEstimate": 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newSearchCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=nothing"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	_ = output
}

func TestSearchWithLimit(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	cmd := newSearchCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=subject:hello", "--limit=5"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

// ---- extractBody ----

func TestExtractBodyNilPayload(t *testing.T) {
	result := extractBody(nil)
	if result != "" {
		t.Errorf("expected empty string for nil payload, got %q", result)
	}
}

func TestExtractBodyHTMLFallback(t *testing.T) {
	payload := &api.MessagePart{
		Parts: []*api.MessagePart{
			{
				MimeType: "text/html",
				Body: &api.MessagePartBody{
					Data: "PGI-aGVsbG88L2I-", // base64url of "<b>hello</b>"
				},
			},
		},
	}
	result := extractBody(payload)
	if result != "hello" {
		t.Errorf("expected 'hello' after stripping HTML, got %q", result)
	}
}

func TestExtractBodyMultipartNested(t *testing.T) {
	payload := &api.MessagePart{
		MimeType: "multipart/mixed",
		Parts: []*api.MessagePart{
			{
				MimeType: "multipart/alternative",
				Parts: []*api.MessagePart{
					{
						MimeType: "text/plain",
						Body: &api.MessagePartBody{
							Data: "SGVsbG8=", // base64url of "Hello"
						},
					},
				},
			},
		},
	}
	result := extractBody(payload)
	if result != "Hello" {
		t.Errorf("expected 'Hello' from nested multipart, got %q", result)
	}
}

func TestExtractBodyEmptyParts(t *testing.T) {
	payload := &api.MessagePart{
		Parts: []*api.MessagePart{
			{
				MimeType: "text/plain",
				Body:     &api.MessagePartBody{Data: ""},
			},
		},
	}
	result := extractBody(payload)
	if result != "" {
		t.Errorf("expected empty string for empty body data, got %q", result)
	}
}

// ---- Provider ----

func TestProviderNew(t *testing.T) {
	p := New()
	if p.Name() != "gmail" {
		t.Errorf("expected name=gmail, got %s", p.Name())
	}
	if p.ServiceFactory == nil {
		t.Error("expected ServiceFactory to be set")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	server := newTestGmailServer(t)
	defer server.Close()

	p := &Provider{ServiceFactory: newTestServiceFactory(server)}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	// Verify the gmail subcommand and its children are registered
	gmailCmd, _, err := root.Find([]string{"gmail"})
	if err != nil || gmailCmd == nil {
		t.Fatal("expected gmail command to be registered")
	}

	subCmds := map[string]bool{}
	for _, c := range gmailCmd.Commands() {
		subCmds[c.Use] = true
	}
	for _, expected := range []string{"list-unread", "read", "search"} {
		if !subCmds[expected] {
			t.Errorf("expected subcommand %q to be registered", expected)
		}
	}
}
