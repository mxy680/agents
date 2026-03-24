package nyscef

import (
	"strings"
	"testing"
)

func TestResolveCountyCode(t *testing.T) {
	tests := []struct {
		input    string
		wantCode string
		wantOK   bool
	}{
		{"bronx", "2", true},
		{"BRONX", "2", true},
		{"Bronx", "2", true},
		{"kings", "24", true},
		{"brooklyn", "24", true},
		{"new york", "31", true},
		{"manhattan", "31", true},
		{"queens", "41", true},
		{"richmond", "43", true},
		{"staten island", "43", true},
		{"2", "2", true},   // numeric pass-through
		{"31", "31", true}, // numeric pass-through
		{"fakecounty", "", false},
		{"", "", false},
	}

	for _, tc := range tests {
		code, ok := resolveCountyCode(tc.input)
		if ok != tc.wantOK {
			t.Errorf("resolveCountyCode(%q) ok=%v, want %v", tc.input, ok, tc.wantOK)
		}
		if ok && code != tc.wantCode {
			t.Errorf("resolveCountyCode(%q) code=%q, want %q", tc.input, code, tc.wantCode)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"123", true},
		{"0", true},
		{"", false},
		{"abc", false},
		{"12a", false},
		{"2", true},
	}
	for _, tc := range tests {
		got := isNumeric(tc.input)
		if got != tc.want {
			t.Errorf("isNumeric(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
	}
	for _, tc := range tests {
		got := truncate(tc.s, tc.max)
		if got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.s, tc.max, got, tc.want)
		}
	}
}

func TestTruncateBody(t *testing.T) {
	short := []byte("short response")
	if got := truncateBody(short); got != "short response" {
		t.Errorf("truncateBody short = %q", got)
	}

	long := []byte(strings.Repeat("x", 300))
	got := truncateBody(long)
	if len(got) > 210 { // 200 + "..."
		t.Errorf("truncateBody long = %d chars, expected <= 210", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("truncateBody long should end with '...', got %q", got[len(got)-5:])
	}
}
