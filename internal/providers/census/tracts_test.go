package census

import (
	"encoding/json"
	"strings"
	"testing"
)

// --------------------------------------------------------------------------
// Helper unit tests
// --------------------------------------------------------------------------

func TestParseTractFIPS_Valid(t *testing.T) {
	state, county, tract, err := parseTractFIPS("36005000100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != "36" {
		t.Errorf("state = %q, want %q", state, "36")
	}
	if county != "005" {
		t.Errorf("county = %q, want %q", county, "005")
	}
	if tract != "000100" {
		t.Errorf("tract = %q, want %q", tract, "000100")
	}
}

func TestParseTractFIPS_InvalidLength(t *testing.T) {
	_, _, _, err := parseTractFIPS("3600500")
	if err == nil {
		t.Error("expected error for short FIPS, got nil")
	}
}

func TestLookupBoroughFIPS(t *testing.T) {
	tests := []struct {
		borough string
		want    string
		wantErr bool
	}{
		{"bronx", "005", false},
		{"brooklyn", "047", false},
		{"manhattan", "061", false},
		{"queens", "081", false},
		{"staten-island", "085", false},
		{"BRONX", "005", false},   // case-insensitive
		{"invalid", "", true},
	}
	for _, tt := range tests {
		got, err := lookupBoroughFIPS(tt.borough)
		if tt.wantErr {
			if err == nil {
				t.Errorf("lookupBoroughFIPS(%q) expected error, got nil", tt.borough)
			}
			continue
		}
		if err != nil {
			t.Errorf("lookupBoroughFIPS(%q) unexpected error: %v", tt.borough, err)
			continue
		}
		if got != tt.want {
			t.Errorf("lookupBoroughFIPS(%q) = %q, want %q", tt.borough, got, tt.want)
		}
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"4521", 4521},
		{"0", 0},
		{"-", 0},
		{"", 0},
		{"52300", 52300},
	}
	for _, tt := range tests {
		got := parseInt(tt.input)
		if got != tt.want {
			t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestRowsToMaps(t *testing.T) {
	data := [][]string{
		{"NAME", "B01001_001E", "state"},
		{"Census Tract 1", "4521", "36"},
		{"Census Tract 2", "3200", "36"},
	}
	maps := rowsToMaps(data)
	if len(maps) != 2 {
		t.Fatalf("expected 2 maps, got %d", len(maps))
	}
	if maps[0]["NAME"] != "Census Tract 1" {
		t.Errorf("expected NAME 'Census Tract 1', got %q", maps[0]["NAME"])
	}
	if maps[1]["B01001_001E"] != "3200" {
		t.Errorf("expected B01001_001E '3200', got %q", maps[1]["B01001_001E"])
	}
}

func TestRowsToMaps_Empty(t *testing.T) {
	result := rowsToMaps([][]string{{"NAME"}}) // header only
	if result != nil {
		t.Errorf("expected nil for header-only data, got %v", result)
	}
}

func TestRowsToMaps_Nil(t *testing.T) {
	result := rowsToMaps(nil)
	if result != nil {
		t.Errorf("expected nil for nil data, got %v", result)
	}
}

func TestComputeVacancyRate(t *testing.T) {
	tests := []struct {
		total  int
		vacant int
		want   float64
	}{
		{2100, 210, 10.0},
		{0, 0, 0.0},
		{1000, 50, 5.0},
	}
	for _, tt := range tests {
		got := computeVacancyRate(tt.total, tt.vacant)
		if got != tt.want {
			t.Errorf("computeVacancyRate(%d, %d) = %.2f, want %.2f", tt.total, tt.vacant, got, tt.want)
		}
	}
}

func TestComputeOwnerRenterPct(t *testing.T) {
	owner := computeOwnerPct(480, 1380)
	renter := computeRenterPct(480, 1380)

	if owner+renter < 99.9 || owner+renter > 100.1 {
		t.Errorf("owner %.2f + renter %.2f should sum to ~100", owner, renter)
	}

	if computeOwnerPct(0, 0) != 0 {
		t.Error("expected 0 for zero occupied units")
	}
}

func TestFormatDollars(t *testing.T) {
	tests := []struct {
		amount int
		want   string
	}{
		{0, "$0"},
		{100, "$100"},
		{1000, "$1,000"},
		{52300, "$52,300"},
		{385000, "$385,000"},
		{1234567, "$1,234,567"},
	}
	for _, tt := range tests {
		got := formatDollars(tt.amount)
		if got != tt.want {
			t.Errorf("formatDollars(%d) = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

func TestIsValidSortField(t *testing.T) {
	valid := []string{"income", "rent", "vacancy", "population"}
	for _, f := range valid {
		if !isValidSortField(f) {
			t.Errorf("isValidSortField(%q) should be true", f)
		}
	}
	if isValidSortField("invalid") {
		t.Error("isValidSortField('invalid') should be false")
	}
}

func TestSortProfiles(t *testing.T) {
	profiles := []TractProfile{
		{Tract: "A", MedianIncome: 30000, MedianRent: 900, VacancyRate: 5.0, Population: 1000},
		{Tract: "B", MedianIncome: 60000, MedianRent: 1800, VacancyRate: 12.0, Population: 5000},
		{Tract: "C", MedianIncome: 45000, MedianRent: 1200, VacancyRate: 8.0, Population: 3000},
	}

	// Sort by income descending.
	sortProfiles(profiles, "income")
	if profiles[0].Tract != "B" {
		t.Errorf("sort by income: expected B first, got %q", profiles[0].Tract)
	}

	// Sort by rent descending.
	sortProfiles(profiles, "rent")
	if profiles[0].Tract != "B" {
		t.Errorf("sort by rent: expected B first, got %q", profiles[0].Tract)
	}

	// Sort by vacancy descending.
	sortProfiles(profiles, "vacancy")
	if profiles[0].Tract != "B" {
		t.Errorf("sort by vacancy: expected B first, got %q", profiles[0].Tract)
	}

	// Sort by population descending.
	sortProfiles(profiles, "population")
	if profiles[0].Tract != "B" {
		t.Errorf("sort by population: expected B first, got %q", profiles[0].Tract)
	}
}

func TestBuildBoroughSummary(t *testing.T) {
	profiles := []TractProfile{
		{Population: 4000, MedianIncome: 50000, MedianRent: 1400, VacancyRate: 5.0},
		{Population: 6000, MedianIncome: 70000, MedianRent: 1800, VacancyRate: 12.0},
		{Population: 3000, MedianIncome: 30000, MedianRent: 1000, VacancyRate: 3.0},
	}

	s := buildBoroughSummary("bronx", profiles)

	if s.Borough != "bronx" {
		t.Errorf("expected borough bronx, got %q", s.Borough)
	}
	if s.TotalPopulation != 13000 {
		t.Errorf("expected total population 13000, got %d", s.TotalPopulation)
	}
	if s.TotalTracts != 3 {
		t.Errorf("expected 3 tracts, got %d", s.TotalTracts)
	}
	if s.HighVacancyTracts != 1 { // only 12% > 10%
		t.Errorf("expected 1 high vacancy tract, got %d", s.HighVacancyTracts)
	}
	expectedAvgIncome := (50000 + 70000 + 30000) / 3
	if s.AvgMedianIncome != expectedAvgIncome {
		t.Errorf("expected avg income %d, got %d", expectedAvgIncome, s.AvgMedianIncome)
	}
}

func TestBuildBoroughSummary_Empty(t *testing.T) {
	s := buildBoroughSummary("manhattan", nil)
	if s.TotalPopulation != 0 {
		t.Errorf("expected 0 population for empty profiles, got %d", s.TotalPopulation)
	}
	if s.TotalTracts != 0 {
		t.Errorf("expected 0 tracts for empty profiles, got %d", s.TotalTracts)
	}
}

func TestBuildYearURL(t *testing.T) {
	base := "https://api.census.gov/data/2023/acs/acs5"
	if got := buildYearURL(base, 2023); got != base {
		t.Errorf("buildYearURL for 2023 should be unchanged, got %q", got)
	}
	if got := buildYearURL(base, 2022); !strings.Contains(got, "2022") {
		t.Errorf("buildYearURL for 2022 should contain 2022, got %q", got)
	}
}

// --------------------------------------------------------------------------
// `tracts profile` command tests (via mock server)
// --------------------------------------------------------------------------

func TestTractsProfile_TextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "profile", "--tract=36005000100"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "36005000100") {
		t.Errorf("expected tract FIPS in output, got: %s", out)
	}
	if !strings.Contains(out, "Population") {
		t.Errorf("expected Population label in output, got: %s", out)
	}
	if !strings.Contains(out, "Median Income") {
		t.Errorf("expected Median Income label in output, got: %s", out)
	}
}

func TestTractsProfile_JSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "profile", "--tract=36005000100", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result TractProfile
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
	}
	if result.Tract != "36005000100" {
		t.Errorf("expected tract 36005000100, got %q", result.Tract)
	}
	if result.Population != 4521 {
		t.Errorf("expected population 4521, got %d", result.Population)
	}
	if result.MedianIncome != 52300 {
		t.Errorf("expected median_income 52300, got %d", result.MedianIncome)
	}
	if result.MedianRent != 1450 {
		t.Errorf("expected median_rent 1450, got %d", result.MedianRent)
	}
	// median_home_value for tract 000100 is 385000
	if result.MedianHomeValue != 385000 {
		t.Errorf("expected median_home_value 385000, got %d", result.MedianHomeValue)
	}
	if result.VacancyRate <= 0 {
		t.Errorf("expected positive vacancy_rate, got %.2f", result.VacancyRate)
	}
}

func TestTractsProfile_SuppressedHomeValue(t *testing.T) {
	// Tract 000200 has median_home_value = "-" (suppressed).
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "profile", "--tract=36005000200", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result TractProfile
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
	}
	if result.MedianHomeValue != 0 {
		t.Errorf("expected median_home_value 0 for suppressed value, got %d", result.MedianHomeValue)
	}
}

func TestTractsProfile_InvalidTractFIPS(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"census", "tracts", "profile", "--tract=12345"})
	if err := root.Execute(); err == nil {
		t.Error("expected error for invalid tract FIPS length, got nil")
	}
}

