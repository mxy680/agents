package github

import (
	"fmt"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newRunsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflow runs for a repository",
		RunE:  makeRunRunsList(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int64("workflow-id", 0, "Filter by workflow ID")
	cmd.Flags().String("branch", "", "Filter by branch name")
	cmd.Flags().String("status", "", "Filter by status (e.g. completed, in_progress, queued)")
	cmd.Flags().Int("limit", 20, "Maximum number of runs to return")
	cmd.Flags().String("page-token", "", "Page number for pagination")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunRunsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		workflowID, _ := cmd.Flags().GetInt64("workflow-id")
		branch, _ := cmd.Flags().GetString("branch")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/actions/runs?per_page=%d", owner, repo, limit)
		if workflowID != 0 {
			path += "&workflow_id=" + strconv.FormatInt(workflowID, 10)
		}
		if branch != "" {
			path += "&branch=" + branch
		}
		if status != "" {
			path += "&status=" + status
		}
		if pageToken != "" {
			path += "&page=" + pageToken
		}

		var raw map[string]any
		if _, err := doGitHub(client, "GET", path, nil, &raw); err != nil {
			return fmt.Errorf("listing workflow runs: %w", err)
		}

		arr, _ := raw["workflow_runs"].([]any)
		summaries := make([]RunSummary, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				summaries = append(summaries, toRunSummary(m))
			}
		}
		return printRunSummaries(cmd, summaries)
	}
}

func newRunsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a workflow run by ID",
		RunE:  makeRunRunsGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int64("run-id", 0, "Workflow run ID (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("run-id")
	return cmd
}

func makeRunRunsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		runID, _ := cmd.Flags().GetInt64("run-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/actions/runs/%d", owner, repo, runID)
		var raw map[string]any
		if _, err := doGitHub(client, "GET", path, nil, &raw); err != nil {
			return fmt.Errorf("getting workflow run %d: %w", runID, err)
		}

		detail := toRunDetail(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:           %d", detail.ID),
			fmt.Sprintf("Name:         %s", detail.Name),
			fmt.Sprintf("Status:       %s", detail.Status),
			fmt.Sprintf("Conclusion:   %s", detail.Conclusion),
			fmt.Sprintf("Branch:       %s", detail.Branch),
			fmt.Sprintf("Event:        %s", detail.Event),
			fmt.Sprintf("Workflow ID:  %d", detail.WorkflowID),
			fmt.Sprintf("Run Number:   %d", detail.RunNumber),
			fmt.Sprintf("Run Attempt:  %d", detail.RunAttempt),
			fmt.Sprintf("URL:          %s", detail.URL),
			fmt.Sprintf("Created:      %s", detail.CreatedAt),
			fmt.Sprintf("Updated:      %s", detail.UpdatedAt),
			fmt.Sprintf("Started:      %s", detail.RunStartedAt),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newRunsRerunCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "re-run",
		Short: "Re-run a workflow run",
		RunE:  makeRunRunsRerun(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int64("run-id", 0, "Workflow run ID (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("run-id")
	return cmd
}

func makeRunRunsRerun(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		runID, _ := cmd.Flags().GetInt64("run-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/actions/runs/%d/rerun", owner, repo, runID)
		if _, err := doGitHub(client, "POST", path, nil, nil); err != nil {
			return fmt.Errorf("re-running workflow run %d: %w", runID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"status": "re-run triggered",
				"runId":  runID,
			})
		}
		fmt.Printf("Re-run triggered for workflow run %d\n", runID)
		return nil
	}
}

func newRunsWorkflowsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflows",
		Short: "List workflows for a repository",
		RunE:  makeRunRunsWorkflows(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunRunsWorkflows(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/actions/workflows", owner, repo)
		var raw map[string]any
		if _, err := doGitHub(client, "GET", path, nil, &raw); err != nil {
			return fmt.Errorf("listing workflows: %w", err)
		}

		arr, _ := raw["workflows"].([]any)
		workflows := make([]WorkflowSummary, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				workflows = append(workflows, WorkflowSummary{
					ID:    jsonInt64(m["id"]),
					Name:  jsonString(m["name"]),
					Path:  jsonString(m["path"]),
					State: jsonString(m["state"]),
				})
			}
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(workflows)
		}

		if len(workflows) == 0 {
			fmt.Println("No workflows found.")
			return nil
		}

		lines := make([]string, 0, len(workflows)+1)
		lines = append(lines, fmt.Sprintf("%-12s  %-30s  %-50s  %s", "ID", "NAME", "PATH", "STATE"))
		for _, w := range workflows {
			lines = append(lines, fmt.Sprintf("%-12d  %-30s  %-50s  %s", w.ID, truncate(w.Name, 30), truncate(w.Path, 50), w.State))
		}
		cli.PrintText(lines)
		return nil
	}
}
