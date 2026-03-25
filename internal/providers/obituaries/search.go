package obituaries

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

// validDateRanges is the set of accepted values for --date-range.
var validDateRanges = map[string]bool{
	"Last7Days":    true,
	"Last30Days":   true,
	"Last6Months":  true,
	"Last12Months": true,
}

// newSearchCmd returns the `obituaries search` subcommand.
func newSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Search obituaries by location",
		Aliases: []string{"s"},
		RunE:    makeRunSearch(factory),
	}
	cmd.Flags().String("city", "", "City to search (required)")
	cmd.Flags().String("state", "New York", "State to search")
	cmd.Flags().String("date-range", "Last30Days", "Date range: Last7Days, Last30Days, Last6Months, Last12Months")
	cmd.Flags().String("keyword", "", "Optional keyword filter")
	cmd.Flags().Int("limit", 50, "Maximum number of results to return")
	cmd.Flags().Int("page", 1, "Page number")
	if err := cmd.MarkFlagRequired("city"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		city, _ := cmd.Flags().GetString("city")
		state, _ := cmd.Flags().GetString("state")
		dateRange, _ := cmd.Flags().GetString("date-range")
		keyword, _ := cmd.Flags().GetString("keyword")
		limit, _ := cmd.Flags().GetInt("limit")
		page, _ := cmd.Flags().GetInt("page")

		if !validDateRanges[dateRange] {
			return fmt.Errorf("invalid --date-range %q; valid options: Last7Days, Last30Days, Last6Months, Last12Months", dateRange)
		}

		params := url.Values{
			"city":      []string{city},
			"state":     []string{state},
			"dateRange": []string{dateRange},
			"keyword":   []string{keyword},
			"page":      []string{fmt.Sprintf("%d", page)},
		}

		body, err := client.Get(ctx, "search", params)
		if err != nil {
			return fmt.Errorf("search obituaries: %w", err)
		}

		var resp SearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		summaries := toSummaries(resp.Obituaries, limit)
		return printObituaries(cmd, summaries)
	}
}

// newNamesCmd returns the `obituaries names` subcommand.
func newNamesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "names",
		Short:   "Extract names from obituaries for ACRIS cross-referencing",
		Long:    "Returns first, last, and full names plus publish date. Pipe these into ACRIS party search to find properties owned by the deceased.",
		Aliases: []string{"n"},
		RunE:    makeRunNames(factory),
	}
	cmd.Flags().String("city", "", "City to search (required)")
	cmd.Flags().String("state", "New York", "State to search")
	cmd.Flags().String("date-range", "Last30Days", "Date range: Last7Days, Last30Days, Last6Months, Last12Months")
	cmd.Flags().String("keyword", "", "Optional keyword filter")
	cmd.Flags().Int("limit", 50, "Maximum number of results to return")
	cmd.Flags().Int("page", 1, "Page number")
	if err := cmd.MarkFlagRequired("city"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunNames(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		city, _ := cmd.Flags().GetString("city")
		state, _ := cmd.Flags().GetString("state")
		dateRange, _ := cmd.Flags().GetString("date-range")
		keyword, _ := cmd.Flags().GetString("keyword")
		limit, _ := cmd.Flags().GetInt("limit")
		page, _ := cmd.Flags().GetInt("page")

		if !validDateRanges[dateRange] {
			return fmt.Errorf("invalid --date-range %q; valid options: Last7Days, Last30Days, Last6Months, Last12Months", dateRange)
		}

		params := url.Values{
			"city":      []string{city},
			"state":     []string{state},
			"dateRange": []string{dateRange},
			"keyword":   []string{keyword},
			"page":      []string{fmt.Sprintf("%d", page)},
		}

		body, err := client.Get(ctx, "search", params)
		if err != nil {
			return fmt.Errorf("search obituaries: %w", err)
		}

		var resp SearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		entries := toNameEntries(resp.Obituaries, limit)
		return printNames(cmd, entries)
	}
}

// toSummaries converts raw obituaries to summaries, applying a limit.
func toSummaries(raw []rawObituary, limit int) []ObituarySummary {
	result := make([]ObituarySummary, 0, min(len(raw), limit))
	for i, r := range raw {
		if i >= limit {
			break
		}
		result = append(result, r.toSummary())
	}
	return result
}

// toNameEntries converts raw obituaries to name entries, applying a limit.
func toNameEntries(raw []rawObituary, limit int) []NameEntry {
	result := make([]NameEntry, 0, min(len(raw), limit))
	for i, r := range raw {
		if i >= limit {
			break
		}
		result = append(result, r.toNameEntry())
	}
	return result
}

// min returns the smaller of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
