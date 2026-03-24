package streeteasy

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestListingsSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"streeteasy", "listings", "search", "--location=nyc"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Riverside") {
			t.Errorf("expected address in output, got: %s", out)
		}
		if !strings.Contains(out, "for_sale") {
			t.Errorf("expected status in output, got: %s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"streeteasy", "listings", "search", "--location=nyc", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []ListingSummary
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Errorf("expected at least one listing")
		}
		if results[0].Address == "" {
			t.Errorf("expected non-empty address")
		}
		if results[0].Price == 0 {
			t.Errorf("expected non-zero price")
		}
	})

	t.Run("for_rent_status", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"streeteasy", "listings", "search", "--location=nyc", "--status=for_rent", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []ListingSummary
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Errorf("expected at least one rent listing")
		}
	})

	t.Run("limit", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"streeteasy", "listings", "search", "--location=nyc", "--limit=1", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []ListingSummary
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 result with --limit=1, got %d", len(results))
		}
	})
}

func TestListingsHistory(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"streeteasy", "listings", "history", "--address=100 Riverside Blvd New York NY"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Listed") && !strings.Contains(out, "Price") && !strings.Contains(out, "found") {
			t.Errorf("expected price history or 'found' in output, got: %s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"streeteasy", "listings", "history", "--address=100 Riverside Blvd New York NY", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []PriceHistoryEntry
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			// An empty array is also valid JSON
			var empty []PriceHistoryEntry
			if err2 := json.Unmarshal([]byte(out), &empty); err2 != nil {
				t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
			}
		}
	})
}

