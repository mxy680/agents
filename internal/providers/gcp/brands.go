package gcp

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newBrandsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List IAP brands (OAuth consent screens) for a project",
		RunE:  makeRunBrandsList(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	return cmd
}

func makeRunBrandsList(factory ClientFactory) func(*cobra.Command, []string) error {
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

		url := fmt.Sprintf("%s/projects/%s/brands", client.iapURL, project)
		var resp struct {
			Brands []map[string]any `json:"brands"`
		}
		if err := client.doJSON(ctx, http.MethodGet, url, nil, &resp); err != nil {
			return fmt.Errorf("listing brands: %w", err)
		}

		summaries := make([]BrandSummary, 0, len(resp.Brands))
		for _, b := range resp.Brands {
			summaries = append(summaries, toBrandSummary(b))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}
		if len(summaries) == 0 {
			fmt.Println("No brands found.")
			return nil
		}
		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-50s  %-30s  %s", "NAME", "TITLE", "SUPPORT EMAIL"))
		for _, b := range summaries {
			lines = append(lines, fmt.Sprintf("%-50s  %-30s  %s",
				truncate(b.Name, 50), truncate(b.ApplicationTitle, 30), b.SupportEmail))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newBrandsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an IAP brand (OAuth consent screen) for a project",
		RunE:  makeRunBrandsCreate(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("title", "", "Application title shown on the consent screen (required)")
	cmd.Flags().String("support-email", "", "Support email shown on the consent screen (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("support-email")
	return cmd
}

func makeRunBrandsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		title, _ := cmd.Flags().GetString("title")
		supportEmail, _ := cmd.Flags().GetString("support-email")

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
			return dryRunResult(cmd, fmt.Sprintf("Would create IAP brand %q on project %q", title, project), map[string]any{
				"action":           "create",
				"project":          project,
				"applicationTitle": title,
				"supportEmail":     supportEmail,
			})
		}

		body := map[string]any{
			"applicationTitle": title,
			"supportEmail":     supportEmail,
		}

		url := fmt.Sprintf("%s/projects/%s/brands", client.iapURL, project)
		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, url, body, &data); err != nil {
			return fmt.Errorf("creating brand: %w", err)
		}

		brand := toBrandSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(brand)
		}
		fmt.Printf("Created brand: %s\n", brand.Name)
		return nil
	}
}

func newBrandsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an IAP brand by ID",
		RunE:  makeRunBrandsGet(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("brand", "", "Brand ID (required, numeric, e.g. 123456789)")
	_ = cmd.MarkFlagRequired("brand")
	return cmd
}

func makeRunBrandsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		brandID, _ := cmd.Flags().GetString("brand")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/projects/%s/brands/%s", client.iapURL, project, brandID)
		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, url, nil, &data); err != nil {
			return fmt.Errorf("getting brand %q: %w", brandID, err)
		}

		brand := toBrandSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(brand)
		}
		lines := []string{
			fmt.Sprintf("Name:             %s", brand.Name),
			fmt.Sprintf("Title:            %s", brand.ApplicationTitle),
			fmt.Sprintf("Support Email:    %s", brand.SupportEmail),
			fmt.Sprintf("Internal Only:    %v", brand.OrgInternalOnly),
		}
		cli.PrintText(lines)
		return nil
	}
}
