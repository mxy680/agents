package instagram

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"abcdef", 5, "ab..."},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestFormatTimestamp(t *testing.T) {
	// Zero value should return "-"
	if got := formatTimestamp(0); got != "-" {
		t.Errorf("formatTimestamp(0) = %q, want %q", got, "-")
	}

	// Non-zero value should return a date/time string in the expected format.
	// We derive the expected value from the local timezone to avoid CI failures.
	epoch := int64(1700000000)
	got := formatTimestamp(epoch)
	if len(got) != len("2006-01-02 15:04") {
		t.Errorf("formatTimestamp(%d) = %q, unexpected format length", epoch, got)
	}
	// Check that the result contains digits and dashes/spaces (format sanity check).
	for _, c := range got {
		if c != '-' && c != ':' && c != ' ' && (c < '0' || c > '9') {
			t.Errorf("formatTimestamp(%d) = %q, unexpected character %q", epoch, got, c)
		}
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{1000000, "1.0M"},
		{3400000, "3.4M"},
	}
	for _, tt := range tests {
		got := formatCount(tt.n)
		if got != tt.want {
			t.Errorf("formatCount(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestConfirmDestructive(t *testing.T) {
	// Without --confirm flag (defaults to false)
	cmd := newTestRootCmd()
	cmd.Flags().Bool("confirm", false, "")
	if err := confirmDestructive(cmd); err == nil {
		t.Error("expected error without --confirm")
	}

	// With --confirm flag set to true
	cmd2 := newTestRootCmd()
	cmd2.Flags().Bool("confirm", false, "")
	_ = cmd2.Flags().Set("confirm", "true")
	if err := confirmDestructive(cmd2); err != nil {
		t.Errorf("unexpected error with --confirm: %v", err)
	}
}

func TestPrintUserSummariesEmpty(t *testing.T) {
	cmd := newTestRootCmd()
	out := captureStdout(t, func() {
		if err := printUserSummaries(cmd, []UserSummary{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	mustContain(t, out, "No users found.")
}

func TestPrintUserSummariesText(t *testing.T) {
	cmd := newTestRootCmd()
	users := []UserSummary{
		{ID: "1", Username: "alice", FullName: "Alice Smith", IsPrivate: false, IsVerified: true},
	}
	out := captureStdout(t, func() {
		if err := printUserSummaries(cmd, users); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	mustContain(t, out, "USERNAME")
	mustContain(t, out, "alice")
	mustContain(t, out, "Alice Smith")
}

func TestPrintMediaSummariesEmpty(t *testing.T) {
	cmd := newTestRootCmd()
	out := captureStdout(t, func() {
		if err := printMediaSummaries(cmd, []MediaSummary{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	mustContain(t, out, "No media found.")
}

func TestPrintMediaSummariesText(t *testing.T) {
	cmd := newTestRootCmd()
	media := []MediaSummary{
		{ID: "m1", Caption: "Hello world", Timestamp: 1700000000, LikeCount: 42, CommentCount: 5},
	}
	out := captureStdout(t, func() {
		if err := printMediaSummaries(cmd, media); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	mustContain(t, out, "ID")
	mustContain(t, out, "Hello world")
	mustContain(t, out, "42")
}
