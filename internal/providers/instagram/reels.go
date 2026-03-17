package instagram

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// clipsUserResponse is the response envelope for GET /api/v1/clips/user/.
type clipsUserResponse struct {
	Items         []clipsUserItem `json:"items"`
	PagingInfo    clipsPagingInfo `json:"paging_info"`
	Status        string          `json:"status"`
}

// clipsUserItem wraps a media item in the clips/user response.
type clipsUserItem struct {
	Media rawMediaItem `json:"media"`
}

// clipsPagingInfo holds pagination data for clips endpoints.
type clipsPagingInfo struct {
	MaxID       string `json:"max_id"`
	MoreAvailable bool `json:"more_available"`
}

// clipsReelsTabResponse is the response envelope for POST /api/v1/clips/reels_tab_feed_items/.
type clipsReelsTabResponse struct {
	Items      []clipsUserItem `json:"items"`
	PagingInfo clipsPagingInfo `json:"paging_info"`
	Status     string          `json:"status"`
}

// toReelSummary converts a rawMediaItem to ReelSummary.
func toReelSummary(item rawMediaItem) ReelSummary {
	thumbnailURL := ""
	if len(item.ImageVersions.Candidates) > 0 {
		thumbnailURL = item.ImageVersions.Candidates[0].URL
	}
	return ReelSummary{
		ID:           item.ID,
		Shortcode:    item.Code,
		Caption:      item.Caption.Text,
		Timestamp:    item.TakenAt,
		LikeCount:    item.LikeCount,
		CommentCount: item.CommentCount,
		PlayCount:    item.PlayCount,
		ThumbnailURL: thumbnailURL,
	}
}

// newReelsCmd builds the `reels` subcommand group.
func newReelsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reels",
		Short:   "View and manage reels",
		Aliases: []string{"reel"},
	}
	cmd.AddCommand(newReelsListCmd(factory))
	cmd.AddCommand(newReelsGetCmd(factory))
	cmd.AddCommand(newReelsFeedCmd(factory))
	cmd.AddCommand(newReelsDeleteCmd(factory))
	return cmd
}

func newReelsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List a user's reels",
		Long:  "List reels (clips) for a user. Defaults to the authenticated user.",
		RunE:  makeRunReelsList(factory),
	}
	cmd.Flags().String("user-id", "", "User ID whose reels to list (defaults to own user)")
	cmd.Flags().Int("limit", 20, "Maximum number of reels to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (max_id from previous response)")
	return cmd
}

func makeRunReelsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if userID == "" {
			userID = client.session.DSUserID
		}

		params := url.Values{}
		params.Set("target_user_id", userID)
		params.Set("page_size", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}

		resp, err := client.Get(ctx, "/api/v1/clips/user/", params)
		if err != nil {
			return fmt.Errorf("listing reels for user %s: %w", userID, err)
		}

		var result clipsUserResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding reels list: %w", err)
		}

		summaries := make([]ReelSummary, 0, len(result.Items))
		for _, item := range result.Items {
			summaries = append(summaries, toReelSummary(item.Media))
		}

		if err := printReelSummaries(cmd, summaries); err != nil {
			return err
		}
		if result.PagingInfo.MoreAvailable && result.PagingInfo.MaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.PagingInfo.MaxID)
		}
		return nil
	}
}

func newReelsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a reel by ID",
		RunE:  makeRunReelsGet(factory),
	}
	cmd.Flags().String("reel-id", "", "Reel ID")
	_ = cmd.MarkFlagRequired("reel-id")
	return cmd
}

func makeRunReelsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		reelID, _ := cmd.Flags().GetString("reel-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/api/v1/media/"+reelID+"/info/", nil)
		if err != nil {
			return fmt.Errorf("getting reel %s: %w", reelID, err)
		}

		var info mediaInfoResponse
		if err := client.DecodeJSON(resp, &info); err != nil {
			return fmt.Errorf("decoding reel info: %w", err)
		}

		if len(info.Items) == 0 {
			return fmt.Errorf("reel %s not found", reelID)
		}

		summary := toReelSummary(info.Items[0])
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summary)
		}

		lines := []string{
			fmt.Sprintf("ID:        %s", summary.ID),
			fmt.Sprintf("Shortcode: %s", summary.Shortcode),
			fmt.Sprintf("Caption:   %s", truncate(summary.Caption, 80)),
			fmt.Sprintf("Date:      %s", formatTimestamp(summary.Timestamp)),
			fmt.Sprintf("Likes:     %s", formatCount(summary.LikeCount)),
			fmt.Sprintf("Comments:  %s", formatCount(summary.CommentCount)),
			fmt.Sprintf("Plays:     %s", formatCount(summary.PlayCount)),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newReelsFeedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feed",
		Short: "Get the reels discovery feed",
		RunE:  makeRunReelsFeed(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of reels to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (max_id from previous response)")
	return cmd
}

func makeRunReelsFeed(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("page_size", strconv.Itoa(limit))
		if cursor != "" {
			body.Set("max_id", cursor)
		}

		resp, err := client.Post(ctx, "/api/v1/clips/reels_tab_feed_items/", body)
		if err != nil {
			return fmt.Errorf("getting reels feed: %w", err)
		}

		var result clipsReelsTabResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding reels feed: %w", err)
		}

		summaries := make([]ReelSummary, 0, len(result.Items))
		for _, item := range result.Items {
			summaries = append(summaries, toReelSummary(item.Media))
		}

		if err := printReelSummaries(cmd, summaries); err != nil {
			return err
		}
		if result.PagingInfo.MoreAvailable && result.PagingInfo.MaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.PagingInfo.MaxID)
		}
		return nil
	}
}

func newReelsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a reel",
		Long:  "Permanently delete one of your own reels.",
		RunE:  makeRunReelsDelete(factory),
	}
	cmd.Flags().String("reel-id", "", "Reel ID to delete")
	_ = cmd.MarkFlagRequired("reel-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunReelsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		reelID, _ := cmd.Flags().GetString("reel-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("delete reel %s", reelID), map[string]string{"reel_id": reelID})
		}
		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Post(ctx, "/api/v1/media/"+reelID+"/delete/?media_type=REELS", nil)
		if err != nil {
			return fmt.Errorf("deleting reel %s: %w", reelID, err)
		}

		var result mediaDeleteResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding delete response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Deleted reel %s\n", reelID)
		return nil
	}
}

// printReelSummaries outputs reel summaries as JSON or a formatted text table.
func printReelSummaries(cmd *cobra.Command, reels []ReelSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(reels)
	}

	if len(reels) == 0 {
		fmt.Println("No reels found.")
		return nil
	}

	lines := make([]string, 0, len(reels)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-40s  %-12s  %-8s  %-8s  %-8s",
		"ID", "CAPTION", "DATE", "LIKES", "COMMENTS", "PLAYS"))
	for _, r := range reels {
		lines = append(lines, fmt.Sprintf("%-20s  %-40s  %-12s  %-8s  %-8s  %-8s",
			truncate(r.ID, 20),
			truncate(r.Caption, 40),
			formatTimestamp(r.Timestamp),
			formatCount(r.LikeCount),
			formatCount(r.CommentCount),
			formatCount(r.PlayCount),
		))
	}
	cli.PrintText(lines)
	return nil
}
