package vercel

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newDeploymentsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployments",
		RunE:  makeRunDeploymentsList(factory),
	}
	cmd.Flags().String("project", "", "Filter by project name or ID")
	cmd.Flags().Int("limit", 20, "Maximum number of deployments to return")
	cmd.Flags().String("page-token", "", "Pagination cursor (until timestamp)")
	cmd.Flags().String("target", "", "Filter by target: production or preview")
	cmd.Flags().String("state", "", "Filter by state: BUILDING, READY, ERROR, CANCELED")
	return cmd
}

func makeRunDeploymentsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")
		target, _ := cmd.Flags().GetString("target")
		state, _ := cmd.Flags().GetString("state")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/v6/deployments?limit=%d", limit)
		if project != "" {
			path += "&projectId=" + project
		}
		if target != "" {
			path += "&target=" + target
		}
		if state != "" {
			path += "&state=" + state
		}
		if pageToken != "" {
			path += "&until=" + pageToken
		}

		var resp struct {
			Deployments []map[string]any `json:"deployments"`
			Pagination  struct {
				Next int64 `json:"next"`
			} `json:"pagination"`
		}
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing deployments: %w", err)
		}

		summaries := make([]DeploymentSummary, 0, len(resp.Deployments))
		for _, d := range resp.Deployments {
			summaries = append(summaries, toDeploymentSummary(d))
		}

		if resp.Pagination.Next != 0 && !cli.IsJSONOutput(cmd) {
			warnf("more results available — use --page-token=%d to fetch next page", resp.Pagination.Next)
		}

		return printDeploymentSummaries(cmd, summaries)
	}
}

func newDeploymentsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get deployment details",
		RunE:  makeRunDeploymentsGet(factory),
	}
	cmd.Flags().String("id", "", "Deployment ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunDeploymentsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v13/deployments/%s", id), nil, &data); err != nil {
			return fmt.Errorf("getting deployment %q: %w", id, err)
		}

		detail := toDeploymentDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", detail.ID),
			fmt.Sprintf("URL:         %s", detail.URL),
			fmt.Sprintf("Name:        %s", detail.Name),
			fmt.Sprintf("State:       %s", detail.State),
			fmt.Sprintf("Ready State: %s", detail.ReadyState),
			fmt.Sprintf("Target:      %s", detail.Target),
			fmt.Sprintf("Type:        %s", detail.Type),
			fmt.Sprintf("Source:      %s", detail.Source),
			fmt.Sprintf("Creator:     %s", detail.Creator),
			fmt.Sprintf("Git Branch:  %s", detail.GitBranch),
			fmt.Sprintf("Git Commit:  %s", detail.GitCommit),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newDeploymentsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new deployment",
		RunE:  makeRunDeploymentsCreate(factory),
	}
	cmd.Flags().String("project", "", "Project name or ID (required)")
	cmd.Flags().String("ref", "", "Git branch or commit SHA to deploy")
	cmd.Flags().String("target", "production", "Deployment target: production or preview")
	cmd.Flags().Bool("force", false, "Force a new deployment even if already up to date")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func makeRunDeploymentsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")
		ref, _ := cmd.Flags().GetString("ref")
		target, _ := cmd.Flags().GetString("target")
		force, _ := cmd.Flags().GetBool("force")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would deploy project %q (target: %s, ref: %s)", project, target, ref), map[string]any{
				"action":  "create",
				"project": project,
				"ref":     ref,
				"target":  target,
				"force":   force,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"name":   project,
			"target": target,
			"source": "cli",
		}
		if ref != "" {
			body["gitSource"] = map[string]any{"ref": ref, "type": "github"}
		}
		if force {
			body["forceNew"] = true
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, "/v13/deployments", body, &data); err != nil {
			return fmt.Errorf("creating deployment for project %q: %w", project, err)
		}

		detail := toDeploymentDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Deployment created: %s (state: %s)\n", detail.ID, detail.State)
		if detail.URL != "" {
			fmt.Printf("URL: https://%s\n", detail.URL)
		}
		return nil
	}
}

func newDeploymentsCancelCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel an in-progress deployment",
		RunE:  makeRunDeploymentsCancel(factory),
	}
	cmd.Flags().String("id", "", "Deployment ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunDeploymentsCancel(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would cancel deployment %q", id), map[string]any{
				"action": "cancel",
				"id":     id,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPatch, fmt.Sprintf("/v12/deployments/%s/cancel", id), nil, &data); err != nil {
			return fmt.Errorf("canceling deployment %q: %w", id, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		fmt.Printf("Canceled deployment: %s\n", id)
		return nil
	}
}

func newDeploymentsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a deployment (irreversible)",
		RunE:  makeRunDeploymentsDelete(factory),
	}
	cmd.Flags().String("id", "", "Deployment ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunDeploymentsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete deployment %q", id), map[string]any{
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/v13/deployments/%s", id), nil); err != nil {
			return fmt.Errorf("deleting deployment %q: %w", id, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "id": id})
		}
		fmt.Printf("Deleted deployment: %s\n", id)
		return nil
	}
}
