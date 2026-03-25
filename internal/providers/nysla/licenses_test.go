package nysla

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --------------------------------------------------------------------------
// lookupCounty unit tests
// --------------------------------------------------------------------------

func TestLookupCounty_Valid(t *testing.T) {
	cases := []struct {
		borough string
		want    string
	}{
		{"bronx", "BRONX"},
		{"brooklyn", "KINGS"},
		{"manhattan", "NEW YORK"},
		{"queens", "QUEENS"},
		{"staten-island", "RICHMOND"},
	}
	for _, tc := range cases {
		got, err := lookupCounty(tc.borough)
		if err != nil {
			t.Errorf("lookupCounty(%q) unexpected error: %v", tc.borough, err)
		}
		if got != tc.want {
			t.Errorf("lookupCounty(%q) = %q, want %q", tc.borough, got, tc.want)
		}
	}
}

func TestLookupCounty_Invalid(t *testing.T) {
	_, err := lookupCounty("jersey-city")
	if err == nil {
		t.Error("expected error for unknown borough, got nil")
	}
	if !strings.Contains(err.Error(), "jersey-city") {
		t.Errorf("expected error to mention borough name, got: %v", err)
	}
}

// --------------------------------------------------------------------------
// buildTypeBreakdown unit tests
// --------------------------------------------------------------------------

func TestBuildTypeBreakdown_Empty(t *testing.T) {
	bd := buildTypeBreakdown(nil)
	if len(bd) != 0 {
		t.Errorf("expected empty breakdown for nil input, got %d items", len(bd))
	}
}

func TestBuildTypeBreakdown_Counts(t *testing.T) {
	licenses := []rawLicense{
		{LicenseTypeName: "RESTAURANT/BAR"},
		{LicenseTypeName: "RESTAURANT/BAR"},
		{LicenseTypeName: "TAVERN"},
	}
	bd := buildTypeBreakdown(licenses)
	if len(bd) != 2 {
		t.Fatalf("expected 2 type entries, got %d", len(bd))
	}
	// Should be sorted by count descending — RESTAURANT/BAR first.
	if bd[0].Type != "RESTAURANT/BAR" {
		t.Errorf("expected RESTAURANT/BAR first, got %q", bd[0].Type)
	}
	if bd[0].Count != 2 {
		t.Errorf("expected count 2 for RESTAURANT/BAR, got %d", bd[0].Count)
	}
	if bd[1].Type != "TAVERN" || bd[1].Count != 1 {
		t.Errorf("unexpected second entry: %+v", bd[1])
	}
}

func TestBuildTypeBreakdown_SortedDescending(t *testing.T) {
	bd := buildTypeBreakdown(testLicenses)
	for i := 1; i < len(bd); i++ {
		if bd[i].Count > bd[i-1].Count {
			t.Errorf("breakdown not sorted descending: bd[%d].Count=%d > bd[%d].Count=%d",
				i, bd[i].Count, i-1, bd[i-1].Count)
		}
	}
}

// --------------------------------------------------------------------------
// toSummary unit tests
// --------------------------------------------------------------------------

func TestToSummary(t *testing.T) {
	r := testLicenses[0]
	s := toSummary(r)
	if s.SerialNumber != r.SerialNumber {
		t.Errorf("SerialNumber mismatch: got %q, want %q", s.SerialNumber, r.SerialNumber)
	}
	if s.PremisesName != r.PremisesName {
		t.Errorf("PremisesName mismatch: got %q, want %q", s.PremisesName, r.PremisesName)
	}
	if s.Address != r.PremisesAddress {
		t.Errorf("Address mismatch: got %q, want %q", s.Address, r.PremisesAddress)
	}
	if s.LicenseType != r.LicenseTypeName {
		t.Errorf("LicenseType mismatch: got %q, want %q", s.LicenseType, r.LicenseTypeName)
	}
}

// --------------------------------------------------------------------------
// truncateStr unit tests
// --------------------------------------------------------------------------

func TestTruncateStr(t *testing.T) {
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
		got := truncateStr(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncateStr(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

// --------------------------------------------------------------------------
// shortDate unit tests
// --------------------------------------------------------------------------

func TestShortDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2025-03-01T00:00:00.000", "2025-03-01"},
		{"2025-03-01", "2025-03-01"},
		{"2025", "2025"},
	}
	for _, tt := range tests {
		got := shortDate(tt.input)
		if got != tt.want {
			t.Errorf("shortDate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --------------------------------------------------------------------------
// `licenses search` command integration tests (via mock server)
// --------------------------------------------------------------------------

func TestLicensesSearch_TextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "search",
			"--borough", "bronx",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "BRONX GRILL") {
		t.Errorf("expected premises name in output, got: %s", out)
	}
}

func TestLicensesSearch_JSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "search",
			"--borough", "bronx",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []LicenseSummary
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) == 0 {
		t.Errorf("expected at least one license in JSON output")
	}
	if results[0].PremisesName == "" {
		t.Errorf("expected non-empty premises_name in first result")
	}
}

func TestLicensesSearch_LicAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "lic", "search",
			"--borough", "bronx",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error via lic alias: %v", err)
		}
	})

	if !strings.Contains(out, "BRONX") {
		t.Errorf("expected output via lic alias, got: %s", out)
	}
}

