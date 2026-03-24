package streeteasy

import (
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"hello world", 20, "hello world"},
		{"hello world", 8, "hello..."},
		{"hello", 5, "hello"},
		{"", 10, ""},
		{"abcdef", 6, "abcdef"},
		{"abcdefg", 6, "abc..."},
	}
	for _, tt := range tests {
		got := truncate(tt.s, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
		}
	}
}

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		price int64
		want  string
	}{
		{0, "-"},
		{500, "$500"},
		{1000, "$1,000"},
		{450000, "$450,000"},
		{1250000, "$1,250,000"},
		{10000000, "$10,000,000"},
		{999, "$999"},
		{1001, "$1,001"},
	}
	for _, tt := range tests {
		got := formatPrice(tt.price)
		if got != tt.want {
			t.Errorf("formatPrice(%d) = %q, want %q", tt.price, got, tt.want)
		}
	}
}

func TestJsonStr(t *testing.T) {
	m := map[string]any{
		"name":   "Broadway",
		"price":  float64(1500000),
		"active": true,
	}

	if got := jsonStr(m, "name"); got != "Broadway" {
		t.Errorf("jsonStr string = %q, want %q", got, "Broadway")
	}
	if got := jsonStr(m, "price"); got != "1500000" {
		t.Errorf("jsonStr float64 = %q, want %q", got, "1500000")
	}
	if got := jsonStr(m, "missing"); got != "" {
		t.Errorf("jsonStr missing = %q, want empty", got)
	}
	// bool is unsupported type — returns empty
	if got := jsonStr(m, "active"); got != "" {
		t.Errorf("jsonStr bool = %q, want empty", got)
	}
}

func TestJsonInt(t *testing.T) {
	m := map[string]any{
		"beds":  float64(3),
		"rooms": 2,
		"name":  "test",
	}

	if got := jsonInt(m, "beds"); got != 3 {
		t.Errorf("jsonInt float64 = %d, want 3", got)
	}
	if got := jsonInt(m, "rooms"); got != 2 {
		t.Errorf("jsonInt int = %d, want 2", got)
	}
	if got := jsonInt(m, "missing"); got != 0 {
		t.Errorf("jsonInt missing = %d, want 0", got)
	}
	if got := jsonInt(m, "name"); got != 0 {
		t.Errorf("jsonInt string = %d, want 0", got)
	}
}

func TestJsonFloat(t *testing.T) {
	m := map[string]any{
		"baths": 2.5,
		"name":  "test",
	}

	if got := jsonFloat(m, "baths"); got != 2.5 {
		t.Errorf("jsonFloat = %f, want 2.5", got)
	}
	if got := jsonFloat(m, "missing"); got != 0 {
		t.Errorf("jsonFloat missing = %f, want 0", got)
	}
	if got := jsonFloat(m, "name"); got != 0 {
		t.Errorf("jsonFloat string = %f, want 0", got)
	}
}

func TestPrintListingSummaries_Empty(t *testing.T) {
	root := newTestRootCmd()
	p := &Provider{ClientFactory: DefaultClientFactory()}
	p.RegisterCommands(root)

	// Directly test the print function with empty list.
	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json=false"})
		_ = printListingSummaries(root, []ListingSummary{})
	})
	if !strings.Contains(out, "No listings") {
		t.Errorf("expected 'No listings' message, got: %s", out)
	}
}

func TestPrintPriceHistory_Empty(t *testing.T) {
	root := newTestRootCmd()
	out := captureStdout(t, func() {
		_ = printPriceHistory(root, []PriceHistoryEntry{})
	})
	if !strings.Contains(out, "No price history") {
		t.Errorf("expected 'No price history' message, got: %s", out)
	}
}

func TestTruncateBody(t *testing.T) {
	short := []byte("short")
	if got := truncateBody(short); got != "short" {
		t.Errorf("truncateBody short = %q, want %q", got, "short")
	}

	long := []byte(strings.Repeat("a", 300))
	got := truncateBody(long)
	if len(got) != 203 { // 200 + "..."
		t.Errorf("truncateBody long len = %d, want 203", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("truncateBody long should end with '...'")
	}
}
