package imessage

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// --- truncate ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string unchanged",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length unchanged",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "long string truncated with ellipsis",
			input:  "hello world",
			maxLen: 8,
			want:   "hello w…",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 5,
			want:   "",
		},
		{
			name:   "single rune maxLen",
			input:  "ab",
			maxLen: 1,
			want:   "…",
		},
		{
			name:   "multibyte runes",
			input:  "日本語テスト",
			maxLen: 4,
			want:   "日本語…",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := truncate(tc.input, tc.maxLen)
			if got != tc.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
			}
		})
	}
}

// --- formatTimestamp ---

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name string
		ts   int64
		want string // partial match check
	}{
		{
			name: "zero returns empty",
			ts:   0,
			want: "",
		},
		{
			name: "negative returns empty",
			ts:   -1,
			want: "",
		},
		{
			name: "valid timestamp returns formatted date",
			ts:   1700000000000, // 2023-11-14 22:13:20 UTC
			want: "2023-11-",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatTimestamp(tc.ts)
			if tc.want == "" {
				if got != "" {
					t.Errorf("formatTimestamp(%d) = %q, want empty", tc.ts, got)
				}
			} else {
				if !strings.Contains(got, tc.want) {
					t.Errorf("formatTimestamp(%d) = %q, want it to contain %q", tc.ts, got, tc.want)
				}
			}
		})
	}
}

// --- getString / getBool / getInt64 ---

func TestGetString(t *testing.T) {
	t.Run("nil map returns empty", func(t *testing.T) {
		got := getString(nil, "key")
		if got != "" {
			t.Errorf("getString(nil, key) = %q, want empty", got)
		}
	})

	t.Run("missing key returns empty", func(t *testing.T) {
		m := map[string]any{"other": "value"}
		got := getString(m, "missing")
		if got != "" {
			t.Errorf("getString with missing key = %q, want empty", got)
		}
	})

	t.Run("existing key returns value", func(t *testing.T) {
		m := map[string]any{"name": "Alice"}
		got := getString(m, "name")
		if got != "Alice" {
			t.Errorf("getString = %q, want %q", got, "Alice")
		}
	})

	t.Run("non-string value returns empty", func(t *testing.T) {
		m := map[string]any{"count": 42}
		got := getString(m, "count")
		if got != "" {
			t.Errorf("getString with non-string value = %q, want empty", got)
		}
	})
}

func TestGetBool(t *testing.T) {
	t.Run("nil map returns false", func(t *testing.T) {
		got := getBool(nil, "key")
		if got {
			t.Error("getBool(nil, key) = true, want false")
		}
	})

	t.Run("missing key returns false", func(t *testing.T) {
		m := map[string]any{"other": true}
		got := getBool(m, "missing")
		if got {
			t.Error("getBool with missing key = true, want false")
		}
	})

	t.Run("true value", func(t *testing.T) {
		m := map[string]any{"active": true}
		got := getBool(m, "active")
		if !got {
			t.Error("getBool with true value = false, want true")
		}
	})

	t.Run("false value", func(t *testing.T) {
		m := map[string]any{"active": false}
		got := getBool(m, "active")
		if got {
			t.Error("getBool with false value = true, want false")
		}
	})

	t.Run("non-bool value returns false", func(t *testing.T) {
		m := map[string]any{"active": "yes"}
		got := getBool(m, "active")
		if got {
			t.Error("getBool with non-bool value = true, want false")
		}
	})
}

func TestGetInt64(t *testing.T) {
	t.Run("nil map returns zero", func(t *testing.T) {
		got := getInt64(nil, "key")
		if got != 0 {
			t.Errorf("getInt64(nil, key) = %d, want 0", got)
		}
	})

	t.Run("missing key returns zero", func(t *testing.T) {
		m := map[string]any{"other": float64(5)}
		got := getInt64(m, "missing")
		if got != 0 {
			t.Errorf("getInt64 missing key = %d, want 0", got)
		}
	})

	t.Run("valid float64 value", func(t *testing.T) {
		m := map[string]any{"total": float64(42)}
		got := getInt64(m, "total")
		if got != 42 {
			t.Errorf("getInt64 = %d, want 42", got)
		}
	})

	t.Run("non-float64 value returns zero", func(t *testing.T) {
		m := map[string]any{"total": "42"}
		got := getInt64(m, "total")
		if got != 0 {
			t.Errorf("getInt64 with string value = %d, want 0", got)
		}
	})
}

