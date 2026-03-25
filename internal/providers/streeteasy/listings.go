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

// jsonLDRe matches the JSON-LD script tag embedded in StreetEasy pages.
var jsonLDRe = regexp.MustCompile(`<script[^>]+type="application/ld\+json"[^>]*>([\s\S]*?)</script>`)

// priceRe matches price strings like "$445,000".
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
		// StreetEasy uses simple area slugs: bronx, brooklyn, manhattan, queens, nyc.
		// For-sale: /for-sale/{area}
		// For-rent: /for-rent/{area}
		var pathPrefix string
		switch status {
		case "for_rent":
			pathPrefix = "/for-rent/"
		default:
			pathPrefix = "/for-sale/"
		}

		areaSlug := locationToSlug(location)
		reqURL := client.baseURL + pathPrefix + url.PathEscape(areaSlug)

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
		areaSlug := locationToSlug(address)
		searchURL := client.baseURL + "/for-sale/" + url.PathEscape(areaSlug)
		body, err := client.Get(ctx, searchURL)
		if err != nil {
			// Also try NYC-wide search as fallback.
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

// locationToSlug converts a location string to a StreetEasy area slug.
// StreetEasy uses simple area names: bronx, brooklyn, manhattan, queens, nyc.
// "Bronx, NY 10452" → "bronx"
// "Brooklyn" → "brooklyn"
// "NYC" → "nyc"
// "Upper West Side" → "upper-west-side"
func locationToSlug(location string) string {
	s := strings.ToLower(strings.TrimSpace(location))

	// Check for known borough/area names first (strip zip codes and state suffixes).
	knownAreas := []string{"bronx", "brooklyn", "manhattan", "queens", "staten-island", "staten island", "nyc", "new york city"}
	for _, area := range knownAreas {
		if strings.Contains(s, strings.ReplaceAll(area, "-", " ")) {
			return strings.ReplaceAll(area, " ", "-")
		}
	}

	// Generic slug: lowercase, remove commas, collapse whitespace to hyphens.
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
	s := strings.ToLower(strings.TrimSpace(address))
	s = strings.ReplaceAll(s, ",", "")
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, "-")
	re2 := regexp.MustCompile(`-+`)
	s = re2.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// jsonLDGraph represents the top-level JSON-LD document structure.
type jsonLDGraph struct {
	Context string       `json:"@context"`
	Graph   []jsonLDItem `json:"@graph"`
}

// jsonLDItem represents a single node in the JSON-LD @graph.
type jsonLDItem struct {
	Type               string           `json:"@type"`
	AdditionalProperty *jsonLDPropValue `json:"additionalProperty,omitempty"`
	Address            *jsonLDAddress   `json:"address,omitempty"`
	Photo              *jsonLDPhoto     `json:"photo,omitempty"`
	URL                string           `json:"url,omitempty"`
	Name               string           `json:"name,omitempty"`
	Description        string           `json:"description,omitempty"`
	NumberOfRooms      any              `json:"numberOfRooms,omitempty"`
}

// jsonLDPropValue represents schema.org PropertyValue.
type jsonLDPropValue struct {
	Type  string `json:"@type"`
	Value string `json:"value"`
}

// jsonLDAddress represents schema.org PostalAddress.
type jsonLDAddress struct {
	Type            string `json:"@type"`
	StreetAddress   string `json:"streetAddress"`
	AddressLocality string `json:"addressLocality"`
	AddressRegion   string `json:"addressRegion"`
	PostalCode      string `json:"postalCode"`
}

// jsonLDPhoto represents a schema.org CreativeWork photo.
type jsonLDPhoto struct {
	Type  string `json:"@type"`
	Image string `json:"image"`
}

// extractJSONLD parses the JSON-LD @graph from an HTML page body.
// Returns the parsed graph items, or an error if no JSON-LD script tag is found.
// An empty @graph is valid and returns an empty slice without error.
func extractJSONLD(body []byte) ([]jsonLDItem, error) {
	matches := jsonLDRe.FindAllSubmatch(body, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no JSON-LD script tag found in page")
	}

	// Try each JSON-LD block — use the first one that parses as a @graph document.
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		var doc jsonLDGraph
		if err := json.Unmarshal(match[1], &doc); err != nil {
			continue
		}
		// A document with a @graph key is a valid graph document even if empty.
		if doc.Graph != nil {
			return doc.Graph, nil
		}
	}

	return nil, fmt.Errorf("no @graph array found in JSON-LD script tags")
}

