package zillow

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newRentalsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rentals",
		Short:   "Search rental listings",
		Aliases: []string{"rental", "rent"},
	}

	cmd.AddCommand(newRentalSearchCmd(factory))
	cmd.AddCommand(newRentalGetCmd(factory))
	cmd.AddCommand(newRentalEstimateCmd(factory))

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

		url := client.baseURL + "/async-create-search-page-state"
		body, err := client.PutJSON(ctx, url, payload)
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

func newRentalGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get rental listing details by ZPID",
		RunE:  makeRunRentalGet(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunRentalGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")

		detail, err := fetchPropertyDetail(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get rental: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		printPropertyDetail(detail)
		return nil
	}
}

func newRentalEstimateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate",
		Short: "Get Rent Zestimate for a property",
		RunE:  makeRunRentalEstimate(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunRentalEstimate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")

		detail, err := fetchPropertyDetail(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get rental estimate: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"zpid":          detail.ZPID,
				"address":       detail.Address,
				"rentZestimate": detail.RentZestimate,
			})
		}

		lines := []string{
			fmt.Sprintf("ZPID:           %s", detail.ZPID),
			fmt.Sprintf("Address:        %s", detail.Address),
			fmt.Sprintf("Rent Zestimate: %s/mo", formatPrice(detail.RentZestimate)),
		}
		cli.PrintText(lines)
		return nil
	}
}
