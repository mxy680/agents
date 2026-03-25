package instagram

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

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

// newCommentsCmd builds the `comments` subcommand group.
func newCommentsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "comments",
		Short:   "View comments",
		Aliases: []string{"comment"},
	}
	cmd.AddCommand(newCommentsListCmd(factory))
	cmd.AddCommand(newCommentsRepliesCmd(factory))
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

		resp, err := client.MobileGet(ctx, "/api/v1/media/"+url.PathEscape(mediaID)+"/comments/", params)
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
			infoResp, infoErr := client.MobileGet(ctx, "/api/v1/media/"+url.PathEscape(mediaID)+"/info/", nil)
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
			parsed, parseErr := parseMediaDetailComments(gqlData, limit)
			if parseErr != nil {
				return fmt.Errorf("comments unavailable for media %s: %w", mediaID, parseErr)
			}
			summaries = parsed
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

		resp, err := client.MobileGet(ctx, "/api/v1/media/"+url.PathEscape(mediaID)+"/comments/"+url.PathEscape(commentID)+"/inline_child_comments/", params)
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
