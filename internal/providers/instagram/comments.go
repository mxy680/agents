package instagram

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// graphQLCommentsNode is a single comment node from the PolarisPostChildCommentsQuery response.
type graphQLCommentsNode struct {
	PK               string `json:"pk"`
	Text             string `json:"text"`
	CreatedAtUTC     int64  `json:"created_at_utc"`
	CommentLikeCount int64  `json:"comment_like_count"`
	HasLikedComment  bool   `json:"has_liked_comment"`
	User             struct {
		PK            string `json:"pk"`
		Username      string `json:"username"`
		ProfilePicURL string `json:"profile_pic_url"`
		IsVerified    bool   `json:"is_verified"`
	} `json:"user"`
}

// commentsDocID is the GraphQL doc_id for PolarisPostChildCommentsQuery.
const commentsDocID = "26914912424764761"

// commentsListResponse is the response for GET /api/v1/media/{media_id}/comments/.
type commentsListResponse struct {
	Comments      []rawComment `json:"comments"`
	NextMinID     string       `json:"next_min_id"`
	NextMaxID     string       `json:"next_max_id"`
	CanViewMorePreviewComments bool `json:"can_view_more_preview_comments"`
	Status        string       `json:"status"`
}

// rawComment is a raw comment object from the Instagram API.
type rawComment struct {
	PK              string  `json:"pk"`
	Text            string  `json:"text"`
	CreatedAt       int64   `json:"created_at"`
	CommentLikeCount int64  `json:"comment_like_count"`
	User            rawUser `json:"user"`
}

// commentActionResponse is a generic response for comment create/delete/like actions.
type commentActionResponse struct {
	Status string `json:"status"`
}

// newCommentsCmd builds the `comments` subcommand group.
func newCommentsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "comments",
		Short:   "View and manage comments",
		Aliases: []string{"comment"},
	}
	cmd.AddCommand(newCommentsListCmd(factory))
	cmd.AddCommand(newCommentsRepliesCmd(factory))
	cmd.AddCommand(newCommentsCreateCmd(factory))
	cmd.AddCommand(newCommentsDeleteCmd(factory))
	cmd.AddCommand(newCommentsLikeCmd(factory))
	cmd.AddCommand(newCommentsUnlikeCmd(factory))
	cmd.AddCommand(newCommentsDisableCmd(factory))
	cmd.AddCommand(newCommentsEnableCmd(factory))
	return cmd
}

func newCommentsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments on a post",
		RunE:  makeRunCommentsList(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().Int("limit", 20, "Maximum number of comments to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunCommentsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("min_id", cursor)
		}

		resp, err := client.MobileGet(ctx, "/api/v1/media/"+mediaID+"/comments/", params)
		if err != nil {
			return fmt.Errorf("listing comments for media %s: %w", mediaID, err)
		}

		var result commentsListResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding comments response: %w", err)
		}

		// Instagram can return HTTP 200 with status:"fail" (e.g. new account restriction).
		// Fall back to the web GraphQL media detail query which includes comments.
		var summaries []CommentSummary
		if result.Status == "fail" {
			// Step 1: Get the shortcode from media info (needed for GraphQL query)
			infoResp, infoErr := client.MobileGet(ctx, "/api/v1/media/"+mediaID+"/info/", nil)
			if infoErr != nil {
				return fmt.Errorf("comments unavailable for media %s: could not fetch media info: %w", mediaID, infoErr)
			}
			var mediaInfo struct {
				Items []struct {
					Code string `json:"code"`
				} `json:"items"`
			}
			if infoErr = client.DecodeJSON(infoResp, &mediaInfo); infoErr != nil || len(mediaInfo.Items) == 0 {
				return fmt.Errorf("comments unavailable for media %s: could not get shortcode", mediaID)
			}
			shortcode := mediaInfo.Items[0].Code

			// Step 2: Use PolarisPostActionLoadPostQueryQuery to get the full media
			// detail including edge_media_to_parent_comment with comments.
			const mediaDetailDocID = "8845758582119845"
			gqlData, gqlErr := client.PostFormGraphQL(ctx, mediaDetailDocID, "PolarisPostActionLoadPostQueryQuery", map[string]any{
				"shortcode": shortcode,
			})
			if gqlErr != nil {
				return fmt.Errorf("comments unavailable for media %s: %w", mediaID, gqlErr)
			}
			summaries, err = parseMediaDetailComments(gqlData, limit)
			if err != nil {
				return fmt.Errorf("comments unavailable for media %s: %w", mediaID, err)
			}
		} else {
			summaries = make([]CommentSummary, 0, len(result.Comments))
			for _, c := range result.Comments {
				summaries = append(summaries, toCommentSummary(c))
			}
		}

		if err := printCommentSummaries(cmd, summaries); err != nil {
			return err
		}
		if result.NextMaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.NextMaxID)
		}
		return nil
	}
}

func newCommentsRepliesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replies",
		Short: "List replies to a comment",
		RunE:  makeRunCommentsReplies(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().String("comment-id", "", "Comment ID to fetch replies for")
	_ = cmd.MarkFlagRequired("comment-id")
	cmd.Flags().Int("limit", 20, "Maximum number of replies to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunCommentsReplies(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")
		commentID, _ := cmd.Flags().GetString("comment-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("min_id", cursor)
		}

		resp, err := client.MobileGet(ctx, "/api/v1/media/"+mediaID+"/comments/"+commentID+"/inline_child_comments/", params)
		if err != nil {
			return fmt.Errorf("listing replies for comment %s: %w", commentID, err)
		}

		var result commentsListResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding replies response: %w", err)
		}

		summaries := make([]CommentSummary, 0, len(result.Comments))
		for _, c := range result.Comments {
			summaries = append(summaries, toCommentSummary(c))
		}

		return printCommentSummaries(cmd, summaries)
	}
}

func newCommentsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Post a comment on a media item",
		RunE:  makeRunCommentsCreate(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID to comment on")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().String("text", "", "Comment text")
	_ = cmd.MarkFlagRequired("text")
	cmd.Flags().String("reply-to", "", "Comment ID to reply to (optional)")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCommentsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")
		text, _ := cmd.Flags().GetString("text")
		replyTo, _ := cmd.Flags().GetString("reply-to")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("post comment on media %s: %q", mediaID, text),
				map[string]string{"media_id": mediaID, "text": text, "reply_to_comment_id": replyTo})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("comment_text", text)
		if replyTo != "" {
			body.Set("replied_to_comment_id", replyTo)
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+mediaID+"/comment/", body)
		if err != nil {
			return fmt.Errorf("creating comment on media %s: %w", mediaID, err)
		}

		var result commentActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding create comment response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Comment posted on media %s\n", mediaID)
		return nil
	}
}

func newCommentsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a comment",
		RunE:  makeRunCommentsDelete(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().String("comment-id", "", "Comment ID to delete")
	_ = cmd.MarkFlagRequired("comment-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCommentsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")
		commentID, _ := cmd.Flags().GetString("comment-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("delete comment %s on media %s", commentID, mediaID),
				map[string]string{"media_id": mediaID, "comment_id": commentID})
		}
		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+mediaID+"/comment/"+commentID+"/delete/", nil)
		if err != nil {
			return fmt.Errorf("deleting comment %s on media %s: %w", commentID, mediaID, err)
		}

		var result commentActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding delete comment response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Deleted comment %s on media %s\n", commentID, mediaID)
		return nil
	}
}

func newCommentsLikeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "like",
		Short: "Like a comment",
		RunE:  makeRunCommentsLike(factory),
	}
	cmd.Flags().String("comment-id", "", "Comment ID to like")
	_ = cmd.MarkFlagRequired("comment-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCommentsLike(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		commentID, _ := cmd.Flags().GetString("comment-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("like comment %s", commentID),
				map[string]string{"comment_id": commentID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+commentID+"/comment_like/", nil)
		if err != nil {
			return fmt.Errorf("liking comment %s: %w", commentID, err)
		}

		var result commentActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding like comment response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Liked comment %s\n", commentID)
		return nil
	}
}

func newCommentsUnlikeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlike",
		Short: "Unlike a comment",
		RunE:  makeRunCommentsUnlike(factory),
	}
	cmd.Flags().String("comment-id", "", "Comment ID to unlike")
	_ = cmd.MarkFlagRequired("comment-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCommentsUnlike(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		commentID, _ := cmd.Flags().GetString("comment-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("unlike comment %s", commentID),
				map[string]string{"comment_id": commentID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+commentID+"/comment_unlike/", nil)
		if err != nil {
			return fmt.Errorf("unliking comment %s: %w", commentID, err)
		}

		var result commentActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding unlike comment response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Unliked comment %s\n", commentID)
		return nil
	}
}

func newCommentsDisableCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable comments on a post",
		RunE:  makeRunCommentsDisable(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCommentsDisable(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("disable comments on media %s", mediaID),
				map[string]string{"media_id": mediaID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+mediaID+"/disable_comments/", nil)
		if err != nil {
			return fmt.Errorf("disabling comments on media %s: %w", mediaID, err)
		}

		var result commentActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding disable comments response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Disabled comments on media %s\n", mediaID)
		return nil
	}
}

func newCommentsEnableCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable comments on a post",
		RunE:  makeRunCommentsEnable(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCommentsEnable(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("enable comments on media %s", mediaID),
				map[string]string{"media_id": mediaID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+mediaID+"/enable_comments/", nil)
		if err != nil {
			return fmt.Errorf("enabling comments on media %s: %w", mediaID, err)
		}

		var result commentActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding enable comments response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Enabled comments on media %s\n", mediaID)
		return nil
	}
}

