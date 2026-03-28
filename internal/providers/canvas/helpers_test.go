package canvas

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{
			name:     "short string unchanged",
			input:    "hello",
			max:      10,
			expected: "hello",
		},
		{
			name:     "exact length unchanged",
			input:    "hello",
			max:      5,
			expected: "hello",
		},
		{
			name:     "long string truncated with ellipsis",
			input:    "hello world this is a long string",
			max:      10,
			expected: "hello w...",
		},
		{
			name:     "empty string unchanged",
			input:    "",
			max:      10,
			expected: "",
		},
		{
			name:     "unicode string truncated correctly",
			input:    "héllo wörld",
			max:      7,
			expected: "héll...",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := truncate(tc.input, tc.max)
			if result != tc.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.max, result, tc.expected)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{int64(1.5 * 1024 * 1024), "1.5 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{int64(2.5 * 1024 * 1024 * 1024), "2.5 GB"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d bytes", tc.bytes), func(t *testing.T) {
			result := formatSize(tc.bytes)
			if result != tc.expected {
				t.Errorf("formatSize(%d) = %q, want %q", tc.bytes, result, tc.expected)
			}
		})
	}
}

func TestConfirmDestructive(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("confirm", false, "confirm")

	// Without --confirm flag, should return an error.
	err := confirmDestructive(cmd, "delete this?")
	if err == nil {
		t.Error("expected error from confirmDestructive without --confirm")
	}
	if !strings.Contains(err.Error(), "--confirm") {
		t.Errorf("error should mention --confirm to proceed, got: %v", err)
	}

	// With --confirm flag, should succeed.
	cmd.Flags().Set("confirm", "true")
	err = confirmDestructive(cmd, "delete this?")
	if err != nil {
		t.Errorf("expected no error with --confirm, got: %v", err)
	}
}

func TestDryRunResult(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignments", "delete", "--course-id", "101", "--assignment-id", "501", "--confirm", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("dryRunResult should output DRY RUN prefix, got: %s", output)
	}
}
