package linkedin

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"", 5, ""},
		{"ab", 3, "ab"},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestFormatTimestamp(t *testing.T) {
	if got := formatTimestamp(0); got != "-" {
		t.Errorf("formatTimestamp(0) = %q, want %q", got, "-")
	}
	// 1704067200000 = 2024-01-01 00:00 UTC (in milliseconds)
	got := formatTimestamp(1704067200000)
	if got == "-" {
		t.Errorf("formatTimestamp(1704067200000) = %q, want non-dash", got)
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{1000000, "1.0M"},
		{2500000, "2.5M"},
	}
	for _, tt := range tests {
		got := formatCount(tt.input)
		if got != tt.want {
			t.Errorf("formatCount(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
