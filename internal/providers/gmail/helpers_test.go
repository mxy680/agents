package gmail

import (
	"fmt"
	"testing"

	api "google.golang.org/api/gmail/v1"
)

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

// ---- extractHeaders ----

func TestExtractHeaders(t *testing.T) {
	headers := []*api.MessagePartHeader{
		{Name: "From", Value: "alice@example.com"},
		{Name: "To", Value: "bob@example.com"},
		{Name: "Subject", Value: "Hello"},
		{Name: "Date", Value: "Mon, 16 Mar 2026 10:00:00 -0500"},
	}

	t.Run("extracts requested headers", func(t *testing.T) {
		got := extractHeaders(headers, "From", "Subject")
		if got["From"] != "alice@example.com" {
			t.Errorf("expected From=alice@example.com, got %s", got["From"])
		}
		if got["Subject"] != "Hello" {
			t.Errorf("expected Subject=Hello, got %s", got["Subject"])
		}
		// To was not requested
		if _, ok := got["To"]; ok {
			t.Error("expected To to be absent from result")
		}
	})

	t.Run("missing header returns empty string", func(t *testing.T) {
		got := extractHeaders(headers, "X-Custom")
		if got["X-Custom"] != "" {
			t.Errorf("expected empty string for missing header, got %q", got["X-Custom"])
		}
	})

	t.Run("empty headers slice", func(t *testing.T) {
		got := extractHeaders(nil, "From")
		if got["From"] != "" {
			t.Errorf("expected empty string from nil headers, got %q", got["From"])
		}
	})
}

// ---- confirmDestructive ----

func TestConfirmDestructive(t *testing.T) {
	t.Run("returns error when --confirm absent", func(t *testing.T) {
		cmd := newTestRootCmd()
		cmd.Flags().Bool("confirm", false, "")
		err := confirmDestructive(cmd)
		if err == nil {
			t.Error("expected error when --confirm not provided")
		}
	})

	t.Run("returns nil when --confirm set", func(t *testing.T) {
		cmd := newTestRootCmd()
		cmd.Flags().Bool("confirm", false, "")
		if err := cmd.ParseFlags([]string{"--confirm"}); err != nil {
			t.Fatalf("ParseFlags: %v", err)
		}
		err := confirmDestructive(cmd)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// ---- dryRunResult ----

func TestDryRunResult(t *testing.T) {
	t.Run("text output prints dry-run prefix", func(t *testing.T) {
		root := newTestRootCmd()
		output := captureStdout(t, func() {
			err := dryRunResult(root, "would delete message", nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if output == "" {
			t.Error("expected non-empty output")
		}
		if len(output) < len("[DRY RUN]") {
			t.Errorf("expected output to contain [DRY RUN], got: %s", output)
		}
	})

	t.Run("json output serialises data", func(t *testing.T) {
		root := newTestRootCmd()
		if err := root.ParseFlags([]string{"--json"}); err != nil {
			t.Fatalf("ParseFlags: %v", err)
		}
		data := map[string]string{"status": "dry-run"}
		output := captureStdout(t, func() {
			err := dryRunResult(root, "would delete message", data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if output == "" {
			t.Error("expected non-empty JSON output")
		}
	})
}
