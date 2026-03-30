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

// newCommentsCreateCmd builds the "comments create" command.
func newCommentsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Post a comment on a LinkedIn post",
		RunE:  makeRunCommentsCreate(factory),
	}
	cmd.Flags().String("post-urn", "", "Activity URN of the post (required)")
	cmd.Flags().String("text", "", "Comment text (required)")
	cmd.Flags().String("reply-to", "", "URN of comment to reply to (optional)")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("post-urn")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func makeRunCommentsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		postURN, _ := cmd.Flags().GetString("post-urn")
		text, _ := cmd.Flags().GetString("text")
		replyTo, _ := cmd.Flags().GetString("reply-to")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("create comment on %s", postURN),
				map[string]string{"post_urn": postURN, "text": text})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"actor":   "urn:li:fs_miniProfile:me",
			"message": map[string]any{"text": text},
		}
		if replyTo != "" {
			body["parentComment"] = replyTo
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
		if err := client.DecodeJSON(resp, &result); err == nil && result.Value.URN != "" {
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]string{"urn": result.Value.URN})
			}
			fmt.Printf("Comment created: %s\n", result.Value.URN)
		} else {
			fmt.Println("Comment created")
		}
		return nil
	}
}

// newCommentsDeleteCmd builds the "comments delete" command.
func newCommentsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a comment",
		RunE:  makeRunCommentsDelete(factory),
	}
	cmd.Flags().String("comment-urn", "", "Comment URN (required)")
	cmd.Flags().Bool("confirm", false, "Confirm the delete action")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("comment-urn")
	return cmd
}

func makeRunCommentsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		commentURN, _ := cmd.Flags().GetString("comment-urn")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("delete comment %s", commentURN),
				map[string]string{"status": "deleted", "comment_urn": commentURN})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/socialActions/" + url.PathEscape(commentURN)
		_, err = client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("deleting comment %s: %w", commentURN, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "comment_urn": commentURN})
		}
		fmt.Printf("Comment %s deleted\n", commentURN)
		return nil
	}
}

// newCommentsLikeCmd builds the "comments like" command.
func newCommentsLikeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "like",
		Short: "Like a comment",
		RunE:  makeRunCommentsLike(factory),
	}
	cmd.Flags().String("comment-urn", "", "Comment URN (required)")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("comment-urn")
	return cmd
}

func makeRunCommentsLike(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		commentURN, _ := cmd.Flags().GetString("comment-urn")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("like comment %s", commentURN),
				map[string]string{"status": "liked", "comment_urn": commentURN})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/socialActions/" + url.PathEscape(commentURN) + "/likes"
		_, err = client.PostJSON(ctx, path, map[string]any{})
		if err != nil {
			return fmt.Errorf("liking comment %s: %w", commentURN, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "liked", "comment_urn": commentURN})
		}
		fmt.Printf("Liked comment %s\n", commentURN)
		return nil
	}
}

// newCommentsUnlikeCmd builds the "comments unlike" command.
func newCommentsUnlikeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlike",
		Short: "Unlike a comment",
		RunE:  makeRunCommentsUnlike(factory),
	}
	cmd.Flags().String("comment-urn", "", "Comment URN (required)")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("comment-urn")
	return cmd
}

func makeRunCommentsUnlike(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		commentURN, _ := cmd.Flags().GetString("comment-urn")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("unlike comment %s", commentURN),
				map[string]string{"status": "unliked", "comment_urn": commentURN})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/socialActions/" + url.PathEscape(commentURN) + "/likes"
		_, err = client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("unliking comment %s: %w", commentURN, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "unliked", "comment_urn": commentURN})
		}
		fmt.Printf("Unliked comment %s\n", commentURN)
		return nil
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