func TestTractsProfile_MissingTractFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"census", "tracts", "profile"})
	if err := root.Execute(); err == nil {
		t.Error("expected error when --tract is missing")
	}
}

func TestTractsProfile_WithYearFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "profile", "--tract=36005000100", "--year=2022", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result TractProfile
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got: %s", out)
	}
	if result.Tract != "36005000100" {
		t.Errorf("expected tract 36005000100, got %q", result.Tract)
	}
}

// --------------------------------------------------------------------------
// `tracts compare` command tests (via mock server)
// --------------------------------------------------------------------------

func TestTractsCompare_TextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "compare", "--borough=bronx"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "TRACT") {
		t.Errorf("expected TRACT column header in output, got: %s", out)
	}
	if !strings.Contains(out, "36005") {
		t.Errorf("expected tract FIPS prefix in output, got: %s", out)
	}
}

func TestTractsCompare_JSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "compare", "--borough=bronx", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []TractProfile
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) == 0 {
		t.Error("expected at least one tract in compare results")
	}
	// Default sort is by income descending — first tract should have highest income.
	if len(results) >= 2 && results[0].MedianIncome < results[1].MedianIncome {
		t.Errorf("expected results sorted by income descending: %d < %d",
			results[0].MedianIncome, results[1].MedianIncome)
	}
}

