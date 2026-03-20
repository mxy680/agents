package linkedin

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

// newFeedCmd builds the "feed" subcommand group.
func newFeedCmd(factory ClientFactory) *cobra.Command {
	feedCmd := &cobra.Command{
		Use:   "feed",
		Short: "Browse your LinkedIn feed",
	}
	feedCmd.AddCommand(newFeedListCmd(factory))
	feedCmd.AddCommand(newFeedHashtagCmd(factory))
	return feedCmd
}

// newFeedListCmd builds the "feed list" command.
func newFeedListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your LinkedIn home feed",
		Long:  "Retrieve posts from your LinkedIn homepage feed.",
		RunE:  makeRunFeedList(factory),
	}
	cmd.Flags().Int("limit", 10, "Maximum number of feed items to return")
	cmd.Flags().String("cursor", "0", "Pagination start offset")
	return cmd
}

// newFeedHashtagCmd builds the "feed hashtag" command.
func newFeedHashtagCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hashtag",
		Short: "Browse a hashtag feed",
		Long:  "Retrieve posts from a LinkedIn hashtag feed.",
		RunE:  makeRunFeedHashtag(factory),
	}
	cmd.Flags().String("tag", "", "Hashtag to browse (without the # prefix)")
	_ = cmd.MarkFlagRequired("tag")
	cmd.Flags().Int("limit", 10, "Maximum number of feed items to return")
	cmd.Flags().String("cursor", "0", "Pagination start offset")
	return cmd
}

func makeRunFeedList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", "feedUpdates")
		params.Set("moduleKey", "HOMEPAGE_FEED")
		params.Set("count", fmt.Sprintf("%d", limit))
		params.Set("start", cursor)

		resp, err := client.Get(ctx, "/voyager/api/feed/dash/feedUpdates", params)
		if err != nil {
			return fmt.Errorf("listing feed: %w", err)
		}

		var raw voyagerFeedUpdatesResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding feed: %w", err)
		}

		posts := make([]PostSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			posts = append(posts, feedElementToPostSummary(el))
		}
		return printPostSummaries(cmd, posts)
	}
}

func makeRunFeedHashtag(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tag, _ := cmd.Flags().GetString("tag")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		hashtagURN := "urn:li:hashtag:" + tag

		params := url.Values{}
		params.Set("q", "feedUpdates")
		params.Set("moduleKey", "HASHTAG_FEED")
		params.Set("hashtagUrn", hashtagURN)
		params.Set("count", fmt.Sprintf("%d", limit))
		params.Set("start", cursor)

		resp, err := client.Get(ctx, "/voyager/api/feed/dash/feedUpdates", params)
		if err != nil {
			return fmt.Errorf("listing hashtag feed for #%s: %w", tag, err)
		}

		var raw voyagerFeedUpdatesResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding hashtag feed: %w", err)
		}

		posts := make([]PostSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			posts = append(posts, feedElementToPostSummary(el))
		}
		return printPostSummaries(cmd, posts)
	}
}
