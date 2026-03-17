package instagram

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// reelsFeedDocID is the GraphQL doc_id for PolarisClipsTabDesktopPaginationQuery.
const reelsFeedDocID = "26136666099278270"

// graphQLReelsFeedNode is a single media node from the PolarisClipsTabDesktopPaginationQuery.
type graphQLReelsFeedNode struct {
	ID            string `json:"pk"`
	Code          string `json:"code"`
	TakenAt       int64  `json:"taken_at"`
	LikeCount     int64  `json:"like_count"`
	CommentCount  int64  `json:"comment_count"`
	PlayCount     int64  `json:"play_count"`
	Caption       struct {
		Text string `json:"text"`
	} `json:"caption"`
	ImageVersions struct {
		Candidates []struct {
			URL string `json:"url"`
		} `json:"candidates"`
	} `json:"image_versions2"`
}

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

		resp, err := client.MobilePost(ctx, "/api/v1/clips/user/", params)
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

		resp, err := client.MobileGet(ctx, "/api/v1/media/"+reelID+"/info/", nil)
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

		// The mobile REST endpoint /api/v1/clips/reels_tab_feed_items/ is no longer
		// available. Use the web GraphQL endpoint instead.
		var afterCursor any
		if cursor != "" {
			afterCursor = cursor
		}
		gqlData, err := client.PostFormGraphQL(ctx, reelsFeedDocID, "PolarisClipsTabDesktopPaginationQuery", map[string]any{
			"after":  afterCursor,
			"before": nil,
			"data": map[string]any{
				"container_module": "clips_tab_desktop_page",
				"seen_reels":       "[]",
			},
			"first": limit,
			"last":  nil,
		})
		if err != nil {
			return fmt.Errorf("getting reels feed: %w", err)
		}

		summaries, nextCursor, err := parseGraphQLReelsFeed(gqlData)
		if err != nil {
			return fmt.Errorf("parsing reels feed response: %w", err)
		}

		if err := printReelSummaries(cmd, summaries); err != nil {
			return err
		}
		if nextCursor != "" {
			fmt.Printf("Next cursor: %s\n", nextCursor)
		}
		return nil
	}
}

// parseGraphQLReelsFeed parses the PolarisClipsTabDesktopPaginationQuery response.
// It returns summaries and an optional next-page cursor.
func parseGraphQLReelsFeed(data json.RawMessage) ([]ReelSummary, string, error) {
	// The response shape wraps media edges inside xdt_api__v1__clips__home__connection.
	// We use a flexible map approach to handle potential schema variations.
	var outer map[string]json.RawMessage
	if err := json.Unmarshal(data, &outer); err != nil {
		return nil, "", fmt.Errorf("unmarshal graphql reels feed data: %w", err)
	}

	// Try the known connection field name first.
	const connectionKey = "xdt_api__v1__clips__home__connection"
	connRaw, ok := outer[connectionKey]
	if !ok {
		// Fall back: look for any key containing "connection" that has edges.
		for k, v := range outer {
			if len(k) > 10 && (containsAny(k, "clips", "reels", "connection")) {
				connRaw = v
				ok = true
				break
			}
		}
	}

	if !ok || connRaw == nil {
		// Return empty result rather than hard error — feed may be empty.
		return []ReelSummary{}, "", nil
	}

	var conn struct {
		Edges []struct {
			Node struct {
				Media graphQLReelsFeedNode `json:"media"`
			} `json:"node"`
		} `json:"edges"`
		PageInfo struct {
			HasNextPage bool   `json:"has_next_page"`
			EndCursor   string `json:"end_cursor"`
		} `json:"page_info"`
	}
	if err := json.Unmarshal(connRaw, &conn); err != nil {
		return nil, "", fmt.Errorf("unmarshal graphql reels connection: %w", err)
	}

	summaries := make([]ReelSummary, 0, len(conn.Edges))
	for _, edge := range conn.Edges {
		n := edge.Node.Media
		thumbnailURL := ""
		if len(n.ImageVersions.Candidates) > 0 {
			thumbnailURL = n.ImageVersions.Candidates[0].URL
		}
		summaries = append(summaries, ReelSummary{
			ID:           n.ID,
			Shortcode:    n.Code,
			Caption:      n.Caption.Text,
			Timestamp:    n.TakenAt,
			LikeCount:    n.LikeCount,
			CommentCount: n.CommentCount,
			PlayCount:    n.PlayCount,
			ThumbnailURL: thumbnailURL,
		})
	}

	nextCursor := ""
	if conn.PageInfo.HasNextPage {
		nextCursor = conn.PageInfo.EndCursor
	}
	return summaries, nextCursor, nil
}

// containsAny returns true if s contains any of the given substrings.
func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
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

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+reelID+"/delete/?media_type=REELS", nil)
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
