package yelp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newAutocompleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "autocomplete",
		Short:   "Get autocomplete suggestions for a search query",
		Aliases: []string{"ac"},
	}

	cmd.AddCommand(newAutocompleteQueryCmd(factory))

	return cmd
}

func newAutocompleteQueryCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Get autocomplete suggestions for a text query",
		RunE:  makeRunAutocompleteQuery(factory),
	}
	cmd.Flags().String("text", "", "Search text to autocomplete")
	cmd.Flags().Float64("latitude", 0, "Latitude for geo-biased results")
	cmd.Flags().Float64("longitude", 0, "Longitude for geo-biased results")
	cmd.Flags().String("locale", "", "Locale code (e.g., en_US)")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func makeRunAutocompleteQuery(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		text, _ := cmd.Flags().GetString("text")
		latitude, _ := cmd.Flags().GetFloat64("latitude")
		longitude, _ := cmd.Flags().GetFloat64("longitude")
		locale, _ := cmd.Flags().GetString("locale")

		params := url.Values{}
		params.Set("text", text)
		if latitude != 0 {
			params.Set("latitude", strconv.FormatFloat(latitude, 'f', -1, 64))
		}
		if longitude != 0 {
			params.Set("longitude", strconv.FormatFloat(longitude, 'f', -1, 64))
		}
		if locale != "" {
			params.Set("locale", locale)
		}

		body, err := client.doYelp(ctx, "GET", "/autocomplete", params)
		if err != nil {
			return fmt.Errorf("autocomplete: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var result AutocompleteResult
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		lines := formatAutocompleteResult(result)
		cli.PrintText(lines)
		return nil
	}
}

// formatAutocompleteResult formats an AutocompleteResult for text output.
func formatAutocompleteResult(r AutocompleteResult) []string {
	lines := []string{}

	if len(r.Terms) > 0 {
		terms := make([]string, 0, len(r.Terms))
		for _, t := range r.Terms {
			terms = append(terms, t.Text)
		}
		lines = append(lines, fmt.Sprintf("Terms:      %s", strings.Join(terms, ", ")))
	}

	if len(r.Businesses) > 0 {
		lines = append(lines, "Businesses:")
		for _, b := range r.Businesses {
			lines = append(lines, fmt.Sprintf("  %-40s  %s", b.Name, b.ID))
		}
	}

	if len(r.Categories) > 0 {
		lines = append(lines, "Categories:")
		for _, c := range r.Categories {
			lines = append(lines, fmt.Sprintf("  %-30s  %s", c.Title, c.Alias))
		}
	}

	if len(lines) == 0 {
		lines = append(lines, "No suggestions found.")
	}

	return lines
}
