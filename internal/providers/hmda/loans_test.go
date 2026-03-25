package hmda

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLoansSummary(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"hmda", "loans", "summary", "--county=bronx"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Bronx") {
			t.Errorf("expected county name in output, got: %s", out)
		}
		if !strings.Contains(out, "36005") {
			t.Errorf("expected FIPS code in output, got: %s", out)
		}
		if !strings.Contains(out, "Total Loans") {
			t.Errorf("expected total loans header in output, got: %s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"hmda", "loans", "summary", "--county=bronx", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result CountySummary
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if result.County != "Bronx" {
			t.Errorf("expected county Bronx, got %q", result.County)
		}
		if result.FIPS != "36005" {
			t.Errorf("expected FIPS 36005, got %q", result.FIPS)
		}
		if result.Year != 2023 {
			t.Errorf("expected year 2023, got %d", result.Year)
		}
		if result.TotalLoans != 45 { // 15 + 8 + 22
			t.Errorf("expected 45 total loans, got %d", result.TotalLoans)
		}
		if result.TotalVolume != 21700000 { // 7500000 + 3200000 + 11000000
			t.Errorf("expected volume 21700000, got %d", result.TotalVolume)
		}
		if len(result.TopTracts) == 0 {
			t.Error("expected at least one tract in top_tracts")
		}
	})

	t.Run("brooklyn_county", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"hmda", "loans", "summary", "--county=brooklyn", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result CountySummary
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		if result.FIPS != "36047" {
			t.Errorf("expected FIPS 36047 for brooklyn, got %q", result.FIPS)
		}
	})

	t.Run("with_year_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"hmda", "loans", "summary", "--county=manhattan", "--year=2022", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result CountySummary
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		if result.Year != 2022 {
			t.Errorf("expected year 2022, got %d", result.Year)
		}
	})

	t.Run("with_purchase_purpose", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"hmda", "loans", "summary", "--county=queens", "--purpose=purchase", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result CountySummary
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		if result.FIPS != "36081" {
			t.Errorf("expected FIPS 36081 for queens, got %q", result.FIPS)
		}
	})

	t.Run("with_refinance_purpose", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"hmda", "loans", "summary", "--county=bronx", "--purpose=refinance", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result CountySummary
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		if result.County != "Bronx" {
			t.Errorf("expected county Bronx, got %q", result.County)
		}
	})

	t.Run("invalid_county", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"hmda", "loans", "summary", "--county=notacounty"})
		if err := root.Execute(); err == nil {
			t.Error("expected error for invalid county, got nil")
		}
	})

	t.Run("invalid_purpose", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"hmda", "loans", "summary", "--county=bronx", "--purpose=invalid"})
		if err := root.Execute(); err == nil {
			t.Error("expected error for invalid purpose, got nil")
		}
	})

	t.Run("missing_county_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"hmda", "loans", "summary"})
		if err := root.Execute(); err == nil {
			t.Error("expected error when --county is missing")
		}
	})
}

func TestLoansTract(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_output_found", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"hmda", "loans", "tract", "--tract=36005000100"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "36005000100") {
			t.Errorf("expected tract code in output, got: %s", out)
		}
		if !strings.Contains(out, "Total Loans") {
			t.Errorf("expected total loans label in output, got: %s", out)
		}
	})

	t.Run("json_output_found", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"hmda", "loans", "tract", "--tract=36005000100", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result TractSummary
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if result.Tract != "36005000100" {
			t.Errorf("expected tract 36005000100, got %q", result.Tract)
		}
		if result.Count != 15 {
			t.Errorf("expected count 15, got %d", result.Count)
		}
		if result.Volume != 7500000 {
			t.Errorf("expected volume 7500000, got %d", result.Volume)
		}
		if result.AvgLoanSize != 500000 { // 7500000 / 15
			t.Errorf("expected avg loan size 500000, got %d", result.AvgLoanSize)
		}
	})

	t.Run("json_output_not_found_returns_zeros", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			// Tract exists in county but not in mock data.
			root.SetArgs([]string{"hmda", "loans", "tract", "--tract=36005999999", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result TractSummary
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		if result.Tract != "36005999999" {
			t.Errorf("expected tract 36005999999, got %q", result.Tract)
		}
		if result.Count != 0 {
			t.Errorf("expected count 0 for unknown tract, got %d", result.Count)
		}
	})

	t.Run("with_year_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"hmda", "loans", "tract", "--tract=36005000200", "--year=2022", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result TractSummary
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		if result.Tract != "36005000200" {
			t.Errorf("expected tract 36005000200, got %q", result.Tract)
		}
	})

	t.Run("invalid_tract_length", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"hmda", "loans", "tract", "--tract=12345"})
		if err := root.Execute(); err == nil {
			t.Error("expected error for invalid tract length")
		}
	})

	t.Run("missing_tract_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"hmda", "loans", "tract"})
		if err := root.Execute(); err == nil {
			t.Error("expected error when --tract is missing")
		}
	})
}

