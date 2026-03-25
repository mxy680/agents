package yelp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- Shared types ---

// BusinessLocation holds address fields returned by the Yelp API.
type BusinessLocation struct {
	Address1   string   `json:"address1"`
	Address2   string   `json:"address2,omitempty"`
	Address3   string   `json:"address3,omitempty"`
	City       string   `json:"city"`
	State      string   `json:"state"`
	ZipCode    string   `json:"zip_code"`
	Country    string   `json:"country"`
	DisplayAddress []string `json:"display_address,omitempty"`
}

// BusinessCategory holds a category alias and display title.
type BusinessCategory struct {
	Alias string `json:"alias"`
	Title string `json:"title"`
}

// BusinessCoordinates holds lat/lng.
type BusinessCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// BusinessHours holds open hours for a business.
type BusinessHours struct {
	HoursType string     `json:"hours_type"`
	Open      []OpenTime `json:"open"`
	IsOpenNow bool       `json:"is_open_now"`
}

// OpenTime is a single open period within a day.
type OpenTime struct {
	IsOvernight bool   `json:"is_overnight"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Day         int    `json:"day"`
}

// BusinessSummary is a simplified business view for list output.
type BusinessSummary struct {
	ID          string            `json:"id"`
	Alias       string            `json:"alias"`
	Name        string            `json:"name"`
	Rating      float64           `json:"rating"`
	ReviewCount int               `json:"review_count"`
	Price       string            `json:"price,omitempty"`
	Phone       string            `json:"phone,omitempty"`
	Location    BusinessLocation  `json:"location"`
	Categories  []BusinessCategory `json:"categories"`
	Distance    float64           `json:"distance,omitempty"`
	IsClosed    bool              `json:"is_closed"`
	URL         string            `json:"url"`
}

// BusinessDetail extends BusinessSummary with additional details.
type BusinessDetail struct {
	BusinessSummary
	Hours         []BusinessHours `json:"hours,omitempty"`
	Photos        []string        `json:"photos,omitempty"`
	IsClaimed     bool            `json:"is_claimed"`
	Coordinates   BusinessCoordinates `json:"coordinates,omitempty"`
}

// ReviewUser holds review author info.
type ReviewUser struct {
	Name     string `json:"name"`
	ImageURL string `json:"image_url,omitempty"`
}

// ReviewSummary is a single Yelp review.
type ReviewSummary struct {
	ID          string     `json:"id"`
	Rating      float64    `json:"rating"`
	Text        string     `json:"text"`
	TimeCreated string     `json:"time_created"`
	User        ReviewUser `json:"user"`
	URL         string     `json:"url"`
}

// EventLocation holds address fields for events.
type EventLocation struct {
	Address1 string `json:"address1"`
	City     string `json:"city"`
	State    string `json:"state"`
	ZipCode  string `json:"zip_code"`
	Country  string `json:"country"`
}

// EventSummary is a single Yelp event.
type EventSummary struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description,omitempty"`
	TimeStart      string        `json:"time_start"`
	TimeEnd        string        `json:"time_end,omitempty"`
	Cost           float64       `json:"cost,omitempty"`
	IsFree         bool          `json:"is_free"`
	AttendingCount int           `json:"attending_count,omitempty"`
	Location       EventLocation `json:"location"`
	EventURL       string        `json:"event_site_url,omitempty"`
}

// CategoryInfo holds Yelp category data.
type CategoryInfo struct {
	Alias         string   `json:"alias"`
	Title         string   `json:"title"`
	ParentAliases []string `json:"parent_aliases,omitempty"`
}

// AutocompleteTerm is a suggested search term.
type AutocompleteTerm struct {
	Text string `json:"text"`
}

// AutocompleteBusinessResult is a business suggested in autocomplete.
type AutocompleteBusinessResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AutocompleteCategory is a category suggested in autocomplete.
type AutocompleteCategory struct {
	Alias string `json:"alias"`
	Title string `json:"title"`
}

// AutocompleteResult holds all autocomplete suggestions.
type AutocompleteResult struct {
	Terms      []AutocompleteTerm           `json:"terms"`
	Businesses []AutocompleteBusinessResult `json:"businesses"`
	Categories []AutocompleteCategory       `json:"categories"`
}

// --- Format helpers ---

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// formatRating formats a float rating as "4.5 stars".
func formatRating(r float64) string {
	return fmt.Sprintf("%.1f", r)
}

// formatAddress formats a BusinessLocation into a single-line address.
func formatAddress(loc BusinessLocation) string {
	if len(loc.DisplayAddress) > 0 {
		return strings.Join(loc.DisplayAddress, ", ")
	}
	parts := []string{}
	if loc.Address1 != "" {
		parts = append(parts, loc.Address1)
	}
	if loc.City != "" {
		parts = append(parts, loc.City)
	}
	if loc.State != "" {
		parts = append(parts, loc.State)
	}
	if loc.ZipCode != "" {
		parts = append(parts, loc.ZipCode)
	}
	return strings.Join(parts, ", ")
}

