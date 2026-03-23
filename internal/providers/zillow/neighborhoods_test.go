package zillow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newNeighborhoodMockServer creates an httptest.Server with neighborhood-specific endpoints.
func newNeighborhoodMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// GraphQL endpoint for neighborhood get and market stats
	mux.HandleFunc("/graphql/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"region": map[string]any{
					"name":            "Capitol Hill",
					"type":            "neighborhood",
					"medianHomeValue": 650000.0,
					"medianRent":      2400.0,
					"medianListPrice": 680000.0,
				},
				"property": map[string]any{
					"zpid":          12345678,
					"price":         450000,
					"bedrooms":      3,
					"bathrooms":     2.0,
					"livingArea":    1800,
					"homeType":      "SINGLE_FAMILY",
					"homeStatus":    "FOR_SALE",
					"zestimate":     460000,
					"rentZestimate": 2200,
					"address": map[string]any{
						"streetAddress": "123 Main St",
						"city":          "Denver",
						"state":         "CO",
						"zipcode":       "80202",
					},
				},
			},
		})
	})

	withAutocompleteMock(mux)
	withSearchMock(mux)
	withMortgageMock(mux)
	withLenderReviewsMock(mux)

	return httptest.NewServer(mux)
}

func TestNeighborhoodGet(t *testing.T) {
	server := newNeighborhoodMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "neighborhoods", "get", "--region-id=12345"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Capitol Hill") {
			t.Errorf("expected neighborhood name in output, got: %s", out)
		}
		if !strings.Contains(out, "neighborhood") {
			t.Errorf("expected type 'neighborhood' in output, got: %s", out)
		}
		if !strings.Contains(out, "650,000") {
			t.Errorf("expected median home value in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "neighborhoods", "get", "--region-id=12345", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result map[string]any
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		data, ok := result["data"].(map[string]any)
		if !ok {
			t.Fatalf("expected 'data' key in JSON response")
		}
		region, ok := data["region"].(map[string]any)
		if !ok {
			t.Fatalf("expected 'region' key in data")
		}
		if region["name"] != "Capitol Hill" {
			t.Errorf("expected region name 'Capitol Hill', got: %v", region["name"])
		}
	})
}

func TestNeighborhoodSearch(t *testing.T) {
	server := newNeighborhoodMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "neighborhoods", "search", "--location=Denver"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Denver") {
			t.Errorf("expected Denver in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "neighborhoods", "search", "--location=Denver", "--json"})
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
	})
}

func TestNeighborhoodMarketStats(t *testing.T) {
	server := newNeighborhoodMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "neighborhoods", "market-stats", "--region-id=12345"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Capitol Hill") {
			t.Errorf("expected neighborhood name in output, got: %s", out)
		}
		if !strings.Contains(out, "650,000") {
			t.Errorf("expected median home value in output, got: %s", out)
		}
		if !strings.Contains(out, "2,400") {
			t.Errorf("expected median rent in output, got: %s", out)
		}
		if !strings.Contains(out, "680,000") {
			t.Errorf("expected median list price in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "neighborhoods", "market-stats", "--region-id=12345", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result map[string]any
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
	})
}
