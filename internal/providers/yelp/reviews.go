package yelp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newReviewsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reviews",
		Short:   "Get reviews for a Yelp business",
		Aliases: []string{"review"},
	}

	cmd.AddCommand(newReviewListCmd(factory))

	return cmd
}

func newReviewListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List reviews for a business",
		RunE:  makeRunReviewList(factory),
	}
	cmd.Flags().String("id", "", "Yelp business ID or alias")
	cmd.Flags().String("locale", "", "Locale code (e.g., en_US)")
	cmd.Flags().String("sort-by", "", "Sort: yelp_sort, newest, oldest, elites")
	cmd.Flags().Int("limit", 20, "Maximum number of reviews (max 50)")
	cmd.Flags().Int("offset", 0, "Offset for pagination")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunReviewList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		id, _ := cmd.Flags().GetString("id")
		locale, _ := cmd.Flags().GetString("locale")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		params := url.Values{}
		if locale != "" {
			params.Set("locale", locale)
		}
		if sortBy != "" {
			params.Set("sort_by", sortBy)
		}
		if limit > 0 {
			params.Set("limit", strconv.Itoa(limit))
		}
		if offset > 0 {
			params.Set("offset", strconv.Itoa(offset))
		}

		body, err := client.doYelp(ctx, "GET", "/biz/"+id+"/review_feed", params)
		if err != nil {
			return fmt.Errorf("list reviews: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var resp struct {
			Reviews []ReviewSummary `json:"reviews"`
			Total   int             `json:"total"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		return printReviewSummaries(cmd, resp.Reviews)
	}
}
