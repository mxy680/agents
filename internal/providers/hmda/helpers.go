package hmda

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// countyFIPS maps human-readable NYC county names to their FIPS codes.
var countyFIPS = map[string]string{
	"bronx":        "36005",
	"brooklyn":     "36047",
	"manhattan":    "36061",
	"queens":       "36081",
	"staten-island": "36085",
}

// countyNames maps FIPS codes back to human-readable names.
var countyNames = map[string]string{
	"36005": "Bronx",
	"36047": "Brooklyn",
	"36061": "Manhattan",
	"36081": "Queens",
	"36085": "Staten Island",
}

// CountySummary is the aggregated loan origination result for a county.
type CountySummary struct {
	County      string         `json:"county"`
	FIPS        string         `json:"fips"`
	Year        int            `json:"year"`
	TotalLoans  int            `json:"total_loans"`
	TotalVolume int64          `json:"total_volume"`
	AvgLoanSize int64          `json:"avg_loan_size"`
	TopTracts   []TractSummary `json:"top_tracts"`
}

// TractSummary is the loan origination result for a single census tract.
type TractSummary struct {
	Tract       string `json:"tract"`
	Count       int    `json:"count"`
	Volume      int64  `json:"volume"`
	AvgLoanSize int64  `json:"avg_loan_size"`
}

// AggregationResponse is the raw API response from the HMDA aggregations endpoint.
type AggregationResponse struct {
	Aggregations []AggregationItem `json:"aggregations"`
}

// AggregationItem represents a single row in the HMDA aggregation response.
type AggregationItem struct {
	CensusTract string  `json:"census_tract"`
	Count       int     `json:"count"`
	Sum         float64 `json:"sum"`
}

// printCountySummary outputs a CountySummary as JSON or formatted text.
func printCountySummary(cmd *cobra.Command, summary CountySummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summary)
	}

	lines := []string{
		fmt.Sprintf("County:        %s (%s)", summary.County, summary.FIPS),
		fmt.Sprintf("Year:          %d", summary.Year),
		fmt.Sprintf("Total Loans:   %d", summary.TotalLoans),
		fmt.Sprintf("Total Volume:  %s", formatDollars(summary.TotalVolume)),
		fmt.Sprintf("Avg Loan Size: %s", formatDollars(summary.AvgLoanSize)),
		"",
		fmt.Sprintf("%-16s  %-8s  %-16s  %-16s", "TRACT", "LOANS", "VOLUME", "AVG LOAN"),
	}
	for _, t := range summary.TopTracts {
		lines = append(lines, fmt.Sprintf("%-16s  %-8d  %-16s  %-16s",
			t.Tract, t.Count, formatDollars(t.Volume), formatDollars(t.AvgLoanSize)))
	}
	cli.PrintText(lines)
	return nil
}

// printTractSummary outputs a TractSummary as JSON or formatted text.
func printTractSummary(cmd *cobra.Command, summary TractSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summary)
	}

	lines := []string{
		fmt.Sprintf("Tract:         %s", summary.Tract),
		fmt.Sprintf("Total Loans:   %d", summary.Count),
		fmt.Sprintf("Total Volume:  %s", formatDollars(summary.Volume)),
		fmt.Sprintf("Avg Loan Size: %s", formatDollars(summary.AvgLoanSize)),
	}
	cli.PrintText(lines)
	return nil
}

// formatDollars formats an int64 dollar amount as "$1,234,567".
func formatDollars(amount int64) string {
	if amount == 0 {
		return "$0"
	}
	s := fmt.Sprintf("%d", amount)
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

// lookupFIPS returns the FIPS code for a county name, or an error if not found.
func lookupFIPS(county string) (string, error) {
	fips, ok := countyFIPS[strings.ToLower(county)]
	if !ok {
		var valid []string
		for k := range countyFIPS {
			valid = append(valid, k)
		}
		return "", fmt.Errorf("unknown county %q; valid options: %s", county, strings.Join(valid, ", "))
	}
	return fips, nil
}

// countyNameForFIPS returns the human-readable county name for a FIPS code.
func countyNameForFIPS(fips string) string {
	if name, ok := countyNames[fips]; ok {
		return name
	}
	return fips
}

// buildCountySummary aggregates raw HMDA items into a CountySummary.
// It limits top_tracts to the top 10 by loan count.
func buildCountySummary(county, fips string, year int, items []AggregationItem) CountySummary {
	var totalLoans int
	var totalVolume float64
	var tracts []TractSummary

	for _, item := range items {
		totalLoans += item.Count
		totalVolume += item.Sum

		avg := int64(0)
		if item.Count > 0 {
			avg = int64(item.Sum) / int64(item.Count)
		}
		tracts = append(tracts, TractSummary{
			Tract:       item.CensusTract,
			Count:       item.Count,
			Volume:      int64(item.Sum),
			AvgLoanSize: avg,
		})
	}

	// Sort tracts by count descending and take top 10.
	sortTractsByCount(tracts)
	topTracts := tracts
	if len(topTracts) > 10 {
		topTracts = topTracts[:10]
	}

	avgLoanSize := int64(0)
	if totalLoans > 0 {
		avgLoanSize = int64(totalVolume) / int64(totalLoans)
	}

	return CountySummary{
		County:      county,
		FIPS:        fips,
		Year:        year,
		TotalLoans:  totalLoans,
		TotalVolume: int64(totalVolume),
		AvgLoanSize: avgLoanSize,
		TopTracts:   topTracts,
	}
}

// sortTractsByCount sorts tracts in descending order by Count (insertion sort — small N).
func sortTractsByCount(tracts []TractSummary) {
	for i := 1; i < len(tracts); i++ {
		key := tracts[i]
		j := i - 1
		for j >= 0 && tracts[j].Count < key.Count {
			tracts[j+1] = tracts[j]
			j--
		}
		tracts[j+1] = key
	}
}
