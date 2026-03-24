package streeteasy

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// nextDataRe matches the __NEXT_DATA__ JSON script tag embedded in StreetEasy pages.
var nextDataRe = regexp.MustCompile(`<script id="__NEXT_DATA__"[^>]*>([\s\S]*?)</script>`)

// priceHistorySectionRe is used to locate the price history JSON block in page HTML.
// StreetEasy embeds listing data in __NEXT_DATA__ as well.
var priceRe = regexp.MustCompile(`\$[\d,]+`)

func newListingsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "listings",
		Short:   "Search and view StreetEasy listings",
		Aliases: []string{"listing", "lst"},
	}

	cmd.AddCommand(newListingsSearchCmd(factory))
	cmd.AddCommand(newListingsHistoryCmd(factory))

	return cmd
}

// newListingsSearchCmd returns the `listings search` subcommand.
func newListingsSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search StreetEasy listings by location",
		RunE:  makeRunListingsSearch(factory),
	}
	cmd.Flags().String("location", "", "Location to search (e.g., 'Bronx, NY 10452' or 'nyc')")
	cmd.Flags().String("status", "for_sale", "Listing status: for_sale, for_rent")
	cmd.Flags().Int("limit", 25, "Maximum results")
	cmd.MarkFlagRequired("location")
	return cmd
}

func makeRunListingsSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		location, _ := cmd.Flags().GetString("location")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")

		// Build StreetEasy search URL.
		// For-sale: /for-sale/{location}
		// For-rent: /for-rent/{location}
		var pathPrefix string
		switch status {
		case "for_rent":
			pathPrefix = "/for-rent/"
		default:
			pathPrefix = "/for-sale/"
		}

		// StreetEasy uses kebab-case URLs: "Bronx, NY 10452" → "bronx-ny-10452"
		slug := locationToSlug(location)
		reqURL := client.baseURL + pathPrefix + url.PathEscape(slug)

		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("search listings: %w", err)
		}

		summaries, err := parseListingsFromPage(body, client.baseURL, limit)
		if err != nil {
			return fmt.Errorf("parse listings: %w", err)
		}

		return printListingSummaries(cmd, summaries)
	}
}

// newListingsHistoryCmd returns the `listings history` subcommand.
func newListingsHistoryCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Get price history for a StreetEasy listing by address",
		RunE:  makeRunListingsHistory(factory),
	}
	cmd.Flags().String("address", "", "Property address (e.g., '1226 Shakespeare Ave Bronx NY')")
	cmd.MarkFlagRequired("address")
	return cmd
}

func makeRunListingsHistory(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		address, _ := cmd.Flags().GetString("address")

		// Step 1: search for the listing by address to get its URL.
		slug := locationToSlug(address)
		searchURL := client.baseURL + "/for-sale/" + url.PathEscape(slug)
		body, err := client.Get(ctx, searchURL)
		if err != nil {
			// Also try NYC-wide search as fallback
			searchURL = client.baseURL + "/for-sale/nyc?q=" + url.QueryEscape(address)
			body, err = client.Get(ctx, searchURL)
			if err != nil {
				return fmt.Errorf("search for address: %w", err)
			}
		}

		// Try to find the first listing URL from search results.
		listingURL := extractFirstListingURL(body, client.baseURL)
		if listingURL == "" {
			// If no listing found in search, try the address as a direct URL slug.
			slug2 := addressToURLSlug(address)
			listingURL = client.baseURL + "/building/" + slug2
		}

		// Step 2: fetch the listing page.
		listingBody, err := client.Get(ctx, listingURL)
		if err != nil {
			return fmt.Errorf("fetch listing page: %w", err)
		}

		// Step 3: parse price history from the listing page.
		entries, err := parsePriceHistory(listingBody)
		if err != nil {
			return fmt.Errorf("parse price history: %w", err)
		}

		return printPriceHistory(cmd, entries)
	}
}

