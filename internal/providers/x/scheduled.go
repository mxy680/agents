package x

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GraphQL query hashes for scheduled tweet operations.
const (
	hashFetchScheduledTweets = "ITtjAzvlZni2wWXwf295Qg"
	hashCreateScheduledTweet = "LCVzRQGxOaGnOnYH01NQXg"
	hashDeleteScheduledTweet = "CTOVqej0JBXAZSwkp1US0g"
)

// ScheduledTweetSummary is a condensed representation of a scheduled tweet.
type ScheduledTweetSummary struct {
	ID               string `json:"id"`
	Text             string `json:"text"`
	ScheduledAt      string `json:"scheduled_at,omitempty"`
	State            string `json:"state,omitempty"`
}

// newScheduledCmd builds the "scheduled" subcommand group.
func newScheduledCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scheduled",
		Short:   "Manage scheduled tweets",
		Aliases: []string{"sched"},
	}
	cmd.AddCommand(newScheduledListCmd(factory))
	cmd.AddCommand(newScheduledCreateCmd(factory))
	cmd.AddCommand(newScheduledDeleteCmd(factory))
	return cmd
}

func newScheduledListCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List scheduled tweets",
		RunE:  makeRunScheduledList(factory),
	}
}

func newScheduledCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a scheduled tweet",
		RunE:  makeRunScheduledCreate(factory),
	}
	cmd.Flags().String("text", "", "Tweet text (required)")
	_ = cmd.MarkFlagRequired("text")
	cmd.Flags().String("date", "", "Scheduled publish time in RFC3339 format (required)")
	_ = cmd.MarkFlagRequired("date")
	cmd.Flags().StringSlice("media-ids", nil, "Comma-separated media IDs to attach")
	cmd.Flags().Bool("dry-run", false, "Print what would be created without creating")
	return cmd
}

func newScheduledDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a scheduled tweet",
		RunE:  makeRunScheduledDelete(factory),
	}
	cmd.Flags().String("tweet-id", "", "Scheduled tweet ID to delete (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without deleting")
	return cmd
}

// --- RunE implementations ---

func makeRunScheduledList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"ascending": true,
		}

		data, err := client.GraphQL(ctx, hashFetchScheduledTweets, "FetchScheduledTweets", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching scheduled tweets: %w", err)
		}

		scheduled, err := parseScheduledTweets(data)
		if err != nil || len(scheduled) == 0 {
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(data)
			}
			fmt.Println("No scheduled tweets found.")
			return nil
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(scheduled)
		}

		lines := make([]string, 0, len(scheduled)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-25s  %-10s  %-50s", "ID", "SCHEDULED AT", "STATE", "TEXT"))
		for _, s := range scheduled {
			lines = append(lines, fmt.Sprintf("%-20s  %-25s  %-10s  %-50s",
				truncate(s.ID, 20),
				truncate(s.ScheduledAt, 25),
				truncate(s.State, 10),
				truncate(s.Text, 50),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunScheduledCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		text, _ := cmd.Flags().GetString("text")
		date, _ := cmd.Flags().GetString("date")
		mediaIDs, _ := cmd.Flags().GetStringSlice("media-ids")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"post":               text,
			"execute_at":         date,
			"dark_request":       false,
		}
		if len(mediaIDs) > 0 {
			mediaEntities := make([]map[string]any, 0, len(mediaIDs))
			for _, id := range mediaIDs {
				id = strings.TrimSpace(id)
				if id != "" {
					mediaEntities = append(mediaEntities, map[string]any{"media_id": id})
				}
			}
			vars["media"] = map[string]any{"media_entities": mediaEntities}
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("create scheduled tweet %q at %s", truncate(text, 60), date), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := client.GraphQLPost(ctx, hashCreateScheduledTweet, "CreateScheduledTweet", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("creating scheduled tweet: %w", err)
		}

		tweet, err := parseCreatedScheduledTweet(data)
		if err != nil {
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(data)
			}
			fmt.Println("Scheduled tweet created.")
			return nil
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(tweet)
		}
		fmt.Printf("Scheduled tweet created: %s (at %s)\n", tweet.ID, tweet.ScheduledAt)
		return nil
	}
}

func makeRunScheduledDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("delete scheduled tweet %s", tweetID), map[string]string{"scheduled_tweet_id": tweetID})
		}

		if err := confirmDestructive(cmd, "this will delete the scheduled tweet"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"scheduled_tweet_id": tweetID,
		}

		_, err = client.GraphQLPost(ctx, hashDeleteScheduledTweet, "DeleteScheduledTweet", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("deleting scheduled tweet %s: %w", tweetID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "tweet_id": tweetID})
		}
		fmt.Printf("Scheduled tweet deleted: %s\n", tweetID)
		return nil
	}
}

// parseScheduledTweets extracts scheduled tweet summaries from GraphQL data.
func parseScheduledTweets(data json.RawMessage) ([]ScheduledTweetSummary, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("parse scheduled tweets data: %w", err)
	}

	var scheduled []ScheduledTweetSummary

	for _, v := range top {
		var items []struct {
			RestID          string `json:"rest_id"`
			ScheduledStatus string `json:"scheduled_status"`
			ExecuteAt       int64  `json:"execute_at"`
			TweetCreate     struct {
				Status string `json:"status"`
				Text   string `json:"text"`
			} `json:"tweet_create_request"`
		}
		if err := json.Unmarshal(v, &items); err != nil {
			continue
		}
		for _, item := range items {
			text := item.TweetCreate.Text
			scheduledAt := ""
			if item.ExecuteAt > 0 {
				scheduledAt = fmt.Sprintf("%d", item.ExecuteAt)
			}
			scheduled = append(scheduled, ScheduledTweetSummary{
				ID:          item.RestID,
				Text:        text,
				ScheduledAt: scheduledAt,
				State:       item.ScheduledStatus,
			})
		}
		if len(scheduled) > 0 {
			return scheduled, nil
		}
	}

	return scheduled, nil
}

// parseCreatedScheduledTweet extracts the created scheduled tweet from GraphQL data.
func parseCreatedScheduledTweet(data json.RawMessage) (*ScheduledTweetSummary, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("parse created scheduled tweet: %w", err)
	}

	for _, v := range top {
		var item struct {
			RestID      string `json:"rest_id"`
			ExecuteAt   int64  `json:"execute_at"`
			TweetCreate struct {
				Text string `json:"status"`
			} `json:"tweet_create_request"`
		}
		if err := json.Unmarshal(v, &item); err != nil {
			continue
		}
		if item.RestID != "" {
			scheduledAt := ""
			if item.ExecuteAt > 0 {
				scheduledAt = fmt.Sprintf("%d", item.ExecuteAt)
			}
			return &ScheduledTweetSummary{
				ID:          item.RestID,
				Text:        item.TweetCreate.Text,
				ScheduledAt: scheduledAt,
			}, nil
		}
	}

	return nil, fmt.Errorf("created scheduled tweet not found in response")
}