func TestTractsCompare_SortByRent(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "compare", "--borough=bronx", "--sort=rent", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []TractProfile
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) >= 2 && results[0].MedianRent < results[1].MedianRent {
		t.Errorf("expected results sorted by rent descending: %d < %d",
			results[0].MedianRent, results[1].MedianRent)
	}
}

func TestTractsCompare_SortByVacancy(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "compare", "--borough=bronx", "--sort=vacancy", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []TractProfile
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) >= 2 && results[0].VacancyRate < results[1].VacancyRate {
		t.Errorf("expected results sorted by vacancy descending: %.2f < %.2f",
			results[0].VacancyRate, results[1].VacancyRate)
	}
}

func TestTractsCompare_SortByPopulation(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "compare", "--borough=bronx", "--sort=population", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []TractProfile
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) >= 2 && results[0].Population < results[1].Population {
		t.Errorf("expected results sorted by population descending: %d < %d",
			results[0].Population, results[1].Population)
	}
}

func TestTractsCompare_LimitFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "compare", "--borough=bronx", "--limit=2", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []TractProfile
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) > 2 {
		t.Errorf("expected at most 2 results with --limit=2, got %d", len(results))
	}
}

func TestTractsCompare_InvalidBorough(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"census", "tracts", "compare", "--borough=notaborough"})
	if err := root.Execute(); err == nil {
		t.Error("expected error for invalid borough, got nil")
	}
}

