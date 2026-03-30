package trends

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// TimePoint represents a single data point in a Google Trends timeline.
type TimePoint struct {
	Date  string `json:"date"`
	Value int    `json:"value"`
}

// CompareResult holds trend data for one keyword in a comparison.
type CompareResult struct {
	Keyword string      `json:"keyword"`
	Data    []TimePoint `json:"data"`
}

// MomentumResult holds the momentum analysis for a keyword.
type MomentumResult struct {
	Keyword     string  `json:"keyword"`
	RecentAvg   float64 `json:"recent_avg"`
	EarlierAvg  float64 `json:"earlier_avg"`
	MomentumPct float64 `json:"momentum_pct"`
	Trend       string  `json:"trend"`
}

// printTimePoints outputs time points as JSON or a formatted text table.
func printTimePoints(cmd *cobra.Command, points []TimePoint) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(points)
	}

	if len(points) == 0 {
		fmt.Println("No data found.")
		return nil
	}

	lines := make([]string, 0, len(points)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-5s", "DATE", "VALUE"))
	for _, p := range points {
		lines = append(lines, fmt.Sprintf("%-20s  %-5d", p.Date, p.Value))
	}
	cli.PrintText(lines)
	return nil
}

// printCompareResults outputs compare results as JSON or a formatted text table.
func printCompareResults(cmd *cobra.Command, results []CompareResult) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(results)
	}

	if len(results) == 0 {
		fmt.Println("No data found.")
		return nil
	}

	// Print each keyword's data as a simple table.
	for _, r := range results {
		fmt.Printf("Keyword: %s\n", r.Keyword)
		lines := make([]string, 0, len(r.Data)+1)
		lines = append(lines, fmt.Sprintf("  %-20s  %-5s", "DATE", "VALUE"))
		for _, p := range r.Data {
			lines = append(lines, fmt.Sprintf("  %-20s  %-5d", p.Date, p.Value))
		}
		cli.PrintText(lines)
	}
	return nil
}

// printMomentumResult outputs a momentum result as JSON or formatted text.
func printMomentumResult(cmd *cobra.Command, result MomentumResult) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(result)
	}

	lines := []string{
		fmt.Sprintf("Keyword:      %s", result.Keyword),
		fmt.Sprintf("Recent avg:   %.1f", result.RecentAvg),
		fmt.Sprintf("Earlier avg:  %.1f", result.EarlierAvg),
		fmt.Sprintf("Momentum:     %.1f%%", result.MomentumPct),
		fmt.Sprintf("Trend:        %s", result.Trend),
	}
	cli.PrintText(lines)
	return nil
}
