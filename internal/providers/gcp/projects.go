package gcp

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newProjectsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List GCP projects",
		RunE:  makeRunProjectsList(factory),
	}
	cmd.Flags().String("parent", "", "Parent resource (e.g. folders/123 or organizations/456)")
	return cmd
}

func makeRunProjectsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		parent, _ := cmd.Flags().GetString("parent")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		url := resourceManagerBaseURL + "/projects"
		if parent != "" {
			url += "?parent=" + parent
		}

		var resp struct {
			Projects []map[string]any `json:"projects"`
		}
		if err := client.doJSON(ctx, http.MethodGet, url, nil, &resp); err != nil {
			return fmt.Errorf("listing projects: %w", err)
		}

		summaries := make([]ProjectSummary, 0, len(resp.Projects))
		for _, p := range resp.Projects {
			summaries = append(summaries, toProjectSummary(p))
		}

		return printProjectSummaries(cmd, summaries)
	}
}

func newProjectsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get project details",
		RunE:  makeRunProjectsGet(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (required)")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func makeRunProjectsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		url := resourceManagerBaseURL + "/projects/" + project
		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, url, nil, &data); err != nil {
			return fmt.Errorf("getting project %q: %w", project, err)
		}

		detail := toProjectDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("Name:         %s", detail.Name),
			fmt.Sprintf("Project ID:   %s", detail.ProjectID),
			fmt.Sprintf("Display Name: %s", detail.DisplayName),
			fmt.Sprintf("State:        %s", detail.State),
			fmt.Sprintf("Parent:       %s", detail.Parent),
			fmt.Sprintf("Created:      %s", detail.CreateTime),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newProjectsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new GCP project",
		RunE:  makeRunProjectsCreate(factory),
	}
	cmd.Flags().String("project", "", "Project ID (required, globally unique)")
	cmd.Flags().String("display-name", "", "Display name (defaults to project ID if omitted)")
	cmd.Flags().String("parent", "", "Parent resource (e.g. folders/123 or organizations/456)")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func makeRunProjectsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		projectID, _ := cmd.Flags().GetString("project")
		displayName, _ := cmd.Flags().GetString("display-name")
		parent, _ := cmd.Flags().GetString("parent")

		if displayName == "" {
			displayName = projectID
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create project %q", projectID), map[string]any{
				"action":      "create",
				"projectId":   projectID,
				"displayName": displayName,
				"parent":      parent,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"projectId":   projectID,
			"displayName": displayName,
		}
		if parent != "" {
			body["parent"] = parent
		}

		url := resourceManagerBaseURL + "/projects"
		var op Operation
		if err := client.doJSON(ctx, http.MethodPost, url, body, &op); err != nil {
			return fmt.Errorf("creating project %q: %w", projectID, err)
		}

		finalOp, err := client.waitForOperation(ctx, &op)
		if err != nil {
			return fmt.Errorf("project creation failed: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"status":    "created",
				"projectId": projectID,
				"operation": finalOp.Name,
			})
		}
		fmt.Printf("Created project: %s\n", projectID)
		return nil
	}
}

func newProjectsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a GCP project (irreversible)",
		RunE:  makeRunProjectsDelete(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func makeRunProjectsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete project %q", project), map[string]any{
				"action":    "delete",
				"projectId": project,
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

		url := resourceManagerBaseURL + "/projects/" + project
		if _, err := client.do(ctx, http.MethodDelete, url, nil); err != nil {
			return fmt.Errorf("deleting project %q: %w", project, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "projectId": project})
		}
		fmt.Printf("Deleted project: %s\n", project)
		return nil
	}
}