func TestTractsCompare_InvalidSortField(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"census", "tracts", "compare", "--borough=bronx", "--sort=invalid"})
	if err := root.Execute(); err == nil {
		t.Error("expected error for invalid sort field, got nil")
	}
}

func TestTractsCompare_MissingBoroughFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"census", "tracts", "compare"})
	if err := root.Execute(); err == nil {
		t.Error("expected error when --borough is missing")
	}
}

// --------------------------------------------------------------------------
// `tracts summary` command tests (via mock server)
// --------------------------------------------------------------------------

func TestTractsSummary_TextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "summary", "--borough=bronx"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Borough") {
		t.Errorf("expected Borough label in output, got: %s", out)
	}
	if !strings.Contains(out, "Total Tracts") {
		t.Errorf("expected Total Tracts label in output, got: %s", out)
	}
}

func TestTractsSummary_JSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "summary", "--borough=bronx", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result BoroughSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
	}
	if result.Borough != "bronx" {
		t.Errorf("expected borough bronx, got %q", result.Borough)
	}
	if result.TotalTracts != len(testTractRows) {
		t.Errorf("expected %d total tracts, got %d", len(testTractRows), result.TotalTracts)
	}
	if result.TotalPopulation <= 0 {
		t.Errorf("expected positive total population, got %d", result.TotalPopulation)
	}
}

func TestTractsSummary_HighVacancyCount(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"census", "tracts", "summary", "--borough=bronx", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result BoroughSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got: %s", out)
	}
	// Tract 2: 200/1600 = 12.5% > 10% — should be high vacancy.
	// Tract 4: 210/1400 = 15% > 10% — should be high vacancy.
	if result.HighVacancyTracts < 1 {
		t.Errorf("expected at least 1 high vacancy tract, got %d", result.HighVacancyTracts)
	}
}

func TestTractsSummary_AllBoroughs(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	boroughs := []string{"bronx", "brooklyn", "manhattan", "queens", "staten-island"}
	for _, borough := range boroughs {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)

		out := captureStdout(t, func() {
			root.SetArgs([]string{"census", "tracts", "summary", "--borough=" + borough, "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("borough %q: unexpected error: %v", borough, err)
			}
		})

		var result BoroughSummary
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("borough %q: expected valid JSON, got: %s", borough, out)
		}
		if result.Borough != borough {
			t.Errorf("borough %q: expected borough %q in result, got %q", borough, borough, result.Borough)
		}
	}
}

func TestTractsSummary_InvalidBorough(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"census", "tracts", "summary", "--borough=notaborough"})
	if err := root.Execute(); err == nil {
		t.Error("expected error for invalid borough, got nil")
	}
}

func TestTractsSummary_MissingBoroughFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"census", "tracts", "summary"})
	if err := root.Execute(); err == nil {
		t.Error("expected error when --borough is missing")
	}
}

// --------------------------------------------------------------------------
// `tracts` alias tests
// --------------------------------------------------------------------------

func TestTractsAliasCmd(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	// Use the "acs" alias for the census command and "tract" alias for tracts.
	out := captureStdout(t, func() {
		root.SetArgs([]string{"acs", "tract", "summary", "--borough=bronx", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error via acs alias: %v", err)
		}
	})

	var result BoroughSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON via alias, got: %s", out)
	}
}
