package fly

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newAppsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List apps",
		RunE:  makeRunAppsList(factory),
	}
	cmd.Flags().String("org", "", "Filter by organization slug")
	return cmd
}

func makeRunAppsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		org, _ := cmd.Flags().GetString("org")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/v1/apps"
		if org != "" {
			path += "?org_slug=" + org
		}

		var resp struct {
			Apps []struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
				OrgID  string `json:"org_slug"`
			} `json:"apps"`
		}
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing apps: %w", err)
		}

		summaries := make([]AppSummary, 0, len(resp.Apps))
		for _, a := range resp.Apps {
			summaries = append(summaries, AppSummary{
				ID:     a.ID,
				Name:   a.Name,
				Status: a.Status,
				OrgID:  a.OrgID,
			})
		}

		return printAppSummaries(cmd, summaries)
	}
}

func newAppsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get app details",
		RunE:  makeRunAppsGet(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	_ = cmd.MarkFlagRequired("app")
	return cmd
}

func makeRunAppsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data AppDetail
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/apps/%s", app), nil, &data); err != nil {
			return fmt.Errorf("getting app %q: %w", app, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		lines := []string{
			fmt.Sprintf("ID:       %s", data.ID),
			fmt.Sprintf("Name:     %s", data.Name),
			fmt.Sprintf("Status:   %s", data.Status),
			fmt.Sprintf("Org:      %s", data.OrgID),
			fmt.Sprintf("Hostname: %s", data.Hostname),
			fmt.Sprintf("App URL:  %s", data.AppURL),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newAppsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new app",
		RunE:  makeRunAppsCreate(factory),
	}
	cmd.Flags().String("name", "", "App name (required)")
	cmd.Flags().String("org", "", "Organization slug (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("org")
	return cmd
}

func makeRunAppsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		org, _ := cmd.Flags().GetString("org")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create app %q in org %q", name, org), map[string]any{
				"action":   "create",
				"app_name": name,
				"org_slug": org,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"app_name": name,
			"org_slug": org,
		}

		var data AppDetail
		if err := client.doJSON(ctx, http.MethodPost, "/v1/apps", body, &data); err != nil {
			return fmt.Errorf("creating app %q: %w", name, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		fmt.Printf("Created app: %s (ID: %s)\n", data.Name, data.ID)
		return nil
	}
}

func newAppsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an app (irreversible)",
		RunE:  makeRunAppsDelete(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	cmd.Flags().Bool("force", false, "Force deletion even if the app has running machines")
	_ = cmd.MarkFlagRequired("app")
	return cmd
}

func makeRunAppsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		force, _ := cmd.Flags().GetBool("force")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete app %q", app), map[string]any{
				"action": "delete",
				"app":    app,
				"force":  force,
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

		path := fmt.Sprintf("/v1/apps/%s", app)
		if force {
			path += "?force=true"
		}

		if _, err := client.do(ctx, http.MethodDelete, path, nil); err != nil {
			return fmt.Errorf("deleting app %q: %w", app, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "app": app})
		}
		fmt.Printf("Deleted app: %s\n", app)
		return nil
	}
}
