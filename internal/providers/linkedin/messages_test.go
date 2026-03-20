package linkedin

import (
	"context"
	"testing"
)

func TestMessagesConversations_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesConversationsCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"conversations"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "conv1") {
		t.Errorf("expected conversation ID in output, got: %s", out)
	}
	if !containsStr(out, "Jane Doe") {
		t.Errorf("expected participant name in output, got: %s", out)
	}
}

func TestMessagesConversations_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesConversationsCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "conversations"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"id"`) {
		t.Errorf("expected JSON output with id field, got: %s", out)
	}
	if !containsStr(out, "conv1") {
		t.Errorf("expected conversation ID in JSON output, got: %s", out)
	}
}

func TestMessagesList_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesListCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"list", "--conversation-id=conv1"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Hello!") {
		t.Errorf("expected message body in output, got: %s", out)
	}
	if !containsStr(out, "Jane Doe") {
		t.Errorf("expected sender name in output, got: %s", out)
	}
}

func TestMessagesList_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesListCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "list", "--conversation-id=conv1"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"text"`) {
		t.Errorf("expected JSON output with text field, got: %s", out)
	}
	if !containsStr(out, "Hello!") {
		t.Errorf("expected message body in JSON output, got: %s", out)
	}
}

func TestMessagesList_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesListCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	root.SetArgs([]string{"list"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for missing --conversation-id, got nil")
	}
}

func TestMessagesSend_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"send", "--conversation-id=conv1", "--text=Hello there", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestMessagesSend_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesSendCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"send", "--conversation-id=conv1", "--text=Hello there"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "conv1") {
		t.Errorf("expected conversation ID in output, got: %s", out)
	}
}

func TestMessagesNew_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesNewCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"new", "--recipients=urn:li:fs_miniProfile:ABC", "--text=Hi", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestMessagesNew_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesNewCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"new", "--recipients=urn:li:fs_miniProfile:ABC", "--text=Hi"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "conversation") {
		t.Errorf("expected success message in output, got: %s", out)
	}
}

func TestMessagesDelete_NoConfirm(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesDeleteCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	root.SetArgs([]string{"delete", "--conversation-id=conv1"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error without --confirm flag, got nil")
	}
}

func TestMessagesDelete_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesDeleteCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"delete", "--conversation-id=conv1", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestMessagesDelete_Confirm(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesDeleteCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"delete", "--conversation-id=conv1", "--confirm"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "conv1") {
		t.Errorf("expected conversation ID in output, got: %s", out)
	}
}

func TestMessagesMarkRead_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesMarkReadCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"mark-read", "--conversation-id=conv1", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestMessagesMarkRead_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newMessagesMarkReadCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"mark-read", "--conversation-id=conv1"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "conv1") {
		t.Errorf("expected conversation ID in output, got: %s", out)
	}
}

func TestToConversationSummary(t *testing.T) {
	el := voyagerConversationElement{
		EntityURN:      "urn:li:fs_conversation:conv1",
		ConversationID: "conv1",
		LastActivityAt: 1704067200000,
		UnreadCount:    3,
		Participants: []voyagerMessagingMemberWrap{
			{
				MessagingMember: voyagerMessagingMember{
					MiniProfile: voyagerMiniProfile{
						EntityURN: "urn:li:fs_miniProfile:A",
						FirstName: "Alice",
						LastName:  "Smith",
					},
				},
			},
		},
	}

	s := toConversationSummary(el)
	if s.ID != "conv1" {
		t.Errorf("ID = %q, want %q", s.ID, "conv1")
	}
	if s.Title != "Alice Smith" {
		t.Errorf("Title = %q, want %q", s.Title, "Alice Smith")
	}
	if s.UnreadCount != 3 {
		t.Errorf("UnreadCount = %d, want 3", s.UnreadCount)
	}
	if len(s.ParticipantURNs) != 1 {
		t.Errorf("ParticipantURNs len = %d, want 1", len(s.ParticipantURNs))
	}
}

func TestToMessageSummary(t *testing.T) {
	el := voyagerMessageElement{
		EntityURN: "urn:li:fs_event:msg1",
		From: voyagerMessagingMemberWrap{
			MessagingMember: voyagerMessagingMember{
				MiniProfile: voyagerMiniProfile{
					EntityURN: "urn:li:fs_miniProfile:A",
					FirstName: "Jane",
					LastName:  "Doe",
				},
			},
		},
		EventContent: voyagerMessageEventWrap{
			MessageEvent: voyagerMessageEvent{Body: "Hello!"},
		},
		CreatedAt: 1704067200000,
	}

	m := toMessageSummary(el)
	if m.ID != "urn:li:fs_event:msg1" {
		t.Errorf("ID = %q, want urn:li:fs_event:msg1", m.ID)
	}
	if m.SenderName != "Jane Doe" {
		t.Errorf("SenderName = %q, want %q", m.SenderName, "Jane Doe")
	}
	if m.Text != "Hello!" {
		t.Errorf("Text = %q, want Hello!", m.Text)
	}
}
