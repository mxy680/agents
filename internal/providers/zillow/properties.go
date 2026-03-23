package zillow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newPropertiesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "properties",
		Short:   "Search and view property listings",
		Aliases: []string{"property", "prop"},
	}

	cmd.AddCommand(newPropertySearchCmd(factory))
	cmd.AddCommand(newPropertySearchMapCmd(factory))

	return cmd
}

func newPropertySearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search properties by location",
		RunE:  makeRunPropertySearch(factory),
	}
	cmd.Flags().String("location", "", "Location to search (e.g., 'Denver, CO')")
	cmd.Flags().String("status", "for_sale", "Listing status: for_sale, for_rent, sold")
	cmd.Flags().Int64("min-price", 0, "Minimum price")
	cmd.Flags().Int64("max-price", 0, "Maximum price")
	cmd.Flags().Int("min-beds", 0, "Minimum bedrooms")
	cmd.Flags().Int("max-beds", 0, "Maximum bedrooms")
	cmd.Flags().Float64("min-baths", 0, "Minimum bathrooms")
	cmd.Flags().Float64("max-baths", 0, "Maximum bathrooms")
	cmd.Flags().Int("min-sqft", 0, "Minimum square footage")
	cmd.Flags().Int("max-sqft", 0, "Maximum square footage")
	cmd.Flags().String("home-type", "", "Home type: house, condo, townhouse, multi_family, land, manufactured, apartment")
	cmd.Flags().String("sort", "", "Sort: newest, price_low, price_high, beds, baths, sqft, lot_size")
	cmd.Flags().Int("days-on-zillow", 0, "Max days on Zillow")
	cmd.Flags().Int("limit", 25, "Maximum results")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.MarkFlagRequired("location")
	return cmd
}

func makeRunPropertySearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		location, _ := cmd.Flags().GetString("location")
		status, _ := cmd.Flags().GetString("status")
		minPrice, _ := cmd.Flags().GetInt64("min-price")
		maxPrice, _ := cmd.Flags().GetInt64("max-price")
		minBeds, _ := cmd.Flags().GetInt("min-beds")
		maxBeds, _ := cmd.Flags().GetInt("max-beds")
		minBaths, _ := cmd.Flags().GetFloat64("min-baths")
		maxBaths, _ := cmd.Flags().GetFloat64("max-baths")
		minSqft, _ := cmd.Flags().GetInt("min-sqft")
		maxSqft, _ := cmd.Flags().GetInt("max-sqft")
		homeType, _ := cmd.Flags().GetString("home-type")
		sortBy, _ := cmd.Flags().GetString("sort")
		daysOnZillow, _ := cmd.Flags().GetInt("days-on-zillow")
		limit, _ := cmd.Flags().GetInt("limit")
		page, _ := cmd.Flags().GetInt("page")

		filterState := buildFilterState(status, minPrice, maxPrice, minBeds, maxBeds, minBaths, maxBaths, minSqft, maxSqft, homeType, daysOnZillow)

		// Resolve location to coordinates via autocomplete, then build map bounds
		bounds, regionID, err := resolveLocationBounds(ctx, client, location)
		if err != nil {
			return fmt.Errorf("resolve location: %w", err)
		}

		searchState := map[string]any{
			"pagination":      map[string]any{"currentPage": page},
			"usersSearchTerm": location,
			"filterState":     filterState,
			"isMapVisible":    true,
		}
		if bounds != nil {
			searchState["mapBounds"] = bounds
		}
		if regionID != "" {
			searchState["regionSelection"] = []map[string]any{
				{"regionId": regionID, "regionType": 6},
			}
		}

		payload := map[string]any{
			"searchQueryState": searchState,
			"wants": map[string]any{
				"cat1": []string{"listResults", "mapResults"},
				"cat2": []string{"total"},
			},
			"requestId": 1,
		}

		if sortBy != "" {
			searchState["sortSelection"] = map[string]string{
				"value": mapSortValue(sortBy),
			}
		}

		reqURL := client.baseURL + "/async-create-search-page-state"
		body, err := client.PutJSON(ctx, reqURL, payload)
		if err != nil {
			return fmt.Errorf("search: %w", err)
		}

		summaries, err := parseSearchResults(body, limit)
		if err != nil {
			return fmt.Errorf("parse search results: %w", err)
		}

		return printPropertySummaries(cmd, summaries)
	}
}

func newPropertySearchMapCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search-map",
		Short: "Search properties by map coordinates",
		RunE:  makeRunPropertySearchMap(factory),
	}
	cmd.Flags().Float64("ne-lat", 0, "Northeast latitude")
	cmd.Flags().Float64("ne-lng", 0, "Northeast longitude")
	cmd.Flags().Float64("sw-lat", 0, "Southwest latitude")
	cmd.Flags().Float64("sw-lng", 0, "Southwest longitude")
	cmd.Flags().String("status", "for_sale", "Listing status: for_sale, for_rent, sold")
	cmd.Flags().Int("zoom", 12, "Map zoom level (1-21)")
	cmd.Flags().Int("limit", 25, "Maximum results")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.MarkFlagRequired("ne-lat")
	cmd.MarkFlagRequired("ne-lng")
	cmd.MarkFlagRequired("sw-lat")
	cmd.MarkFlagRequired("sw-lng")
	return cmd
}

func makeRunPropertySearchMap(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		neLat, _ := cmd.Flags().GetFloat64("ne-lat")
		neLng, _ := cmd.Flags().GetFloat64("ne-lng")
		swLat, _ := cmd.Flags().GetFloat64("sw-lat")
		swLng, _ := cmd.Flags().GetFloat64("sw-lng")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")

		filterState := buildFilterState(status, 0, 0, 0, 0, 0, 0, 0, 0, "", 0)

		payload := map[string]any{
			"searchQueryState": map[string]any{
				"pagination": map[string]any{"currentPage": page},
				"mapBounds": map[string]any{
					"north": neLat,
					"east":  neLng,
					"south": swLat,
					"west":  swLng,
				},
				"filterState":  filterState,
				"isMapVisible": true,
			},
			"wants": map[string]any{
				"cat1": []string{"listResults", "mapResults"},
				"cat2": []string{"total"},
			},
			"requestId": 1,
		}

		url := client.baseURL + "/async-create-search-page-state"
		body, err := client.PutJSON(ctx, url, payload)
		if err != nil {
			return fmt.Errorf("search-map: %w", err)
		}

		summaries, err := parseSearchResults(body, limit)
		if err != nil {
			return fmt.Errorf("parse search results: %w", err)
		}

		return printPropertySummaries(cmd, summaries)
	}
}

// parseSearchResults extracts property summaries from the search API response.
func parseSearchResults(body []byte, limit int) ([]PropertySummary, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	cat1, _ := resp["cat1"].(map[string]any)
	if cat1 == nil {
		return nil, nil
	}
	searchResults, _ := cat1["searchResults"].(map[string]any)
	if searchResults == nil {
		return nil, nil
	}
	listResults, _ := searchResults["listResults"].([]any)

	var summaries []PropertySummary
	for _, item := range listResults {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		s := PropertySummary{
			ZPID:    jsonStr(m, "zpid"),
			Address: jsonStr(m, "address"),
			Status:  jsonStr(m, "statusText"),
		}

		if price, ok := m["unformattedPrice"].(float64); ok {
			s.Price = int64(price)
		}
		if beds, ok := m["beds"].(float64); ok {
			s.Beds = int(beds)
		}
		if baths, ok := m["baths"].(float64); ok {
			s.Baths = baths
		}
		if area, ok := m["area"].(float64); ok {
			s.Sqft = int(area)
		}

		if latLong, ok := m["latLong"].(map[string]any); ok {
			if lat, ok := latLong["latitude"].(float64); ok {
				s.Latitude = lat
			}
			if lng, ok := latLong["longitude"].(float64); ok {
				s.Longitude = lng
			}
		}

		if detailURL := jsonStr(m, "detailUrl"); detailURL != "" {
			if strings.HasPrefix(detailURL, "http") {
				s.ZillowURL = detailURL
			} else {
				s.ZillowURL = "https://www.zillow.com" + detailURL
			}
		}

		if hdpData, ok := m["hdpData"].(map[string]any); ok {
			if homeInfo, ok := hdpData["homeInfo"].(map[string]any); ok {
				if ht := jsonStr(homeInfo, "homeType"); ht != "" {
					s.HomeType = ht
				}
				if days, ok := homeInfo["daysOnZillow"].(float64); ok {
					s.DaysOnMarket = int(days)
				}
			}
		}

		summaries = append(summaries, s)
		if limit > 0 && len(summaries) >= limit {
			break
		}
	}

	return summaries, nil
}

