package x

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GraphQL query hashes for likes operations.
const (
	hashLikes           = "IohM3gxQHfvWePH5E3KuNA"
	hashFavoriteTweet   = "lI07N6Otwv1PhnEgXILM7A"
	hashUnfavoriteTweet = "ZYKSe-w7KEslx3JhSIk5LA"
)

// newLikesCmd builds the "likes" subcommand group.
func newLikesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "likes",
		Short:   "Manage tweet likes",
		Aliases: []string{"like"},
	}
	cmd.AddCommand(newLikesListCmd(factory))
	cmd.AddCommand(newLikesLikeCmd(factory))
	cmd.AddCommand(newLikesUnlikeCmd(factory))
	return cmd
}

// newLikesListCmd builds the "likes list" command.
func newLikesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tweets liked by a user",
		RunE:  makeRunLikesList(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newLikesLikeCmd builds the "likes like" command.
func newLikesLikeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "like",
		Short: "Like a tweet",
		RunE:  makeRunLikesLike(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID to like (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be liked without liking")
	return cmd
}

// newLikesUnlikeCmd builds the "likes unlike" command.
func newLikesUnlikeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlike",
		Short: "Unlike a tweet",
		RunE:  makeRunLikesUnlike(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID to unlike (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be unliked without unliking")
	return cmd
}

// --- RunE implementations ---

func makeRunLikesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"userId":                 userID,
			"count":                  limit,
			"includePromotedContent": false,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashLikes, "Likes", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching likes for user %s: %w", userID, err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunLikesLike(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"tweet_id": tweetID,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("like tweet %s", tweetID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := client.GraphQLPost(ctx, hashFavoriteTweet, "FavoriteTweet", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("liking tweet %s: %w", tweetID, err)
		}

		var payload struct {
			FavoriteTweet struct {
				TweetResults json.RawMessage `json:"tweet_results"`
			} `json:"favorite_tweet"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("parse like response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "liked", "tweet_id": tweetID})
		}
		fmt.Printf("Tweet liked: %s\n", tweetID)
		return nil
	}
}

func makeRunLikesUnlike(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"tweet_id": tweetID,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("unlike tweet %s", tweetID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashUnfavoriteTweet, "UnfavoriteTweet", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("unliking tweet %s: %w", tweetID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "unliked", "tweet_id": tweetID})
		}
		fmt.Printf("Tweet unliked: %s\n", tweetID)
		return nil
	}
}