// --- toChatSummary ---

func TestToChatSummary(t *testing.T) {
	t.Run("basic chat", func(t *testing.T) {
		raw := json.RawMessage(`{
			"guid": "iMessage;-;chat123",
			"displayName": "My Group",
			"chatIdentifier": "iMessage;-;chat123",
			"isArchived": false
		}`)
		s := toChatSummary(raw)
		if s.GUID != "iMessage;-;chat123" {
			t.Errorf("GUID = %q, want iMessage;-;chat123", s.GUID)
		}
		if s.DisplayName != "My Group" {
			t.Errorf("DisplayName = %q, want My Group", s.DisplayName)
		}
		if s.IsArchived {
			t.Error("IsArchived = true, want false")
		}
	})

	t.Run("with participants", func(t *testing.T) {
		raw := json.RawMessage(`{
			"guid": "iMessage;-;chat123",
			"chatIdentifier": "iMessage;-;chat123",
			"participants": [
				{"handle": {"address": "+1234567890"}},
				{"handle": {"address": "+0987654321"}}
			]
		}`)
		s := toChatSummary(raw)
		if len(s.Participants) != 2 {
			t.Fatalf("Participants len = %d, want 2", len(s.Participants))
		}
		if s.Participants[0] != "+1234567890" {
			t.Errorf("Participants[0] = %q, want +1234567890", s.Participants[0])
		}
	})

	t.Run("with lastMessage", func(t *testing.T) {
		raw := json.RawMessage(`{
			"guid": "iMessage;-;chat123",
			"chatIdentifier": "iMessage;-;chat123",
			"lastMessage": {"text": "Hello world"}
		}`)
		s := toChatSummary(raw)
		if s.LastMessage != "Hello world" {
			t.Errorf("LastMessage = %q, want Hello world", s.LastMessage)
		}
	})

	t.Run("lastMessage truncated at 80", func(t *testing.T) {
		longText := strings.Repeat("a", 100)
		raw := json.RawMessage(`{"guid": "g", "chatIdentifier": "g", "lastMessage": {"text": "` + longText + `"}}`)
		s := toChatSummary(raw)
		runes := []rune(s.LastMessage)
		if len(runes) > 80 {
			t.Errorf("LastMessage rune len = %d, want <= 80", len(runes))
		}
	})
}

// --- toMessageSummary ---

func TestToMessageSummary(t *testing.T) {
	t.Run("basic message", func(t *testing.T) {
		raw := json.RawMessage(`{
			"guid": "msg-001",
			"text": "Hello",
			"isFromMe": true,
			"dateCreated": 1700000000000,
			"subject": "Re:"
		}`)
		s := toMessageSummary(raw)
		if s.GUID != "msg-001" {
			t.Errorf("GUID = %q, want msg-001", s.GUID)
		}
		if !s.IsFromMe {
			t.Error("IsFromMe = false, want true")
		}
		if s.DateCreated != 1700000000000 {
			t.Errorf("DateCreated = %d, want 1700000000000", s.DateCreated)
		}
		if s.Subject != "Re:" {
			t.Errorf("Subject = %q, want Re:", s.Subject)
		}
	})

	t.Run("with handle", func(t *testing.T) {
		raw := json.RawMessage(`{
			"guid": "msg-001",
			"isFromMe": false,
			"handle": {"address": "+1234567890"}
		}`)
		s := toMessageSummary(raw)
		if s.Handle != "+1234567890" {
			t.Errorf("Handle = %q, want +1234567890", s.Handle)
		}
	})

	t.Run("with chats extracts first guid", func(t *testing.T) {
		raw := json.RawMessage(`{
			"guid": "msg-001",
			"isFromMe": false,
			"chats": [{"guid": "iMessage;-;chat123"}, {"guid": "iMessage;-;chat456"}]
		}`)
		s := toMessageSummary(raw)
		if s.ChatGUID != "iMessage;-;chat123" {
			t.Errorf("ChatGUID = %q, want iMessage;-;chat123", s.ChatGUID)
		}
	})

	t.Run("with attachments sets HasAttachment", func(t *testing.T) {
		raw := json.RawMessage(`{
			"guid": "msg-001",
			"isFromMe": false,
			"attachments": [{"guid": "att-001"}]
		}`)
		s := toMessageSummary(raw)
		if !s.HasAttachment {
			t.Error("HasAttachment = false, want true")
		}
	})

	t.Run("empty attachments", func(t *testing.T) {
		raw := json.RawMessage(`{
			"guid": "msg-001",
			"isFromMe": false,
			"attachments": []
		}`)
		s := toMessageSummary(raw)
		if s.HasAttachment {
			t.Error("HasAttachment = true, want false for empty attachments")
		}
	})
}

