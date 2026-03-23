package zillow

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSearchAutocomplete(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_city", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "search", "autocomplete", "--query=Denver"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Denver") {
			t.Errorf("expected Denver in output, got: %s", out)
		}
		if !strings.Contains(out, "city") {
			t.Errorf("expected type 'city' in output, got: %s", out)
		}
	})

	t.Run("text_address", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "search", "autocomplete", "--query=123 Main"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "123 Main St") {
			t.Errorf("expected address in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "search", "autocomplete", "--query=Denver", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []AutocompleteResult
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Errorf("expected at least one result")
		}
		if results[0].Display == "" {
			t.Errorf("expected non-empty display")
		}
		if results[0].Type != "city" {
			t.Errorf("expected type 'city', got: %s", results[0].Type)
		}
	})
}

func TestSearchByAddress(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("address_with_zpid", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		// Query with "123" triggers ZPID result in the mock
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "search", "by-address", "--address=123 Main St Denver"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		// Should fetch property detail since mock returns a ZPID for "123"
		if !strings.Contains(out, "123 Main St") {
			t.Errorf("expected address in output, got: %s", out)
		}
	})

	t.Run("address_no_zpid", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		// Query without "123" returns city result with no ZPID
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "search", "by-address", "--address=Denver"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		// Should show suggestions since no ZPID found
		if !strings.Contains(out, "Denver") {
			t.Errorf("expected Denver in output, got: %s", out)
		}
	})

	t.Run("json_with_zpid", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "search", "by-address", "--address=123 Main St Denver", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		// Should return autocomplete results as JSON array
		var results []AutocompleteResult
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Error("expected at least one result")
		}
	})
}

func TestParseAutocompleteResults(t *testing.T) {
	t.Run("address_result", func(t *testing.T) {
		body := []byte(`{
			"results": [
				{
					"display": "123 Main St, Denver, CO 80202",
					"resultType": "address",
					"metaData": {
						"zpid": "12345678",
						"lat": 39.7392,
						"lng": -104.9903
					}
				},
				{
					"display": "Denver, CO",
					"resultType": "city",
					"metaData": {
						"zpid": "",
						"lat": 39.7392,
						"lng": -104.9903
					}
				}
			]
		}`)

		results, err := parseAutocompleteResults(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}

		addr := results[0]
		if addr.Display != "123 Main St, Denver, CO 80202" {
			t.Errorf("unexpected display: %s", addr.Display)
		}
		if addr.Type != "address" {
			t.Errorf("expected type 'address', got %s", addr.Type)
		}
		if addr.ZPID != "12345678" {
			t.Errorf("expected ZPID 12345678, got %s", addr.ZPID)
		}
		if addr.Latitude != 39.7392 {
			t.Errorf("expected lat 39.7392, got %f", addr.Latitude)
		}
		if addr.Longitude != -104.9903 {
			t.Errorf("expected lng -104.9903, got %f", addr.Longitude)
		}

		city := results[1]
		if city.Type != "city" {
			t.Errorf("expected type 'city', got %s", city.Type)
		}
		if city.ZPID != "" {
			t.Errorf("expected empty ZPID for city, got %s", city.ZPID)
		}
	})

	t.Run("empty_results", func(t *testing.T) {
		body := []byte(`{"results": []}`)
		results, err := parseAutocompleteResults(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("missing_results_key", func(t *testing.T) {
		body := []byte(`{}`)
		results, err := parseAutocompleteResults(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if results != nil {
			t.Errorf("expected nil results, got %v", results)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		_, err := parseAutocompleteResults([]byte(`not-json`))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}
