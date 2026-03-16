package gmail

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// ---- messages list ----

func TestMessagesListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	messagesCmd := buildTestMessagesCmd(factory)
	root.AddCommand(messagesCmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list", "--json"})
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

func TestMessagesListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesListEmpty(t *testing.T) {
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
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

func TestMessagesListInvalidSince(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	root.SetArgs([]string{"messages", "list", "--since=notaduration"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --since")
	}
}

func TestMessagesListWithLimit(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list", "--limit=5", "--since=7d"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

func TestMessagesListWithQuery(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list", "--query=from:alice", "--json"})
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

func TestMessagesListWithQueryText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list", "--query=from:alice"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesListWithQueryEmpty(t *testing.T) {
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
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list", "--query=nothing"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

func TestMessagesListWithQueryLimit(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list", "--query=subject:hello", "--limit=5"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

// ---- messages get ----

func TestMessagesGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "get", "--id=msg1", "--json"})
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

func TestMessagesGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "get", "--id=msg1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesGetMsg2(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "get", "--id=msg2", "--json"})
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

// ---- messages send ----

func TestMessagesSendDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "send", "--to=test@example.com", "--subject=Test", "--body=Hello", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if len(output) == 0 {
		t.Error("expected dry-run output")
	}
}

func TestMessagesSendDryRunJSON(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "send", "--to=test@example.com", "--subject=Test", "--body=Hello", "--dry-run", "--json"})
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

func TestMessagesSendRequiresBodyOrFile(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	root.SetArgs([]string{"messages", "send", "--to=test@example.com", "--subject=Test", "--dry-run"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when neither --body nor --body-file provided")
	}
}

func TestMessagesSendBodyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "body-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("Body from file")
	tmpFile.Close()

	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "send", "--to=test@example.com", "--subject=Test", "--body-file=" + tmpFile.Name(), "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if len(output) == 0 {
		t.Error("expected output from body-file dry run")
	}
}

func TestMessagesSendActual(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "send", "--to=test@example.com", "--subject=Hello", "--body=World"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected output after sending")
	}
}

func TestMessagesSendActualJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "send", "--to=test@example.com", "--subject=Hello", "--body=World", "--json"})
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

func TestMessagesSendWithCC(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"messages", "send", "--to=test@example.com", "--subject=Hello", "--body=World", "--cc=cc@example.com", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

func TestMessagesSendWithReplyTo(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"messages", "send", "--to=test@example.com", "--subject=Re: Test", "--body=Reply body", "--reply-to=msg1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

func TestMessagesSendBodyFileMissing(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	root.SetArgs([]string{"messages", "send", "--to=test@example.com", "--subject=Test", "--body-file=/nonexistent/path.txt"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing body file")
	}
}

// ---- messages trash ----

func TestMessagesTrashJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "trash", "--id=msg1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "msg1" {
		t.Errorf("expected id=msg1, got %s", result["id"])
	}
	if result["status"] != "trashed" {
		t.Errorf("expected status=trashed, got %s", result["status"])
	}
}

func TestMessagesTrashText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "trash", "--id=msg1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesTrashDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "trash", "--id=msg1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- messages untrash ----

func TestMessagesUntrashJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "untrash", "--id=msg1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "msg1" {
		t.Errorf("expected id=msg1, got %s", result["id"])
	}
	if result["status"] != "untrashed" {
		t.Errorf("expected status=untrashed, got %s", result["status"])
	}
}

func TestMessagesUntrashText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "untrash", "--id=msg1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesUntrashDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "untrash", "--id=msg1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- messages delete ----

func TestMessagesDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "delete", "--id=msg1", "--confirm", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "msg1" {
		t.Errorf("expected id=msg1, got %s", result["id"])
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
}

func TestMessagesDeleteText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "delete", "--id=msg1", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesDeleteDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "delete", "--id=msg1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

func TestMessagesDeleteRequiresConfirm(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	root.SetArgs([]string{"messages", "delete", "--id=msg1"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --confirm not provided")
	}
}

// ---- messages modify ----

func TestMessagesModifyJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "modify", "--id=msg1", "--add-labels=STARRED", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "msg1" {
		t.Errorf("expected id=msg1, got %v", result["id"])
	}
	if result["status"] != "modified" {
		t.Errorf("expected status=modified, got %v", result["status"])
	}
}