// locationToSlug converts a location string to a StreetEasy URL slug.
// "Bronx, NY 10452" → "bronx-ny-10452"
// "Upper West Side" → "upper-west-side"
func locationToSlug(location string) string {
	// Lowercase, replace commas and spaces with hyphens, collapse multiple hyphens.
	s := strings.ToLower(location)
	s = strings.ReplaceAll(s, ",", "")
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, "-")
	re2 := regexp.MustCompile(`-+`)
	s = re2.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// addressToURLSlug converts an address to a building URL slug (best-effort).
func addressToURLSlug(address string) string {
	return locationToSlug(address)
}

// parseListingsFromPage extracts ListingSummary values from a StreetEasy page's
// __NEXT_DATA__ JSON blob.
func parseListingsFromPage(body []byte, baseURL string, limit int) ([]ListingSummary, error) {
	nextData, err := extractNextData(body)
	if err != nil {
		return nil, err
	}

	// Navigate: props → pageProps → listings (array) or searchListings
	props, _ := nextData["props"].(map[string]any)
	if props == nil {
		return nil, nil
	}
	pageProps, _ := props["pageProps"].(map[string]any)
	if pageProps == nil {
		return nil, nil
	}

	// StreetEasy may store listings under different keys depending on page type.
	// Try common keys: listings, searchResults, homes
	var rawListings []any
	for _, key := range []string{"listings", "searchResults", "homes", "results"} {
		if v, ok := pageProps[key]; ok {
			if arr, ok := v.([]any); ok {
				rawListings = arr
				break
			}
		}
	}

	// Also try pageProps → data → listings
	if rawListings == nil {
		if data, ok := pageProps["data"].(map[string]any); ok {
			for _, key := range []string{"listings", "searchResults", "homes"} {
				if v, ok := data[key]; ok {
					if arr, ok := v.([]any); ok {
						rawListings = arr
						break
					}
				}
			}
		}
	}

	var summaries []ListingSummary
	for _, item := range rawListings {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		s := extractListingSummary(m, baseURL)
		summaries = append(summaries, s)
		if limit > 0 && len(summaries) >= limit {
			break
		}
	}

	return summaries, nil
}

// extractListingSummary maps a raw listing map to a ListingSummary.
func extractListingSummary(m map[string]any, baseURL string) ListingSummary {
	s := ListingSummary{
		ID:     jsonStr(m, "id"),
		Status: jsonStr(m, "status"),
	}

	// Address may be nested or flat.
	if addr := jsonStr(m, "address"); addr != "" {
		s.Address = addr
	} else if addrMap, ok := m["address"].(map[string]any); ok {
		parts := []string{}
		for _, k := range []string{"streetAddress", "street", "line1"} {
			if v := jsonStr(addrMap, k); v != "" {
				parts = append(parts, v)
			}
		}
		for _, k := range []string{"city", "borough"} {
			if v := jsonStr(addrMap, k); v != "" {
				parts = append(parts, v)
			}
		}
		if state := jsonStr(addrMap, "state"); state != "" {
			parts = append(parts, state)
		}
		s.Address = strings.Join(parts, ", ")
	}

	// Price
	if price := jsonInt(m, "price"); price > 0 {
		s.Price = int64(price)
	} else if priceStr := jsonStr(m, "price"); priceStr != "" {
		if p, err := parseRawPrice(priceStr); err == nil {
			s.Price = p
		}
	}

	// Beds/baths/sqft
	s.Beds = jsonInt(m, "bedrooms")
	if s.Beds == 0 {
		s.Beds = jsonInt(m, "beds")
	}
	s.Baths = jsonFloat(m, "bathrooms")
	if s.Baths == 0 {
		s.Baths = jsonFloat(m, "baths")
	}
	s.Sqft = jsonInt(m, "sqft")
	if s.Sqft == 0 {
		s.Sqft = jsonInt(m, "squareFeet")
	}

	s.DaysOnMarket = jsonInt(m, "daysOnMarket")

	// URL
	if listingURL := jsonStr(m, "url"); listingURL != "" {
		if strings.HasPrefix(listingURL, "http") {
			s.URL = listingURL
		} else {
			s.URL = baseURL + listingURL
		}
	} else if slug := jsonStr(m, "slug"); slug != "" {
		s.URL = baseURL + "/" + slug
	}

	return s
}

