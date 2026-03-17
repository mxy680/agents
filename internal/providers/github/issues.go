package github

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newIssuesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues in a repository",
		RunE:  makeRunIssuesList(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("state", "open", "Issue state: open, closed, all")
	cmd.Flags().String("labels", "", "Comma-separated list of label names to filter by")
	cmd.Flags().String("assignee", "", "Filter by assignee login")
	cmd.Flags().String("sort", "created", "Sort field: created, updated, comments")
	cmd.Flags().Int("limit", 20, "Maximum number of issues to return")
	cmd.Flags().String("page-token", "", "Page number for pagination")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunIssuesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		state, _ := cmd.Flags().GetString("state")
		labels, _ := cmd.Flags().GetString("labels")
		assignee, _ := cmd.Flags().GetString("assignee")
		sort, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/issues?state=%s&sort=%s&per_page=%d", owner, repo, state, sort, limit)
		if labels != "" {
			path += "&labels=" + labels
		}
		if assignee != "" {
			path += "&assignee=" + assignee
		}
		if pageToken != "" {
			path += "&page=" + pageToken
		}

		var raw []map[string]any
		if _, err := doGitHub(client, "GET", path, nil, &raw); err != nil {
			return fmt.Errorf("listing issues for %s/%s: %w", owner, repo, err)
		}

		summaries := make([]IssueSummary, 0, len(raw))
		for _, item := range raw {
			summaries = append(summaries, toIssueSummary(item))
		}
		return printIssueSummaries(cmd, summaries)
	}
}

func newIssuesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an issue by number",
		RunE:  makeRunIssuesGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int("number", 0, "Issue number (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("number")
	return cmd
}

func makeRunIssuesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		number, _ := cmd.Flags().GetInt("number")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
		var raw map[string]any
		if _, err := doGitHub(client, "GET", path, nil, &raw); err != nil {
			return fmt.Errorf("getting issue #%d for %s/%s: %w", number, owner, repo, err)
		}

		detail := toIssueDetail(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("Number:    #%d", detail.Number),
			fmt.Sprintf("Title:     %s", detail.Title),
			fmt.Sprintf("State:     %s", detail.State),
			fmt.Sprintf("User:      %s", detail.User),
			fmt.Sprintf("Comments:  %d", detail.Comments),
			fmt.Sprintf("Created:   %s", detail.CreatedAt),
			fmt.Sprintf("Updated:   %s", detail.UpdatedAt),
		}
		if detail.ClosedAt != "" {
			lines = append(lines, fmt.Sprintf("Closed:    %s", detail.ClosedAt))
		}
		if len(detail.Labels) > 0 {
			lines = append(lines, fmt.Sprintf("Labels:    %s", strings.Join(detail.Labels, ", ")))
		}
		if len(detail.Assignees) > 0 {
			lines = append(lines, fmt.Sprintf("Assignees: %s", strings.Join(detail.Assignees, ", ")))
		}
		if detail.URL != "" {
			lines = append(lines, fmt.Sprintf("URL:       %s", detail.URL))
		}
		if detail.Body != "" {
			lines = append(lines, fmt.Sprintf("Body:\n%s", detail.Body))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newIssuesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		RunE:  makeRunIssuesCreate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("title", "", "Issue title (required)")
	cmd.Flags().String("body", "", "Issue body")
	cmd.Flags().String("labels", "", "Comma-separated list of label names")
	cmd.Flags().String("assignees", "", "Comma-separated list of assignee logins")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func makeRunIssuesCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		labelsStr, _ := cmd.Flags().GetString("labels")
		assigneesStr, _ := cmd.Flags().GetString("assignees")

		reqBody := map[string]any{
			"title": title,
		}
		if body != "" {
			reqBody["body"] = body
		}
		if labelsStr != "" {
			reqBody["labels"] = strings.Split(labelsStr, ",")
		}
		if assigneesStr != "" {
			reqBody["assignees"] = strings.Split(assigneesStr, ",")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
		var raw map[string]any
		if _, err := doGitHub(client, "POST", path, reqBody, &raw); err != nil {
			return fmt.Errorf("creating issue in %s/%s: %w", owner, repo, err)
		}

		detail := toIssueDetail(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Created issue #%d: %s\n%s\n", detail.Number, detail.Title, detail.URL)
		return nil
	}
}

func newIssuesUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing issue",
		RunE:  makeRunIssuesUpdate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int("number", 0, "Issue number (required)")
	cmd.Flags().String("title", "", "New issue title")
	cmd.Flags().String("body", "", "New issue body")
	cmd.Flags().String("state", "", "New issue state: open, closed")
	cmd.Flags().String("labels", "", "Comma-separated list of label names")
	cmd.Flags().String("assignees", "", "Comma-separated list of assignee logins")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("number")
	return cmd
}

func makeRunIssuesUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		number, _ := cmd.Flags().GetInt("number")

		reqBody := map[string]any{}
		if cmd.Flags().Changed("title") {
			title, _ := cmd.Flags().GetString("title")
			reqBody["title"] = title
		}
		if cmd.Flags().Changed("body") {
			body, _ := cmd.Flags().GetString("body")
			reqBody["body"] = body
		}
		if cmd.Flags().Changed("state") {
			state, _ := cmd.Flags().GetString("state")
			reqBody["state"] = state
		}
		if cmd.Flags().Changed("labels") {
			labelsStr, _ := cmd.Flags().GetString("labels")
			reqBody["labels"] = strings.Split(labelsStr, ",")
		}
		if cmd.Flags().Changed("assignees") {
			assigneesStr, _ := cmd.Flags().GetString("assignees")
			reqBody["assignees"] = strings.Split(assigneesStr, ",")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
		var raw map[string]any
		if _, err := doGitHub(client, "PATCH", path, reqBody, &raw); err != nil {
			return fmt.Errorf("updating issue #%d in %s/%s: %w", number, owner, repo, err)
		}

		detail := toIssueDetail(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Updated issue #%d: %s\n", detail.Number, detail.Title)
		return nil
	}
}

func newIssuesCloseCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close an issue",
		RunE:  makeRunIssuesClose(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int("number", 0, "Issue number (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("number")
	return cmd
}

func makeRunIssuesClose(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		number, _ := cmd.Flags().GetInt("number")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would close issue #%d in %s/%s", number, owner, repo), map[string]any{
				"action": "close",
				"owner":  owner,
				"repo":   repo,
				"number": number,
				"state":  "closed",
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
		var raw map[string]any
		if _, err := doGitHub(client, "PATCH", path, map[string]any{"state": "closed"}, &raw); err != nil {
			return fmt.Errorf("closing issue #%d in %s/%s: %w", number, owner, repo, err)
		}

		detail := toIssueDetail(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Closed issue #%d: %s\n", detail.Number, detail.Title)
		return nil
	}
}

func newIssuesCommentCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Add a comment to an issue",
		RunE:  makeRunIssuesComment(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int("number", 0, "Issue number (required)")
	cmd.Flags().String("body", "", "Comment body (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("body")
	return cmd
}

func makeRunIssuesComment(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		number, _ := cmd.Flags().GetInt("number")
		body, _ := cmd.Flags().GetString("body")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would add comment to issue #%d in %s/%s", number, owner, repo), map[string]any{
				"action": "comment",
				"owner":  owner,
				"repo":   repo,
				"number": number,
				"body":   body,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, number)
		var raw map[string]any
		if _, err := doGitHub(client, "POST", path, map[string]any{"body": body}, &raw); err != nil {
			return fmt.Errorf("adding comment to issue #%d in %s/%s: %w", number, owner, repo, err)
		}

		info := IssueCommentInfo{
			ID:        jsonInt64(raw["id"]),
			User:      jsonNestedString(raw["user"], "login"),
			Body:      jsonString(raw["body"]),
			CreatedAt: jsonString(raw["created_at"]),
		}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Printf("Added comment %d to issue #%d\n", info.ID, number)
		return nil
	}
}