func TestMessagesModifyText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "modify", "--id=msg1", "--add-labels=STARRED"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesModifyDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "modify", "--id=msg1", "--add-labels=STARRED", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- messages import ----

func TestMessagesImportJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "raw-*.eml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("From: test@example.com\r\nSubject: Test\r\n\r\nBody")
	tmpFile.Close()

	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "import", "--raw-file=" + tmpFile.Name(), "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["status"] != "imported" {
		t.Errorf("expected status=imported, got %s", result["status"])
	}
}

func TestMessagesImportText(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "raw-*.eml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("From: test@example.com\r\nSubject: Test\r\n\r\nBody")
	tmpFile.Close()

	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "import", "--raw-file=" + tmpFile.Name()})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesImportDryRun(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "raw-*.eml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("From: test@example.com\r\nSubject: Test\r\n\r\nBody")
	tmpFile.Close()

	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "import", "--raw-file=" + tmpFile.Name(), "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- messages insert ----

func TestMessagesInsertJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "raw-*.eml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("From: test@example.com\r\nSubject: Test\r\n\r\nBody")
	tmpFile.Close()

	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "insert", "--raw-file=" + tmpFile.Name(), "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["status"] != "inserted" {
		t.Errorf("expected status=inserted, got %s", result["status"])
	}
}

func TestMessagesInsertText(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "raw-*.eml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("From: test@example.com\r\nSubject: Test\r\n\r\nBody")
	tmpFile.Close()

	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "insert", "--raw-file=" + tmpFile.Name()})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesInsertDryRun(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "raw-*.eml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("From: test@example.com\r\nSubject: Test\r\n\r\nBody")
	tmpFile.Close()

	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "insert", "--raw-file=" + tmpFile.Name(), "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- messages batch-modify ----

func TestMessagesBatchModifyJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "batch-modify", "--ids=msg1,msg2", "--add-labels=STARRED", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["status"] != "modified" {
		t.Errorf("expected status=modified, got %v", result["status"])
	}
}

func TestMessagesBatchModifyText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "batch-modify", "--ids=msg1,msg2", "--add-labels=STARRED"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesBatchModifyDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "batch-modify", "--ids=msg1,msg2", "--add-labels=STARRED", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- messages batch-delete ----

func TestMessagesBatchDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "batch-delete", "--ids=msg1,msg2", "--confirm", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %v", result["status"])
	}
}

func TestMessagesBatchDeleteText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "batch-delete", "--ids=msg1,msg2", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestMessagesBatchDeleteDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "batch-delete", "--ids=msg1,msg2", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

func TestMessagesBatchDeleteRequiresConfirm(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestMessagesCmd(factory))

	root.SetArgs([]string{"messages", "batch-delete", "--ids=msg1,msg2"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --confirm not provided")
	}
}

// buildTestMessagesCmd creates a `messages` subcommand tree for use in tests.
func buildTestMessagesCmd(factory ServiceFactory) *cobra.Command {
	messagesCmd := &cobra.Command{Use: "messages"}
	messagesCmd.AddCommand(newMessagesListCmd(factory))
	messagesCmd.AddCommand(newMessagesGetCmd(factory))
	messagesCmd.AddCommand(newMessagesSendCmd(factory))
	messagesCmd.AddCommand(newMessagesTrashCmd(factory))
	messagesCmd.AddCommand(newMessagesUntrashCmd(factory))
	messagesCmd.AddCommand(newMessagesDeleteCmd(factory))
	messagesCmd.AddCommand(newMessagesModifyCmd(factory))
	messagesCmd.AddCommand(newMessagesImportCmd(factory))
	messagesCmd.AddCommand(newMessagesInsertCmd(factory))
	messagesCmd.AddCommand(newMessagesBatchModifyCmd(factory))
	messagesCmd.AddCommand(newMessagesBatchDeleteCmd(factory))
	return messagesCmd
}
