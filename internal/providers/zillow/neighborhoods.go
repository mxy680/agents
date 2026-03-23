package zillow

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newNeighborhoodsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "neighborhoods",
		Short:   "View neighborhood data and market stats",
		Aliases: []string{"neighborhood", "hood"},
	}

	cmd.AddCommand(newNeighborhoodGetCmd(factory))
	cmd.AddCommand(newNeighborhoodSearchCmd(factory))
	cmd.AddCommand(newNeighborhoodMarketStatsCmd(factory))

	return cmd
}

func newNeighborhoodGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get neighborhood details by region ID",
		RunE:  makeRunNeighborhoodGet(factory),
	}
	cmd.Flags().String("region-id", "", "Zillow region ID")
	cmd.MarkFlagRequired("region-id")
	return cmd
}

func makeRunNeighborhoodGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		regionID, _ := cmd.Flags().GetString("region-id")

		reqURL := fmt.Sprintf("%s/graphql/?regionId=%s", client.baseURL, regionID)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get neighborhood: %w", err)
		}

		var resp map[string]any
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(resp)
		}

		data, _ := resp["data"].(map[string]any)
		if data == nil {
			fmt.Println("Neighborhood not found.")
			return nil
		}
		region, _ := data["region"].(map[string]any)
		if region == nil {
			fmt.Println("Neighborhood not found.")
			return nil
		}

		lines := []string{
			fmt.Sprintf("Name:    %s", jsonStr(region, "name")),
			fmt.Sprintf("Type:    %s", jsonStr(region, "type")),
		}
		if mhv, ok := region["medianHomeValue"].(float64); ok && mhv > 0 {
			lines = append(lines, fmt.Sprintf("Median Home Value: %s", formatPrice(int64(mhv))))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newNeighborhoodSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for neighborhoods by location",
		RunE:  makeRunNeighborhoodSearch(factory),
	}
	cmd.Flags().String("location", "", "Location to search")
	cmd.MarkFlagRequired("location")
	return cmd
}

func makeRunNeighborhoodSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		location, _ := cmd.Flags().GetString("location")

		// Use autocomplete API to find neighborhood regions
		reqURL := client.staticURL + "/autocomplete/v3/suggestions?q=" + url.QueryEscape(location)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("search neighborhoods: %w", err)
		}

		results, err := parseAutocompleteResults(body)
		if err != nil {
			return fmt.Errorf("parse results: %w", err)
		}

		// Filter to neighborhood-type results
		var neighborhoods []AutocompleteResult
		for _, r := range results {
			neighborhoods = append(neighborhoods, r)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(neighborhoods)
		}

		if len(neighborhoods) == 0 {
			fmt.Println("No neighborhoods found.")
			return nil
		}
		lines := []string{fmt.Sprintf("Neighborhoods matching %q:", location)}
		for _, n := range neighborhoods {
			lines = append(lines, fmt.Sprintf("  %-40s  [%s]", truncate(n.Display, 40), n.Type))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newNeighborhoodMarketStatsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market-stats",
		Short: "Get market statistics for a neighborhood",
		RunE:  makeRunNeighborhoodMarketStats(factory),
	}
	cmd.Flags().String("region-id", "", "Zillow region ID")
	cmd.MarkFlagRequired("region-id")
	return cmd
}

func makeRunNeighborhoodMarketStats(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		regionID, _ := cmd.Flags().GetString("region-id")

		reqURL := fmt.Sprintf("%s/graphql/?regionId=%s&marketStats=true", client.baseURL, regionID)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get market stats: %w", err)
		}

		var resp map[string]any
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(resp)
		}

		data, _ := resp["data"].(map[string]any)
		if data == nil {
			fmt.Println("No market stats available.")
			return nil
		}
		region, _ := data["region"].(map[string]any)
		if region == nil {
			fmt.Println("No market stats available.")
			return nil
		}

		lines := []string{fmt.Sprintf("Market Stats for %s:", jsonStr(region, "name"))}
		if mhv, ok := region["medianHomeValue"].(float64); ok {
			lines = append(lines, fmt.Sprintf("  Median Home Value:  %s", formatPrice(int64(mhv))))
		}
		if mr, ok := region["medianRent"].(float64); ok {
			lines = append(lines, fmt.Sprintf("  Median Rent:        %s/mo", formatPrice(int64(mr))))
		}
		if mls, ok := region["medianListPrice"].(float64); ok {
			lines = append(lines, fmt.Sprintf("  Median List Price:  %s", formatPrice(int64(mls))))
		}
		cli.PrintText(lines)
		return nil
	}
}