// formatCategories formats a slice of categories as a comma-separated string.
func formatCategories(cats []BusinessCategory) string {
	names := make([]string, 0, len(cats))
	for _, c := range cats {
		names = append(names, c.Title)
	}
	return strings.Join(names, ", ")
}

// formatDistance formats a distance in meters as "0.5 mi".
func formatDistance(meters float64) string {
	if meters == 0 {
		return "-"
	}
	miles := meters / 1609.344
	return fmt.Sprintf("%.1f mi", miles)
}

// printBusinessSummaries outputs business summaries as JSON or a formatted text table.
func printBusinessSummaries(cmd *cobra.Command, summaries []BusinessSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No businesses found.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-8s  %-35s  %-5s  %-7s  %-5s  %-30s  %-20s",
		"RATING", "NAME", "PRICE", "REVIEWS", "DIST", "ADDRESS", "CATEGORIES"))
	for _, b := range summaries {
		name := truncate(b.Name, 35)
		addr := truncate(formatAddress(b.Location), 30)
		cats := truncate(formatCategories(b.Categories), 20)
		price := b.Price
		if price == "" {
			price = "-"
		}
		closed := ""
		if b.IsClosed {
			closed = " [CLOSED]"
		}
		lines = append(lines, fmt.Sprintf("%-8s  %-35s  %-5s  %-7d  %-5s  %-30s  %-20s%s",
			formatRating(b.Rating), name, price, b.ReviewCount,
			formatDistance(b.Distance), addr, cats, closed))
	}
	cli.PrintText(lines)
	return nil
}

// printReviewSummaries outputs review summaries as JSON or a formatted text table.
func printReviewSummaries(cmd *cobra.Command, reviews []ReviewSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(reviews)
	}

	if len(reviews) == 0 {
		fmt.Println("No reviews found.")
		return nil
	}

	lines := make([]string, 0, len(reviews)+1)
	lines = append(lines, fmt.Sprintf("%-5s  %-20s  %-19s  %s",
		"STARS", "USER", "DATE", "REVIEW"))
	for _, r := range reviews {
		user := truncate(r.User.Name, 20)
		date := ""
		if len(r.TimeCreated) >= 10 {
			date = r.TimeCreated[:10]
		}
		text := truncate(r.Text, 80)
		lines = append(lines, fmt.Sprintf("%-5s  %-20s  %-19s  %s",
			formatRating(r.Rating), user, date, text))
	}
	cli.PrintText(lines)
	return nil
}

// printEventSummaries outputs event summaries as JSON or a formatted text table.
func printEventSummaries(cmd *cobra.Command, events []EventSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(events)
	}

	if len(events) == 0 {
		fmt.Println("No events found.")
		return nil
	}

	lines := make([]string, 0, len(events)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-5s  %-19s  %s",
		"NAME", "FREE", "START", "LOCATION"))
	for _, e := range events {
		name := truncate(e.Name, 40)
		free := "no"
		if e.IsFree {
			free = "yes"
		}
		start := ""
		if len(e.TimeStart) >= 10 {
			start = e.TimeStart[:10]
		}
		loc := strings.Join([]string{e.Location.City, e.Location.State}, ", ")
		lines = append(lines, fmt.Sprintf("%-40s  %-5s  %-19s  %s",
			name, free, start, loc))
	}
	cli.PrintText(lines)
	return nil
}

// --- Search snippet parser ---

