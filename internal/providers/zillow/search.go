package zillow

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Search Zillow autocomplete and address lookup",
		Aliases: []string{"find"},
	}

	cmd.AddCommand(newSearchAutocompleteCmd(factory))
	cmd.AddCommand(newSearchByAddressCmd(factory))

	return cmd
}

func newSearchAutocompleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autocomplete",
		Short: "Get autocomplete suggestions for a search query",
		RunE:  makeRunSearchAutocomplete(factory),
	}
	cmd.Flags().String("query", "", "Search query")
	cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunSearchAutocomplete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		query, _ := cmd.Flags().GetString("query")

		reqURL := client.staticURL + "/autocomplete/v3/suggestions?q=" + url.QueryEscape(query)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("autocomplete: %w", err)
		}

		results, err := parseAutocompleteResults(body)
		if err != nil {
			return fmt.Errorf("parse autocomplete: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(results)
		}

		if len(results) == 0 {
			fmt.Println("No suggestions found.")
			return nil
		}
		lines := []string{fmt.Sprintf("Suggestions for %q:", query)}
		for _, r := range results {
			line := fmt.Sprintf("  %-40s  [%s]", truncate(r.Display, 40), r.Type)
			if r.ZPID != "" {
				line += fmt.Sprintf("  zpid:%s", r.ZPID)
			}
			lines = append(lines, line)
		}
		cli.PrintText(lines)
		return nil
	}
}

func newSearchByAddressCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "by-address",
		Short: "Search for a property by address",
		RunE:  makeRunSearchByAddress(factory),
	}
	cmd.Flags().String("address", "", "Property address")
	cmd.MarkFlagRequired("address")
	return cmd
}

func makeRunSearchByAddress(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		address, _ := cmd.Flags().GetString("address")

		// Use autocomplete to resolve address to ZPID, then fetch details
		reqURL := client.staticURL + "/autocomplete/v3/suggestions?q=" + url.QueryEscape(address)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("search by address: %w", err)
		}

		results, err := parseAutocompleteResults(body)
		if err != nil {
			return fmt.Errorf("parse autocomplete: %w", err)
		}

		// Find the first result with a ZPID
		for _, r := range results {
			if r.ZPID != "" {
				detail, err := fetchPropertyDetail(ctx, client, r.ZPID)
				if err != nil {
					return fmt.Errorf("fetch property: %w", err)
				}
				if cli.IsJSONOutput(cmd) {
					return cli.PrintJSON(detail)
				}
				printPropertyDetail(detail)
				return nil
			}
		}

		// If no ZPID found, show autocomplete results
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(results)
		}
		if len(results) == 0 {
			fmt.Println("No results found for that address.")
			return nil
		}
		fmt.Println("No exact property match. Did you mean:")
		for _, r := range results {
			fmt.Printf("  %s [%s]\n", r.Display, r.Type)
		}
		return nil
	}
}

// parseAutocompleteResults extracts results from Zillow's autocomplete API.
func parseAutocompleteResults(body []byte) ([]AutocompleteResult, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	results, _ := resp["results"].([]any)
	var autocomplete []AutocompleteResult
	for _, item := range results {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		r := AutocompleteResult{
			Display: jsonStr(m, "display"),
			Type:    jsonStr(m, "resultType"),
		}
		if meta, ok := m["metaData"].(map[string]any); ok {
			r.ZPID = jsonStr(meta, "zpid")
			if lat, ok := meta["lat"].(float64); ok {
				r.Latitude = lat
			}
			if lng, ok := meta["lng"].(float64); ok {
				r.Longitude = lng
			}
		}
		autocomplete = append(autocomplete, r)
	}
	return autocomplete, nil
}
