package instagram

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// liveListResponse is the response for GET /api/v1/live/reels_tray_broadcasts/.
type liveListResponse struct {
	Broadcasts []rawBroadcast `json:"broadcasts"`
	Status     string         `json:"status"`
}

// rawBroadcast is the raw live broadcast object from the Instagram API.
type rawBroadcast struct {
	ID              string `json:"id"`
	BroadcastStatus string `json:"broadcast_status"`
	CoverFrameURL   string `json:"cover_frame_url"`
	ViewerCount     int64  `json:"viewer_count"`
	PublishedTime   int64  `json:"published_time"`
}

// broadcastInfoResponse is the response for GET /api/v1/live/{id}/info/.
type broadcastInfoResponse struct {
	Broadcast rawBroadcast `json:"broadcast"`
	Status    string       `json:"status"`
}

// liveCommentsResponse is the response for GET /api/v1/live/{id}/get_comment/.
type liveCommentsResponse struct {
	Comments []rawLiveComment `json:"comments"`
	Status   string           `json:"status"`
}

// rawLiveComment is a comment in a live broadcast.
type rawLiveComment struct {
	PK        string  `json:"pk"`
	Text      string  `json:"text"`
	CreatedAt float64 `json:"created_at"`
	User      rawUser `json:"user"`
}

// liveHeartbeatResponse is the response for POST /api/v1/live/{id}/heartbeat_and_get_viewer_count/.
type liveHeartbeatResponse struct {
	ViewerCount int64  `json:"viewer_count"`
	Status      string `json:"status"`
}

// liveLikeResponse is the response for POST /api/v1/live/{id}/like/.
type liveLikeResponse struct {
	LikeTs int64  `json:"like_ts"`
	Status string `json:"status"`
}

// liveCommentResponse is the response for POST /api/v1/live/{id}/comment/.
type liveCommentResponse struct {
	Comment rawLiveComment `json:"comment"`
	Status  string         `json:"status"`
}

// newLiveCmd builds the `live` subcommand group.
func newLiveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "live",
		Short:   "Browse live broadcasts",
		Aliases: []string{"broadcast"},
	}
	cmd.AddCommand(newLiveListCmd(factory))
	cmd.AddCommand(newLiveGetCmd(factory))
	cmd.AddCommand(newLiveCommentsCmd(factory))
	cmd.AddCommand(newLiveHeartbeatCmd(factory))
	cmd.AddCommand(newLiveLikeCmd(factory))
	cmd.AddCommand(newLivePostCommentCmd(factory))
	return cmd
}

func newLiveListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List live broadcasts from followed users",
		RunE:  makeRunLiveList(factory),
	}
	return cmd
}