// --- toAttachmentSummary ---

func TestToAttachmentSummary(t *testing.T) {
	raw := json.RawMessage(`{
		"guid": "att-guid-001",
		"transferName": "photo.jpg",
		"mimeType": "image/jpeg",
		"totalBytes": 204800,
		"isOutgoing": false,
		"createdDate": 1700000000000
	}`)

	s := toAttachmentSummary(raw)
	if s.GUID != "att-guid-001" {
		t.Errorf("GUID = %q, want att-guid-001", s.GUID)
	}
	if s.FileName != "photo.jpg" {
		t.Errorf("FileName = %q, want photo.jpg", s.FileName)
	}
	if s.MIMEType != "image/jpeg" {
		t.Errorf("MIMEType = %q, want image/jpeg", s.MIMEType)
	}
	if s.TotalBytes != 204800 {
		t.Errorf("TotalBytes = %d, want 204800", s.TotalBytes)
	}
	if s.IsOutgoing {
		t.Error("IsOutgoing = true, want false")
	}
	if s.DateCreated != 1700000000000 {
		t.Errorf("DateCreated = %d, want 1700000000000", s.DateCreated)
	}
}

// --- toHandleSummary ---

func TestToHandleSummary(t *testing.T) {
	raw := json.RawMessage(`{
		"address": "+1234567890",
		"service": "iMessage",
		"country": "US",
		"uncanonicalizedId": "+1 (234) 567-890"
	}`)

	s := toHandleSummary(raw)
	if s.Address != "+1234567890" {
		t.Errorf("Address = %q, want +1234567890", s.Address)
	}
	if s.Service != "iMessage" {
		t.Errorf("Service = %q, want iMessage", s.Service)
	}
	if s.Country != "US" {
		t.Errorf("Country = %q, want US", s.Country)
	}
	if s.UncanonID != "+1 (234) 567-890" {
		t.Errorf("UncanonID = %q, want +1 (234) 567-890", s.UncanonID)
	}
}

// --- toContactSummary ---

func TestToContactSummary(t *testing.T) {
	t.Run("basic contact", func(t *testing.T) {
		raw := json.RawMessage(`{
			"id": "contact-001",
			"firstName": "John",
			"lastName": "Smith",
			"displayName": "John Smith"
		}`)
		s := toContactSummary(raw)
		if s.ID != "contact-001" {
			t.Errorf("ID = %q, want contact-001", s.ID)
		}
		if s.FirstName != "John" {
			t.Errorf("FirstName = %q, want John", s.FirstName)
		}
		if s.LastName != "Smith" {
			t.Errorf("LastName = %q, want Smith", s.LastName)
		}
		if s.DisplayName != "John Smith" {
			t.Errorf("DisplayName = %q, want John Smith", s.DisplayName)
		}
	})

	t.Run("with phones and emails", func(t *testing.T) {
		raw := json.RawMessage(`{
			"id": "c-002",
			"firstName": "Jane",
			"lastName": "Doe",
			"phoneNumbers": [
				{"address": "+1234567890"},
				{"address": "+0987654321"}
			],
			"emails": [
				{"address": "jane@example.com"}
			]
		}`)
		s := toContactSummary(raw)
		if len(s.Phones) != 2 {
			t.Fatalf("Phones len = %d, want 2", len(s.Phones))
		}
		if s.Phones[0] != "+1234567890" {
			t.Errorf("Phones[0] = %q, want +1234567890", s.Phones[0])
		}
		if len(s.Emails) != 1 {
			t.Fatalf("Emails len = %d, want 1", len(s.Emails))
		}
		if s.Emails[0] != "jane@example.com" {
			t.Errorf("Emails[0] = %q, want jane@example.com", s.Emails[0])
		}
	})
}

