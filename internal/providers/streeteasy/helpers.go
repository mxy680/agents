package streeteasy

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// ListingSummary is a simplified listing view for search output.
type ListingSummary struct {
	ID           string `json:"id,omitempty"`
	Address      string `json:"address"`
	Price        int64  `json:"price,omitempty"`
	Beds         int    `json:"beds,omitempty"`
	Baths        float64 `json:"baths,omitempty"`
	Sqft         int    `json:"sqft,omitempty"`
	DaysOnMarket int    `json:"daysOnMarket,omitempty"`
	Status       string `json:"status,omitempty"`
	URL          string `json:"url,omitempty"`
}

// PriceHistoryEntry represents a single price history event for a listing.
type PriceHistoryEntry struct {
	Date  string `json:"date"`
	Event string `json:"event"`
	Price int64  `json:"price,omitempty"`
}

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// formatPrice formats a price as "$1,234,567".
func formatPrice(price int64) string {
	if price == 0 {
		return "-"
	}
	s := fmt.Sprintf("%d", price)
	n := len(s)
	if n <= 3 {
		return "$" + s
	}
	var parts []string
	for n > 3 {
		parts = append([]string{s[n-3 : n]}, parts...)
		n -= 3
	}
	parts = append([]string{s[:n]}, parts...)
	return "$" + strings.Join(parts, ",")
}

// printListingSummaries outputs listing summaries as JSON or a formatted text table.
func printListingSummaries(cmd *cobra.Command, summaries []ListingSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No listings found.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-12s  %-4s  %-5s  %-8s  %-12s",
		"ADDRESS", "PRICE", "BEDS", "BATHS", "SQFT", "STATUS"))
	for _, s := range summaries {
		addr := truncate(s.Address, 40)
		price := formatPrice(s.Price)
		beds := "-"
		if s.Beds > 0 {
			beds = fmt.Sprintf("%d", s.Beds)
		}
		baths := "-"
		if s.Baths > 0 {
			baths = fmt.Sprintf("%.1f", s.Baths)
		}
		sqft := "-"
		if s.Sqft > 0 {
			sqft = fmt.Sprintf("%d", s.Sqft)
		}
		lines = append(lines, fmt.Sprintf("%-40s  %-12s  %-4s  %-5s  %-8s  %-12s",
			addr, price, beds, baths, sqft, s.Status))
	}
	cli.PrintText(lines)
	return nil
}

// printPriceHistory outputs price history as JSON or a formatted text table.
func printPriceHistory(cmd *cobra.Command, entries []PriceHistoryEntry) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(entries)
	}

	if len(entries) == 0 {
		fmt.Println("No price history found.")
		return nil
	}

	lines := make([]string, 0, len(entries)+1)
	lines = append(lines, fmt.Sprintf("%-12s  %-20s  %-12s", "DATE", "EVENT", "PRICE"))
	for _, e := range entries {
		price := formatPrice(e.Price)
		lines = append(lines, fmt.Sprintf("%-12s  %-20s  %-12s", e.Date, e.Event, price))
	}
	cli.PrintText(lines)
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

// jsonInt safely extracts an int from a map.
func jsonInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		}
	}
	return 0
}

// jsonFloat safely extracts a float64 from a map.
func jsonFloat(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		if val, ok := v.(float64); ok {
			return val
		}
	}
	return 0
}
