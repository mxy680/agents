package linkedin

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerProfileViewsResponse is the response for GET /voyager/api/identity/wvmpCards.
type voyagerProfileViewsResponse struct {
	ViewsCount int    `json:"viewsCount"`
	TimePeriod string `json:"timePeriod"`
}

// voyagerSearchAppearancesResponse is the response for GET /voyager/api/identity/searchAppearances.
type voyagerSearchAppearancesResponse struct {
	SearchAppearanceCount int    `json:"searchAppearanceCount"`
	TimePeriod            string `json:"timePeriod"`
}

// voyagerPostImpressionsResponse is the response for GET /voyager/api/socialActions/{postUrn}/impressions.
type voyagerPostImpressionsResponse struct {
	ImpressionCount       int `json:"impressionCount"`
	UniqueImpressionsCount int `json:"uniqueImpressionsCount"`
}

// AnalyticsSearchAppearances holds search appearance analytics.
type AnalyticsSearchAppearances struct {
	Count      int    `json:"count"`
	TimePeriod string `json:"time_period"`
}

// AnalyticsPostImpressions holds post impression analytics.
type AnalyticsPostImpressions struct {
	Impressions       int `json:"impressions"`
	UniqueImpressions int `json:"unique_impressions"`
}

// newAnalyticsCmd builds the "analytics" subcommand group.
func newAnalyticsCmd(factory ClientFactory) *cobra.Command {
	analyticsCmd := &cobra.Command{
		Use:   "analytics",
		Short: "View LinkedIn analytics",
		Long:  "View profile views, search appearances, and post impressions.",
	}
	analyticsCmd.AddCommand(newAnalyticsProfileViewsCmd(factory))
	analyticsCmd.AddCommand(newAnalyticsSearchAppearancesCmd(factory))
	analyticsCmd.AddCommand(newAnalyticsPostImpressionsCmd(factory))
	return analyticsCmd
}

// newAnalyticsProfileViewsCmd builds the "analytics profile-views" command.
func newAnalyticsProfileViewsCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "profile-views",
		Short: "Show who viewed your profile",
		Long:  "Retrieve profile view count and time period via the LinkedIn Voyager API.",
		RunE:  makeRunAnalyticsProfileViews(factory),
	}
}

// newAnalyticsSearchAppearancesCmd builds the "analytics search-appearances" command.
func newAnalyticsSearchAppearancesCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "search-appearances",
		Short: "Show how often you appeared in search",
		Long:  "Retrieve the number of times you appeared in LinkedIn search results.",
		RunE:  makeRunAnalyticsSearchAppearances(factory),
	}
}

// newAnalyticsPostImpressionsCmd builds the "analytics post-impressions" command.
func newAnalyticsPostImpressionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post-impressions",
		Short: "Show impressions for a post",
		Long:  "Retrieve impression and unique impression counts for a specific post.",
		RunE:  makeRunAnalyticsPostImpressions(factory),
	}
	cmd.Flags().String("post-urn", "", "Activity URN of the post (e.g. urn:li:activity:1234)")
	_ = cmd.MarkFlagRequired("post-urn")
	return cmd
}

func makeRunAnalyticsProfileViews(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/voyager/api/identity/wvmpCards", nil)
		if err != nil {
			return fmt.Errorf("fetching profile views: %w", err)
		}

		var raw voyagerProfileViewsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding profile views: %w", err)
		}

		views := AnalyticsProfileViews{
			TotalViews: raw.ViewsCount,
			TimePeriod: raw.TimePeriod,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(views)
		}
		cli.PrintText([]string{
			fmt.Sprintf("Profile Views: %s", formatCount(views.TotalViews)),
			fmt.Sprintf("Time Period:   %s", views.TimePeriod),
		})
		return nil
	}
}

func makeRunAnalyticsSearchAppearances(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/voyager/api/identity/searchAppearances", nil)
		if err != nil {
			return fmt.Errorf("fetching search appearances: %w", err)
		}

		var raw voyagerSearchAppearancesResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding search appearances: %w", err)
		}

		appearances := AnalyticsSearchAppearances{
			Count:      raw.SearchAppearanceCount,
			TimePeriod: raw.TimePeriod,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(appearances)
		}
		cli.PrintText([]string{
			fmt.Sprintf("Search Appearances: %s", formatCount(appearances.Count)),
			fmt.Sprintf("Time Period:        %s", appearances.TimePeriod),
		})
		return nil
	}
}

func makeRunAnalyticsPostImpressions(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		postURN, _ := cmd.Flags().GetString("post-urn")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/socialActions/" + url.PathEscape(postURN) + "/impressions"
		resp, err := client.Get(ctx, path, nil)
		if err != nil {
			return fmt.Errorf("fetching impressions for %s: %w", postURN, err)
		}

		var raw voyagerPostImpressionsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding post impressions: %w", err)
		}

		impressions := AnalyticsPostImpressions{
			Impressions:       raw.ImpressionCount,
			UniqueImpressions: raw.UniqueImpressionsCount,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(impressions)
		}
		cli.PrintText([]string{
			fmt.Sprintf("Post URN:           %s", postURN),
			fmt.Sprintf("Impressions:        %s", formatCount(impressions.Impressions)),
			fmt.Sprintf("Unique Impressions: %s", formatCount(impressions.UniqueImpressions)),
		})
		return nil
	}
}
