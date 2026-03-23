package zillow

import (
	"strings"
	"testing"
)

func TestPropertySearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "properties", "search", "--location=Denver, CO"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "123 Main St") {
			t.Errorf("expected address in output, got: %s", out)
		}
		if !strings.Contains(out, "12345678") {
			t.Errorf("expected zpid in output, got: %s", out)
		}
		if !strings.Contains(out, "FOR SALE") {
			t.Errorf("expected status in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "properties", "search", "--location=Denver, CO", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "12345678") {
			t.Errorf("expected zpid in JSON output, got: %s", out)
		}
		if !strings.Contains(out, `"address"`) {
			t.Errorf("expected address field in JSON output, got: %s", out)
		}
	})
}

func TestPropertySearchWithFilters(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"zillow", "properties", "search",
			"--location=Denver, CO",
			"--min-price=300000",
			"--max-price=500000",
			"--min-beds=2",
			"--home-type=house",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "12345678") {
		t.Errorf("expected zpid in filtered JSON output, got: %s", out)
	}
}

func TestPropertySearchMap(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"zillow", "properties", "search-map",
			"--ne-lat=39.8",
			"--ne-lng=-104.9",
			"--sw-lat=39.6",
			"--sw-lng=-105.1",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "123 Main St") {
		t.Errorf("expected address in text output, got: %s", out)
	}
}

func TestPropertySearchMapJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"zillow", "properties", "search-map",
			"--ne-lat=39.8",
			"--ne-lng=-104.9",
			"--sw-lat=39.6",
			"--sw-lng=-105.1",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, `"zpid"`) {
		t.Errorf("expected zpid field in JSON output, got: %s", out)
	}
}