func makeRunLiveList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/live/reels_tray_broadcasts/", nil)
		if err != nil {
			return fmt.Errorf("listing live broadcasts: %w", err)
		}

		var result liveListResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding live list response: %w", err)
		}

		broadcasts := make([]LiveBroadcast, 0, len(result.Broadcasts))
		for _, b := range result.Broadcasts {
			broadcasts = append(broadcasts, LiveBroadcast{
				ID:              b.ID,
				BroadcastStatus: b.BroadcastStatus,
				CoverFrameURL:   b.CoverFrameURL,
				ViewerCount:     b.ViewerCount,
				StartedAt:       b.PublishedTime,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(broadcasts)
		}

		if len(broadcasts) == 0 {
			fmt.Println("No live broadcasts.")
			return nil
		}

		lines := make([]string, 0, len(broadcasts)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-12s  %-12s  %-12s", "ID", "STATUS", "VIEWERS", "STARTED"))
		for _, b := range broadcasts {
			lines = append(lines, fmt.Sprintf("%-20s  %-12s  %-12s  %-12s",
				truncate(b.ID, 20),
				truncate(b.BroadcastStatus, 12),
				formatCount(b.ViewerCount),
				formatTimestamp(b.StartedAt),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newLiveGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get live broadcast info",
		RunE:  makeRunLiveGet(factory),
	}
	cmd.Flags().String("broadcast-id", "", "Broadcast ID")
	_ = cmd.MarkFlagRequired("broadcast-id")
	return cmd
}

func makeRunLiveGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		broadcastID, _ := cmd.Flags().GetString("broadcast-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/live/"+url.PathEscape(broadcastID)+"/info/", nil)
		if err != nil {
			return fmt.Errorf("getting broadcast %s: %w", broadcastID, err)
		}

		var result broadcastInfoResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding broadcast info: %w", err)
		}

		b := LiveBroadcast{
			ID:              result.Broadcast.ID,
			BroadcastStatus: result.Broadcast.BroadcastStatus,
			CoverFrameURL:   result.Broadcast.CoverFrameURL,
			ViewerCount:     result.Broadcast.ViewerCount,
			StartedAt:       result.Broadcast.PublishedTime,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(b)
		}

		lines := []string{
			fmt.Sprintf("ID:      %s", b.ID),
			fmt.Sprintf("Status:  %s", b.BroadcastStatus),
			fmt.Sprintf("Viewers: %s", formatCount(b.ViewerCount)),
			fmt.Sprintf("Started: %s", formatTimestamp(b.StartedAt)),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newLiveCommentsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comments",
		Short: "Get comments for a live broadcast",
		RunE:  makeRunLiveComments(factory),
	}
	cmd.Flags().String("broadcast-id", "", "Broadcast ID")
	_ = cmd.MarkFlagRequired("broadcast-id")
	return cmd
}

func makeRunLiveComments(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		broadcastID, _ := cmd.Flags().GetString("broadcast-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/live/"+url.PathEscape(broadcastID)+"/get_comment/", nil)
		if err != nil {
			return fmt.Errorf("getting comments for broadcast %s: %w", broadcastID, err)
		}

		var result liveCommentsResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding live comments: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result.Comments)
		}

		if len(result.Comments) == 0 {
			fmt.Println("No comments.")
			return nil
		}

		lines := make([]string, 0, len(result.Comments)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-50s", "USER", "COMMENT"))
		for _, c := range result.Comments {
			lines = append(lines, fmt.Sprintf("%-20s  %-50s",
				truncate(c.User.Username, 20),
				truncate(c.Text, 50),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newLiveHeartbeatCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "heartbeat",
		Short: "Send a heartbeat and get viewer count",
		RunE:  makeRunLiveHeartbeat(factory),
	}
	cmd.Flags().String("broadcast-id", "", "Broadcast ID")
	_ = cmd.MarkFlagRequired("broadcast-id")
	return cmd
}

func makeRunLiveHeartbeat(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		broadcastID, _ := cmd.Flags().GetString("broadcast-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/live/"+url.PathEscape(broadcastID)+"/heartbeat_and_get_viewer_count/", nil)
		if err != nil {
			return fmt.Errorf("heartbeat for broadcast %s: %w", broadcastID, err)
		}

		var result liveHeartbeatResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding heartbeat response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Viewer count: %s\n", formatCount(result.ViewerCount))
		return nil
	}
}

func newLiveLikeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "like",
		Short: "Like a live broadcast",
		RunE:  makeRunLiveLike(factory),
	}
	cmd.Flags().String("broadcast-id", "", "Broadcast ID")
	_ = cmd.MarkFlagRequired("broadcast-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunLiveLike(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		broadcastID, _ := cmd.Flags().GetString("broadcast-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("like broadcast %s", broadcastID),
				map[string]string{"broadcast_id": broadcastID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/live/"+url.PathEscape(broadcastID)+"/like/", nil)
		if err != nil {
			return fmt.Errorf("liking broadcast %s: %w", broadcastID, err)
		}

		var result liveLikeResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding like response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Liked broadcast %s\n", broadcastID)
		return nil
	}
}

func newLivePostCommentCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post-comment",
		Short: "Post a comment on a live broadcast",
		RunE:  makeRunLivePostComment(factory),
	}
	cmd.Flags().String("broadcast-id", "", "Broadcast ID")
	_ = cmd.MarkFlagRequired("broadcast-id")
	cmd.Flags().String("text", "", "Comment text")
	_ = cmd.MarkFlagRequired("text")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunLivePostComment(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		broadcastID, _ := cmd.Flags().GetString("broadcast-id")
		text, _ := cmd.Flags().GetString("text")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("post comment on broadcast %s", broadcastID),
				map[string]string{"broadcast_id": broadcastID, "text": text})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("comment_text", text)

		resp, err := client.MobilePost(ctx, "/api/v1/live/"+url.PathEscape(broadcastID)+"/comment/", body)
		if err != nil {
			return fmt.Errorf("posting comment on broadcast %s: %w", broadcastID, err)
		}

		var result liveCommentResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding post comment response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Posted comment on broadcast %s\n", broadcastID)
		return nil
	}
}
