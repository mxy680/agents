package places

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLookupJSON(t *testing.T) {
	scraper := mockScraper(testEntries()[:1], nil)

	root := newTestRootCmd()
	lookupCmd := newLookupCmd(scraper)
	root.AddCommand(lookupCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lookup", "--url=https://maps.google.com/?cid=123", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var entry Entry
	if err := json.Unmarshal([]byte(out), &entry); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if entry.Title != "Coffee Corner" {
		t.Errorf("Title = %q", entry.Title)
	}
	if entry.Rating != 4.5 {
		t.Errorf("Rating = %f", entry.Rating)
	}
}

func TestLookupText(t *testing.T) {
	scraper := mockScraper(testEntries()[:1], nil)

	root := newTestRootCmd()
	lookupCmd := newLookupCmd(scraper)
	root.AddCommand(lookupCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lookup", "--url=https://maps.google.com/?cid=123"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	expected := []string{
		"Coffee Corner",
		"123 Main St, Cleveland",
		"+1 216-555-0100",
		"coffeecorner.example.com",
		"4.5",
		"$$",
		"info@coffeecorner.example.com",
		"Monday",
		"Alice",
		"Best coffee in town!",
	}
	for _, e := range expected {
		if !strings.Contains(out, e) {
			t.Errorf("expected output to contain %q, got: %s", e, out)
		}
	}
}

func TestLookupWithEmail(t *testing.T) {
	scraper, captured := mockScraperWithCapture(testEntries()[:1])

	root := newTestRootCmd()
	lookupCmd := newLookupCmd(scraper)
	root.AddCommand(lookupCmd)

	captureStdout(t, func() {
		root.SetArgs([]string{"lookup", "--url=https://maps.google.com/?cid=123", "--email", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !captured.Email {
		t.Error("Email should be true")
	}
	if len(captured.Queries) != 1 || captured.Queries[0] != "https://maps.google.com/?cid=123" {
		t.Errorf("Queries = %v", captured.Queries)
	}
}

func TestLookupNoResult(t *testing.T) {
	scraper := mockScraper([]Entry{}, nil)

	root := newTestRootCmd()
	lookupCmd := newLookupCmd(scraper)
	root.AddCommand(lookupCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lookup", "--url=https://maps.google.com/?cid=999"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "No place found") {
		t.Errorf("expected 'No place found' in output, got: %s", out)
	}
}

func TestLookupMissingURL(t *testing.T) {
	scraper := mockScraper(testEntries(), nil)

	root := newTestRootCmd()
	lookupCmd := newLookupCmd(scraper)
	root.AddCommand(lookupCmd)

	root.SetArgs([]string{"lookup"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing --url")
	}
}

func TestLookupScraperError(t *testing.T) {
	scraper := errScraper("timeout")

	root := newTestRootCmd()
	lookupCmd := newLookupCmd(scraper)
	root.AddCommand(lookupCmd)

	root.SetArgs([]string{"lookup", "--url=https://maps.google.com/?cid=123"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error from scraper")
	}
}
