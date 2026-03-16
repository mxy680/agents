package gmail

import (
	"encoding/json"
	"testing"
)

// ---- drafts list ----

func TestDraftsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var drafts []DraftSummary
	if err := json.Unmarshal([]byte(output), &drafts); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if len(drafts) != 2 {
		t.Fatalf("expected 2 drafts, got %d", len(drafts))
	}
	if drafts[0].ID != "draft1" {
		t.Errorf("expected first draft ID=draft1, got %s", drafts[0].ID)
	}
	if drafts[0].Subject != "Draft Subject" {
		t.Errorf("expected first draft Subject=Draft Subject, got %s", drafts[0].Subject)
	}
	if drafts[0].To != "recipient@example.com" {
		t.Errorf("expected first draft To=recipient@example.com, got %s", drafts[0].To)
	}
}

func TestDraftsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "list"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

// ---- drafts get ----

func TestDraftsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "get", "--id=draft1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var detail DraftDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if detail.ID != "draft1" {
		t.Errorf("expected ID=draft1, got %s", detail.ID)
	}
	if detail.MessageID != "draftmsg1" {
		t.Errorf("expected MessageID=draftmsg1, got %s", detail.MessageID)
	}
	if detail.Subject != "Draft Subject" {
		t.Errorf("expected Subject=Draft Subject, got %s", detail.Subject)
	}
}

func TestDraftsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "get", "--id=draft1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

// ---- drafts create ----

func TestDraftsCreateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "create", "--to=recipient@example.com", "--subject=Test Draft", "--body=Hello", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "draft1" {
		t.Errorf("expected id=draft1, got %s", result["id"])
	}
	if result["status"] != "created" {
		t.Errorf("expected status=created, got %s", result["status"])
	}
}

func TestDraftsCreateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "create", "--to=recipient@example.com", "--subject=Test Draft", "--body=Hello"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestDraftsCreateDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "create", "--to=recipient@example.com", "--subject=Test Draft", "--body=Hello", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

func TestDraftsCreateRequiresBody(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	root.SetArgs([]string{"drafts", "create", "--to=recipient@example.com", "--subject=Test Draft"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --body or --body-file not provided")
	}
}

// ---- drafts update ----

func TestDraftsUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "update", "--id=draft1", "--to=recipient@example.com", "--subject=Updated Subject", "--body=Updated body", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "draft1" {
		t.Errorf("expected id=draft1, got %s", result["id"])
	}
	if result["status"] != "updated" {
		t.Errorf("expected status=updated, got %s", result["status"])
	}
}

func TestDraftsUpdateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "update", "--id=draft1", "--to=recipient@example.com", "--subject=Updated Subject", "--body=Updated body"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestDraftsUpdateDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "update", "--id=draft1", "--to=recipient@example.com", "--subject=Updated Subject", "--body=Updated body", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

func TestDraftsUpdateRequiresBody(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	root.SetArgs([]string{"drafts", "update", "--id=draft1", "--to=recipient@example.com", "--subject=Updated Subject"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --body or --body-file not provided")
	}
}

// ---- drafts send ----

func TestDraftsSendJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "send", "--id=draft1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "sentmsg1" {
		t.Errorf("expected id=sentmsg1, got %s", result["id"])
	}
	if result["status"] != "sent" {
		t.Errorf("expected status=sent, got %s", result["status"])
	}
}

func TestDraftsSendText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "send", "--id=draft1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestDraftsSendDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "send", "--id=draft1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- drafts delete ----

func TestDraftsDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "delete", "--id=draft1", "--confirm", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "draft1" {
		t.Errorf("expected id=draft1, got %s", result["id"])
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
}

func TestDraftsDeleteText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "delete", "--id=draft1", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestDraftsDeleteDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"drafts", "delete", "--id=draft1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

func TestDraftsDeleteRequiresConfirm(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestDraftsCmd(factory))

	root.SetArgs([]string{"drafts", "delete", "--id=draft1"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --confirm not provided")
	}
}