// buildFilterState constructs the filterState for Zillow's search API.
func buildFilterState(status string, minPrice, maxPrice int64, minBeds, maxBeds int, minBaths, maxBaths float64, minSqft, maxSqft int, homeType string, daysOnZillow int) map[string]any {
	fs := map[string]any{}

	switch status {
	case "for_sale":
		fs["isForSaleByAgent"] = map[string]any{"value": true}
		fs["isForSaleByOwner"] = map[string]any{"value": true}
		fs["isNewConstruction"] = map[string]any{"value": true}
		fs["isComingSoon"] = map[string]any{"value": true}
		fs["isAuction"] = map[string]any{"value": true}
		fs["isForRent"] = map[string]any{"value": false}
		fs["isRecentlySold"] = map[string]any{"value": false}
	case "for_rent":
		fs["isForRent"] = map[string]any{"value": true}
		fs["isForSaleByAgent"] = map[string]any{"value": false}
		fs["isForSaleByOwner"] = map[string]any{"value": false}
		fs["isNewConstruction"] = map[string]any{"value": false}
		fs["isComingSoon"] = map[string]any{"value": false}
		fs["isAuction"] = map[string]any{"value": false}
		fs["isRecentlySold"] = map[string]any{"value": false}
	case "sold":
		fs["isRecentlySold"] = map[string]any{"value": true}
		fs["isForSaleByAgent"] = map[string]any{"value": false}
		fs["isForSaleByOwner"] = map[string]any{"value": false}
		fs["isNewConstruction"] = map[string]any{"value": false}
		fs["isComingSoon"] = map[string]any{"value": false}
		fs["isAuction"] = map[string]any{"value": false}
		fs["isForRent"] = map[string]any{"value": false}
	}

	if minPrice > 0 {
		fs["price"] = map[string]any{"min": minPrice}
	}
	if maxPrice > 0 {
		if existing, ok := fs["price"].(map[string]any); ok {
			existing["max"] = maxPrice
		} else {
			fs["price"] = map[string]any{"max": maxPrice}
		}
	}
	if minBeds > 0 {
		fs["beds"] = map[string]any{"min": minBeds}
	}
	if maxBeds > 0 {
		if existing, ok := fs["beds"].(map[string]any); ok {
			existing["max"] = maxBeds
		} else {
			fs["beds"] = map[string]any{"max": maxBeds}
		}
	}
	if minBaths > 0 {
		fs["baths"] = map[string]any{"min": minBaths}
	}
	if maxBaths > 0 {
		if existing, ok := fs["baths"].(map[string]any); ok {
			existing["max"] = maxBaths
		} else {
			fs["baths"] = map[string]any{"max": maxBaths}
		}
	}
	if minSqft > 0 {
		fs["sqft"] = map[string]any{"min": minSqft}
	}
	if maxSqft > 0 {
		if existing, ok := fs["sqft"].(map[string]any); ok {
			existing["max"] = maxSqft
		} else {
			fs["sqft"] = map[string]any{"max": maxSqft}
		}
	}
	if daysOnZillow > 0 {
		fs["doz"] = map[string]any{"value": strconv.Itoa(daysOnZillow)}
	}

	if homeType != "" {
		switch strings.ToLower(homeType) {
		case "house":
			fs["isAllHomes"] = map[string]any{"value": false}
			fs["isSingleFamily"] = map[string]any{"value": true}
		case "condo":
			fs["isAllHomes"] = map[string]any{"value": false}
			fs["isCondo"] = map[string]any{"value": true}
		case "townhouse":
			fs["isAllHomes"] = map[string]any{"value": false}
			fs["isTownhouse"] = map[string]any{"value": true}
		case "multi_family":
			fs["isAllHomes"] = map[string]any{"value": false}
			fs["isMultiFamily"] = map[string]any{"value": true}
		case "land":
			fs["isAllHomes"] = map[string]any{"value": false}
			fs["isLotLand"] = map[string]any{"value": true}
		case "manufactured":
			fs["isAllHomes"] = map[string]any{"value": false}
			fs["isManufactured"] = map[string]any{"value": true}
		case "apartment":
			fs["isAllHomes"] = map[string]any{"value": false}
			fs["isApartment"] = map[string]any{"value": true}
		}
	}

	return fs
}

