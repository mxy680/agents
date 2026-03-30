package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newPagesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Pages projects",
		RunE:  makeRunPagesList(factory),
	}
	return cmd
}

func makeRunPagesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath("/pages/projects")
		if err != nil {
			return err
		}

		var resp []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing Pages projects: %w", err)
		}

		projects := make([]PagesSummary, 0, len(resp))
		for _, p := range resp {
			projects = append(projects, toPagesSummary(p))
		}

		return printPagesProjects(cmd, projects)
	}
}

func newPagesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Pages project",
		RunE:  makeRunPagesGet(factory),
	}
	cmd.Flags().String("project", "", "Pages project name (required)")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func makeRunPagesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath(fmt.Sprintf("/pages/projects/%s", project))
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("getting Pages project %q: %w", project, err)
		}

		p := toPagesSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(p)
		}

		lines := []string{
			fmt.Sprintf("ID:                %s", p.ID),
			fmt.Sprintf("Name:              %s", p.Name),
			fmt.Sprintf("Subdomain:         %s", p.SubDomain),
			fmt.Sprintf("Production Branch: %s", p.ProductionBranch),
			fmt.Sprintf("Created:           %s", p.CreatedOn),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newPagesDeploymentsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployments for a Pages project",
		RunE:  makeRunPagesDeploymentsList(factory),
	}
	cmd.Flags().String("project", "", "Pages project name (required)")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func makeRunPagesDeploymentsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath(fmt.Sprintf("/pages/projects/%s/deployments", project))
		if err != nil {
			return err
		}

		var resp []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing deployments for Pages project %q: %w", project, err)
		}

		deployments := make([]PagesDeploymentSummary, 0, len(resp))
		for _, d := range resp {
			deployments = append(deployments, toPagesDeploymentSummary(d))
		}

		return printPagesDeployments(cmd, deployments)
	}
}

func newPagesDeploymentsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Pages deployment",
		RunE:  makeRunPagesDeploymentsGet(factory),
	}
	cmd.Flags().String("project", "", "Pages project name (required)")
	cmd.Flags().String("deployment", "", "Deployment ID (required)")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("deployment")
	return cmd
}

func makeRunPagesDeploymentsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")
		deploymentID, _ := cmd.Flags().GetString("deployment")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath(fmt.Sprintf("/pages/projects/%s/deployments/%s", project, deploymentID))
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("getting deployment %q: %w", deploymentID, err)
		}

		d := toPagesDeploymentSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(d)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", d.ID),
			fmt.Sprintf("URL:         %s", d.URL),
			fmt.Sprintf("Environment: %s", d.Environment),
			fmt.Sprintf("Stage:       %s", d.Stage),
			fmt.Sprintf("Created:     %s", d.CreatedOn),
		}
		cli.PrintText(lines)
		return nil
	}
}