func TestBuildCountySummary(t *testing.T) {
	items := []AggregationItem{
		{CensusTract: "36005000100", Count: 10, Sum: 5000000},
		{CensusTract: "36005000200", Count: 20, Sum: 8000000},
	}

	summary := buildCountySummary("Bronx", "36005", 2023, items)

	if summary.County != "Bronx" {
		t.Errorf("expected county Bronx, got %q", summary.County)
	}
	if summary.FIPS != "36005" {
		t.Errorf("expected FIPS 36005, got %q", summary.FIPS)
	}
	if summary.Year != 2023 {
		t.Errorf("expected year 2023, got %d", summary.Year)
	}
	if summary.TotalLoans != 30 {
		t.Errorf("expected 30 total loans, got %d", summary.TotalLoans)
	}
	if summary.TotalVolume != 13000000 {
		t.Errorf("expected volume 13000000, got %d", summary.TotalVolume)
	}
	if summary.AvgLoanSize != 433333 { // 13000000 / 30
		t.Errorf("expected avg loan size 433333, got %d", summary.AvgLoanSize)
	}
	if len(summary.TopTracts) != 2 {
		t.Errorf("expected 2 top tracts, got %d", len(summary.TopTracts))
	}
	// Top tract should be sorted by count descending: 20 first.
	if summary.TopTracts[0].Count != 20 {
		t.Errorf("expected top tract to have count 20, got %d", summary.TopTracts[0].Count)
	}
}

func TestBuildCountySummary_Empty(t *testing.T) {
	summary := buildCountySummary("Manhattan", "36061", 2023, nil)
	if summary.TotalLoans != 0 {
		t.Errorf("expected 0 loans for empty items, got %d", summary.TotalLoans)
	}
	if summary.AvgLoanSize != 0 {
		t.Errorf("expected 0 avg loan size for empty items, got %d", summary.AvgLoanSize)
	}
	if len(summary.TopTracts) != 0 {
		t.Errorf("expected 0 top tracts for empty items, got %d", len(summary.TopTracts))
	}
}

func TestBuildCountySummary_Top10Cap(t *testing.T) {
	// Create 15 tracts to verify the cap at 10.
	items := make([]AggregationItem, 15)
	for i := range items {
		items[i] = AggregationItem{
			CensusTract: "36005" + strings.Repeat("0", 6-len(string(rune('0'+i)))) + string(rune('0'+i)),
			Count:       i + 1,
			Sum:         int64((i + 1) * 100000),
		}
	}

	summary := buildCountySummary("Bronx", "36005", 2023, items)
	if len(summary.TopTracts) != 10 {
		t.Errorf("expected 10 top tracts (cap), got %d", len(summary.TopTracts))
	}
}

func TestLookupFIPS(t *testing.T) {
	tests := []struct {
		county  string
		want    string
		wantErr bool
	}{
		{"bronx", "36005", false},
		{"brooklyn", "36047", false},
		{"manhattan", "36061", false},
		{"queens", "36081", false},
		{"staten-island", "36085", false},
		{"BRONX", "36005", false},   // case-insensitive
		{"invalid", "", true},
	}
	for _, tt := range tests {
		got, err := lookupFIPS(tt.county)
		if tt.wantErr {
			if err == nil {
				t.Errorf("lookupFIPS(%q) expected error, got nil", tt.county)
			}
			continue
		}
		if err != nil {
			t.Errorf("lookupFIPS(%q) unexpected error: %v", tt.county, err)
			continue
		}
		if got != tt.want {
			t.Errorf("lookupFIPS(%q) = %q, want %q", tt.county, got, tt.want)
		}
	}
}

func TestFormatDollars(t *testing.T) {
	tests := []struct {
		amount int64
		want   string
	}{
		{0, "$0"},
		{100, "$100"},
		{1000, "$1,000"},
		{1000000, "$1,000,000"},
		{7500000, "$7,500,000"},
		{21700000, "$21,700,000"},
	}
	for _, tt := range tests {
		got := formatDollars(tt.amount)
		if got != tt.want {
			t.Errorf("formatDollars(%d) = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

func TestPurposeCode(t *testing.T) {
	tests := []struct {
		purpose string
		want    string
		wantErr bool
	}{
		{"purchase", "1", false},
		{"refinance", "31", false},
		{"invalid", "", true},
	}
	for _, tt := range tests {
		got, err := purposeCode(tt.purpose)
		if tt.wantErr {
			if err == nil {
				t.Errorf("purposeCode(%q) expected error", tt.purpose)
			}
			continue
		}
		if err != nil {
			t.Errorf("purposeCode(%q) unexpected error: %v", tt.purpose, err)
			continue
		}
		if got != tt.want {
			t.Errorf("purposeCode(%q) = %q, want %q", tt.purpose, got, tt.want)
		}
	}
}

func TestCountyNameForFIPS(t *testing.T) {
	tests := []struct {
		fips string
		want string
	}{
		{"36005", "Bronx"},
		{"36047", "Brooklyn"},
		{"36061", "Manhattan"},
		{"36081", "Queens"},
		{"36085", "Staten Island"},
		{"99999", "99999"}, // unknown FIPS returns the code itself
	}
	for _, tt := range tests {
		got := countyNameForFIPS(tt.fips)
		if got != tt.want {
			t.Errorf("countyNameForFIPS(%q) = %q, want %q", tt.fips, got, tt.want)
		}
	}
}
