package x

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GraphQL query hashes for retweet operations.
const (
	hashCreateRetweet = "ojPdsZsimiJrUGLR1sjUtA"
	hashDeleteRetweet = "iQtK4dl5hBmXewYZuEOKVw"
)

// newRetweetsCmd builds the "retweets" subcommand group.
func newRetweetsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retweets",
		Short:   "Manage retweets",
		Aliases: []string{"retweet", "rt"},
	}
	cmd.AddCommand(newRetweetsRetweetCmd(factory))
	cmd.AddCommand(newRetweetsUndoCmd(factory))
	return cmd
}

// newRetweetsRetweetCmd builds the "retweets retweet" command.
func newRetweetsRetweetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retweet",
		Short: "Retweet a tweet",
		RunE:  makeRunRetweetsRetweet(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID to retweet (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be retweeted without retweeting")
	return cmd
}

// newRetweetsUndoCmd builds the "retweets undo" command.
func newRetweetsUndoCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Undo a retweet",
		RunE:  makeRunRetweetsUndo(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID to undo retweet (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be undone without undoing")
	return cmd
}

// --- RunE implementations ---

func makeRunRetweetsRetweet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"tweet_id":     tweetID,
			"dark_request": false,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("retweet tweet %s", tweetID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashCreateRetweet, "CreateRetweet", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("retweeting tweet %s: %w", tweetID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "retweeted", "tweet_id": tweetID})
		}
		fmt.Printf("Tweet retweeted: %s\n", tweetID)
		return nil
	}
}

func makeRunRetweetsUndo(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"source_tweet_id": tweetID,
			"dark_request":    false,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("undo retweet of tweet %s", tweetID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashDeleteRetweet, "DeleteRetweet", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("undoing retweet of tweet %s: %w", tweetID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "retweet_undone", "tweet_id": tweetID})
		}
		fmt.Printf("Retweet undone: %s\n", tweetID)
		return nil
	}
}
