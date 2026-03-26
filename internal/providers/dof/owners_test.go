package dof

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOwnersLookup(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "lookup", "--bbl=2029640028"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "2029640028") {
			t.Errorf("expected BBL in output, got: %s", out)
		}
		if !strings.Contains(out, "ACME REALTY LLC") {
			t.Errorf("expected owner name in output, got: %s", out)
		}
		if !strings.Contains(out, "Owner") {
			t.Errorf("expected Owner label in output, got: %s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "lookup", "--bbl=2029640028", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result OwnerRecord
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if result.BBL != "2029640028" {
			t.Errorf("expected BBL 2029640028, got %q", result.BBL)
		}
		if result.OwnerName != "ACME REALTY LLC" {
			t.Errorf("expected owner ACME REALTY LLC, got %q", result.OwnerName)
		}
		if result.TaxClass != "4" {
			t.Errorf("expected tax class 4, got %q", result.TaxClass)
		}
		if result.AssessedValue != "1200000" {
			t.Errorf("expected assessed value 1200000, got %q", result.AssessedValue)
		}
		if result.Borough != "2" {
			t.Errorf("expected borough 2, got %q", result.Borough)
		}
	})

	t.Run("not_found_returns_error", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"dof", "owners", "lookup", "--bbl=9999999999"})
		if err := root.Execute(); err == nil {
			t.Error("expected error for unknown BBL, got nil")
		}
	})

	t.Run("missing_bbl_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"dof", "owners", "lookup"})
		if err := root.Execute(); err == nil {
			t.Error("expected error when --bbl is missing")
		}
	})
}

func TestOwnersSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "search", "--name=LLC"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "LLC") {
			t.Errorf("expected LLC in output, got: %s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "search", "--name=LLC", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []OwnerRecord
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Error("expected at least one result for LLC search")
		}
		for _, r := range results {
			if !strings.Contains(r.OwnerName, "LLC") {
				t.Errorf("expected LLC in owner name, got %q", r.OwnerName)
			}
		}
	})

	t.Run("with_borough_filter", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "search", "--name=LLC", "--borough=bronx", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []OwnerRecord
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s", out)
		}
		// Results may be empty or filtered — just verify no error and valid JSON.
	})

	t.Run("with_limit_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "search", "--name=LLC", "--limit=10", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []OwnerRecord
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s", out)
		}
	})

	t.Run("invalid_borough", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"dof", "owners", "search", "--name=LLC", "--borough=notaborough"})
		if err := root.Execute(); err == nil {
			t.Error("expected error for invalid borough")
		}
	})

	t.Run("missing_name_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"dof", "owners", "search"})
		if err := root.Execute(); err == nil {
			t.Error("expected error when --name is missing")
		}
	})

	t.Run("no_results_text_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "search", "--name=XYZUNKNOWN"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "No results") {
			t.Errorf("expected 'No results' message, got: %s", out)
		}
	})
}

func TestOwnersByEntity(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "by-entity", "--pattern=LLC"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "LLC") {
			t.Errorf("expected LLC in output, got: %s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "by-entity", "--pattern=LLC", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []OwnerRecord
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Error("expected at least one result for LLC pattern")
		}
	})

	t.Run("with_borough_filter", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "by-entity", "--pattern=LLC", "--borough=brooklyn", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []OwnerRecord
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s", out)
		}
	})

	t.Run("with_limit_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"dof", "owners", "by-entity", "--pattern=ACME", "--limit=5", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []OwnerRecord
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s", out)
		}
	})

	t.Run("invalid_borough", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"dof", "owners", "by-entity", "--pattern=LLC", "--borough=invalid"})
		if err := root.Execute(); err == nil {
			t.Error("expected error for invalid borough")
		}
	})

	t.Run("missing_pattern_flag", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		root.SetArgs([]string{"dof", "owners", "by-entity"})
		if err := root.Execute(); err == nil {
			t.Error("expected error when --pattern is missing")
		}
	})
}

func TestToOwnerRecord(t *testing.T) {
	raw := rawOwnerRecord{
		BBLE:          "2029640028",
		OwnerName:     "ACME LLC",
		TaxClass:      "4",
		AssessedValue: "500000",
		Borough:       "2",
		Block:         "02964",
		Lot:           "0028",
		Address:       "123 MAIN ST",
	}
	rec := toOwnerRecord(raw)
	if rec.BBL != "2029640028" {
		t.Errorf("BBL = %q, want %q", rec.BBL, "2029640028")
	}
	if rec.OwnerName != "ACME LLC" {
		t.Errorf("OwnerName = %q, want %q", rec.OwnerName, "ACME LLC")
	}
	if rec.TaxClass != "4" {
		t.Errorf("TaxClass = %q, want %q", rec.TaxClass, "4")
	}
	if rec.Borough != "2" {
		t.Errorf("Borough = %q, want %q", rec.Borough, "2")
	}
}

func TestToOwnerRecords(t *testing.T) {
	raw := defaultOwnerRecords()
	recs := toOwnerRecords(raw)
	if len(recs) != len(raw) {
		t.Errorf("len = %d, want %d", len(recs), len(raw))
	}
}

func TestLookupBoroughCode(t *testing.T) {
	tests := []struct {
		borough string
		want    string
		wantErr bool
	}{
		{"manhattan", "1", false},
		{"bronx", "2", false},
		{"brooklyn", "3", false},
		{"queens", "4", false},
		{"staten-island", "5", false},
		{"MANHATTAN", "1", false}, // case-insensitive
		{"invalid", "", true},
	}
	for _, tt := range tests {
		got, err := lookupBoroughCode(tt.borough)
		if tt.wantErr {
			if err == nil {
				t.Errorf("lookupBoroughCode(%q) expected error, got nil", tt.borough)
			}
			continue
		}
		if err != nil {
			t.Errorf("lookupBoroughCode(%q) unexpected error: %v", tt.borough, err)
			continue
		}
		if got != tt.want {
			t.Errorf("lookupBoroughCode(%q) = %q, want %q", tt.borough, got, tt.want)
		}
	}
}

func TestBoroughLabel(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"1", "Manhattan"},
		{"2", "Bronx"},
		{"3", "Brooklyn"},
		{"4", "Queens"},
		{"5", "Staten Island"},
		{"9", "9"}, // unknown returns the code itself
	}
	for _, tt := range tests {
		got := boroughLabel(tt.code)
		if got != tt.want {
			t.Errorf("boroughLabel(%q) = %q, want %q", tt.code, got, tt.want)
		}
	}
}

func TestTruncateStr(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 5, "hi"},
		{"", 5, ""},
	}
	for _, tt := range tests {
		got := truncateStr(tt.s, tt.n)
		if got != tt.want {
			t.Errorf("truncateStr(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
		}
	}
}