// --- formatChatLine ---

func TestFormatChatLine(t *testing.T) {
	t.Run("with display name", func(t *testing.T) {
		c := ChatSummary{
			GUID:        "iMessage;-;chat123",
			DisplayName: "My Group",
			ChatID:      "iMessage;-;chat123",
		}
		line := formatChatLine(c)
		if !strings.Contains(line, "My Group") {
			t.Errorf("formatChatLine output %q does not contain display name", line)
		}
		if !strings.Contains(line, "iMessage;-;chat123") {
			t.Errorf("formatChatLine output %q does not contain GUID", line)
		}
	})

	t.Run("without display name uses ChatID", func(t *testing.T) {
		c := ChatSummary{
			GUID:   "iMessage;-;chat123",
			ChatID: "+1234567890",
		}
		line := formatChatLine(c)
		if !strings.Contains(line, "+1234567890") {
			t.Errorf("formatChatLine output %q does not contain ChatID as fallback", line)
		}
	})

	t.Run("with last message", func(t *testing.T) {
		c := ChatSummary{
			GUID:        "iMessage;-;chat123",
			DisplayName: "Group",
			ChatID:      "iMessage;-;chat123",
			LastMessage: "Hey there!",
		}
		line := formatChatLine(c)
		if !strings.Contains(line, "Hey there!") {
			t.Errorf("formatChatLine output %q does not contain last message", line)
		}
	})
}

// --- formatMessageLine ---

func TestFormatMessageLine(t *testing.T) {
	t.Run("from me", func(t *testing.T) {
		m := MessageSummary{
			GUID:        "msg-001",
			Text:        "Hello!",
			IsFromMe:    true,
			DateCreated: 1700000000000,
			Handle:      "+1234567890",
		}
		line := formatMessageLine(m)
		if !strings.Contains(line, "->") {
			t.Errorf("formatMessageLine for IsFromMe=true should contain '->': %q", line)
		}
		if !strings.Contains(line, "Hello!") {
			t.Errorf("formatMessageLine should contain message text: %q", line)
		}
	})

	t.Run("from other", func(t *testing.T) {
		m := MessageSummary{
			GUID:        "msg-002",
			Text:        "Hi there",
			IsFromMe:    false,
			DateCreated: 1700000000000,
			Handle:      "+0987654321",
		}
		line := formatMessageLine(m)
		if !strings.Contains(line, "<-") {
			t.Errorf("formatMessageLine for IsFromMe=false should contain '<-': %q", line)
		}
		if !strings.Contains(line, "+0987654321") {
			t.Errorf("formatMessageLine should contain handle: %q", line)
		}
	})

	t.Run("empty handle uses me", func(t *testing.T) {
		m := MessageSummary{
			GUID:        "msg-003",
			Text:        "test",
			IsFromMe:    false,
			DateCreated: 1700000000000,
		}
		line := formatMessageLine(m)
		if !strings.Contains(line, "me") {
			t.Errorf("formatMessageLine with empty handle should contain 'me': %q", line)
		}
	})
}

// --- confirmDestructive ---

func TestConfirmDestructive(t *testing.T) {
	t.Run("without confirm flag returns error", func(t *testing.T) {
		root := newTestRootCmd()
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("confirm", false, "")
		root.AddCommand(cmd)

		err := confirmDestructive(cmd, "delete thing")
		if err == nil {
			t.Error("confirmDestructive without --confirm should return error")
		}
		if !strings.Contains(err.Error(), "delete thing") {
			t.Errorf("error should mention action name, got: %v", err)
		}
	})

	t.Run("with confirm flag returns nil", func(t *testing.T) {
		root := newTestRootCmd()
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("confirm", false, "")
		root.AddCommand(cmd)
		_ = cmd.Flags().Set("confirm", "true")

		err := confirmDestructive(cmd, "delete thing")
		if err != nil {
			t.Errorf("confirmDestructive with --confirm should return nil, got: %v", err)
		}
	})
}
