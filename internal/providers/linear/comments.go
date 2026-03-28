package linear

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newCommentsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments on an issue",
		RunE:  makeRunCommentsList(factory),
	}
	cmd.Flags().String("issue", "", "Issue ID (required)")
	_ = cmd.MarkFlagRequired("issue")
	return cmd
}

func makeRunCommentsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		issueID, _ := cmd.Flags().GetString("issue")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($issueId: String!) {
  issue(id: $issueId) {
    comments {
      nodes {
        id
        body
        user { name }
        createdAt
      }
    }
  }
}`

		var resp struct {
			Issue struct {
				Comments struct {
					Nodes []struct {
						ID   string `json:"id"`
						Body string `json:"body"`
						User *struct {
							Name string `json:"name"`
						} `json:"user"`
						CreatedAt string `json:"createdAt"`
					} `json:"nodes"`
				} `json:"comments"`
			} `json:"issue"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"issueId": issueID}, &resp); err != nil {
			return fmt.Errorf("listing comments: %w", err)
		}

		comments := make([]CommentSummary, 0, len(resp.Issue.Comments.Nodes))
		for _, n := range resp.Issue.Comments.Nodes {
			c := CommentSummary{
				ID:        n.ID,
				Body:      n.Body,
				CreatedAt: n.CreatedAt,
			}
			if n.User != nil {
				c.User = n.User.Name
			}
			comments = append(comments, c)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(comments)
		}
		if len(comments) == 0 {
			fmt.Println("No comments found.")
			return nil
		}
		lines := make([]string, 0, len(comments)+1)
		lines = append(lines, fmt.Sprintf("%-28s  %-20s  %-24s  %s", "ID", "USER", "CREATED", "BODY"))
		for _, c := range comments {
			lines = append(lines, fmt.Sprintf("%-28s  %-20s  %-24s  %s",
				truncate(c.ID, 28), truncate(c.User, 20), c.CreatedAt, truncate(c.Body, 60)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newCommentsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a comment on an issue",
		RunE:  makeRunCommentsCreate(factory),
	}
	cmd.Flags().String("issue", "", "Issue ID (required)")
	cmd.Flags().String("body", "", "Comment body (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would be created without making changes")
	_ = cmd.MarkFlagRequired("issue")
	_ = cmd.MarkFlagRequired("body")
	return cmd
}

func makeRunCommentsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		issueID, _ := cmd.Flags().GetString("issue")
		body, _ := cmd.Flags().GetString("body")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create comment on issue %s", issueID), map[string]any{
				"action":  "create",
				"issueId": issueID,
				"body":    body,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
mutation($input: CommentCreateInput!) {
  commentCreate(input: $input) {
    comment {
      id
      body
    }
  }
}`

		var resp struct {
			CommentCreate struct {
				Comment struct {
					ID   string `json:"id"`
					Body string `json:"body"`
				} `json:"comment"`
			} `json:"commentCreate"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"input": map[string]any{
			"issueId": issueID,
			"body":    body,
		}}, &resp); err != nil {
			return fmt.Errorf("creating comment: %w", err)
		}

		comment := resp.CommentCreate.Comment
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(comment)
		}
		fmt.Printf("Created comment: %s\n", comment.ID)
		return nil
	}
}

func newCommentsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a comment (irreversible)",
		RunE:  makeRunCommentsDelete(factory),
	}
	cmd.Flags().String("id", "", "Comment ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without making changes")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunCommentsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete comment %q", id), map[string]any{
				"action": "delete",
				"id":     id,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
mutation($id: String!) {
  commentDelete(id: $id) {
    success
  }
}`

		var resp struct {
			CommentDelete struct {
				Success bool `json:"success"`
			} `json:"commentDelete"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("deleting comment %q: %w", id, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"success": resp.CommentDelete.Success, "id": id})
		}
		fmt.Printf("Deleted comment: %s\n", id)
		return nil
	}
}
