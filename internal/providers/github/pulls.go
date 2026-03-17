package github

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newPullsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests",
		RunE:  makeRunPullsList(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("state", "open", "PR state: open, closed, all")
	cmd.Flags().String("head", "", "Filter by head branch (user:branch)")
	cmd.Flags().String("base", "", "Filter by base branch")
	cmd.Flags().String("sort", "created", "Sort by: created, updated, popularity, long-running")
	cmd.Flags().Int("limit", 20, "Maximum number of pull requests to return")
	cmd.Flags().String("page-token", "", "Page number for pagination")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunPullsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		state, _ := cmd.Flags().GetString("state")
		head, _ := cmd.Flags().GetString("head")
		base, _ := cmd.Flags().GetString("base")
		sort, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls?state=%s&sort=%s&per_page=%d", owner, repo, state, sort, limit)
		if head != "" {
			path += "&head=" + head
		}
		if base != "" {
			path += "&base=" + base
		}
		if pageToken != "" {
			path += "&page=" + pageToken
		}

		var data []map[string]any
		if _, err := doGitHub(client, "GET", path, nil, &data); err != nil {
			return fmt.Errorf("listing pull requests: %w", err)
		}

		summaries := make([]PullSummary, 0, len(data))
		for _, item := range data {
			summaries = append(summaries, toPullSummary(item))
		}
		return printPullSummaries(cmd, summaries)
	}
}

func newPullsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a pull request by number",
		RunE:  makeRunPullsGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int("number", 0, "Pull request number (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("number")
	return cmd
}

func makeRunPullsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		number, _ := cmd.Flags().GetInt("number")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number)
		var data map[string]any
		if _, err := doGitHub(client, "GET", path, nil, &data); err != nil {
			return fmt.Errorf("getting pull request #%d: %w", number, err)
		}

		detail := toPullDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("Number:     #%d", detail.Number),
			fmt.Sprintf("Title:      %s", detail.Title),
			fmt.Sprintf("State:      %s", detail.State),
			fmt.Sprintf("User:       %s", detail.User),
			fmt.Sprintf("Head:       %s", detail.Head),
			fmt.Sprintf("Base:       %s", detail.Base),
			fmt.Sprintf("Draft:      %v", detail.Draft),
			fmt.Sprintf("Additions:  %d", detail.Additions),
			fmt.Sprintf("Deletions:  %d", detail.Deletions),
			fmt.Sprintf("Commits:    %d", detail.Commits),
			fmt.Sprintf("URL:        %s", detail.URL),
			fmt.Sprintf("Created:    %s", detail.CreatedAt),
			fmt.Sprintf("Updated:    %s", detail.UpdatedAt),
		}
		if detail.MergedAt != "" {
			lines = append(lines, fmt.Sprintf("Merged:     %s", detail.MergedAt))
		}
		if detail.Body != "" {
			lines = append(lines, fmt.Sprintf("Body:\n%s", detail.Body))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newPullsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request",
		RunE:  makeRunPullsCreate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("title", "", "Pull request title (required)")
	cmd.Flags().String("head", "", "Head branch (required)")
	cmd.Flags().String("base", "", "Base branch (required)")
	cmd.Flags().String("body", "", "Pull request body")
	cmd.Flags().Bool("draft", false, "Create as draft pull request")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("head")
	_ = cmd.MarkFlagRequired("base")
	return cmd
}

func makeRunPullsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		title, _ := cmd.Flags().GetString("title")
		head, _ := cmd.Flags().GetString("head")
		base, _ := cmd.Flags().GetString("base")
		body, _ := cmd.Flags().GetString("body")
		draft, _ := cmd.Flags().GetBool("draft")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create PR %q from %q into %q", title, head, base), map[string]any{
				"action": "create_pull",
				"owner":  owner,
				"repo":   repo,
				"title":  title,
				"head":   head,
				"base":   base,
				"body":   body,
				"draft":  draft,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"title": title,
			"head":  head,
			"base":  base,
			"body":  body,
			"draft": draft,
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)
		var data map[string]any
		if _, err := doGitHub(client, "POST", path, reqBody, &data); err != nil {
			return fmt.Errorf("creating pull request: %w", err)
		}

		detail := toPullDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Created PR #%d: %s\n%s\n", detail.Number, detail.Title, detail.URL)
		return nil
	}
}

func newPullsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a pull request",
		RunE:  makeRunPullsUpdate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int("number", 0, "Pull request number (required)")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("body", "", "New body")
	cmd.Flags().String("state", "", "New state: open or closed")
	cmd.Flags().String("base", "", "New base branch")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("number")
	return cmd
}

func makeRunPullsUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		number, _ := cmd.Flags().GetInt("number")

		// Only include fields that were explicitly set
		reqBody := map[string]any{}
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			reqBody["title"] = v
		}
		if cmd.Flags().Changed("body") {
			v, _ := cmd.Flags().GetString("body")
			reqBody["body"] = v
		}
		if cmd.Flags().Changed("state") {
			v, _ := cmd.Flags().GetString("state")
			reqBody["state"] = v
		}
		if cmd.Flags().Changed("base") {
			v, _ := cmd.Flags().GetString("base")
			reqBody["base"] = v
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update PR #%d in %s/%s", number, owner, repo), map[string]any{
				"action": "update_pull",
				"owner":  owner,
				"repo":   repo,
				"number": number,
				"fields": reqBody,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number)
		var data map[string]any
		if _, err := doGitHub(client, "PATCH", path, reqBody, &data); err != nil {
			return fmt.Errorf("updating pull request #%d: %w", number, err)
		}

		detail := toPullDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Updated PR #%d: %s\n", detail.Number, detail.Title)
		return nil
	}
}

func newPullsMergeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge a pull request",
		RunE:  makeRunPullsMerge(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int("number", 0, "Pull request number (required)")
	cmd.Flags().String("method", "merge", "Merge method: merge, squash, rebase")
	cmd.Flags().String("commit-title", "", "Title for the merge commit")
	cmd.Flags().String("commit-message", "", "Message for the merge commit")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("number")
	return cmd
}

func makeRunPullsMerge(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		number, _ := cmd.Flags().GetInt("number")
		method, _ := cmd.Flags().GetString("method")
		commitTitle, _ := cmd.Flags().GetString("commit-title")
		commitMessage, _ := cmd.Flags().GetString("commit-message")

		// Validate merge method
		switch method {
		case "merge", "squash", "rebase":
			// valid
		default:
			return fmt.Errorf("invalid merge method %q; must be one of: merge, squash, rebase", method)
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would merge PR #%d in %s/%s using method %q", number, owner, repo, method), map[string]any{
				"action":        "merge_pull",
				"owner":         owner,
				"repo":          repo,
				"number":        number,
				"mergeMethod":   method,
				"commitTitle":   commitTitle,
				"commitMessage": commitMessage,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"merge_method": method,
		}
		if commitTitle != "" {
			reqBody["commit_title"] = commitTitle
		}
		if commitMessage != "" {
			reqBody["commit_message"] = commitMessage
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, number)
		var data map[string]any
		if _, err := doGitHub(client, "PUT", path, reqBody, &data); err != nil {
			return fmt.Errorf("merging pull request #%d: %w", number, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		sha, _ := data["sha"].(string)
		msg, _ := data["message"].(string)
		merged, _ := data["merged"].(bool)
		fmt.Printf("Merged: %v\nSHA: %s\nMessage: %s\n", merged, sha, msg)
		return nil
	}
}

func newPullsReviewCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Submit a review on a pull request",
		RunE:  makeRunPullsReview(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int("number", 0, "Pull request number (required)")
	cmd.Flags().String("event", "", "Review event: APPROVE, REQUEST_CHANGES, COMMENT (required)")
	cmd.Flags().String("body", "", "Review body text")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("event")
	return cmd
}

func makeRunPullsReview(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		number, _ := cmd.Flags().GetInt("number")
		event, _ := cmd.Flags().GetString("event")
		body, _ := cmd.Flags().GetString("body")

		// Validate review event
		switch event {
		case "APPROVE", "REQUEST_CHANGES", "COMMENT":
			// valid
		default:
			return fmt.Errorf("invalid review event %q; must be one of: APPROVE, REQUEST_CHANGES, COMMENT", event)
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would submit %s review on PR #%d in %s/%s", event, number, owner, repo), map[string]any{
				"action": "review_pull",
				"owner":  owner,
				"repo":   repo,
				"number": number,
				"event":  event,
				"body":   body,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"event": event,
			"body":  body,
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", owner, repo, number)
		var data map[string]any
		if _, err := doGitHub(client, "POST", path, reqBody, &data); err != nil {
			return fmt.Errorf("submitting review on pull request #%d: %w", number, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		id := jsonInt64(data["id"])
		state, _ := data["state"].(string)
		reviewBody, _ := data["body"].(string)
		fmt.Printf("Review submitted: ID=%d state=%s\n", id, state)
		if reviewBody != "" {
			fmt.Printf("Body: %s\n", reviewBody)
		}
		return nil
	}
}
