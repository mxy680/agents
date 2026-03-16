package gmail

import (
	"encoding/json"
	"testing"
)

// ---- threads list ----

func TestThreadsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var summaries []ThreadSummary
	if err := json.Unmarshal([]byte(output), &summaries); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].ID != "thread1" {
		t.Errorf("expected first thread ID=thread1, got %s", summaries[0].ID)
	}
	if summaries[0].Snippet != "First thread snippet" {
		t.Errorf("expected Snippet='First thread snippet', got %s", summaries[0].Snippet)
	}
}

func TestThreadsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "list"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

// ---- threads get ----

func TestThreadsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "get", "--id=thread1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var detail ThreadDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if detail.ID != "thread1" {
		t.Errorf("expected ID=thread1, got %s", detail.ID)
	}
	if len(detail.Messages) != 2 {
		t.Fatalf("expected 2 messages in thread, got %d", len(detail.Messages))
	}
	if detail.Messages[0].From != "alice@example.com" {
		t.Errorf("expected first message From=alice@example.com, got %s", detail.Messages[0].From)
	}
}

func TestThreadsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "get", "--id=thread1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

// ---- threads trash ----

func TestThreadsTrashJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "trash", "--id=thread1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "thread1" {
		t.Errorf("expected id=thread1, got %s", result["id"])
	}
	if result["status"] != "trashed" {
		t.Errorf("expected status=trashed, got %s", result["status"])
	}
}

func TestThreadsTrashText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "trash", "--id=thread1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestThreadsTrashDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "trash", "--id=thread1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- threads untrash ----

func TestThreadsUntrashJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "untrash", "--id=thread1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "thread1" {
		t.Errorf("expected id=thread1, got %s", result["id"])
	}
	if result["status"] != "untrashed" {
		t.Errorf("expected status=untrashed, got %s", result["status"])
	}
}

func TestThreadsUntrashText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "untrash", "--id=thread1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestThreadsUntrashDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "untrash", "--id=thread1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- threads delete ----

func TestThreadsDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "delete", "--id=thread1", "--confirm", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "thread1" {
		t.Errorf("expected id=thread1, got %s", result["id"])
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
}

func TestThreadsDeleteText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "delete", "--id=thread1", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestThreadsDeleteDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "delete", "--id=thread1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

func TestThreadsDeleteRequiresConfirm(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	root.SetArgs([]string{"threads", "delete", "--id=thread1"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --confirm not provided")
	}
}

// ---- threads modify ----

func TestThreadsModifyJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "modify", "--id=thread1", "--add-labels=STARRED", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "thread1" {
		t.Errorf("expected id=thread1, got %s", result["id"])
	}
	if result["status"] != "modified" {
		t.Errorf("expected status=modified, got %s", result["status"])
	}
}

func TestThreadsModifyText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "modify", "--id=thread1", "--add-labels=STARRED"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestThreadsModifyDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestThreadsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "modify", "--id=thread1", "--add-labels=STARRED", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}