// parseListingsFromPage extracts ListingSummary values from a StreetEasy page's
// JSON-LD <script type="application/ld+json"> tag.
func parseListingsFromPage(body []byte, baseURL string, limit int) ([]ListingSummary, error) {
	items, err := extractJSONLD(body)
	if err != nil {
		return nil, err
	}

	var summaries []ListingSummary
	for _, item := range items {
		// JSON-LD listing items are typed as ApartmentComplex, SingleFamilyResidence, etc.
		// Skip items that are clearly not listings (e.g. WebSite, Organization).
		if item.Type == "WebSite" || item.Type == "Organization" || item.Type == "BreadcrumbList" {
			continue
		}

		s := jsonLDItemToListingSummary(item, baseURL)
		if s.Address == "" && s.Price == 0 {
			continue
		}
		summaries = append(summaries, s)
		if limit > 0 && len(summaries) >= limit {
			break
		}
	}

	return summaries, nil
}

// jsonLDItemToListingSummary converts a JSON-LD graph item to a ListingSummary.
func jsonLDItemToListingSummary(item jsonLDItem, baseURL string) ListingSummary {
	s := ListingSummary{}

	// Build address from schema.org PostalAddress fields.
	if item.Address != nil {
		var parts []string
		if item.Address.StreetAddress != "" {
			parts = append(parts, item.Address.StreetAddress)
		}
		if item.Address.AddressLocality != "" {
			parts = append(parts, item.Address.AddressLocality)
		}
		if item.Address.AddressRegion != "" {
			parts = append(parts, item.Address.AddressRegion)
		}
		s.Address = strings.Join(parts, ", ")
	}

	// Price from additionalProperty.value (e.g. "$445,000").
	if item.AdditionalProperty != nil && item.AdditionalProperty.Value != "" {
		if p, err := parseRawPrice(item.AdditionalProperty.Value); err == nil {
			s.Price = p
		}
	}

	// URL from the item's url field.
	if item.URL != "" {
		if strings.HasPrefix(item.URL, "http") {
			s.URL = item.URL
		} else {
			s.URL = baseURL + item.URL
		}
	}

	return s
}

// extractListingSummary maps a raw listing map to a ListingSummary.
// Used by tests that pass raw map data.
func extractListingSummary(m map[string]any, baseURL string) ListingSummary {
	s := ListingSummary{
		ID:     jsonStr(m, "id"),
		Status: jsonStr(m, "status"),
	}

	// Address may be nested or flat.
	if addr := jsonStr(m, "address"); addr != "" {
		s.Address = addr
	} else if addrMap, ok := m["address"].(map[string]any); ok {
		var parts []string
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

// parsePriceHistory extracts price history from a StreetEasy listing detail page.
// NOTE: StreetEasy listing detail pages embed price history in JavaScript state
// that is not available in the JSON-LD tag. Parsing requires a dedicated listing
// detail page scraper. This function returns an empty array for now.
func parsePriceHistory(body []byte) ([]PriceHistoryEntry, error) {
	// Price history is not available in the JSON-LD schema.org data.
	// It would require parsing a different data structure from the listing detail page.
	return []PriceHistoryEntry{}, nil
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

// extractFirstListingURL finds the first listing URL from the JSON-LD data on a page.
func extractFirstListingURL(body []byte, baseURL string) string {
	items, err := extractJSONLD(body)
	if err != nil {
		return ""
	}

	for _, item := range items {
		if item.Type == "WebSite" || item.Type == "Organization" || item.Type == "BreadcrumbList" {
			continue
		}
		if item.URL != "" {
			if strings.HasPrefix(item.URL, "http") {
				return item.URL
			}
			return baseURL + item.URL
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
