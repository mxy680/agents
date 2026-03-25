package yelp

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newCategoriesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "categories",
		Short:   "List and get Yelp categories",
		Aliases: []string{"category", "cat"},
	}

	cmd.AddCommand(newCategoryListCmd(factory))
	cmd.AddCommand(newCategoryGetCmd(factory))

	return cmd
}

func newCategoryListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all Yelp categories",
		RunE:  makeRunCategoryList(factory),
	}
	cmd.Flags().String("locale", "", "Locale code (e.g., en_US)")
	return cmd
}

func makeRunCategoryList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		locale, _ := cmd.Flags().GetString("locale")

		params := url.Values{}
		if locale != "" {
			params.Set("locale", locale)
		}

		body, err := client.doYelp(ctx, "GET", "/categories", params)
		if err != nil {
			return fmt.Errorf("list categories: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var resp struct {
			Categories []CategoryInfo `json:"categories"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		return printCategoryList(cmd, resp.Categories)
	}
}

func newCategoryGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific category by alias",
		RunE:  makeRunCategoryGet(factory),
	}
	cmd.Flags().String("alias", "", "Category alias (e.g., 'pizza', 'coffee')")
	_ = cmd.MarkFlagRequired("alias")
	return cmd
}

func makeRunCategoryGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		alias, _ := cmd.Flags().GetString("alias")

		body, err := client.doYelp(ctx, "GET", "/categories/"+alias, nil)
		if err != nil {
			return fmt.Errorf("get category: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var resp struct {
			Category CategoryInfo `json:"category"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		cat := resp.Category
		lines := []string{
			fmt.Sprintf("Alias:   %s", cat.Alias),
			fmt.Sprintf("Title:   %s", cat.Title),
			fmt.Sprintf("Parents: %s", orDash(joinStrings(cat.ParentAliases, ", "))),
		}
		cli.PrintText(lines)
		return nil
	}
}

// joinStrings joins a string slice with a separator, returning "" if empty.
func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for _, s := range ss[1:] {
		result += sep + s
	}
	return result
}
