package places

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSearchJSON(t *testing.T) {
	scraper := mockScraper(testEntries(), nil)

	root := newTestRootCmd()
	searchCmd := newSearchCmd(scraper)
	root.AddCommand(searchCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=coffee in cleveland", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var entries []Entry
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Title != "Coffee Corner" {
		t.Errorf("first entry title = %q", entries[0].Title)
	}
}

func TestSearchText(t *testing.T) {
	scraper := mockScraper(testEntries(), nil)

	root := newTestRootCmd()
	searchCmd := newSearchCmd(scraper)
	root.AddCommand(searchCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=coffee in cleveland"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Coffee Corner") {
		t.Errorf("expected 'Coffee Corner' in output, got: %s", out)
	}
	if !strings.Contains(out, "Bean & Leaf") {
		t.Errorf("expected 'Bean & Leaf' in output, got: %s", out)
	}
	if !strings.Contains(out, "216-555-0100") {
		t.Errorf("expected phone in output, got: %s", out)
	}
}

func TestSearchWithLimit(t *testing.T) {
	scraper := mockScraper(testEntries(), nil)

	root := newTestRootCmd()
	searchCmd := newSearchCmd(scraper)
	root.AddCommand(searchCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=coffee", "--limit=1", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var entries []Entry
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (limit=1), got %d", len(entries))
	}
}

func TestSearchPassesOptions(t *testing.T) {
	scraper, captured := mockScraperWithCapture(testEntries())

	root := newTestRootCmd()
	searchCmd := newSearchCmd(scraper)
	root.AddCommand(searchCmd)

	captureStdout(t, func() {
		root.SetArgs([]string{"search",
			"--query=dentists in cleveland",
			"--geo=41.499,-81.694",
			"--zoom=14",
			"--depth=2",
			"--email",
			"--concurrency=4",
			"--lang=en",
			"--limit=10",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if len(captured.Queries) != 1 || captured.Queries[0] != "dentists in cleveland" {
		t.Errorf("Queries = %v", captured.Queries)
	}
	if captured.Geo != "41.499,-81.694" {
		t.Errorf("Geo = %q", captured.Geo)
	}
	if captured.Zoom != 14 {
		t.Errorf("Zoom = %d", captured.Zoom)
	}
	if captured.Depth != 2 {
		t.Errorf("Depth = %d", captured.Depth)
	}
	if !captured.Email {
		t.Error("Email should be true")
	}
	if captured.Concurrency != 4 {
		t.Errorf("Concurrency = %d", captured.Concurrency)
	}
	if captured.Lang != "en" {
		t.Errorf("Lang = %q", captured.Lang)
	}
	if captured.Limit != 10 {
		t.Errorf("Limit = %d", captured.Limit)
	}
}

func TestSearchInvalidGeo(t *testing.T) {
	scraper := mockScraper(testEntries(), nil)

	root := newTestRootCmd()
	searchCmd := newSearchCmd(scraper)
	root.AddCommand(searchCmd)

	root.SetArgs([]string{"search", "--query=coffee", "--geo=invalid"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --geo")
	}
}

func TestSearchMissingQuery(t *testing.T) {
	scraper := mockScraper(testEntries(), nil)

	root := newTestRootCmd()
	searchCmd := newSearchCmd(scraper)
	root.AddCommand(searchCmd)

	root.SetArgs([]string{"search"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing --query")
	}
}

func TestSearchScraperError(t *testing.T) {
	scraper := errScraper("connection refused")

	root := newTestRootCmd()
	searchCmd := newSearchCmd(scraper)
	root.AddCommand(searchCmd)

	root.SetArgs([]string{"search", "--query=coffee"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error from scraper")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error = %q, expected 'connection refused'", err.Error())
	}
}

func TestSearchNoResults(t *testing.T) {
	scraper := mockScraper([]Entry{}, nil)

	root := newTestRootCmd()
	searchCmd := newSearchCmd(scraper)
	root.AddCommand(searchCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=nonexistent"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "No places found") {
		t.Errorf("expected 'No places found' in output, got: %s", out)
	}
}
