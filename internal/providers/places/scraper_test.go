package places

import (
	"os"
	"testing"
)

func TestDefaultScraperBinaryFromEnv(t *testing.T) {
	t.Setenv("GOOGLE_MAPS_SCRAPER_BIN", "/usr/local/bin/custom-scraper")
	got := defaultScraperBinary()
	if got != "/usr/local/bin/custom-scraper" {
		t.Errorf("got %q, want /usr/local/bin/custom-scraper", got)
	}
}

func TestDefaultScraperBinaryFallback(t *testing.T) {
	t.Setenv("GOOGLE_MAPS_SCRAPER_BIN", "")
	got := defaultScraperBinary()
	// Should return either a PATH-found binary or the default name
	if got == "" {
		t.Error("should not be empty")
	}
}

func TestDefaultScraperFuncMissingBinary(t *testing.T) {
	fn := defaultScraperFunc("/nonexistent/scraper/binary")
	_, err := fn(t.Context(), ScraperOptions{
		Queries: []string{"test query"},
	})
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestDefaultScraperFuncWritesInputFile(t *testing.T) {
	// Create a mock binary that just touches the output file
	tmpDir := t.TempDir()
	mockBin := tmpDir + "/mock-scraper"
	// Write a shell script that creates the expected output file
	script := `#!/bin/sh
# Parse -results flag to find output path
while [ "$#" -gt 0 ]; do
  case "$1" in
    -results) shift; echo '[]' > "$1"; shift ;;
    *) shift ;;
  esac
done
`
	if err := os.WriteFile(mockBin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	fn := defaultScraperFunc(mockBin)
	result, err := fn(t.Context(), ScraperOptions{
		Queries:     []string{"coffee in cleveland"},
		Geo:         "41.499,-81.694",
		Zoom:        14,
		Depth:       2,
		Email:       true,
		Concurrency: 4,
		Lang:        "en",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries from empty output, got %d", len(result.Entries))
	}
}

func TestParseOutputFileNotFound(t *testing.T) {
	_, err := parseOutputFile("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseOutputFileInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := tmpDir + "/bad.json"
	if err := os.WriteFile(path, []byte("[invalid json}"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := parseOutputFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseOutputFileInvalidNDJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := tmpDir + "/bad.ndjson"
	if err := os.WriteFile(path, []byte("{invalid json}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := parseOutputFile(path)
	if err == nil {
		t.Fatal("expected error for invalid NDJSON")
	}
}