func TestLicensesSearch_UnknownBorough(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "search",
			"--borough", "hoboken",
		})
		err := root.Execute()
		if err == nil {
			t.Error("expected error for unknown borough, got nil")
		}
	})
}

func TestLicensesSearch_UnknownType(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "search",
			"--borough", "bronx",
			"--type", "nightclub",
		})
		err := root.Execute()
		if err == nil {
			t.Error("expected error for unknown license type, got nil")
		}
	})
}

func TestLicensesSearch_EmptyResult(t *testing.T) {
	// Use a server that returns an empty array.
	mux := newEmptyMockServer(t)
	defer mux.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(mux)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "search",
			"--borough", "bronx",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No licenses found") {
		t.Errorf("expected empty-result message, got: %s", out)
	}
}

// --------------------------------------------------------------------------
// `licenses count` command integration tests
// --------------------------------------------------------------------------

func TestLicensesCount_TextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "count",
			"--borough", "bronx",
			"--since", "2024-01-01",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "New licenses:") {
		t.Errorf("expected 'New licenses:' in output, got: %s", out)
	}
	if !strings.Contains(out, "bronx") {
		t.Errorf("expected borough in output, got: %s", out)
	}
}

func TestLicensesCount_JSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "count",
			"--borough", "bronx",
			"--since", "2024-01-01",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result CountResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON object, got: %s, error: %v", out, err)
	}
	if result.Borough != "bronx" {
		t.Errorf("expected borough 'bronx', got %q", result.Borough)
	}
	if result.Since != "2024-01-01" {
		t.Errorf("expected since '2024-01-01', got %q", result.Since)
	}
	if result.NewLicenses != len(testLicenses) {
		t.Errorf("expected new_licenses %d (mock returns all), got %d", len(testLicenses), result.NewLicenses)
	}
	if len(result.Breakdown) == 0 {
		t.Errorf("expected non-empty breakdown")
	}
}

func TestLicensesCount_WithZIP(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "count",
			"--borough", "bronx",
			"--zip", "10451",
			"--since", "2024-01-01",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result CountResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
	}
	if result.ZIP != "10451" {
		t.Errorf("expected zip '10451', got %q", result.ZIP)
	}
}

func TestLicensesCount_MissingSince(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "count",
			"--borough", "bronx",
			// --since omitted
		})
		err := root.Execute()
		if err == nil {
			t.Error("expected error when --since is missing, got nil")
		}
	})
}

// --------------------------------------------------------------------------
// `licenses density` command integration tests
// --------------------------------------------------------------------------

func TestLicensesDensity_TextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "density",
			"--borough", "bronx",
			"--zip", "10451",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Total active:") {
		t.Errorf("expected 'Total active:' in output, got: %s", out)
	}
	if !strings.Contains(out, "10451") {
		t.Errorf("expected ZIP in output, got: %s", out)
	}
}

func TestLicensesDensity_JSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "density",
			"--borough", "bronx",
			"--zip", "10451",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result DensityResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON object, got: %s, error: %v", out, err)
	}
	if result.ZIP != "10451" {
		t.Errorf("expected zip '10451', got %q", result.ZIP)
	}
	if result.TotalActive != len(testLicenses) {
		t.Errorf("expected total_active %d (mock returns all), got %d", len(testLicenses), result.TotalActive)
	}
	if len(result.ByType) == 0 {
		t.Errorf("expected non-empty by_type breakdown")
	}
}

func TestLicensesDensity_MissingZIP(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "density",
			"--borough", "bronx",
			// --zip omitted
		})
		err := root.Execute()
		if err == nil {
			t.Error("expected error when --zip is missing, got nil")
		}
	})
}

// --------------------------------------------------------------------------
// Provider alias integration test
// --------------------------------------------------------------------------

func TestLiquorAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"liquor", "licenses", "search",
			"--borough", "bronx",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error via liquor alias: %v", err)
		}
	})

	if !strings.Contains(out, "BRONX") {
		t.Errorf("expected output via liquor alias, got: %s", out)
	}
}

// --------------------------------------------------------------------------
// HTTP error handling
// --------------------------------------------------------------------------

func TestLicensesSearch_HTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	captureStdout(t, func() {
		root.SetArgs([]string{
			"nysla", "licenses", "search",
			"--borough", "bronx",
		})
		err := root.Execute()
		if err == nil {
			t.Error("expected error on HTTP 500, got nil")
		}
	})
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// newEmptyMockServer returns a server that always responds with an empty JSON array.
func newEmptyMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	})
	return httptest.NewServer(mux)
}