// mapSortValue converts user-facing sort names to Zillow's internal values.
func mapSortValue(sort string) string {
	switch strings.ToLower(sort) {
	case "newest":
		return "days"
	case "price_low":
		return "pricea"
	case "price_high":
		return "priced"
	case "beds":
		return "beds"
	case "baths":
		return "baths"
	case "sqft":
		return "size"
	case "lot_size":
		return "lot"
	default:
		return sort
	}
}

// extractZPIDFromURL extracts a ZPID from a Zillow URL like /homedetails/.../12345678_zpid/
func extractZPIDFromURL(rawURL string) string {
	parts := strings.Split(rawURL, "/")
	for _, part := range parts {
		if strings.HasSuffix(part, "_zpid") {
			return strings.TrimSuffix(part, "_zpid")
		}
	}
	return ""
}

// resolveLocationBounds resolves a location string to map bounds via the autocomplete API.
// Returns (mapBounds, regionID, error). The regionID is used for Zillow's regionSelection.
func resolveLocationBounds(ctx context.Context, client *Client, location string) (map[string]any, string, error) {
	reqURL := client.staticURL + "/autocomplete/v3/suggestions?q=" + url.QueryEscape(location)
	body, err := client.Get(ctx, reqURL)
	if err != nil {
		return nil, "", fmt.Errorf("autocomplete: %w", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, "", nil
	}

	results, _ := resp["results"].([]any)
	if len(results) == 0 {
		return nil, "", nil
	}

	// Use the first result
	first, _ := results[0].(map[string]any)
	if first == nil {
		return nil, "", nil
	}

	meta, _ := first["metaData"].(map[string]any)
	if meta == nil {
		return nil, "", nil
	}

	lat, latOK := meta["lat"].(float64)
	lng, lngOK := meta["lng"].(float64)
	if !latOK || !lngOK {
		return nil, "", nil
	}

	regionID := jsonStr(meta, "regionId")

	// Create map bounds around the coordinates (~0.4 degree offset for city-level search)
	offset := 0.4
	bounds := map[string]any{
		"north": lat + offset,
		"south": lat - offset,
		"east":  lng + offset,
		"west":  lng - offset,
	}

	return bounds, regionID, nil
}

// jsonStr safely extracts a string from a map.
func jsonStr(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return strconv.FormatFloat(val, 'f', -1, 64)
		case json.Number:
			return val.String()
		}
	}
	return ""
}
