package canvas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConversationsListText(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"conversations", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Welcome to class") {
		t.Errorf("expected conversation subject in output, got: %s", output)
	}
	if !strings.Contains(output, "Question about homework") {
		t.Errorf("expected second conversation subject in output, got: %s", output)
	}
	if !strings.Contains(output, "read") {
		t.Errorf("expected workflow state in output, got: %s", output)
	}
}

func TestConversationsListJSON(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"conversations", "list", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"subject"`) {
		t.Errorf("expected subject field in JSON, got: %s", output)
	}
	if !strings.Contains(output, "Welcome to class") {
		t.Errorf("expected conversation subject in JSON output, got: %s", output)
	}
	if !strings.Contains(output, `"workflow_state"`) {
		t.Errorf("expected workflow_state field in JSON, got: %s", output)
	}
}

func TestConversationsGetText(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"conversations", "get", "--conversation-id", "1001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Welcome to class") {
		t.Errorf("expected conversation subject in output, got: %s", output)
	}
	if !strings.Contains(output, "1001") {
		t.Errorf("expected conversation ID in output, got: %s", output)
	}
	if !strings.Contains(output, "read") {
		t.Errorf("expected workflow state in output, got: %s", output)
	}
	if !strings.Contains(output, "Prof Smith") {
		t.Errorf("expected participant in output, got: %s", output)
	}
}

func TestConversationsGetMissingID(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"conversations", "get"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --conversation-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--conversation-id") {
		t.Errorf("error should mention --conversation-id, got: %v", execErr)
	}
}

func TestConversationsCreateDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"conversations", "create",
			"--recipients", "42",
			"--subject", "Hello",
			"--body", "Hi there!",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
	if !strings.Contains(output, "Hello") {
		t.Errorf("expected subject in dry-run output, got: %s", output)
	}
}

func TestConversationsCreateSuccess(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"conversations", "create",
			"--recipients", "42",
			"--subject", "Question about homework",
			"--body", "What is due Friday?",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1002") {
		t.Errorf("expected conversation ID in output, got: %s", output)
	}
}

func TestConversationsReplyDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"conversations", "reply",
			"--conversation-id", "1001",
			"--body", "Thanks!",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
	if !strings.Contains(output, "1001") {
		t.Errorf("expected conversation ID in dry-run output, got: %s", output)
	}
}

func TestConversationsReplySuccess(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"conversations", "reply",
			"--conversation-id", "1001",
			"--body", "Thanks for the update!",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1001") {
		t.Errorf("expected conversation ID in reply output, got: %s", output)
	}
}

func TestConversationsDeleteNoConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"conversations", "delete", "--conversation-id", "1001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is absent")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestConversationsDeleteConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"conversations", "delete",
			"--conversation-id", "1001",
			"--confirm",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1001") {
		t.Errorf("expected conversation ID in deletion output, got: %s", output)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}

func TestConversationsMarkAllReadSuccess(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"conversations", "mark-all-read"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "marked as read") {
		t.Errorf("expected confirmation message, got: %s", output)
	}
}

func TestConversationsUnreadCountText(t *testing.T) {
	mux := http.NewServeMux()
	withConversationsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"conversations", "unread-count"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "5") {
		t.Errorf("expected unread count (5) in output, got: %s", output)
	}
	if !strings.Contains(output, "Unread") {
		t.Errorf("expected 'Unread' label in output, got: %s", output)
	}
}

func TestConversationsCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"conversations", "create",
			"--recipients", "42",
			"--subject", "Question about homework",
			"--body", "When is it due?",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "1002") && !strings.Contains(output, "Question") {
		t.Errorf("expected conversation creation output, got: %s", output)
	}
}

func TestConversationsReplyLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"conversations", "reply",
			"--conversation-id", "1001",
			"--body", "Thanks for the update!",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1001") && !strings.Contains(output, "replied") && !strings.Contains(output, "reply") {
		t.Errorf("expected reply output, got: %s", output)
	}
}

func TestConversationsUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"conversations", "update",
			"--conversation-id", "1001",
			"--starred",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1001") {
		t.Errorf("expected conversation ID 1001 in update output, got: %s", output)
	}
}

func TestConversationsMarkAllReadLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"conversations", "mark-all-read"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "marked") && !strings.Contains(output, "read") && !strings.Contains(output, "All") {
		t.Errorf("expected mark-all-read output, got: %s", output)
	}
}

func TestConversationsUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newConversationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"conversations", "update",
			"--conversation-id", "1001",
			"--workflow-state", "read",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1001") {
		t.Errorf("expected conversation ID in JSON output, got: %s", output)
	}
}