// searchSnippetResult represents the structure of Yelp's /search/snippet response.
// The response nests search results under searchPageProps.mainContentComponentsListProps.
type searchSnippetResult struct {
	BizID       string  `json:"bizId"`
	Alias       string  `json:"alias"`
	Name        string  `json:"name"`
	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"reviewCount"`
	PriceRange  string  `json:"priceRange,omitempty"`
	Phone       string  `json:"phone,omitempty"`
	Neighborhoods []string `json:"neighborhoods,omitempty"`
	Address     string  `json:"formattedAddress,omitempty"`
	City        string  `json:"city,omitempty"`
	State       string  `json:"state,omitempty"`
	ZipCode     string  `json:"zipCode,omitempty"`
	Categories  []struct {
		Alias string `json:"alias"`
		Title string `json:"title"`
	} `json:"categories"`
	IsClosed bool   `json:"isClosed"`
	BizURL   string `json:"businessUrl,omitempty"`
}

// parseSearchSnippet extracts BusinessSummary items from Yelp's /search/snippet JSON response.
func parseSearchSnippet(body []byte) ([]BusinessSummary, error) {
	// The snippet response is a complex nested structure. Try multiple extraction strategies.

	// Strategy 1: Look for searchPageProps.mainContentComponentsListProps
	var topLevel map[string]json.RawMessage
	if err := json.Unmarshal(body, &topLevel); err != nil {
		return nil, fmt.Errorf("parse top-level: %w", err)
	}

	// Try to find the search results in the nested structure
	var results []BusinessSummary

	// Check if this is a direct array of businesses (some endpoints)
	if rawProps, ok := topLevel["searchPageProps"]; ok {
		var pageProps map[string]json.RawMessage
		if err := json.Unmarshal(rawProps, &pageProps); err == nil {
			if rawComponents, ok := pageProps["mainContentComponentsListProps"]; ok {
				results = extractBusinessesFromComponents(rawComponents)
			}
		}
	}

	// Fallback: try to find bizId-containing objects anywhere in the response
	if len(results) == 0 {
		results = extractBusinessesFromRaw(body)
	}

	return results, nil
}

// extractBusinessesFromComponents extracts businesses from the mainContentComponentsListProps array.
func extractBusinessesFromComponents(raw json.RawMessage) []BusinessSummary {
	var components []json.RawMessage
	if err := json.Unmarshal(raw, &components); err != nil {
		return nil
	}

	var results []BusinessSummary
	for _, comp := range components {
		var item struct {
			SearchResultLayoutType string              `json:"searchResultLayoutType"`
			BizID                  string              `json:"bizId"`
			SearchResultBusiness   searchSnippetResult `json:"searchResultBusiness"`
		}
		if err := json.Unmarshal(comp, &item); err != nil {
			continue
		}
		if item.SearchResultLayoutType != "iaResult" && item.BizID == "" {
			continue
		}

		biz := item.SearchResultBusiness
		if biz.Name == "" {
			continue
		}

		cats := make([]BusinessCategory, 0, len(biz.Categories))
		for _, c := range biz.Categories {
			cats = append(cats, BusinessCategory{Alias: c.Alias, Title: c.Title})
		}

		results = append(results, BusinessSummary{
			ID:          biz.BizID,
			Alias:       biz.Alias,
			Name:        biz.Name,
			Rating:      biz.Rating,
			ReviewCount: biz.ReviewCount,
			Price:       biz.PriceRange,
			Phone:       biz.Phone,
			Location: BusinessLocation{
				Address1: biz.Address,
				City:     biz.City,
				State:    biz.State,
				ZipCode:  biz.ZipCode,
			},
			Categories: cats,
			IsClosed:   biz.IsClosed,
			URL:        biz.BizURL,
		})
	}
	return results
}

// extractBusinessesFromRaw attempts to find business data in a raw JSON response
// by looking for objects with "bizId" keys.
func extractBusinessesFromRaw(body []byte) []BusinessSummary {
	// Simple heuristic: unmarshal as a generic structure and look for business patterns
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(body, &generic); err != nil {
		return nil
	}

	// Try to find any array that contains objects with business-like fields
	var results []BusinessSummary
	for _, val := range generic {
		var arr []searchSnippetResult
		if err := json.Unmarshal(val, &arr); err != nil {
			continue
		}
		for _, biz := range arr {
			if biz.Name == "" {
				continue
			}
			cats := make([]BusinessCategory, 0, len(biz.Categories))
			for _, c := range biz.Categories {
				cats = append(cats, BusinessCategory{Alias: c.Alias, Title: c.Title})
			}
			results = append(results, BusinessSummary{
				ID:          biz.BizID,
				Alias:       biz.Alias,
				Name:        biz.Name,
				Rating:      biz.Rating,
				ReviewCount: biz.ReviewCount,
				Price:       biz.PriceRange,
				Phone:       biz.Phone,
				Location: BusinessLocation{
					Address1: biz.Address,
					City:     biz.City,
					State:    biz.State,
					ZipCode:  biz.ZipCode,
				},
				Categories: cats,
				IsClosed:   biz.IsClosed,
				URL:        biz.BizURL,
			})
		}
	}
	return results
}

// printCategoryList outputs categories as JSON or a formatted text table.
func printCategoryList(cmd *cobra.Command, cats []CategoryInfo) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(cats)
	}

	if len(cats) == 0 {
		fmt.Println("No categories found.")
		return nil
	}

	lines := make([]string, 0, len(cats)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-40s  %s",
		"ALIAS", "TITLE", "PARENT ALIASES"))
	for _, c := range cats {
		parents := strings.Join(c.ParentAliases, ", ")
		lines = append(lines, fmt.Sprintf("%-40s  %-40s  %s",
			truncate(c.Alias, 40), truncate(c.Title, 40), parents))
	}
	cli.PrintText(lines)
	return nil
}
