package vercel

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newProjectsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE:  makeRunProjectsList(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of projects to return")
	cmd.Flags().String("page-token", "", "Pagination cursor (next page token)")
	return cmd
}

func makeRunProjectsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/v10/projects?limit=%d", limit)
		if pageToken != "" {
			path += "&until=" + pageToken
		}

		var resp struct {
			Projects   []map[string]any `json:"projects"`
			Pagination struct {
				Next int64 `json:"next"`
			} `json:"pagination"`
		}
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing projects: %w", err)
		}

		summaries := make([]ProjectSummary, 0, len(resp.Projects))
		for _, p := range resp.Projects {
			summaries = append(summaries, toProjectSummary(p))
		}

		if resp.Pagination.Next != 0 && !cli.IsJSONOutput(cmd) {
			warnf("more results available — use --page-token=%d to fetch next page", resp.Pagination.Next)
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
	cmd.Flags().String("project", "", "Project name or ID (required)")
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

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v9/projects/%s", project), nil, &data); err != nil {
			return fmt.Errorf("getting project %q: %w", project, err)
		}

		detail := toProjectDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:               %s", detail.ID),
			fmt.Sprintf("Name:             %s", detail.Name),
			fmt.Sprintf("Framework:        %s", detail.Framework),
			fmt.Sprintf("Node.js:          %s", detail.NodeJS),
			fmt.Sprintf("Root Directory:   %s", detail.RootDirectory),
			fmt.Sprintf("Build Command:    %s", detail.BuildCommand),
			fmt.Sprintf("Output Directory: %s", detail.OutputDirectory),
			fmt.Sprintf("Install Command:  %s", detail.InstallCommand),
			fmt.Sprintf("Dev Command:      %s", detail.DevCommand),
			fmt.Sprintf("Account ID:       %s", detail.AccountID),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newProjectsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		RunE:  makeRunProjectsCreate(factory),
	}
	cmd.Flags().String("name", "", "Project name (required)")
	cmd.Flags().String("framework", "", "Framework preset (e.g. nextjs, vite, remix)")
	cmd.Flags().String("git-repo", "", "Git repository URL to link (optional)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunProjectsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		framework, _ := cmd.Flags().GetString("framework")
		gitRepo, _ := cmd.Flags().GetString("git-repo")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create project %q", name), map[string]any{
				"action":    "create",
				"name":      name,
				"framework": framework,
				"gitRepo":   gitRepo,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{"name": name}
		if framework != "" {
			body["framework"] = framework
		}
		if gitRepo != "" {
			body["gitRepository"] = map[string]any{"url": gitRepo}
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, "/v11/projects", body, &data); err != nil {
			return fmt.Errorf("creating project %q: %w", name, err)
		}

		detail := toProjectDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Created project: %s (ID: %s)\n", detail.Name, detail.ID)
		return nil
	}
}

func newProjectsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a project",
		RunE:  makeRunProjectsUpdate(factory),
	}
	cmd.Flags().String("project", "", "Project name or ID (required)")
	cmd.Flags().String("name", "", "New project name")
	cmd.Flags().String("framework", "", "Framework preset (e.g. nextjs, vite)")
	cmd.Flags().String("build-command", "", "Build command override")
	cmd.Flags().String("output-directory", "", "Output directory override")
	cmd.Flags().String("install-command", "", "Install command override")
	cmd.Flags().String("root-directory", "", "Root directory override")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func makeRunProjectsUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")
		name, _ := cmd.Flags().GetString("name")
		framework, _ := cmd.Flags().GetString("framework")
		buildCmd, _ := cmd.Flags().GetString("build-command")
		outputDir, _ := cmd.Flags().GetString("output-directory")
		installCmd, _ := cmd.Flags().GetString("install-command")
		rootDir, _ := cmd.Flags().GetString("root-directory")

		body := map[string]any{}
		if name != "" {
			body["name"] = name
		}
		if framework != "" {
			body["framework"] = framework
		}
		if buildCmd != "" {
			body["buildCommand"] = buildCmd
		}
		if outputDir != "" {
			body["outputDirectory"] = outputDir
		}
		if installCmd != "" {
			body["installCommand"] = installCmd
		}
		if rootDir != "" {
			body["rootDirectory"] = rootDir
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update project %q", project), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPatch, fmt.Sprintf("/v9/projects/%s", project), body, &data); err != nil {
			return fmt.Errorf("updating project %q: %w", project, err)
		}

		detail := toProjectDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Updated project: %s (ID: %s)\n", detail.Name, detail.ID)
		return nil
	}
}

func newProjectsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a project (irreversible)",
		RunE:  makeRunProjectsDelete(factory),
	}
	cmd.Flags().String("project", "", "Project name or ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func makeRunProjectsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete project %q", project), map[string]any{
				"action":  "delete",
				"project": project,
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/v9/projects/%s", project), nil); err != nil {
			return fmt.Errorf("deleting project %q: %w", project, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "project": project})
		}
		fmt.Printf("Deleted project: %s\n", project)
		return nil
	}
}