func TestParseSearchResults(t *testing.T) {
	body := []byte(`{
		"cat1": {
			"searchResults": {
				"listResults": [
					{
						"zpid": "12345678",
						"address": "123 Main St, Denver, CO 80202",
						"statusText": "FOR SALE",
						"unformattedPrice": 450000,
						"beds": 3,
						"baths": 2.0,
						"area": 1800,
						"latLong": {"latitude": 39.7392, "longitude": -104.9903},
						"detailUrl": "/homedetails/123-Main-St/12345678_zpid/",
						"hdpData": {
							"homeInfo": {
								"homeType": "SINGLE_FAMILY",
								"daysOnZillow": 14
							}
						}
					},
					{
						"zpid": "87654321",
						"address": "456 Oak Ave, Denver, CO 80203",
						"statusText": "FOR SALE",
						"unformattedPrice": 325000,
						"beds": 2,
						"baths": 1.5,
						"area": 1200
					}
				]
			}
		}
	}`)

	t.Run("no limit", func(t *testing.T) {
		summaries, err := parseSearchResults(body, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(summaries) != 2 {
			t.Fatalf("expected 2 results, got %d", len(summaries))
		}
		if summaries[0].ZPID != "12345678" {
			t.Errorf("first ZPID = %q, want %q", summaries[0].ZPID, "12345678")
		}
		if summaries[0].Price != 450000 {
			t.Errorf("first Price = %d, want 450000", summaries[0].Price)
		}
		if summaries[0].Beds != 3 {
			t.Errorf("first Beds = %d, want 3", summaries[0].Beds)
		}
		if summaries[0].Baths != 2.0 {
			t.Errorf("first Baths = %f, want 2.0", summaries[0].Baths)
		}
		if summaries[0].Sqft != 1800 {
			t.Errorf("first Sqft = %d, want 1800", summaries[0].Sqft)
		}
		if summaries[0].HomeType != "SINGLE_FAMILY" {
			t.Errorf("first HomeType = %q, want %q", summaries[0].HomeType, "SINGLE_FAMILY")
		}
		if summaries[0].DaysOnMarket != 14 {
			t.Errorf("first DaysOnMarket = %d, want 14", summaries[0].DaysOnMarket)
		}
		if summaries[0].Latitude != 39.7392 {
			t.Errorf("first Latitude = %f, want 39.7392", summaries[0].Latitude)
		}
		if !strings.Contains(summaries[0].ZillowURL, "12345678_zpid") {
			t.Errorf("first ZillowURL = %q, expected zpid suffix", summaries[0].ZillowURL)
		}
	})

	t.Run("with limit", func(t *testing.T) {
		summaries, err := parseSearchResults(body, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(summaries) != 1 {
			t.Fatalf("expected 1 result with limit, got %d", len(summaries))
		}
	})

	t.Run("empty cat1", func(t *testing.T) {
		summaries, err := parseSearchResults([]byte(`{}`), 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summaries != nil {
			t.Errorf("expected nil summaries for empty response, got %v", summaries)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := parseSearchResults([]byte(`not-json`), 0)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestBuildFilterState(t *testing.T) {
	t.Run("for_sale sets correct flags", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 0, 0, "", 0)
		if v, ok := fs["isForRent"].(map[string]any); !ok || v["value"] != false {
			t.Error("expected isForRent=false for for_sale status")
		}
		if v, ok := fs["isForSaleByAgent"].(map[string]any); !ok || v["value"] != true {
			t.Error("expected isForSaleByAgent=true for for_sale status")
		}
		if v, ok := fs["isRecentlySold"].(map[string]any); !ok || v["value"] != false {
			t.Error("expected isRecentlySold=false for for_sale status")
		}
	})

	t.Run("for_rent sets correct flags", func(t *testing.T) {
		fs := buildFilterState("for_rent", 0, 0, 0, 0, 0, 0, 0, 0, "", 0)
		if v, ok := fs["isForRent"].(map[string]any); !ok || v["value"] != true {
			t.Error("expected isForRent=true for for_rent status")
		}
		if v, ok := fs["isForSaleByAgent"].(map[string]any); !ok || v["value"] != false {
			t.Error("expected isForSaleByAgent=false for for_rent status")
		}
	})

	t.Run("sold sets correct flags", func(t *testing.T) {
		fs := buildFilterState("sold", 0, 0, 0, 0, 0, 0, 0, 0, "", 0)
		if v, ok := fs["isRecentlySold"].(map[string]any); !ok || v["value"] != true {
			t.Error("expected isRecentlySold=true for sold status")
		}
	})

	t.Run("price range", func(t *testing.T) {
		fs := buildFilterState("for_sale", 300000, 500000, 0, 0, 0, 0, 0, 0, "", 0)
		priceFilter, ok := fs["price"].(map[string]any)
		if !ok {
			t.Fatal("expected price filter")
		}
		if priceFilter["min"] != int64(300000) {
			t.Errorf("min price = %v, want 300000", priceFilter["min"])
		}
		if priceFilter["max"] != int64(500000) {
			t.Errorf("max price = %v, want 500000", priceFilter["max"])
		}
	})

	t.Run("min price only", func(t *testing.T) {
		fs := buildFilterState("for_sale", 200000, 0, 0, 0, 0, 0, 0, 0, "", 0)
		priceFilter, ok := fs["price"].(map[string]any)
		if !ok {
			t.Fatal("expected price filter")
		}
		if priceFilter["min"] != int64(200000) {
			t.Errorf("min price = %v, want 200000", priceFilter["min"])
		}
		if _, hasMax := priceFilter["max"]; hasMax {
			t.Error("expected no max price when maxPrice=0")
		}
	})

	t.Run("beds range", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 2, 4, 0, 0, 0, 0, "", 0)
		bedsFilter, ok := fs["beds"].(map[string]any)
		if !ok {
			t.Fatal("expected beds filter")
		}
		if bedsFilter["min"] != 2 {
			t.Errorf("min beds = %v, want 2", bedsFilter["min"])
		}
		if bedsFilter["max"] != 4 {
			t.Errorf("max beds = %v, want 4", bedsFilter["max"])
		}
	})

	t.Run("baths filter", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 2.0, 0, 0, 0, "", 0)
		bathsFilter, ok := fs["baths"].(map[string]any)
		if !ok {
			t.Fatal("expected baths filter")
		}
		if bathsFilter["min"] != 2.0 {
			t.Errorf("min baths = %v, want 2.0", bathsFilter["min"])
		}
	})

	t.Run("sqft range", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 1000, 3000, "", 0)
		sqftFilter, ok := fs["sqft"].(map[string]any)
		if !ok {
			t.Fatal("expected sqft filter")
		}
		if sqftFilter["min"] != 1000 {
			t.Errorf("min sqft = %v, want 1000", sqftFilter["min"])
		}
		if sqftFilter["max"] != 3000 {
			t.Errorf("max sqft = %v, want 3000", sqftFilter["max"])
		}
	})

	t.Run("days on zillow", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 0, 0, "", 7)
		dozFilter, ok := fs["doz"].(map[string]any)
		if !ok {
			t.Fatal("expected doz filter")
		}
		if dozFilter["value"] != "7" {
			t.Errorf("doz value = %v, want \"7\"", dozFilter["value"])
		}
	})

	t.Run("home type house", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 0, 0, "house", 0)
		if v, ok := fs["isSingleFamily"].(map[string]any); !ok || v["value"] != true {
			t.Error("expected isSingleFamily=true for house type")
		}
		if v, ok := fs["isAllHomes"].(map[string]any); !ok || v["value"] != false {
			t.Error("expected isAllHomes=false for house type")
		}
	})

	t.Run("home type condo", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 0, 0, "condo", 0)
		if v, ok := fs["isCondo"].(map[string]any); !ok || v["value"] != true {
			t.Error("expected isCondo=true for condo type")
		}
	})

	t.Run("home type townhouse", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 0, 0, "townhouse", 0)
		if v, ok := fs["isTownhouse"].(map[string]any); !ok || v["value"] != true {
			t.Error("expected isTownhouse=true for townhouse type")
		}
	})

	t.Run("home type multi_family", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 0, 0, "multi_family", 0)
		if v, ok := fs["isMultiFamily"].(map[string]any); !ok || v["value"] != true {
			t.Error("expected isMultiFamily=true for multi_family type")
		}
	})

	t.Run("home type land", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 0, 0, "land", 0)
		if v, ok := fs["isLotLand"].(map[string]any); !ok || v["value"] != true {
			t.Error("expected isLotLand=true for land type")
		}
	})

	t.Run("home type manufactured", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 0, 0, "manufactured", 0)
		if v, ok := fs["isManufactured"].(map[string]any); !ok || v["value"] != true {
			t.Error("expected isManufactured=true for manufactured type")
		}
	})

	t.Run("home type apartment", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 0, 0, "apartment", 0)
		if v, ok := fs["isApartment"].(map[string]any); !ok || v["value"] != true {
			t.Error("expected isApartment=true for apartment type")
		}
	})

	t.Run("no home type leaves no type filter", func(t *testing.T) {
		fs := buildFilterState("for_sale", 0, 0, 0, 0, 0, 0, 0, 0, "", 0)
		if _, ok := fs["isSingleFamily"]; ok {
			t.Error("expected no isSingleFamily filter when homeType is empty")
		}
	})
}

func TestMapSortValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"newest", "days"},
		{"price_low", "pricea"},
		{"price_high", "priced"},
		{"beds", "beds"},
		{"baths", "baths"},
		{"sqft", "size"},
		{"lot_size", "lot"},
		{"NEWEST", "days"},
		{"unknown_sort", "unknown_sort"},
		{"", ""},
	}
	for _, tt := range tests {
		got := mapSortValue(tt.input)
		if got != tt.want {
			t.Errorf("mapSortValue(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPropertySearchSortFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"zillow", "properties", "search",
			"--location=Denver, CO",
			"--sort=price_low",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "12345678") {
		t.Errorf("expected zpid in sorted JSON output, got: %s", out)
	}
}