// parseMediaDetailComments parses comments from the PolarisPostActionLoadPostQueryQuery
// response. The data contains xdt_shortcode_media.edge_media_to_parent_comment.edges.
func parseMediaDetailComments(data json.RawMessage, limit int) ([]CommentSummary, error) {
	var outer map[string]json.RawMessage
	if err := json.Unmarshal(data, &outer); err != nil {
		return nil, fmt.Errorf("unmarshal graphql data: %w", err)
	}

	mediaRaw, ok := outer["xdt_shortcode_media"]
	if !ok {
		return nil, fmt.Errorf("graphql response missing xdt_shortcode_media")
	}

	var media struct {
		EdgeMediaToParentComment struct {
			Count int64 `json:"count"`
			Edges []struct {
				Node struct {
					ID        string `json:"id"`
					Text      string `json:"text"`
					CreatedAt int64  `json:"created_at"`
					Owner     struct {
						ID       string `json:"id"`
						Username string `json:"username"`
					} `json:"owner"`
					EdgeLikedBy struct {
						Count int64 `json:"count"`
					} `json:"edge_liked_by"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"edge_media_to_parent_comment"`
	}
	if err := json.Unmarshal(mediaRaw, &media); err != nil {
		return nil, fmt.Errorf("unmarshal media detail: %w", err)
	}

	edges := media.EdgeMediaToParentComment.Edges
	if limit > 0 && len(edges) > limit {
		edges = edges[:limit]
	}

	summaries := make([]CommentSummary, 0, len(edges))
	for _, edge := range edges {
		n := edge.Node
		summaries = append(summaries, CommentSummary{
			PK:        n.ID,
			Text:      n.Text,
			Timestamp: n.CreatedAt,
			LikeCount: n.EdgeLikedBy.Count,
			UserID:    n.Owner.ID,
			Username:  n.Owner.Username,
		})
	}
	return summaries, nil
}

// parseGraphQLComments parses the GraphQL PolarisPostChildCommentsQuery response data
// into a slice of CommentSummary. The data field contains a deeply nested connection object.
func parseGraphQLComments(data json.RawMessage) ([]CommentSummary, error) {
	// The GraphQL response wraps everything in a long connection field name.
	// We unmarshal into a map first and extract the connection dynamically.
	var outer map[string]json.RawMessage
	if err := json.Unmarshal(data, &outer); err != nil {
		return nil, fmt.Errorf("unmarshal graphql data: %w", err)
	}

	// The connection field name is the long xdt_api__v1__... key.
	const connectionKey = "xdt_api__v1__media__media_id__comments__parent_comment_id__child_comments__connection"
	connRaw, ok := outer[connectionKey]
	if !ok {
		return nil, fmt.Errorf("graphql response missing connection field")
	}

	var conn struct {
		Edges []struct {
			Node graphQLCommentsNode `json:"node"`
		} `json:"edges"`
	}
	if err := json.Unmarshal(connRaw, &conn); err != nil {
		return nil, fmt.Errorf("unmarshal graphql connection: %w", err)
	}

	summaries := make([]CommentSummary, 0, len(conn.Edges))
	for _, edge := range conn.Edges {
		n := edge.Node
		summaries = append(summaries, CommentSummary{
			PK:        n.PK,
			Text:      n.Text,
			Timestamp: n.CreatedAtUTC,
			LikeCount: n.CommentLikeCount,
			UserID:    n.User.PK,
			Username:  n.User.Username,
		})
	}
	return summaries, nil
}

// toCommentSummary converts a rawComment to CommentSummary.
func toCommentSummary(c rawComment) CommentSummary {
	return CommentSummary{
		PK:        c.PK,
		Text:      c.Text,
		Timestamp: c.CreatedAt,
		LikeCount: c.CommentLikeCount,
		UserID:    c.User.PK,
		Username:  c.User.Username,
	}
}

// printCommentSummaries outputs comment summaries as JSON or text.
func printCommentSummaries(cmd *cobra.Command, comments []CommentSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(comments)
	}

	if len(comments) == 0 {
		fmt.Println("No comments found.")
		return nil
	}

	lines := make([]string, 0, len(comments)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-15s  %-16s  %-40s", "ID", "USERNAME", "DATE", "TEXT"))
	for _, c := range comments {
		lines = append(lines, fmt.Sprintf("%-20s  %-15s  %-16s  %-40s",
			truncate(c.PK, 20),
			truncate(c.Username, 15),
			formatTimestamp(c.Timestamp),
			truncate(c.Text, 40),
		))
	}
	cli.PrintText(lines)
	return nil
}
