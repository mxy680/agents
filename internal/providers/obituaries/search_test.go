package obituaries

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --------------------------------------------------------------------------
// toSummaries unit tests
// --------------------------------------------------------------------------

func TestToSummaries_Limit(t *testing.T) {
	got := toSummaries(testObituaries, 2)
	if len(got) != 2 {
		t.Errorf("expected 2 summaries with limit=2, got %d", len(got))
	}
}

func TestToSummaries_LimitExceedsCount(t *testing.T) {
	got := toSummaries(testObituaries, 100)
	if len(got) != len(testObituaries) {
		t.Errorf("expected %d summaries, got %d", len(testObituaries), len(got))
	}
}

func TestToSummaries_Fields(t *testing.T) {
	got := toSummaries(testObituaries[:1], 10)
	if len(got) == 0 {
		t.Fatal("expected at least one summary")
	}
	s := got[0]
	if s.Name.First != "Maria" {
		t.Errorf("Name.First = %q, want %q", s.Name.First, "Maria")
	}
	if s.Name.Last != "Gonzalez" {
		t.Errorf("Name.Last = %q, want %q", s.Name.Last, "Gonzalez")
	}
	if s.Age != 82 {
		t.Errorf("Age = %d, want 82", s.Age)
	}
	if s.City != "Bronx" {
		t.Errorf("City = %q, want %q", s.City, "Bronx")
	}
	if s.PublishDate != "2026-03-20" {
		t.Errorf("PublishDate = %q, want %q", s.PublishDate, "2026-03-20")
	}
}

// --------------------------------------------------------------------------
// toNameEntries unit tests
// --------------------------------------------------------------------------

func TestToNameEntries_Limit(t *testing.T) {
	got := toNameEntries(testObituaries, 3)
	if len(got) != 3 {
		t.Errorf("expected 3 entries with limit=3, got %d", len(got))
	}
}

func TestToNameEntries_Fields(t *testing.T) {
	got := toNameEntries(testObituaries[:1], 10)
	if len(got) == 0 {
		t.Fatal("expected at least one entry")
	}
	e := got[0]
	if e.First != "Maria" {
		t.Errorf("First = %q, want %q", e.First, "Maria")
	}
	if e.Last != "Gonzalez" {
		t.Errorf("Last = %q, want %q", e.Last, "Gonzalez")
	}
	if e.Full != "Maria Gonzalez" {
		t.Errorf("Full = %q, want %q", e.Full, "Maria Gonzalez")
	}
	if e.PublishDate != "2026-03-20" {
		t.Errorf("PublishDate = %q, want %q", e.PublishDate, "2026-03-20")
	}
}

// --------------------------------------------------------------------------
// min unit tests
// --------------------------------------------------------------------------

func TestMin(t *testing.T) {
	tests := []struct{ a, b, want int }{
		{3, 5, 3},
		{5, 3, 3},
		{4, 4, 4},
	}
	for _, tt := range tests {
		if got := min(tt.a, tt.b); got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

// --------------------------------------------------------------------------
// `obituaries search` command integration tests (via mock server)
// --------------------------------------------------------------------------

func TestSearch_TextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"obituaries", "search", "--city", "Bronx"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Gonzalez") {
		t.Errorf("expected 'Gonzalez' in output, got: %s", out)
	}
	if !strings.Contains(out, "Washington") {
		t.Errorf("expected 'Washington' in output, got: %s", out)
	}
}

func TestSearch_JSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"obituaries", "search",
			"--city", "Bronx",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []ObituarySummary
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) == 0 {
		t.Errorf("expected at least one obituary in JSON output")
	}
	if results[0].Name.Full == "" {
		t.Errorf("expected non-empty Name.Full")
	}
}

func TestSearch_Alias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		// Use provider alias "obit" and subcommand alias "s".
		root.SetArgs([]string{"obit", "s", "--city", "Bronx"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Gonzalez") {
		t.Errorf("expected 'Gonzalez' via alias, got: %s", out)
	}
}

func TestSearch_WithState(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"obituaries", "search",
			"--city", "Bronx", "--state", "New York",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []ObituarySummary
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
	}
	if results[0].State != "New York" {
		t.Errorf("State = %q, want %q", results[0].State, "New York")
	}
}

func TestSearch_Limit(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"obituaries", "search",
			"--city", "Bronx", "--limit", "2",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []ObituarySummary
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results with --limit 2, got %d", len(results))
	}
}

func TestSearch_InvalidDateRange(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{
		"obituaries", "search",
		"--city", "Bronx", "--date-range", "BadValue",
	})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid --date-range, got nil")
	}
}

func TestSearch_EmptyResult(t *testing.T) {
	// Serve an empty result set.
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(buildSearchResponse(nil))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"obituaries", "search", "--city", "Bronx"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No obituaries found") {
		t.Errorf("expected empty-results message, got: %s", out)
	}
}

func TestSearch_HTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"obituaries", "search", "--city", "Bronx"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
}

// --------------------------------------------------------------------------
// `obituaries names` command integration tests (via mock server)
// --------------------------------------------------------------------------

func TestNames_TextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"obituaries", "names", "--city", "Bronx"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Gonzalez") {
		t.Errorf("expected 'Gonzalez' in output, got: %s", out)
	}
	if !strings.Contains(out, "FIRST") {
		t.Errorf("expected 'FIRST' header in output, got: %s", out)
	}
}

func TestNames_JSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"obituaries", "names",
			"--city", "Bronx",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var entries []NameEntry
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(entries) == 0 {
		t.Errorf("expected at least one name entry in JSON output")
	}
	if entries[0].First == "" || entries[0].Last == "" {
		t.Errorf("expected non-empty first/last names")
	}
	if entries[0].PublishDate == "" {
		t.Errorf("expected non-empty publish_date")
	}
}

func TestNames_Alias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"obit", "n", "--city", "Bronx", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var entries []NameEntry
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("expected valid JSON via alias, got: %s, error: %v", out, err)
	}
	if len(entries) == 0 {
		t.Errorf("expected entries via alias")
	}
}

func TestNames_EmptyResult(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(buildSearchResponse(nil))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"obituaries", "names", "--city", "Bronx"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No obituaries found") {
		t.Errorf("expected empty-results message, got: %s", out)
	}
}

func TestNames_InvalidDateRange(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{
		"obituaries", "names",
		"--city", "Bronx", "--date-range", "InvalidRange",
	})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid --date-range, got nil")
	}
}

// --------------------------------------------------------------------------
// truncateName unit tests
// --------------------------------------------------------------------------

func TestTruncateName(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"Short", 10, "Short"},
		{"Exactly ten", 11, "Exactly ten"},
		{"A very long name here", 10, "A very ..."},
	}
	for _, tt := range tests {
		got := truncateName(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncateName(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}
