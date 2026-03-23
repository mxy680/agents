package zillow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newAgentMockServer creates an httptest.Server with agent-specific endpoints
// as well as all base mock endpoints.
func newAgentMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Agent search endpoint — uses /search/ prefix match since the URL
	// contains JSON query params with curly braces that don't match exact patterns.
	mux.HandleFunc("/search/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"cat3": map[string]any{
				"agentResults": []map[string]any{
					{
						"agentId":     "agent-abc123",
						"name":        "Jane Smith",
						"phone":       "720-555-1234",
						"rating":      4.9,
						"reviewCount": 52.0,
						"recentSales": 18.0,
						"profileUrl":  "https://www.zillow.com/profile/janesmith/",
						"photo":       "https://photos.zillowstatic.com/agent1.jpg",
					},
					{
						"agentId": "agent-def456",
						"name":    "Bob Jones",
						"phone":   "303-555-5678",
						"rating":  4.5,
					},
				},
			},
		})
	})

	// GraphQL covers agent get/reviews/listings AND property detail/walkscore/schools
	mux.HandleFunc("/graphql/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"agent": map[string]any{
					"name":        "Jane Smith",
					"phone":       "720-555-1234",
					"rating":      4.9,
					"reviewCount": 52.0,
					"reviews": []map[string]any{
						{
							"rating":      5.0,
							"date":        "2024-11-01",
							"description": "Jane was fantastic throughout the whole process.",
							"reviewer":    "Happy Buyer",
						},
					},
					"listings": []map[string]any{
						{
							"zpid":    "12345678",
							"address": "123 Main St, Denver, CO",
							"price":   450000.0,
							"beds":    3.0,
							"baths":   2.0,
							"sqft":    1800.0,
							"status":  "FOR_SALE",
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
					"walkScore":    map[string]any{"walkscore": 82.0, "description": "Very Walkable"},
					"transitScore": map[string]any{"transit_score": 65.0, "description": "Excellent Transit"},
					"bikeScore":    map[string]any{"bike_score": 70.0, "description": "Very Bikeable"},
					"schools": []map[string]any{
						{"name": "Lincoln Elementary", "rating": 8.0, "level": "Elementary", "type": "Public", "grades": "K-5", "distance": 0.3},
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

func TestAgentSearch(t *testing.T) {
	server := newAgentMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "agents", "search", "--location=Denver, CO"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Jane Smith") {
			t.Errorf("expected agent name in output, got: %s", out)
		}
		if !strings.Contains(out, "agent-abc123") {
			t.Errorf("expected agent ID in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "agents", "search", "--location=Denver, CO", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []AgentSummary
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Errorf("expected at least one agent, got empty array")
		}
		if results[0].Name != "Jane Smith" {
			t.Errorf("expected name Jane Smith, got: %s", results[0].Name)
		}
	})
}

func TestAgentGet(t *testing.T) {
	server := newAgentMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "agents", "get", "--agent-id=agent-abc123"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Jane Smith") {
			t.Errorf("expected agent name in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "agents", "get", "--agent-id=agent-abc123", "--json"})
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

func TestAgentReviews(t *testing.T) {
	server := newAgentMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "agents", "reviews", "--agent-id=agent-abc123"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "agent-abc123") {
			t.Errorf("expected agent ID in output, got: %s", out)
		}
		if !strings.Contains(out, "Happy Buyer") {
			t.Errorf("expected reviewer name in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "agents", "reviews", "--agent-id=agent-abc123", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result map[string]any
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if result["agentId"] != "agent-abc123" {
			t.Errorf("expected agentId in JSON, got: %v", result)
		}
		reviews, ok := result["reviews"].([]any)
		if !ok || len(reviews) == 0 {
			t.Errorf("expected reviews array in JSON, got: %v", result)
		}
	})
}

func TestAgentListings(t *testing.T) {
	server := newAgentMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "agents", "listings", "--agent-id=agent-abc123"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "123 Main St") {
			t.Errorf("expected listing address in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "agents", "listings", "--agent-id=agent-abc123", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []PropertySummary
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Errorf("expected at least one listing")
		}
	})
}

func TestParseAgentSearchResults(t *testing.T) {
	t.Run("valid_results", func(t *testing.T) {
		body := []byte(`{
			"cat3": {
				"agentResults": [
					{
						"agentId": "agent-abc123",
						"name": "Jane Smith",
						"phone": "720-555-1234",
						"rating": 4.9,
						"reviewCount": 52,
						"recentSales": 18,
						"profileUrl": "https://www.zillow.com/profile/janesmith/"
					},
					{
						"agentId": "agent-def456",
						"name": "Bob Jones",
						"rating": 4.5
					}
				]
			}
		}`)

		results, err := parseAgentSearchResults(body, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if results[0].AgentID != "agent-abc123" {
			t.Errorf("expected agentId agent-abc123, got %s", results[0].AgentID)
		}
		if results[0].Name != "Jane Smith" {
			t.Errorf("expected name Jane Smith, got %s", results[0].Name)
		}
		if results[0].Rating != 4.9 {
			t.Errorf("expected rating 4.9, got %f", results[0].Rating)
		}
		if results[0].ReviewCount != 52 {
			t.Errorf("expected reviewCount 52, got %d", results[0].ReviewCount)
		}
		if results[0].RecentSales != 18 {
			t.Errorf("expected recentSales 18, got %d", results[0].RecentSales)
		}
	})

	t.Run("limit_applied", func(t *testing.T) {
		body := []byte(`{
			"cat3": {
				"agentResults": [
					{"agentId": "a1", "name": "Agent 1"},
					{"agentId": "a2", "name": "Agent 2"},
					{"agentId": "a3", "name": "Agent 3"}
				]
			}
		}`)

		results, err := parseAgentSearchResults(body, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results (limit), got %d", len(results))
		}
	})

	t.Run("empty_cat3", func(t *testing.T) {
		body := []byte(`{}`)
		results, err := parseAgentSearchResults(body, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if results != nil {
			t.Errorf("expected nil results for empty body, got %v", results)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		body := []byte(`not-json`)
		_, err := parseAgentSearchResults(body, 10)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})
}
