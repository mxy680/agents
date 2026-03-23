package zillow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newBuilderMockServer creates an httptest.Server with builder-specific endpoints.
func newBuilderMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// GraphQL endpoint: handles builder search, get, communities, and reviews
	mux.HandleFunc("/graphql/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"builders": []map[string]any{
					{
						"builderId":   "builder-001",
						"name":        "Toll Brothers",
						"rating":      4.2,
						"reviewCount": 315.0,
						"url":         "https://www.zillow.com/builder/toll-brothers/",
					},
					{
						"builderId":   "builder-002",
						"name":        "KB Home",
						"rating":      3.8,
						"reviewCount": 200.0,
					},
				},
				"builder": map[string]any{
					"name":        "Toll Brothers",
					"rating":      4.2,
					"reviewCount": 315.0,
					"communities": []map[string]any{
						{
							"name":      "Highland Estates",
							"location":  "Denver, CO",
							"priceFrom": 650000.0,
							"priceTo":   950000.0,
							"url":       "https://www.zillow.com/community/highland-estates/",
						},
					},
					"reviews": []map[string]any{
						{
							"rating":      4.0,
							"description": "Good quality construction.",
						},
					},
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

func TestBuilderSearch(t *testing.T) {
	server := newBuilderMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "builders", "search", "--location=Denver, CO"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Toll Brothers") {
			t.Errorf("expected builder name in output, got: %s", out)
		}
		if !strings.Contains(out, "builder-001") {
			t.Errorf("expected builder ID in output, got: %s", out)
		}
		if !strings.Contains(out, "4.2") {
			t.Errorf("expected rating in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "builders", "search", "--location=Denver, CO", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []BuilderSummary
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Errorf("expected at least one builder")
		}
		if results[0].Name != "Toll Brothers" {
			t.Errorf("expected name 'Toll Brothers', got %s", results[0].Name)
		}
		if results[0].BuilderID != "builder-001" {
			t.Errorf("expected builderID 'builder-001', got %s", results[0].BuilderID)
		}
	})
}

func TestBuilderGet(t *testing.T) {
	server := newBuilderMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "builders", "get", "--builder-id=builder-001"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Toll Brothers") {
			t.Errorf("expected builder name in output, got: %s", out)
		}
		if !strings.Contains(out, "4.2") {
			t.Errorf("expected rating in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "builders", "get", "--builder-id=builder-001", "--json"})
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

func TestBuilderCommunities(t *testing.T) {
	server := newBuilderMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "builders", "communities", "--builder-id=builder-001"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Highland Estates") {
			t.Errorf("expected community name in output, got: %s", out)
		}
		if !strings.Contains(out, "Denver") {
			t.Errorf("expected location in output, got: %s", out)
		}
		if !strings.Contains(out, "650,000") {
			t.Errorf("expected price range in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "builders", "communities", "--builder-id=builder-001", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result map[string]any
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if result["builderId"] != "builder-001" {
			t.Errorf("expected builderId in JSON, got: %v", result)
		}
		communities, ok := result["communities"].([]any)
		if !ok || len(communities) == 0 {
			t.Errorf("expected communities array in JSON, got: %v", result)
		}
	})
}

func TestBuilderReviews(t *testing.T) {
	server := newBuilderMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "builders", "reviews", "--builder-id=builder-001"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Builder reviews loaded") {
			t.Errorf("expected 'Builder reviews loaded' in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "builders", "reviews", "--builder-id=builder-001", "--json"})
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

func TestParseBuilderSearchResults(t *testing.T) {
	t.Run("valid_results", func(t *testing.T) {
		body := []byte(`{
			"data": {
				"builders": [
					{
						"builderId": "builder-001",
						"name": "Toll Brothers",
						"rating": 4.2,
						"reviewCount": 315,
						"url": "https://www.zillow.com/builder/toll-brothers/"
					},
					{
						"builderId": "builder-002",
						"name": "KB Home",
						"rating": 3.8,
						"reviewCount": 200
					}
				]
			}
		}`)

		results, err := parseBuilderSearchResults(body, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if results[0].BuilderID != "builder-001" {
			t.Errorf("expected builderID 'builder-001', got %s", results[0].BuilderID)
		}
		if results[0].Name != "Toll Brothers" {
			t.Errorf("expected name 'Toll Brothers', got %s", results[0].Name)
		}
		if results[0].Rating != 4.2 {
			t.Errorf("expected rating 4.2, got %f", results[0].Rating)
		}
		if results[0].ReviewCount != 315 {
			t.Errorf("expected reviewCount 315, got %d", results[0].ReviewCount)
		}
	})

	t.Run("limit_applied", func(t *testing.T) {
		body := []byte(`{
			"data": {
				"builders": [
					{"builderId": "b1", "name": "Builder 1"},
					{"builderId": "b2", "name": "Builder 2"},
					{"builderId": "b3", "name": "Builder 3"}
				]
			}
		}`)

		results, err := parseBuilderSearchResults(body, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results (limit), got %d", len(results))
		}
	})

	t.Run("no_data", func(t *testing.T) {
		body := []byte(`{}`)
		results, err := parseBuilderSearchResults(body, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if results != nil {
			t.Errorf("expected nil results for empty body")
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		_, err := parseBuilderSearchResults([]byte(`not-json`), 10)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}
