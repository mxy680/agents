package places

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newSearchCmd(scraper ScraperFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search Google Maps for businesses and places",
		Long: `Search Google Maps by text query. Returns rich data including address, phone,
hours, reviews, ratings, and optionally emails.

Examples:
  integrations places search --query="coffee shops in Cleveland OH"
  integrations places search --query="dentist" --geo=41.499,-81.694 --zoom=14
  integrations places search --query="florists near 90210" --email --limit=10 --json`,
		Aliases: []string{"find"},
		RunE:    makeRunSearch(scraper),
	}
	cmd.Flags().String("query", "", "Search query, e.g. 'coffee shops in Cleveland' (required)")
	cmd.Flags().String("geo", "", "Geo-target: lat,lng (e.g. '41.499,-81.694')")
	cmd.Flags().Int("zoom", 0, "Google Maps zoom level (affects search area, 1-21)")
	cmd.Flags().Int("depth", 1, "Pagination depth (1=first page ~20 results, 2+=more)")
	cmd.Flags().Bool("email", false, "Extract emails from business websites (slower)")
	cmd.Flags().Int("concurrency", 1, "Number of concurrent scrapers")
	cmd.Flags().String("lang", "", "Language code (e.g. en, es, fr)")
	cmd.Flags().Int("limit", 20, "Maximum results to return")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunSearch(scraper ScraperFunc) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		query, _ := cmd.Flags().GetString("query")
		geo, _ := cmd.Flags().GetString("geo")
		zoom, _ := cmd.Flags().GetInt("zoom")
		depth, _ := cmd.Flags().GetInt("depth")
		email, _ := cmd.Flags().GetBool("email")
		concurrency, _ := cmd.Flags().GetInt("concurrency")
		lang, _ := cmd.Flags().GetString("lang")
		limit, _ := cmd.Flags().GetInt("limit")

		// Validate geo format if provided
		if geo != "" {
			if _, _, err := parseLatLng(geo); err != nil {
				return fmt.Errorf("invalid --geo: %w", err)
			}
		}

		opts := ScraperOptions{
			Queries:     []string{query},
			Geo:         geo,
			Zoom:        zoom,
			Depth:       depth,
			Email:       email,
			Concurrency: concurrency,
			Lang:        lang,
			Limit:       limit,
		}

		result, err := scraper(ctx, opts)
		if err != nil {
			return fmt.Errorf("search: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result.Entries)
		}

		if len(result.Entries) == 0 {
			fmt.Println("No places found.")
			return nil
		}

		summaries := make([]PlaceSummary, 0, len(result.Entries))
		for _, e := range result.Entries {
			summaries = append(summaries, toPlaceSummary(e))
		}
		return printPlaceSummaries(cmd, summaries)
	}
}