// parsePriceHistory extracts price history from a StreetEasy listing page.
func parsePriceHistory(body []byte) ([]PriceHistoryEntry, error) {
	nextData, err := extractNextData(body)
	if err != nil {
		return nil, err
	}

	// Navigate into the page data for price history.
	// Common path: props → pageProps → listing → priceHistory
	// or: props → pageProps → priceHistory
	props, _ := nextData["props"].(map[string]any)
	if props == nil {
		return nil, nil
	}
	pageProps, _ := props["pageProps"].(map[string]any)
	if pageProps == nil {
		return nil, nil
	}

	var rawHistory []any
	for _, path := range [][]string{
		{"priceHistory"},
		{"listing", "priceHistory"},
		{"data", "priceHistory"},
		{"home", "priceHistory"},
	} {
		rawHistory = navigatePath(pageProps, path)
		if rawHistory != nil {
			break
		}
	}

	var entries []PriceHistoryEntry
	for _, item := range rawHistory {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		e := PriceHistoryEntry{
			Date:  jsonStr(m, "date"),
			Event: jsonStr(m, "event"),
		}
		if e.Event == "" {
			e.Event = jsonStr(m, "eventType")
		}
		if price := jsonInt(m, "price"); price > 0 {
			e.Price = int64(price)
		} else if priceStr := jsonStr(m, "price"); priceStr != "" {
			if p, err := parseRawPrice(priceStr); err == nil {
				e.Price = p
			}
		}
		if e.Date != "" || e.Price > 0 {
			entries = append(entries, e)
		}
	}

	return entries, nil
}

// navigatePath traverses a nested map following the given key path and returns
// the []any at the leaf, or nil if not found.
func navigatePath(m map[string]any, path []string) []any {
	if len(path) == 0 {
		return nil
	}
	if len(path) == 1 {
		if v, ok := m[path[0]]; ok {
			if arr, ok := v.([]any); ok {
				return arr
			}
		}
		return nil
	}
	next, ok := m[path[0]].(map[string]any)
	if !ok {
		return nil
	}
	return navigatePath(next, path[1:])
}

// extractNextData parses the __NEXT_DATA__ JSON blob from an HTML page body.
func extractNextData(body []byte) (map[string]any, error) {
	matches := nextDataRe.FindSubmatch(body)
	if matches == nil || len(matches) < 2 {
		return nil, fmt.Errorf("__NEXT_DATA__ not found in page (PerimeterX may have blocked the request)")
	}

	var data map[string]any
	if err := json.Unmarshal(matches[1], &data); err != nil {
		return nil, fmt.Errorf("parse __NEXT_DATA__ JSON: %w", err)
	}

	return data, nil
}

// extractFirstListingURL finds the first listing detail URL in the page HTML.
// StreetEasy listing URLs look like /nyc/real_estate/12345678 or /building/...
func extractFirstListingURL(body []byte, baseURL string) string {
	// Look for detail page links in __NEXT_DATA__
	nextData, err := extractNextData(body)
	if err != nil {
		return ""
	}

	props, _ := nextData["props"].(map[string]any)
	if props == nil {
		return ""
	}
	pageProps, _ := props["pageProps"].(map[string]any)
	if pageProps == nil {
		return ""
	}

	// Try to find any listings array and get the URL of the first item.
	for _, key := range []string{"listings", "searchResults", "homes", "results"} {
		if v, ok := pageProps[key]; ok {
			if arr, ok := v.([]any); ok && len(arr) > 0 {
				if m, ok := arr[0].(map[string]any); ok {
					if u := jsonStr(m, "url"); u != "" {
						if strings.HasPrefix(u, "http") {
							return u
						}
						return baseURL + u
					}
					if slug := jsonStr(m, "slug"); slug != "" {
						return baseURL + "/" + slug
					}
				}
			}
		}
	}

	return ""
}

// parseRawPrice strips "$", commas from a price string and parses to int64.
func parseRawPrice(s string) (int64, error) {
	// Handle formatted strings like "$1,250,000" or "1250000"
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	return strconv.ParseInt(s, 10, 64)
}
