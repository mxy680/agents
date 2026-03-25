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
