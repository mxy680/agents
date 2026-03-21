package places

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newLookupCmd(scraper ScraperFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lookup",
		Short: "Get details for a specific Google Maps URL",
		Long: `Scrape full details for a business from its Google Maps URL.

Examples:
  integrations places lookup --url="https://maps.google.com/?cid=12345"
  integrations places lookup --url="https://www.google.com/maps/place/..." --email --json`,
		RunE: makeRunLookup(scraper),
	}
	cmd.Flags().String("url", "", "Google Maps URL (required)")
	cmd.Flags().Bool("email", false, "Extract emails from business website")
	_ = cmd.MarkFlagRequired("url")
	return cmd
}

func makeRunLookup(scraper ScraperFunc) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		url, _ := cmd.Flags().GetString("url")
		email, _ := cmd.Flags().GetBool("email")

		opts := ScraperOptions{
			Queries:     []string{url},
			Depth:       1,
			Email:       email,
			Concurrency: 1,
			Limit:       1,
		}

		result, err := scraper(ctx, opts)
		if err != nil {
			return fmt.Errorf("lookup: %w", err)
		}

		if len(result.Entries) == 0 {
			fmt.Println("No place found at that URL.")
			return nil
		}

		entry := result.Entries[0]

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(entry)
		}

		printEntryDetail(entry)
		return nil
	}
}
