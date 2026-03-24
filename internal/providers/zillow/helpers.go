package zillow

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// PropertySummary is a simplified property view for list output.
type PropertySummary struct {
	ZPID         string  `json:"zpid"`
	Address      string  `json:"address"`
	Price        int64   `json:"price,omitempty"`
	Beds         int     `json:"beds,omitempty"`
	Baths        float64 `json:"baths,omitempty"`
	Sqft         int     `json:"sqft,omitempty"`
	HomeType     string  `json:"homeType,omitempty"`
	Status       string  `json:"status,omitempty"`
	ZillowURL    string  `json:"zillowUrl,omitempty"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
	DaysOnMarket int     `json:"daysOnMarket,omitempty"`
}

// AutocompleteResult is a single autocomplete suggestion.
type AutocompleteResult struct {
	Display   string  `json:"display"`
	ZPID      string  `json:"zpid,omitempty"`
	Type      string  `json:"type,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	URL       string  `json:"url,omitempty"`
}

// MortgageRate holds mortgage rate data.
type MortgageRate struct {
	Program  string  `json:"program"`
	Rate     float64 `json:"rate"`
	APR      float64 `json:"apr,omitempty"`
	LoanType string  `json:"loanType,omitempty"`
	State    string  `json:"state,omitempty"`
	Date     string  `json:"date,omitempty"`
}

// MortgageCalculation holds mortgage payment calculation results.
type MortgageCalculation struct {
	MonthlyPayment float64 `json:"monthlyPayment"`
	Principal      float64 `json:"principal"`
	Interest       float64 `json:"interest"`
	Tax            float64 `json:"tax,omitempty"`
	Insurance      float64 `json:"insurance,omitempty"`
	TotalCost      float64 `json:"totalCost"`
}

// LenderReview is a single lender review from Zillow's mortgage API.
type LenderReview struct {
	Rating                   float64 `json:"rating"`
	Title                    string  `json:"title,omitempty"`
	Content                  string  `json:"content,omitempty"`
	LoanType                 string  `json:"loanType,omitempty"`
	LoanProgram              string  `json:"loanProgram,omitempty"`
	ClosingCostsSatisfaction float64 `json:"closingCostsSatisfaction,omitempty"`
	InterestRateSatisfaction float64 `json:"interestRateSatisfaction,omitempty"`
	VerifiedReviewer         bool    `json:"verifiedReviewer,omitempty"`
}

// LenderInfo holds lender overview data returned by the reviews endpoint.
type LenderInfo struct {
	ProfileURL   string        `json:"profileUrl,omitempty"`
	ReviewURL    string        `json:"reviewUrl,omitempty"`
	TotalReviews int           `json:"totalReviews,omitempty"`
	Rating       float64       `json:"rating,omitempty"`
	Reviews      []LenderReview `json:"reviews,omitempty"`
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
	// Insert commas
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

// printPropertySummaries outputs property summaries as JSON or a formatted text table.
func printPropertySummaries(cmd *cobra.Command, summaries []PropertySummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No properties found.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-12s  %-40s  %-12s  %-4s  %-5s  %-8s  %-12s",
		"ZPID", "ADDRESS", "PRICE", "BEDS", "BATHS", "SQFT", "STATUS"))
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
		lines = append(lines, fmt.Sprintf("%-12s  %-40s  %-12s  %-4s  %-5s  %-8s  %-12s",
			s.ZPID, addr, price, beds, baths, sqft, s.Status))
	}
	cli.PrintText(lines)
	return nil
}
