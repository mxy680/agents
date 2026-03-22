package x

import (
	"testing"
)

func TestDMInbox_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "inbox", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"conversations"`) {
		t.Errorf("expected JSON conversations field in output, got: %s", out)
	}
	if !containsStr(out, "conv-abc-123") {
		t.Errorf("expected conversation ID in output, got: %s", out)
	}
}

func TestDMInbox_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "inbox"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "conv-abc-123") && !containsStr(out, "ID") {
		t.Errorf("expected conversation in text output, got: %s", out)
	}
}

func TestDMConversation_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "conversation", "--conversation-id=conv-abc-123", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"messages"`) {
		t.Errorf("expected JSON messages field in output, got: %s", out)
	}
	if !containsStr(out, "msg-001") {
		t.Errorf("expected message ID in output, got: %s", out)
	}
}

func TestDMConversation_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "conversation", "--conversation-id=conv-abc-123"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Hello there!") && !containsStr(out, "ID") {
		t.Errorf("expected message content in text output, got: %s", out)
	}
}

func TestDMConversation_WithCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "conversation", "--conversation-id=conv-abc-123", "--cursor=msg-100", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"messages"`) {
		t.Errorf("expected JSON messages field in output, got: %s", out)
	}
}

func TestDMSend_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "send", "--user-id=111", "--text=Hello!", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "sent") {
		t.Errorf("expected 'sent' in output, got: %s", out)
	}
}

func TestDMSend_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "send", "--user-id=111", "--text=Hello!"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "111") {
		t.Errorf("expected user ID in output, got: %s", out)
	}
}

func TestDMSend_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "send", "--user-id=111", "--text=Hello!", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestDMSendGroup_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "send-group", "--conversation-id=conv-abc-123", "--text=Group message", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "sent") {
		t.Errorf("expected 'sent' in output, got: %s", out)
	}
}

func TestDMSendGroup_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "send-group", "--conversation-id=conv-abc-123", "--text=Group message", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestDMDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "delete", "--message-id=msg-001", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestDMDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "delete", "--message-id=msg-001", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestDMDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"dm", "delete", "--message-id=msg-001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm not provided, got nil")
	}
}

func TestDMReact_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "react", "--message-id=msg-001", "--emoji=❤️", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "reacted") {
		t.Errorf("expected 'reacted' in output, got: %s", out)
	}
}

func TestDMReact_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "react", "--message-id=msg-001", "--emoji=❤️", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestDMUnreact_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "unreact", "--message-id=msg-001", "--emoji=❤️", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "unreacted") {
		t.Errorf("expected 'unreacted' in output, got: %s", out)
	}
}

func TestDMUnreact_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "unreact", "--message-id=msg-001", "--emoji=❤️", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestDMAddMembers_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "add-members", "--conversation-id=conv-abc-123", "--user-ids=333,444", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "added") {
		t.Errorf("expected 'added' in output, got: %s", out)
	}
}

func TestDMAddMembers_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "add-members", "--conversation-id=conv-abc-123", "--user-ids=333,444", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestDMRenameGroup_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "rename-group", "--conversation-id=conv-abc-123", "--name=New Group Name", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "renamed") {
		t.Errorf("expected 'renamed' in output, got: %s", out)
	}
}

func TestDMRenameGroup_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "rename-group", "--conversation-id=conv-abc-123", "--name=New Group Name"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "conv-abc-123") && !containsStr(out, "New Group Name") {
		t.Errorf("expected conversation details in output, got: %s", out)
	}
}

func TestDMRenameGroup_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newDMCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"dm", "rename-group", "--conversation-id=conv-abc-123", "--name=New Name", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}