func TestParseListingsFromPage(t *testing.T) {
	t.Run("with_listings", func(t *testing.T) {
		listings := []map[string]any{
			{
				"id":      "111",
				"address": "1 Broadway, New York, NY 10004",
				"price":   float64(1500000),
				"bedrooms": float64(2),
				"bathrooms": float64(1.5),
				"sqft":    float64(1000),
				"status":  "for_sale",
				"url":     "/nyc/real_estate/111",
			},
		}
		body := buildListingsPage(listings)
		summaries, err := parseListingsFromPage(body, "https://streeteasy.com", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(summaries) != 1 {
			t.Fatalf("expected 1 summary, got %d", len(summaries))
		}
		s := summaries[0]
		if s.Address != "1 Broadway, New York, NY 10004" {
			t.Errorf("unexpected address: %s", s.Address)
		}
		if s.Price != 1500000 {
			t.Errorf("expected price 1500000, got %d", s.Price)
		}
		if s.Beds != 2 {
			t.Errorf("expected 2 beds, got %d", s.Beds)
		}
		if s.Baths != 1.5 {
			t.Errorf("expected 1.5 baths, got %f", s.Baths)
		}
		if s.URL != "https://streeteasy.com/nyc/real_estate/111" {
			t.Errorf("unexpected URL: %s", s.URL)
		}
	})

	t.Run("empty_listings", func(t *testing.T) {
		body := buildListingsPage([]map[string]any{})
		summaries, err := parseListingsFromPage(body, "https://streeteasy.com", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(summaries) != 0 {
			t.Errorf("expected 0 summaries, got %d", len(summaries))
		}
	})

	t.Run("no_next_data", func(t *testing.T) {
		body := []byte(`<html><body>No data here</body></html>`)
		_, err := parseListingsFromPage(body, "https://streeteasy.com", 10)
		if err == nil {
			t.Error("expected error when __NEXT_DATA__ is missing")
		}
	})

	t.Run("limit_respected", func(t *testing.T) {
		listings := []map[string]any{
			{"id": "1", "address": "Addr 1"},
			{"id": "2", "address": "Addr 2"},
			{"id": "3", "address": "Addr 3"},
		}
		body := buildListingsPage(listings)
		summaries, err := parseListingsFromPage(body, "https://streeteasy.com", 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(summaries) != 2 {
			t.Errorf("expected 2 summaries (limit=2), got %d", len(summaries))
		}
	})
}

func TestParsePriceHistory(t *testing.T) {
	t.Run("with_history", func(t *testing.T) {
		history := []map[string]any{
			{"date": "2024-01-01", "event": "Listed", "price": float64(1500000)},
			{"date": "2024-03-15", "event": "Price Change", "price": float64(1400000)},
		}
		body := buildListingDetailPage(history)
		entries, err := parsePriceHistory(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].Date != "2024-01-01" {
			t.Errorf("unexpected date: %s", entries[0].Date)
		}
		if entries[0].Event != "Listed" {
			t.Errorf("unexpected event: %s", entries[0].Event)
		}
		if entries[0].Price != 1500000 {
			t.Errorf("expected price 1500000, got %d", entries[0].Price)
		}
	})

	t.Run("empty_history", func(t *testing.T) {
		body := buildListingDetailPage([]map[string]any{})
		entries, err := parsePriceHistory(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("no_next_data", func(t *testing.T) {
		body := []byte(`<html><body>Not a Next.js page</body></html>`)
		_, err := parsePriceHistory(body)
		if err == nil {
			t.Error("expected error when __NEXT_DATA__ is missing")
		}
	})
}

func TestLocationToSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"nyc", "nyc"},
		{"Bronx, NY 10452", "bronx-ny-10452"},
		{"Upper West Side", "upper-west-side"},
		{"New York, NY", "new-york-ny"},
		{"brooklyn", "brooklyn"},
	}
	for _, tt := range tests {
		got := locationToSlug(tt.input)
		if got != tt.want {
			t.Errorf("locationToSlug(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNavigatePath(t *testing.T) {
	m := map[string]any{
		"a": map[string]any{
			"b": []any{"x", "y"},
		},
		"flat": []any{"z"},
	}

	// Two-level path
	result := navigatePath(m, []string{"a", "b"})
	if len(result) != 2 {
		t.Errorf("navigatePath two-level: expected 2 results, got %d", len(result))
	}

	// One-level path
	result = navigatePath(m, []string{"flat"})
	if len(result) != 1 {
		t.Errorf("navigatePath one-level: expected 1 result, got %d", len(result))
	}

	// Missing key
	result = navigatePath(m, []string{"missing"})
	if result != nil {
		t.Errorf("navigatePath missing: expected nil, got %v", result)
	}

	// Empty path
	result = navigatePath(m, []string{})
	if result != nil {
		t.Errorf("navigatePath empty: expected nil, got %v", result)
	}

	// Intermediate not a map
	result = navigatePath(m, []string{"flat", "b"})
	if result != nil {
		t.Errorf("navigatePath non-map intermediate: expected nil, got %v", result)
	}
}

func TestExtractListingSummary_NestedAddress(t *testing.T) {
	m := map[string]any{
		"id": "999",
		"address": map[string]any{
			"streetAddress": "123 Main St",
			"city":          "New York",
			"state":         "NY",
		},
		"price":      float64(750000),
		"beds":       float64(2),
		"baths":      float64(1.0),
		"squareFeet": float64(900),
		"url":        "https://streeteasy.com/nyc/real_estate/999",
	}

	s := extractListingSummary(m, "https://streeteasy.com")
	if s.ID != "999" {
		t.Errorf("expected id 999, got %s", s.ID)
	}
	if s.Address == "" {
		t.Error("expected non-empty address from nested map")
	}
	if s.Price != 750000 {
		t.Errorf("expected price 750000, got %d", s.Price)
	}
	if s.Sqft != 900 {
		t.Errorf("expected sqft 900, got %d", s.Sqft)
	}
	if s.URL != "https://streeteasy.com/nyc/real_estate/999" {
		t.Errorf("unexpected URL: %s", s.URL)
	}
}

func TestExtractListingSummary_SlugURL(t *testing.T) {
	m := map[string]any{
		"address": "200 West St",
		"slug":    "nyc/real_estate/200-west-st",
	}
	s := extractListingSummary(m, "https://streeteasy.com")
	if s.URL != "https://streeteasy.com/nyc/real_estate/200-west-st" {
		t.Errorf("unexpected slug URL: %s", s.URL)
	}
}

func TestExtractListingSummary_StringPrice(t *testing.T) {
	m := map[string]any{
		"address": "300 East 57th St",
		"price":   "$1,200,000",
	}
	s := extractListingSummary(m, "https://streeteasy.com")
	if s.Price != 1200000 {
		t.Errorf("expected price 1200000, got %d", s.Price)
	}
}

func TestExtractFirstListingURL_WithSlug(t *testing.T) {
	listings := []map[string]any{
		{"slug": "nyc/real_estate/12345"},
	}
	body := buildListingsPage(listings)
	url := extractFirstListingURL(body, "https://streeteasy.com")
	if url != "https://streeteasy.com/nyc/real_estate/12345" {
		t.Errorf("unexpected URL from slug: %s", url)
	}
}

func TestExtractFirstListingURL_WithAbsoluteURL(t *testing.T) {
	listings := []map[string]any{
		{"url": "https://streeteasy.com/nyc/real_estate/99999"},
	}
	body := buildListingsPage(listings)
	url := extractFirstListingURL(body, "https://streeteasy.com")
	if url != "https://streeteasy.com/nyc/real_estate/99999" {
		t.Errorf("unexpected absolute URL: %s", url)
	}
}

func TestExtractFirstListingURL_NoListings(t *testing.T) {
	body := buildListingsPage([]map[string]any{})
	url := extractFirstListingURL(body, "https://streeteasy.com")
	if url != "" {
		t.Errorf("expected empty URL for no listings, got: %s", url)
	}
}

func TestExtractFirstListingURL_NoNextData(t *testing.T) {
	body := []byte(`<html><body>no data</body></html>`)
	url := extractFirstListingURL(body, "https://streeteasy.com")
	if url != "" {
		t.Errorf("expected empty URL when no __NEXT_DATA__, got: %s", url)
	}
}

func TestParsePriceHistory_NestedPath(t *testing.T) {
	// Test the nested props -> pageProps -> listing -> priceHistory path
	priceHistory := []map[string]any{
		{"date": "2024-05-01", "event": "Listed", "price": float64(900000)},
	}
	nextData := map[string]any{
		"props": map[string]any{
			"pageProps": map[string]any{
				"listing": map[string]any{
					"priceHistory": priceHistory,
				},
			},
		},
	}
	nextDataJSON, _ := json.Marshal(nextData)
	page := `<script id="__NEXT_DATA__" type="application/json">` + string(nextDataJSON) + `</script>`

	entries, err := parsePriceHistory([]byte(page))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Price != 900000 {
		t.Errorf("expected price 900000, got %d", entries[0].Price)
	}
}

func TestParsePriceHistory_EventTypeKey(t *testing.T) {
	// Test fallback to "eventType" key when "event" is missing
	history := []map[string]any{
		{"date": "2024-02-01", "eventType": "Price Reduction", "price": float64(800000)},
	}
	body := buildListingDetailPage(history)
	entries, err := parsePriceHistory(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Event != "Price Reduction" {
		t.Errorf("expected event 'Price Reduction', got %s", entries[0].Event)
	}
}

func TestParseListingsFromPage_DataPath(t *testing.T) {
	// Test the props -> pageProps -> data -> listings path
	listings := []map[string]any{
		{"id": "555", "address": "5 Beekman St"},
	}
	nextData := map[string]any{
		"props": map[string]any{
			"pageProps": map[string]any{
				"data": map[string]any{
					"listings": listings,
				},
			},
		},
	}
	nextDataJSON, _ := json.Marshal(nextData)
	page := `<script id="__NEXT_DATA__" type="application/json">` + string(nextDataJSON) + `</script>`

	summaries, err := parseListingsFromPage([]byte(page), "https://streeteasy.com", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
}

func TestParseRawPrice(t *testing.T) {
	tests := []struct {
		input string
		want  int64
		err   bool
	}{
		{"1500000", 1500000, false},
		{"$1,500,000", 1500000, false},
		{"4,750,000", 4750000, false},
		{"abc", 0, true},
	}
	for _, tt := range tests {
		got, err := parseRawPrice(tt.input)
		if tt.err {
			if err == nil {
				t.Errorf("parseRawPrice(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseRawPrice(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseRawPrice(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
