package zillow

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
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
	cmd.AddCommand(newPropertyGetCmd(factory))
	cmd.AddCommand(newPropertyGetByURLCmd(factory))
	cmd.AddCommand(newPropertyPhotosCmd(factory))
	cmd.AddCommand(newPropertyPriceHistoryCmd(factory))
	cmd.AddCommand(newPropertyTaxHistoryCmd(factory))
	cmd.AddCommand(newPropertySimilarCmd(factory))
	cmd.AddCommand(newPropertyNearbyCmd(factory))

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

		payload := map[string]any{
			"searchQueryState": map[string]any{
				"pagination":      map[string]any{"currentPage": page},
				"usersSearchTerm": location,
				"filterState":     filterState,
			},
			"wants": map[string]any{
				"cat1": []string{"listResults", "mapResults"},
				"cat2": []string{"total"},
			},
			"requestId": 1,
		}

		if sortBy != "" {
			payload["searchQueryState"].(map[string]any)["sortSelection"] = map[string]string{
				"value": mapSortValue(sortBy),
			}
		}

		url := client.baseURL + "/async-create-search-page-state"
		body, err := client.PutJSON(ctx, url, payload)
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
				"filterState": filterState,
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

func newPropertyGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get property details by ZPID",
		RunE:  makeRunPropertyGet(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunPropertyGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")

		detail, err := fetchPropertyDetail(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get property: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		printPropertyDetail(detail)
		return nil
	}
}

func newPropertyGetByURLCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-by-url",
		Short: "Get property details by Zillow URL",
		RunE:  makeRunPropertyGetByURL(factory),
	}
	cmd.Flags().String("url", "", "Zillow property URL")
	cmd.MarkFlagRequired("url")
	return cmd
}

func makeRunPropertyGetByURL(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		rawURL, _ := cmd.Flags().GetString("url")

		// Extract ZPID from URL (format: .../12345678_zpid/)
		zpid := extractZPIDFromURL(rawURL)
		if zpid == "" {
			return fmt.Errorf("could not extract ZPID from URL: %s", rawURL)
		}

		detail, err := fetchPropertyDetail(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get property by URL: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		printPropertyDetail(detail)
		return nil
	}
}

func newPropertyPhotosCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "photos",
		Short: "Get property photos",
		RunE:  makeRunPropertyPhotos(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunPropertyPhotos(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")

		detail, err := fetchPropertyDetail(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get photos: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"zpid":   detail.ZPID,
				"photos": detail.Photos,
			})
		}

		if len(detail.Photos) == 0 {
			fmt.Println("No photos found.")
			return nil
		}
		lines := []string{fmt.Sprintf("Photos for %s (%d):", detail.Address, len(detail.Photos))}
		for _, p := range detail.Photos {
			lines = append(lines, "  "+p)
		}
		cli.PrintText(lines)
		return nil
	}
}

func newPropertyPriceHistoryCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price-history",
		Short: "Get property price history",
		RunE:  makeRunPropertyPriceHistory(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunPropertyPriceHistory(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")

		detail, err := fetchPropertyDetail(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get price history: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"zpid":         detail.ZPID,
				"priceHistory": detail.PriceHistory,
			})
		}

		if len(detail.PriceHistory) == 0 {
			fmt.Println("No price history found.")
			return nil
		}
		lines := []string{fmt.Sprintf("Price History for %s:", detail.Address)}
		lines = append(lines, fmt.Sprintf("  %-12s  %-15s  %-12s  %-10s", "DATE", "EVENT", "PRICE", "SOURCE"))
		for _, e := range detail.PriceHistory {
			lines = append(lines, fmt.Sprintf("  %-12s  %-15s  %-12s  %-10s",
				e.Date, e.Event, formatPrice(e.Price), e.Source))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newPropertyTaxHistoryCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tax-history",
		Short: "Get property tax history",
		RunE:  makeRunPropertyTaxHistory(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunPropertyTaxHistory(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")

		detail, err := fetchPropertyDetail(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get tax history: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"zpid":       detail.ZPID,
				"taxHistory": detail.TaxHistory,
			})
		}

		if len(detail.TaxHistory) == 0 {
			fmt.Println("No tax history found.")
			return nil
		}
		lines := []string{fmt.Sprintf("Tax History for %s:", detail.Address)}
		lines = append(lines, fmt.Sprintf("  %-6s  %-12s  %-12s", "YEAR", "TAX PAID", "ASSESSED"))
		for _, t := range detail.TaxHistory {
			lines = append(lines, fmt.Sprintf("  %-6d  %-12s  %-12s",
				t.Year, formatPrice(t.TaxPaid), formatPrice(t.TaxAssessed)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newPropertySimilarCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "similar",
		Short: "Get similar/comparable properties",
		RunE:  makeRunPropertySimilar(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.Flags().Int("limit", 10, "Maximum results")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunPropertySimilar(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")
		limit, _ := cmd.Flags().GetInt("limit")

		detail, err := fetchPropertyDetail(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get similar: %w", err)
		}

		// Comps come from the property detail response
		summaries := parsePropertyDetailComps(detail)
		if limit > 0 && len(summaries) > limit {
			summaries = summaries[:limit]
		}

		return printPropertySummaries(cmd, summaries)
	}
}

func newPropertyNearbyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nearby",
		Short: "Get nearby properties",
		RunE:  makeRunPropertyNearby(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.Flags().Int("limit", 10, "Maximum results")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunPropertyNearby(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")
		limit, _ := cmd.Flags().GetInt("limit")

		// Use the property detail endpoint which includes nearbyHomes
		url := client.baseURL + "/graphql/?zpid=" + zpid
		body, err := client.Get(ctx, url)
		if err != nil {
			return fmt.Errorf("get nearby: %w", err)
		}

		summaries, err := parseNearbyFromGraphQL(body)
		if err != nil {
			return fmt.Errorf("parse nearby: %w", err)
		}

		if limit > 0 && len(summaries) > limit {
			summaries = summaries[:limit]
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
			s.ZillowURL = "https://www.zillow.com" + detailURL
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

// fetchPropertyDetail retrieves full property details via Zillow's GraphQL API.
func fetchPropertyDetail(ctx context.Context, client *Client, zpid string) (PropertyDetail, error) {
	reqURL := client.baseURL + "/graphql/?zpid=" + zpid
	body, err := client.Get(ctx, reqURL)
	if err != nil {
		return PropertyDetail{}, fmt.Errorf("fetch property %s: %w", zpid, err)
	}
	return parsePropertyDetailResponse(body, zpid)
}

// parsePropertyDetailResponse parses the GraphQL property detail response.
func parsePropertyDetailResponse(body []byte, zpid string) (PropertyDetail, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return PropertyDetail{}, fmt.Errorf("unmarshal: %w", err)
	}

	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return PropertyDetail{}, fmt.Errorf("no data in response for zpid %s", zpid)
	}
	prop, _ := data["property"].(map[string]any)
	if prop == nil {
		return PropertyDetail{}, fmt.Errorf("no property data for zpid %s", zpid)
	}

	detail := PropertyDetail{
		ZPID:        zpid,
		Description: jsonStr(prop, "description"),
		HomeType:    jsonStr(prop, "homeType"),
		Status:      jsonStr(prop, "homeStatus"),
	}

	if price, ok := prop["price"].(float64); ok {
		detail.Price = int64(price)
	}
	if beds, ok := prop["bedrooms"].(float64); ok {
		detail.Beds = int(beds)
	}
	if baths, ok := prop["bathrooms"].(float64); ok {
		detail.Baths = baths
	}
	if sqft, ok := prop["livingArea"].(float64); ok {
		detail.Sqft = int(sqft)
	}
	if lot, ok := prop["lotSize"].(float64); ok {
		detail.LotSize = int(lot)
	}
	if year, ok := prop["yearBuilt"].(float64); ok {
		detail.YearBuilt = int(year)
	}
	if z, ok := prop["zestimate"].(float64); ok {
		detail.Zestimate = int64(z)
	}
	if rz, ok := prop["rentZestimate"].(float64); ok {
		detail.RentZestimate = int64(rz)
	}
	if lat, ok := prop["latitude"].(float64); ok {
		detail.Latitude = lat
	}
	if lng, ok := prop["longitude"].(float64); ok {
		detail.Longitude = lng
	}
	if days, ok := prop["daysOnZillow"].(float64); ok {
		detail.DaysOnMarket = int(days)
	}
	if hoa, ok := prop["monthlyHoaFee"].(float64); ok {
		detail.MonthlyHOA = int(hoa)
	}

	// Address
	if addr, ok := prop["address"].(map[string]any); ok {
		detail.StreetAddress = jsonStr(addr, "streetAddress")
		detail.City = jsonStr(addr, "city")
		detail.State = jsonStr(addr, "state")
		detail.Zipcode = jsonStr(addr, "zipcode")
		detail.Address = fmt.Sprintf("%s, %s, %s %s",
			detail.StreetAddress, detail.City, detail.State, detail.Zipcode)
	}

	// Listing agent
	if agent, ok := prop["listingAgent"].(map[string]any); ok {
		detail.ListingAgent = jsonStr(agent, "name")
	}
	detail.ListingBrokerage = jsonStr(prop, "brokerageName")

	// Photos
	if photos, ok := prop["responsivePhotos"].([]any); ok {
		for _, p := range photos {
			if pm, ok := p.(map[string]any); ok {
				if url := jsonStr(pm, "url"); url != "" {
					detail.Photos = append(detail.Photos, url)
				}
			}
		}
	}

	// Price history
	if history, ok := prop["priceHistory"].([]any); ok {
		for _, h := range history {
			if hm, ok := h.(map[string]any); ok {
				pe := PriceEvent{
					Date:   jsonStr(hm, "date"),
					Event:  jsonStr(hm, "event"),
					Source: jsonStr(hm, "source"),
				}
				if price, ok := hm["price"].(float64); ok {
					pe.Price = int64(price)
				}
				detail.PriceHistory = append(detail.PriceHistory, pe)
			}
		}
	}

	// Tax history
	if taxes, ok := prop["taxHistory"].([]any); ok {
		for _, t := range taxes {
			if tm, ok := t.(map[string]any); ok {
				tr := TaxRecord{}
				if year, ok := tm["year"].(float64); ok {
					tr.Year = int(year)
				}
				if paid, ok := tm["taxPaid"].(float64); ok {
					tr.TaxPaid = int64(paid)
				}
				if assessed, ok := tm["taxAssessment"].(float64); ok {
					tr.TaxAssessed = int64(assessed)
				}
				detail.TaxHistory = append(detail.TaxHistory, tr)
			}
		}
	}

	// Schools
	if schools, ok := prop["schools"].([]any); ok {
		for _, s := range schools {
			if sm, ok := s.(map[string]any); ok {
				ss := SchoolSummary{
					Name:   jsonStr(sm, "name"),
					Level:  jsonStr(sm, "level"),
					Type:   jsonStr(sm, "type"),
					Grades: jsonStr(sm, "grades"),
					Link:   jsonStr(sm, "link"),
				}
				if rating, ok := sm["rating"].(float64); ok {
					ss.Rating = int(rating)
				}
				if dist, ok := sm["distance"].(float64); ok {
					ss.Distance = dist
				}
				detail.Schools = append(detail.Schools, ss)
			}
		}
	}

	// Build Zillow URL
	detail.ZillowURL = fmt.Sprintf("https://www.zillow.com/homedetails/%s_zpid/", zpid)

	return detail, nil
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
	// Look for pattern: digits followed by _zpid
	parts := strings.Split(rawURL, "/")
	for _, part := range parts {
		if strings.HasSuffix(part, "_zpid") {
			return strings.TrimSuffix(part, "_zpid")
		}
	}
	return ""
}

// parseNearbyFromGraphQL extracts nearby property summaries from the GraphQL response.
func parseNearbyFromGraphQL(body []byte) ([]PropertySummary, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return nil, nil
	}
	prop, _ := data["property"].(map[string]any)
	if prop == nil {
		return nil, nil
	}

	nearby, _ := prop["nearbyHomes"].([]any)
	var summaries []PropertySummary
	for _, item := range nearby {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		s := PropertySummary{
			ZPID: jsonStr(m, "zpid"),
		}
		if price, ok := m["price"].(float64); ok {
			s.Price = int64(price)
		}
		if beds, ok := m["bedrooms"].(float64); ok {
			s.Beds = int(beds)
		}
		if baths, ok := m["bathrooms"].(float64); ok {
			s.Baths = baths
		}
		if sqft, ok := m["livingArea"].(float64); ok {
			s.Sqft = int(sqft)
		}
		if addr, ok := m["address"].(map[string]any); ok {
			s.Address = fmt.Sprintf("%s, %s, %s %s",
				jsonStr(addr, "streetAddress"),
				jsonStr(addr, "city"),
				jsonStr(addr, "state"),
				jsonStr(addr, "zipcode"))
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

// parsePropertyDetailComps extracts comparable property summaries from a detail response.
func parsePropertyDetailComps(detail PropertyDetail) []PropertySummary {
	// Comps are populated from the GraphQL response during fetchPropertyDetail.
	// For now return empty — they'll be populated when we implement the full parser.
	return nil
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
