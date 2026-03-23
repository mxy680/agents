package zillow

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRentalsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rentals",
		Short:   "Search rental listings",
		Aliases: []string{"rental", "rent"},
	}

	cmd.AddCommand(newRentalSearchCmd(factory))

	return cmd
}

func newRentalSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search rental listings by location",
		RunE:  makeRunRentalSearch(factory),
	}
	cmd.Flags().String("location", "", "Location to search (e.g., 'Denver, CO')")
	cmd.Flags().Int64("min-price", 0, "Minimum rent")
	cmd.Flags().Int64("max-price", 0, "Maximum rent")
	cmd.Flags().Int("min-beds", 0, "Minimum bedrooms")
	cmd.Flags().Int("max-beds", 0, "Maximum bedrooms")
	cmd.Flags().String("home-type", "", "Type: apartment, house, condo, townhouse")
	cmd.Flags().Int("limit", 25, "Maximum results")
	cmd.Flags().Int("page", 1, "Page number")
	cmd.MarkFlagRequired("location")
	return cmd
}

func makeRunRentalSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		location, _ := cmd.Flags().GetString("location")
		minPrice, _ := cmd.Flags().GetInt64("min-price")
		maxPrice, _ := cmd.Flags().GetInt64("max-price")
		minBeds, _ := cmd.Flags().GetInt("min-beds")
		maxBeds, _ := cmd.Flags().GetInt("max-beds")
		homeType, _ := cmd.Flags().GetString("home-type")
		limit, _ := cmd.Flags().GetInt("limit")
		page, _ := cmd.Flags().GetInt("page")

		filterState := buildFilterState("for_rent", minPrice, maxPrice, minBeds, maxBeds, 0, 0, 0, 0, homeType, 0)

		payload := map[string]any{
			"searchQueryState": map[string]any{
				"pagination":      map[string]any{"currentPage": page},
				"usersSearchTerm": location,
				"filterState":     filterState,
			},
			"wants": map[string]any{
				"cat1": []string{"listResults", "mapResults"},
				"cat2": []string{"total"},
			},
			"requestId": 1,
		}

		reqURL := client.baseURL + "/async-create-search-page-state"
		body, err := client.PutJSON(ctx, reqURL, payload)
		if err != nil {
			return fmt.Errorf("rental search: %w", err)
		}

		summaries, err := parseSearchResults(body, limit)
		if err != nil {
			return fmt.Errorf("parse results: %w", err)
		}

		return printPropertySummaries(cmd, summaries)
	}
}
