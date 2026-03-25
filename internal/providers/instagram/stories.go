package instagram

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// storyFeedResponse is the response envelope for GET /api/v1/feed/user/{id}/story/.
type storyFeedResponse struct {
	Reel   *storyReel `json:"reel"`
	Status string     `json:"status"`
}

// storyReel contains the list of story items for a user.
type storyReel struct {
	Items  []rawStoryItem `json:"items"`
	Status string         `json:"status"`
}

// rawStoryItem is the raw story/media item from the stories endpoint.
type rawStoryItem struct {
	ID        string `json:"id"`
	MediaType int    `json:"media_type"`
	TakenAt   int64  `json:"taken_at"`
	ExpiringAt int64 `json:"expiring_at"`
	ImageVersions struct {
		Candidates []struct {
			URL string `json:"url"`
		} `json:"candidates"`
	} `json:"image_versions2"`
}

// toStorySummary converts a rawStoryItem to StorySummary.
func toStorySummary(item rawStoryItem) StorySummary {
	thumbnailURL := ""
	if len(item.ImageVersions.Candidates) > 0 {
		thumbnailURL = item.ImageVersions.Candidates[0].URL
	}
	return StorySummary{
		ID:           item.ID,
		MediaType:    item.MediaType,
		Timestamp:    item.TakenAt,
		ExpiresAt:    item.ExpiringAt,
		ThumbnailURL: thumbnailURL,
	}
}

// storyViewersResponse is the response for GET /api/v1/media/{id}/list_reel_media_viewer/.
type storyViewersResponse struct {
	Users      []rawUser `json:"users"`
	UserCount  int64     `json:"user_count"`
	NextMaxID  string    `json:"next_max_id"`
	Status     string    `json:"status"`
}

// reelsTrayResponse is the response for GET /api/v1/feed/reels_tray/.
type reelsTrayResponse struct {
	Trays  []reelsTrayItem `json:"tray"`
	Status string          `json:"status"`
}

// reelsTrayItem represents a single user's story tray entry.
type reelsTrayItem struct {
	ID     string         `json:"id"`
	User   rawUser        `json:"user"`
	Items  []rawStoryItem `json:"items"`
	Seen   int64          `json:"seen"`
}

// storiesTrayEntry is the output shape for a stories tray item.
type storiesTrayEntry struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Seen     bool   `json:"seen"`
	Count    int    `json:"story_count"`
}

// newStoriesCmd builds the `stories` subcommand group.
func newStoriesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stories",
		Short:   "View stories",
		Aliases: []string{"story", "st"},
	}
	cmd.AddCommand(newStoriesListCmd(factory))
	cmd.AddCommand(newStoriesGetCmd(factory))
	cmd.AddCommand(newStoriesViewersCmd(factory))
	cmd.AddCommand(newStoriesFeedCmd(factory))
	return cmd
}

func newStoriesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List a user's active stories",
		Long:  "List active story items for a user. Defaults to the authenticated user.",
		RunE:  makeRunStoriesList(factory),
	}
	cmd.Flags().String("user-id", "", "User ID whose stories to list (defaults to own user)")
	return cmd
}

func makeRunStoriesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if userID == "" {
			userID = client.SelfUserID()
		}

		resp, err := client.MobileGet(ctx, "/api/v1/feed/user/"+url.PathEscape(userID)+"/story/", nil)
		if err != nil {
			return fmt.Errorf("listing stories for user %s: %w", userID, err)
		}

		var feed storyFeedResponse
		if err := client.DecodeJSON(resp, &feed); err != nil {
			return fmt.Errorf("decoding stories list: %w", err)
		}

		var items []rawStoryItem
		if feed.Reel != nil {
			items = feed.Reel.Items
		}

		summaries := make([]StorySummary, 0, len(items))
		for _, item := range items {
			summaries = append(summaries, toStorySummary(item))
		}

		return printStorySummaries(cmd, summaries)
	}
}

func newStoriesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a story item by ID",
		RunE:  makeRunStoriesGet(factory),
	}
	cmd.Flags().String("story-id", "", "Story item ID")
	_ = cmd.MarkFlagRequired("story-id")
	return cmd
}

func makeRunStoriesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		storyID, _ := cmd.Flags().GetString("story-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/media/"+url.PathEscape(storyID)+"/info/", nil)
		if err != nil {
			return fmt.Errorf("getting story %s: %w", storyID, err)
		}

		var info mediaInfoResponse
		if err := client.DecodeJSON(resp, &info); err != nil {
			return fmt.Errorf("decoding story info: %w", err)
		}

		if len(info.Items) == 0 {
			return fmt.Errorf("story %s not found", storyID)
		}

		item := info.Items[0]
		thumbnailURL := ""
		if len(item.ImageVersions.Candidates) > 0 {
			thumbnailURL = item.ImageVersions.Candidates[0].URL
		}
		summary := StorySummary{
			ID:           item.ID,
			MediaType:    item.MediaType,
			Timestamp:    item.TakenAt,
			ThumbnailURL: thumbnailURL,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summary)
		}

		lines := []string{
			fmt.Sprintf("ID:        %s", summary.ID),
			fmt.Sprintf("Type:      %d", summary.MediaType),
			fmt.Sprintf("Taken At:  %s", formatTimestamp(summary.Timestamp)),
			fmt.Sprintf("Expires:   %s", formatTimestamp(summary.ExpiresAt)),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newStoriesViewersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "viewers",
		Short: "List viewers of a story item",
		RunE:  makeRunStoriesViewers(factory),
	}
	cmd.Flags().String("story-id", "", "Story item ID")
	_ = cmd.MarkFlagRequired("story-id")
	cmd.Flags().Int("limit", 50, "Maximum number of viewers to return")
	return cmd
}

func makeRunStoriesViewers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		storyID, _ := cmd.Flags().GetString("story-id")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/media/"+url.PathEscape(storyID)+"/list_reel_media_viewer/", nil)
		if err != nil {
			return fmt.Errorf("getting viewers for story %s: %w", storyID, err)
		}

		var result storyViewersResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding viewers response: %w", err)
		}

		users := result.Users
		if len(users) > limit {
			users = users[:limit]
		}

		summaries := make([]UserSummary, 0, len(users))
		for _, u := range users {
			summaries = append(summaries, UserSummary{
				ID:            u.PK,
				Username:      u.Username,
				FullName:      u.FullName,
				ProfilePicURL: u.ProfilePicURL,
				IsPrivate:     u.IsPrivate,
				IsVerified:    u.IsVerified,
			})
		}
		return printUserSummaries(cmd, summaries)
	}
}

func newStoriesFeedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feed",
		Short: "Get the stories tray (stories from followed accounts)",
		RunE:  makeRunStoriesFeed(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of tray entries to return")
	return cmd
}

func makeRunStoriesFeed(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/feed/reels_tray/", nil)
		if err != nil {
			return fmt.Errorf("getting stories feed: %w", err)
		}

		var tray reelsTrayResponse
		if err := client.DecodeJSON(resp, &tray); err != nil {
			return fmt.Errorf("decoding stories tray: %w", err)
		}

		entries := tray.Trays
		if len(entries) > limit {
			entries = entries[:limit]
		}

		summaries := make([]storiesTrayEntry, 0, len(entries))
		for _, e := range entries {
			summaries = append(summaries, storiesTrayEntry{
				UserID:   e.User.PK,
				Username: e.User.Username,
				Seen:     e.Seen > 0,
				Count:    len(e.Items),
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No stories in feed.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-10s  %-6s", "USERNAME", "STORIES", "SEEN"))
		for _, s := range summaries {
			lines = append(lines, fmt.Sprintf("%-20s  %-10d  %-6v",
				truncate(s.Username, 20),
				s.Count,
				s.Seen,
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

// printStorySummaries outputs story summaries as JSON or a formatted text table.
func printStorySummaries(cmd *cobra.Command, stories []StorySummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(stories)
	}

	if len(stories) == 0 {
		fmt.Println("No stories found.")
		return nil
	}

	lines := make([]string, 0, len(stories)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-5s  %-16s  %-16s", "ID", "TYPE", "TAKEN AT", "EXPIRES"))
	for _, s := range stories {
		lines = append(lines, fmt.Sprintf("%-20s  %-5d  %-16s  %-16s",
			truncate(s.ID, 20),
			s.MediaType,
			formatTimestamp(s.Timestamp),
			formatTimestamp(s.ExpiresAt),
		))
	}
	cli.PrintText(lines)
	return nil
}
