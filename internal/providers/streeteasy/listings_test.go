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
		// parsePriceHistory returns empty; expect "No price history found." message.
		if !strings.Contains(out, "found") && !strings.Contains(out, "Price") && !strings.Contains(out, "Listed") {
			t.Errorf("expected price history message in output, got: %s", out)
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
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		// parsePriceHistory returns empty — that is expected behavior.
	})
}

func TestParseListingsFromPage(t *testing.T) {
	t.Run("with_listings", func(t *testing.T) {
		listings := []map[string]any{
			{
				"streetAddress":   "1 Broadway",
				"addressLocality": "New York",
				"addressRegion":   "NY",
				"price":           float64(1500000),
				"url":             "/nyc/real_estate/111",
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
		if !strings.Contains(s.Address, "1 Broadway") {
			t.Errorf("unexpected address: %s", s.Address)
		}
		if s.Price != 1500000 {
			t.Errorf("expected price 1500000, got %d", s.Price)
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

	t.Run("no_json_ld", func(t *testing.T) {
		body := []byte(`<html><body>No data here</body></html>`)
		_, err := parseListingsFromPage(body, "https://streeteasy.com", 10)
		if err == nil {
			t.Error("expected error when JSON-LD is missing")
		}
	})

	t.Run("limit_respected", func(t *testing.T) {
		listings := []map[string]any{
			{"streetAddress": "Addr 1", "addressLocality": "NYC", "price": float64(100000)},
			{"streetAddress": "Addr 2", "addressLocality": "NYC", "price": float64(200000)},
			{"streetAddress": "Addr 3", "addressLocality": "NYC", "price": float64(300000)},
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
	t.Run("returns_empty_array", func(t *testing.T) {
		// parsePriceHistory always returns an empty array because price history
		// is not available in the JSON-LD tag on StreetEasy pages.
		body := buildListingDetailPage(nil)
		entries, err := parsePriceHistory(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries (history not in JSON-LD), got %d", len(entries))
		}
	})

	t.Run("returns_empty_for_any_page", func(t *testing.T) {
		body := []byte(`<html><body>Some page content</body></html>`)
		entries, err := parsePriceHistory(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entries == nil {
			t.Error("expected non-nil empty slice")
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})
}

func TestLocationToSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"nyc", "nyc"},
		{"Bronx, NY 10452", "bronx"},
		{"Upper West Side", "upper-west-side"},
		{"Brooklyn", "brooklyn"},
		{"brooklyn", "brooklyn"},
		{"Manhattan", "manhattan"},
		{"Queens", "queens"},
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

func TestExtractFirstListingURL_WithURL(t *testing.T) {
	listings := []map[string]any{
		{
			"streetAddress":   "100 Main St",
			"addressLocality": "New York",
			"price":           float64(500000),
			"url":             "/nyc/real_estate/12345",
		},
	}
	body := buildListingsPage(listings)
	u := extractFirstListingURL(body, "https://streeteasy.com")
	if u != "https://streeteasy.com/nyc/real_estate/12345" {
		t.Errorf("unexpected URL from JSON-LD: %s", u)
	}
}

func TestExtractFirstListingURL_WithAbsoluteURL(t *testing.T) {
	listings := []map[string]any{
		{
			"streetAddress":   "100 Main St",
			"addressLocality": "New York",
			"price":           float64(500000),
			"url":             "https://streeteasy.com/nyc/real_estate/99999",
		},
	}
	body := buildListingsPage(listings)
	u := extractFirstListingURL(body, "https://streeteasy.com")
	if u != "https://streeteasy.com/nyc/real_estate/99999" {
		t.Errorf("unexpected absolute URL: %s", u)
	}
}

func TestExtractFirstListingURL_NoListings(t *testing.T) {
	body := buildListingsPage([]map[string]any{})
	u := extractFirstListingURL(body, "https://streeteasy.com")
	if u != "" {
		t.Errorf("expected empty URL for no listings, got: %s", u)
	}
}

func TestExtractFirstListingURL_NoJSONLD(t *testing.T) {
	body := []byte(`<html><body>no data</body></html>`)
	u := extractFirstListingURL(body, "https://streeteasy.com")
	if u != "" {
		t.Errorf("expected empty URL when no JSON-LD, got: %s", u)
	}
}

func TestExtractJSONLD_MultipleScriptTags(t *testing.T) {
	// Page with two JSON-LD blocks — first one has no @graph, second has listings.
	page := `<html><head>
<script type="application/ld+json">{"@context":"http://schema.org","@type":"WebSite","name":"StreetEasy"}</script>
<script type="application/ld+json">{"@context":"http://schema.org","@graph":[{"@type":"ApartmentComplex","address":{"@type":"PostalAddress","streetAddress":"5 Park Ave","addressLocality":"New York","addressRegion":"NY"},"additionalProperty":{"@type":"PropertyValue","value":"$900,000"},"url":"/nyc/real_estate/5"}]}</script>
</head><body></body></html>`

	items, err := extractJSONLD([]byte(page))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 graph item, got %d", len(items))
	}
	if items[0].Type != "ApartmentComplex" {
		t.Errorf("expected ApartmentComplex type, got %s", items[0].Type)
	}
}

func TestJSONLDItemToListingSummary(t *testing.T) {
	item := jsonLDItem{
		Type: "ApartmentComplex",
		Address: &jsonLDAddress{
			Type:            "PostalAddress",
			StreetAddress:   "5700 Arlington Avenue",
			AddressLocality: "Riverdale",
			AddressRegion:   "NY",
		},
		AdditionalProperty: &jsonLDPropValue{
			Type:  "PropertyValue",
			Value: "$445,000",
		},
		URL: "/nyc/real_estate/5700",
	}

	s := jsonLDItemToListingSummary(item, "https://streeteasy.com")
	if !strings.Contains(s.Address, "5700 Arlington Avenue") {
		t.Errorf("expected street address in address, got: %s", s.Address)
	}
	if !strings.Contains(s.Address, "Riverdale") {
		t.Errorf("expected locality in address, got: %s", s.Address)
	}
	if s.Price != 445000 {
		t.Errorf("expected price 445000, got %d", s.Price)
	}
	if s.URL != "https://streeteasy.com/nyc/real_estate/5700" {
		t.Errorf("unexpected URL: %s", s.URL)
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

// TestParseListingsFromPage_JSONLDStructure verifies parsing of the exact JSON-LD
// structure documented in the task spec.
func TestParseListingsFromPage_JSONLDStructure(t *testing.T) {
	page := `<!DOCTYPE html><html><head>
<script type="application/ld+json">
{
  "@context": "http://schema.org",
  "@graph": [
    {
      "@type": "ApartmentComplex",
      "additionalProperty": {
        "@type": "PropertyValue",
        "value": "$445,000"
      },
      "address": {
        "@type": "PostalAddress",
        "addressLocality": "Riverdale",
        "addressRegion": "NY",
        "postalCode": "10471",
        "streetAddress": "5700 Arlington Avenue"
      },
      "photo": {
        "@type": "CreativeWork",
        "image": "https://photos.zillowstatic.com/example.jpg"
      },
      "url": "/nyc/real_estate/99901"
    }
  ]
}
</script>
</head><body></body></html>`

	summaries, err := parseListingsFromPage([]byte(page), "https://streeteasy.com", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if !strings.Contains(s.Address, "5700 Arlington Avenue") {
		t.Errorf("expected street address, got: %s", s.Address)
	}
	if !strings.Contains(s.Address, "Riverdale") {
		t.Errorf("expected locality in address, got: %s", s.Address)
	}
	if s.Price != 445000 {
		t.Errorf("expected price 445000, got %d", s.Price)
	}
	if s.URL != "https://streeteasy.com/nyc/real_estate/99901" {
		t.Errorf("unexpected URL: %s", s.URL)
	}
}

// Ensure json import is used — needed by some test helpers that use json.Marshal.
var _ = json.Marshal
