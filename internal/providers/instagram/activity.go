package instagram

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// activityFeedResponse is the response for GET /api/v1/news/inbox/.
type activityFeedResponse struct {
	NewStories   []rawActivityItem `json:"new_stories"`
	OldStories   []rawActivityItem `json:"old_stories"`
	Continuation string            `json:"continuation"`
	Status       string            `json:"status"`
}

// rawActivityItem is a single notification entry from the activity feed.
type rawActivityItem struct {
	PK   string `json:"pk"`
	Type int    `json:"type"`
	Args struct {
		Timestamp   float64 `json:"timestamp"` // Instagram returns float timestamps
		Text        string  `json:"text"`
		ProfileID   string  `json:"profile_id"`
		ProfileName string  `json:"profile_name"`
	} `json:"args"`
}

// activityMarkCheckedResponse is the response for POST /api/v1/news/inbox_seen/.
type activityMarkCheckedResponse struct {
	Status string `json:"status"`
}

// newActivityCmd builds the `activity` subcommand group.
func newActivityCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "activity",
		Short:   "View and manage activity notifications",
		Aliases: []string{"notifications", "notif"},
	}
	cmd.AddCommand(newActivityFeedCmd(factory))
	cmd.AddCommand(newActivityMarkCheckedCmd(factory))
	return cmd
}

func newActivityFeedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feed",
		Short: "List activity notifications",
		RunE:  makeRunActivityFeed(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of notifications to return")
	return cmd
}

func makeRunActivityFeed(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("mark_as_seen", "false")
		params.Set("count", strconv.Itoa(limit))

		resp, err := client.MobileGet(ctx, "/api/v1/news/inbox/", params)
		if err != nil {
			return fmt.Errorf("fetching activity feed: %w", err)
		}

		var result activityFeedResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding activity feed: %w", err)
		}

		all := append(result.NewStories, result.OldStories...)
		if len(all) > limit {
			all = all[:limit]
		}

		items := make([]ActivityItem, 0, len(all))
		for _, s := range all {
			items = append(items, ActivityItem{
				PK:          s.PK,
				Type:        s.Type,
				Timestamp:   int64(s.Args.Timestamp),
				Text:        s.Args.Text,
				ProfileID:   s.Args.ProfileID,
				ProfileName: s.Args.ProfileName,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(items)
		}

		if len(items) == 0 {
			fmt.Println("No activity.")
			return nil
		}

		lines := make([]string, 0, len(items)+1)
		lines = append(lines, fmt.Sprintf("%-12s  %-20s  %-50s", "DATE", "USER", "TEXT"))
		for _, item := range items {
			lines = append(lines, fmt.Sprintf("%-12s  %-20s  %-50s",
				formatTimestamp(item.Timestamp),
				truncate(item.ProfileName, 20),
				truncate(item.Text, 50),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newActivityMarkCheckedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-checked",
		Short: "Mark activity feed as seen",
		RunE:  makeRunActivityMarkChecked(factory),
	}
	return cmd
}

func makeRunActivityMarkChecked(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/news/inbox_seen/", nil)
		if err != nil {
			return fmt.Errorf("marking activity as seen: %w", err)
		}

		var result activityMarkCheckedResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding mark-checked response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Println("Activity feed marked as seen.")
		return nil
	}
}
