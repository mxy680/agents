package gcp

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newServicesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List enabled GCP APIs for a project",
		RunE:  makeRunServicesList(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	return cmd
}

func makeRunServicesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/projects/%s/services?filter=state:ENABLED", client.serviceUsageURL, project)
		var resp struct {
			Services []map[string]any `json:"services"`
		}
		if err := client.doJSON(ctx, http.MethodGet, url, nil, &resp); err != nil {
			return fmt.Errorf("listing services: %w", err)
		}

		summaries := make([]ServiceSummary, 0, len(resp.Services))
		for _, s := range resp.Services {
			summaries = append(summaries, toServiceSummary(s))
		}

		return printServiceSummaries(cmd, summaries)
	}
}

func newServicesEnableCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable a GCP API for a project",
		RunE:  makeRunServicesEnable(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("service", "", "Service name to enable (e.g. iap.googleapis.com)")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("service")
	return cmd
}

func makeRunServicesEnable(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		service, _ := cmd.Flags().GetString("service")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would enable service %q on project %q", service, project), map[string]any{
				"action":  "enable",
				"project": project,
				"service": service,
			})
		}

		url := fmt.Sprintf("%s/projects/%s/services/%s:enable", client.serviceUsageURL, project, service)
		var op Operation
		if err := client.doJSON(ctx, http.MethodPost, url, map[string]any{}, &op); err != nil {
			return fmt.Errorf("enabling service %q: %w", service, err)
		}

		if _, err := client.waitForOperation(ctx, &op); err != nil {
			return fmt.Errorf("service enable operation failed: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "enabled", "service": service, "project": project})
		}
		fmt.Printf("Enabled %s on project %s\n", service, project)
		return nil
	}
}

func newServicesDisableCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable a GCP API for a project (irreversible if dependents exist)",
		RunE:  makeRunServicesDisable(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("service", "", "Service name to disable (e.g. iap.googleapis.com)")
	cmd.Flags().Bool("confirm", false, "Confirm disabling the service")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("service")
	return cmd
}

func makeRunServicesDisable(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		service, _ := cmd.Flags().GetString("service")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would disable service %q on project %q", service, project), map[string]any{
				"action":  "disable",
				"project": project,
				"service": service,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		url := fmt.Sprintf("%s/projects/%s/services/%s:disable", client.serviceUsageURL, project, service)
		var op Operation
		if err := client.doJSON(ctx, http.MethodPost, url, map[string]any{}, &op); err != nil {
			return fmt.Errorf("disabling service %q: %w", service, err)
		}

		if _, err := client.waitForOperation(ctx, &op); err != nil {
			return fmt.Errorf("service disable operation failed: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "disabled", "service": service, "project": project})
		}
		fmt.Printf("Disabled %s on project %s\n", service, project)
		return nil
	}
}
