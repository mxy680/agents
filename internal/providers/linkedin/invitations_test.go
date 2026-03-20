package linkedin

import (
	"testing"
)

func TestInvitationsList_Received_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Carol") {
		t.Errorf("expected 'Carol' in output, got: %s", out)
	}
	if !containsStr(out, "123456") {
		t.Errorf("expected invitation ID '123456' in output, got: %s", out)
	}
}

func TestInvitationsList_Received_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"id"`) {
		t.Errorf("expected JSON field 'id' in output, got: %s", out)
	}
	if !containsStr(out, "123456") {
		t.Errorf("expected invitation ID in JSON output, got: %s", out)
	}
	if !containsStr(out, "received") {
		t.Errorf("expected 'received' direction in JSON output, got: %s", out)
	}
}

func TestInvitationsList_Sent_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "list", "--direction", "sent"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "789012") {
		t.Errorf("expected invitation ID '789012' in output, got: %s", out)
	}
}

func TestInvitationsList_WithAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invite", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Carol") {
		t.Errorf("expected 'Carol' in output via alias, got: %s", out)
	}
}

func TestInvitationsSend_MissingURN(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"invitations", "send"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --urn is missing")
	}
}

func TestInvitationsSend_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "send", "--urn", "alice-smith", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestInvitationsSend_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "send", "--urn", "alice-smith", "--message", "Hi there"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "sent") {
		t.Errorf("expected 'sent' in output, got: %s", out)
	}
}

func TestInvitationsAccept_MissingID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"invitations", "accept"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --invitation-id is missing")
	}
}

func TestInvitationsAccept_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "accept", "--invitation-id", "123456", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestInvitationsAccept_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "accept", "--invitation-id", "123456"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "accepted") {
		t.Errorf("expected 'accepted' in output, got: %s", out)
	}
}

func TestInvitationsReject_MissingID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"invitations", "reject"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --invitation-id is missing")
	}
}

func TestInvitationsReject_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "reject", "--invitation-id", "123456", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestInvitationsReject_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "reject", "--invitation-id", "123456"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "rejected") {
		t.Errorf("expected 'rejected' in output, got: %s", out)
	}
}

func TestInvitationsWithdraw_MissingID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"invitations", "withdraw"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --invitation-id is missing")
	}
}

func TestInvitationsWithdraw_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "withdraw", "--invitation-id", "789012", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestInvitationsWithdraw_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "withdraw", "--invitation-id", "789012"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "withdrawn") {
		t.Errorf("expected 'withdrawn' in output, got: %s", out)
	}
}

func TestInvitationsAccept_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newInvitationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"invitations", "accept", "--invitation-id", "123456", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON 'status' field in output, got: %s", out)
	}
	if !containsStr(out, "accepted") {
		t.Errorf("expected 'accepted' in JSON output, got: %s", out)
	}
}
