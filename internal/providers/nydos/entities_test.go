package nydos

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEntitiesRecent(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "recent", "--since=2026-03-01"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "SEMINOLE") {
			t.Errorf("expected entity name in output, got: %s", out)
		}
		if !strings.Contains(out, "DOS ID") {
			t.Errorf("expected header in output, got: %s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "recent", "--since=2026-03-01", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var entities []EntitySummary
		if err := json.Unmarshal([]byte(out), &entities); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if len(entities) == 0 {
			t.Fatal("expected at least one entity")
		}
		// Verify key fields are populated.
		e := entities[0]
		if e.DOSID == "" {
			t.Error("expected dos_id to be populated")
		}
		if e.Name == "" {
			t.Error("expected name to be populated")
		}
		if e.FilingDate == "" {
			t.Error("expected filing_date to be populated")
		}
		// Date should be truncated to YYYY-MM-DD.
		if len(e.FilingDate) != 10 {
			t.Errorf("expected filing_date to be 10 chars (YYYY-MM-DD), got %q", e.FilingDate)
		}
	})

	t.Run("with_type_filter_llc", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "recent", "--since=2026-03-01", "--type=llc", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var entities []EntitySummary
		if err := json.Unmarshal([]byte(out), &entities); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		// Mock returns all records; we just verify it doesn't error with valid type.
		_ = entities
	})

	t.Run("with_county_filter", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "recent", "--since=2026-03-01", "--county=BRONX", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var entities []EntitySummary
		if err := json.Unmarshal([]byte(out), &entities); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		_ = entities
	})

	t.Run("with_limit_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "recent", "--since=2026-03-01", "--limit=10", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var entities []EntitySummary
		if err := json.Unmarshal([]byte(out), &entities); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		_ = entities
	})

	t.Run("invalid_type", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"nydos", "entities", "recent", "--since=2026-03-01", "--type=invalid"})
		if err := root.Execute(); err == nil {
			t.Error("expected error for invalid entity type, got nil")
		}
	})

	t.Run("missing_since_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"nydos", "entities", "recent"})
		if err := root.Execute(); err == nil {
			t.Error("expected error when --since is missing")
		}
	})
}

func TestEntitiesSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "search", "--name=SEMINOLE"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "SEMINOLE") {
			t.Errorf("expected entity name in output, got: %s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "search", "--name=SEMINOLE", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var entities []EntitySummary
		if err := json.Unmarshal([]byte(out), &entities); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if len(entities) == 0 {
			t.Fatal("expected at least one entity")
		}
		e := entities[0]
		if e.DOSID == "" {
			t.Error("expected dos_id to be populated")
		}
		if e.County == "" {
			t.Error("expected county to be populated for active corp")
		}
	})

	t.Run("with_type_filter", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "search", "--name=SEMINOLE", "--type=lp", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var entities []EntitySummary
		if err := json.Unmarshal([]byte(out), &entities); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		_ = entities
	})

	t.Run("with_limit_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "search", "--name=ACME", "--limit=10", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var entities []EntitySummary
		if err := json.Unmarshal([]byte(out), &entities); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		_ = entities
	})

	t.Run("invalid_type", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"nydos", "entities", "search", "--name=TEST", "--type=badtype"})
		if err := root.Execute(); err == nil {
			t.Error("expected error for invalid entity type, got nil")
		}
	})

	t.Run("missing_name_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"nydos", "entities", "search"})
		if err := root.Execute(); err == nil {
			t.Error("expected error when --name is missing")
		}
	})
}

func TestEntitiesMatchAddress(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "match-address", "--address=1776 SEMINOLE"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "SEMINOLE") {
			t.Errorf("expected entity name in output, got: %s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "match-address", "--address=1776 SEMINOLE AVE", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var entities []EntitySummary
		if err := json.Unmarshal([]byte(out), &entities); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if len(entities) == 0 {
			t.Fatal("expected at least one entity")
		}
	})

	t.Run("with_since_filter", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "match-address", "--address=SEMINOLE", "--since=2026-01-01", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var entities []EntitySummary
		if err := json.Unmarshal([]byte(out), &entities); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		_ = entities
	})

	t.Run("with_limit_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nydos", "entities", "match-address", "--address=1776 SEMINOLE", "--limit=10", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var entities []EntitySummary
		if err := json.Unmarshal([]byte(out), &entities); err != nil {
			t.Fatalf("expected valid JSON, got: %s", out)
		}
		_ = entities
	})

	t.Run("stop_words_only_address_errors", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"nydos", "entities", "match-address", "--address=AVE ST RD"})
		if err := root.Execute(); err == nil {
			t.Error("expected error when address produces no tokens after stop word filtering")
		}
	})

	t.Run("missing_address_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"nydos", "entities", "match-address"})
		if err := root.Execute(); err == nil {
			t.Error("expected error when --address is missing")
		}
	})
}

func TestAddressTokens(t *testing.T) {
	tests := []struct {
		address string
		want    []string
	}{
		{
			address: "1776 SEMINOLE AVE",
			want:    []string{"1776", "SEMINOLE"},
		},
		{
			address: "100 MAIN STREET",
			want:    []string{"100", "MAIN"},
		},
		{
			address: "55 Water St",
			want:    []string{"55", "WATER"},
		},
		{
			address: "AVE ST RD",
			want:    nil,
		},
		{
			address: "200 PARK BLVD",
			want:    []string{"200", "PARK"},
		},
	}
	for _, tt := range tests {
		got := addressTokens(tt.address)
		if len(got) != len(tt.want) {
			t.Errorf("addressTokens(%q) = %v, want %v", tt.address, got, tt.want)
			continue
		}
		for i, tok := range got {
			if tok != tt.want[i] {
				t.Errorf("addressTokens(%q)[%d] = %q, want %q", tt.address, i, tok, tt.want[i])
			}
		}
	}
}

