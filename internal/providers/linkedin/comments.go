package linkedin

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerCommentsResponse is the response envelope for listing comments on a post.
type voyagerCommentsResponse struct {
	Elements []voyagerCommentElement `json:"elements"`
	Paging   voyagerPaging           `json:"paging"`
}

// voyagerCommentElement represents a single comment in the list response.
type voyagerCommentElement struct {
	URN       string `json:"urn"`
	Commenter *struct {
		// LinkedIn wraps the commenter in a type-keyed field.
		MemberActor *struct {
			MiniProfile struct {
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
				EntityURN string `json:"entityUrn"`
			} `json:"miniProfile"`
		} `json:"com.linkedin.voyager.feed.MemberActor"`
	} `json:"commenter"`
	Comment *struct {
		Values []struct {
			Value string `json:"value"`
		} `json:"values"`
	} `json:"comment"`
	SocialDetail *struct {
		TotalSocialActivityCounts struct {
			NumLikes int `json:"numLikes"`
		} `json:"totalSocialActivityCounts"`
	} `json:"socialDetail"`
	CreatedAt int64 `json:"createdAt"`
}

// newCommentsCmd builds the "comments" subcommand group.
func newCommentsCmd(factory ClientFactory) *cobra.Command {
	commentsCmd := &cobra.Command{
		Use:     "comments",
		Short:   "Manage LinkedIn post comments",
		Aliases: []string{"comment"},
	}
	commentsCmd.AddCommand(newCommentsListCmd(factory))
	commentsCmd.AddCommand(newCommentsCreateCmd(factory))
	commentsCmd.AddCommand(newCommentsDeleteCmd(factory))
	commentsCmd.AddCommand(newCommentsLikeCmd(factory))
	commentsCmd.AddCommand(newCommentsUnlikeCmd(factory))
	return commentsCmd
}

// newCommentsListCmd builds the "comments list" command.
func newCommentsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments on a LinkedIn post",
		Long:  "List comments on a post by its activity URN.",
		RunE:  makeRunCommentsList(factory),
	}
	cmd.Flags().String("post-urn", "", "Activity URN of the post (e.g. urn:li:activity:1234)")
	_ = cmd.MarkFlagRequired("post-urn")
	cmd.Flags().Int("limit", 10, "Maximum number of comments to return")
	cmd.Flags().String("cursor", "0", "Pagination start offset")
	return cmd
}

// newCommentsCreateCmd builds the "comments create" command.
func newCommentsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a comment on a LinkedIn post",
		Long:  "Post a new comment on a post.",
		RunE:  makeRunCommentsCreate(factory),
	}
	cmd.Flags().String("post-urn", "", "Activity URN of the post (e.g. urn:li:activity:1234)")
	_ = cmd.MarkFlagRequired("post-urn")
	cmd.Flags().String("text", "", "Comment text")
	_ = cmd.MarkFlagRequired("text")
	cmd.Flags().String("reply-to", "", "Comment URN to reply to (optional)")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without creating the comment")
	return cmd
}

// newCommentsDeleteCmd builds the "comments delete" command.
func newCommentsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a comment on a LinkedIn post",
		Long:  "Delete a comment by its URN.",
		RunE:  makeRunCommentsDelete(factory),
	}
	cmd.Flags().String("comment-urn", "", "URN of the comment to delete")
	_ = cmd.MarkFlagRequired("comment-urn")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without deleting")
	return cmd
}

// newCommentsLikeCmd builds the "comments like" command.
func newCommentsLikeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "like",
		Short: "Like a comment",
		Long:  "Like a comment by its URN.",
		RunE:  makeRunCommentsLike(factory),
	}
	cmd.Flags().String("comment-urn", "", "URN of the comment to like")
	_ = cmd.MarkFlagRequired("comment-urn")
	cmd.Flags().Bool("dry-run", false, "Print what would be liked without sending")
	return cmd
}

// newCommentsUnlikeCmd builds the "comments unlike" command.
func newCommentsUnlikeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlike",
		Short: "Unlike a comment",
		Long:  "Unlike a comment by its URN.",
		RunE:  makeRunCommentsUnlike(factory),
	}
	cmd.Flags().String("comment-urn", "", "URN of the comment to unlike")
	_ = cmd.MarkFlagRequired("comment-urn")
	cmd.Flags().Bool("dry-run", false, "Print what would be unliked without sending")
	return cmd
}

func makeRunCommentsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		postURN, _ := cmd.Flags().GetString("post-urn")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("count", fmt.Sprintf("%d", limit))
		params.Set("start", cursor)

		path := "/voyager/api/socialActions/" + url.PathEscape(postURN) + "/comments"
		resp, err := client.Get(ctx, path, params)
		if err != nil {
			return fmt.Errorf("listing comments for %s: %w", postURN, err)
		}

		var raw voyagerCommentsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding comments: %w", err)
		}

		comments := make([]CommentSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			comments = append(comments, commentElementToSummary(el))
		}
		return printCommentSummaries(cmd, comments)
	}
}

func makeRunCommentsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		postURN, _ := cmd.Flags().GetString("post-urn")
		text, _ := cmd.Flags().GetString("text")
		replyTo, _ := cmd.Flags().GetString("reply-to")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		body := map[string]any{
			"comment": map[string]any{
				"values": []map[string]string{
					{"value": text},
				},
			},
		}
		if replyTo != "" {
			body["parentCommentUrn"] = replyTo
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("create comment %q on post %s", truncate(text, 60), postURN), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/socialActions/" + url.PathEscape(postURN) + "/comments"
		resp, err := client.PostJSON(ctx, path, body)
		if err != nil {
			return fmt.Errorf("creating comment on %s: %w", postURN, err)
		}

		var result struct {
			Value struct {
				URN string `json:"urn"`
			} `json:"value"`
		}
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding create comment response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"urn": result.Value.URN})
		}
		fmt.Printf("Comment created: %s\n", result.Value.URN)
		return nil
	}
}

func makeRunCommentsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		commentURN, _ := cmd.Flags().GetString("comment-urn")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("delete comment %s", commentURN), map[string]string{"comment_urn": commentURN})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// Delete uses the comment URN as the path segment directly.
		path := "/voyager/api/socialActions/" + url.PathEscape(commentURN)
		resp, err := client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("deleting comment %s: %w", commentURN, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "comment_urn": commentURN})
		}
		fmt.Printf("Comment deleted: %s\n", commentURN)
		return nil
	}
}

func makeRunCommentsLike(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		commentURN, _ := cmd.Flags().GetString("comment-urn")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("like comment %s", commentURN), map[string]string{"comment_urn": commentURN})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/socialActions/" + url.PathEscape(commentURN) + "/likes"
		resp, err := client.PostJSON(ctx, path, map[string]any{})
		if err != nil {
			return fmt.Errorf("liking comment %s: %w", commentURN, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "liked", "comment_urn": commentURN})
		}
		fmt.Printf("Liked comment: %s\n", commentURN)
		return nil
	}
}

func makeRunCommentsUnlike(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		commentURN, _ := cmd.Flags().GetString("comment-urn")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("unlike comment %s", commentURN), map[string]string{"comment_urn": commentURN})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/socialActions/" + url.PathEscape(commentURN) + "/likes"
		resp, err := client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("unliking comment %s: %w", commentURN, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "unliked", "comment_urn": commentURN})
		}
		fmt.Printf("Unliked comment: %s\n", commentURN)
		return nil
	}
}

// commentElementToSummary converts a voyagerCommentElement to CommentSummary.
func commentElementToSummary(el voyagerCommentElement) CommentSummary {
	c := CommentSummary{
		URN:       el.URN,
		Timestamp: el.CreatedAt,
	}
	if el.Commenter != nil && el.Commenter.MemberActor != nil {
		mp := el.Commenter.MemberActor.MiniProfile
		c.AuthorURN = mp.EntityURN
		c.AuthorName = mp.FirstName + " " + mp.LastName
	}
	if el.Comment != nil && len(el.Comment.Values) > 0 {
		c.Text = el.Comment.Values[0].Value
	}
	if el.SocialDetail != nil {
		c.LikeCount = el.SocialDetail.TotalSocialActivityCounts.NumLikes
	}
	return c
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
	lines = append(lines, fmt.Sprintf("%-40s  %-25s  %-50s  %-12s  %-6s", "URN", "AUTHOR", "TEXT", "DATE", "LIKES"))
	for _, c := range comments {
		lines = append(lines, fmt.Sprintf("%-40s  %-25s  %-50s  %-12s  %-6s",
			truncate(c.URN, 40),
			truncate(c.AuthorName, 25),
			truncate(c.Text, 50),
			formatTimestamp(c.Timestamp),
			formatCount(c.LikeCount),
		))
	}
	cli.PrintText(lines)
	return nil
}
