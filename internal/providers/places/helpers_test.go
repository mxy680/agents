package places

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"", 5, ""},
	}
	for _, tt := range tests {
		got := truncate(tt.s, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
		}
	}
}

func TestParseLatLng(t *testing.T) {
	lat, lng, err := parseLatLng("41.4993,-81.6944")
	if err != nil {
		t.Fatal(err)
	}
	if lat != 41.4993 || lng != -81.6944 {
		t.Errorf("got (%f, %f), want (41.4993, -81.6944)", lat, lng)
	}

	_, _, err = parseLatLng("invalid")
	if err == nil {
		t.Fatal("expected error for invalid lat,lng")
	}

	_, _, err = parseLatLng("abc,def")
	if err == nil {
		t.Fatal("expected error for non-numeric lat,lng")
	}
}

func TestToPlaceSummary(t *testing.T) {
	e := Entry{
		Title:       "Test Place",
		Address:     "123 Test St",
		Phone:       "+1 555-0100",
		Rating:      4.5,
		ReviewCount: 100,
		Website:     "https://example.com",
		PriceRange:  "$$",
		Status:      "OPERATIONAL",
	}

	s := toPlaceSummary(e)
	if s.Title != "Test Place" {
		t.Errorf("Title = %q", s.Title)
	}
	if s.Rating != 4.5 {
		t.Errorf("Rating = %f", s.Rating)
	}
	if s.Phone != "+1 555-0100" {
		t.Errorf("Phone = %q", s.Phone)
	}
	if s.PriceRange != "$$" {
		t.Errorf("PriceRange = %q", s.PriceRange)
	}
}

func TestEntryJSONRoundTrip(t *testing.T) {
	entries := testEntries()
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}

	var parsed []Entry
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	if len(parsed) != len(entries) {
		t.Fatalf("got %d entries, want %d", len(parsed), len(entries))
	}
	if parsed[0].Title != "Coffee Corner" {
		t.Errorf("Title = %q", parsed[0].Title)
	}
	if parsed[0].Rating != 4.5 {
		t.Errorf("Rating = %f", parsed[0].Rating)
	}
	if len(parsed[0].Emails) != 1 {
		t.Errorf("Emails = %v", parsed[0].Emails)
	}
}

func TestParseOutputFileJSONArray(t *testing.T) {
	entries := testEntries()[:2]
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "output.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	parsed, err := parseOutputFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 2 {
		t.Fatalf("got %d entries, want 2", len(parsed))
	}
	if parsed[0].Title != "Coffee Corner" {
		t.Errorf("Title = %q", parsed[0].Title)
	}
}

func TestParseOutputFileNDJSON(t *testing.T) {
	entries := testEntries()[:2]
	var content string
	for _, e := range entries {
		line, _ := json.Marshal(e)
		content += string(line) + "\n"
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "output.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	parsed, err := parseOutputFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 2 {
		t.Fatalf("got %d entries, want 2", len(parsed))
	}
}

func TestParseOutputFileEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "output.json")
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	parsed, err := parseOutputFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 0 {
		t.Fatalf("expected empty, got %d", len(parsed))
	}
}

func TestParseOutputFileWithBOM(t *testing.T) {
	entries := testEntries()[:1]
	data, _ := json.Marshal(entries)
	bom := []byte{0xef, 0xbb, 0xbf}
	withBOM := append(bom, data...)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "output.json")
	if err := os.WriteFile(path, withBOM, 0o600); err != nil {
		t.Fatal(err)
	}

	parsed, err := parseOutputFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 {
		t.Fatalf("got %d entries, want 1", len(parsed))
	}
}

func TestTrimBOM(t *testing.T) {
	bom := []byte{0xef, 0xbb, 0xbf, 'h', 'i'}
	got := trimBOM(bom)
	if string(got) != "hi" {
		t.Errorf("trimBOM = %q, want 'hi'", got)
	}

	noBom := []byte("hi")
	got = trimBOM(noBom)
	if string(got) != "hi" {
		t.Errorf("trimBOM = %q, want 'hi'", got)
	}
}