func TestBuildAddress(t *testing.T) {
	tests := []struct {
		addr1, city, state, zip string
		want                    string
	}{
		{"1776 SEMINOLE AVE", "BRONX", "NY", "10462", "1776 SEMINOLE AVE, BRONX, NY, 10462"},
		{"100 MAIN ST", "NEW YORK", "NY", "", "100 MAIN ST, NEW YORK, NY"},
		{"", "NEW YORK", "NY", "10001", "NEW YORK, NY, 10001"},
		{"", "", "", "", ""},
	}
	for _, tt := range tests {
		got := buildAddress(tt.addr1, tt.city, tt.state, tt.zip)
		if got != tt.want {
			t.Errorf("buildAddress(%q, %q, %q, %q) = %q, want %q",
				tt.addr1, tt.city, tt.state, tt.zip, got, tt.want)
		}
	}
}

func TestTruncateDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2026-03-10T00:00:00.000", "2026-03-10"},
		{"2026-03-10", "2026-03-10"},
		{"2026", "2026"},
		{"", ""},
	}
	for _, tt := range tests {
		got := truncateDate(tt.input)
		if got != tt.want {
			t.Errorf("truncateDate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLookupEntityType(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"llc", "DOMESTIC LIMITED LIABILITY COMPANY", false},
		{"corp", "DOMESTIC BUSINESS CORPORATION", false},
		{"lp", "DOMESTIC LIMITED PARTNERSHIP", false},
		{"LLC", "DOMESTIC LIMITED LIABILITY COMPANY", false}, // case-insensitive
		{"invalid", "", true},
	}
	for _, tt := range tests {
		got, err := lookupEntityType(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("lookupEntityType(%q) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("lookupEntityType(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("lookupEntityType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDailyFilingToSummary(t *testing.T) {
	r := DailyFilingRecord{
		DOSID:      "9876543",
		CorpName:   "TEST LLC",
		FilingDate: "2026-03-10T00:00:00.000",
		EntityType: "DOMESTIC LIMITED LIABILITY COMPANY",
		SOPName:    "PROCESS AGENT",
		SOPAddr1:   "10 TEST ST",
		SOPCity:    "ALBANY",
		SOPState:   "NY",
		SOPZip5:    "12207",
		FilerName:  "FILER NAME",
		FilerAddr1: "20 FILER RD",
		FilerCity:  "BUFFALO",
		FilerState: "NY",
	}

	s := dailyFilingToSummary(r)

	if s.DOSID != "9876543" {
		t.Errorf("DOSID = %q, want %q", s.DOSID, "9876543")
	}
	if s.Name != "TEST LLC" {
		t.Errorf("Name = %q, want %q", s.Name, "TEST LLC")
	}
	if s.FilingDate != "2026-03-10" {
		t.Errorf("FilingDate = %q, want %q", s.FilingDate, "2026-03-10")
	}
	if s.EntityType != "DOMESTIC LIMITED LIABILITY COMPANY" {
		t.Errorf("EntityType = %q", s.EntityType)
	}
	if s.ProcessName != "PROCESS AGENT" {
		t.Errorf("ProcessName = %q, want %q", s.ProcessName, "PROCESS AGENT")
	}
	if !strings.Contains(s.ProcessAddress, "10 TEST ST") {
		t.Errorf("ProcessAddress = %q, expected to contain address", s.ProcessAddress)
	}
	if !strings.Contains(s.ProcessAddress, "12207") {
		t.Errorf("ProcessAddress = %q, expected to contain zip", s.ProcessAddress)
	}
	if s.FilerName != "FILER NAME" {
		t.Errorf("FilerName = %q, want %q", s.FilerName, "FILER NAME")
	}
	if !strings.Contains(s.FilerAddress, "20 FILER RD") {
		t.Errorf("FilerAddress = %q, expected to contain address", s.FilerAddress)
	}
}

func TestActiveCorpToSummary(t *testing.T) {
	r := ActiveCorpRecord{
		DOSID:                "1111111",
		CurrentEntityName:    "ACTIVE CORP LLC",
		InitialDOSFilingDate: "2019-07-04T00:00:00.000",
		County:               "QUEENS",
		EntityTypeCode:       "DOMESTIC LIMITED LIABILITY COMPANY",
		DOSProcessName:       "PROCESS CO",
		DOSProcessAddr1:      "500 QUEENS BLVD",
		DOSProcessCity:       "QUEENS",
		DOSProcessState:      "NY",
		DOSProcessZip:        "11354",
	}

	s := activeCorpToSummary(r)

	if s.DOSID != "1111111" {
		t.Errorf("DOSID = %q, want %q", s.DOSID, "1111111")
	}
	if s.Name != "ACTIVE CORP LLC" {
		t.Errorf("Name = %q, want %q", s.Name, "ACTIVE CORP LLC")
	}
	if s.FilingDate != "2019-07-04" {
		t.Errorf("FilingDate = %q, want %q", s.FilingDate, "2019-07-04")
	}
	if s.County != "QUEENS" {
		t.Errorf("County = %q, want %q", s.County, "QUEENS")
	}
	if s.ProcessName != "PROCESS CO" {
		t.Errorf("ProcessName = %q, want %q", s.ProcessName, "PROCESS CO")
	}
	if !strings.Contains(s.ProcessAddress, "500 QUEENS BLVD") {
		t.Errorf("ProcessAddress = %q, expected to contain address", s.ProcessAddress)
	}
	if s.FilerName != "" {
		t.Errorf("FilerName should be empty for active corp records, got %q", s.FilerName)
	}
}
