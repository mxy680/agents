package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newProjectsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"proj"},
		Short:   "Manage Supabase projects",
	}
	cmd.AddCommand(
		newProjectsListCmd(factory),
		newProjectsGetCmd(factory),
		newProjectsCreateCmd(factory),
		newProjectsUpdateCmd(factory),
		newProjectsDeleteCmd(factory),
		newProjectsPauseCmd(factory),
		newProjectsRestoreCmd(factory),
		newProjectsHealthCmd(factory),
		newProjectsRegionsCmd(factory),
	)
	return cmd
}

// --- Converters (snake_case API → camelCase structs) ---

func toProjectSummary(data map[string]any) ProjectSummary {
	s := func(key string) string {
		v, _ := data[key].(string)
		return v
	}
	return ProjectSummary{
		ID:             s("id"),
		Name:           s("name"),
		OrganizationID: s("organization_id"),
		Region:         s("region"),
		Status:         s("status"),
		CreatedAt:      s("created_at"),
	}
}

func toProjectDetail(data map[string]any) ProjectDetail {
	s := func(key string) string {
		v, _ := data[key].(string)
		return v
	}
	return ProjectDetail{
		ProjectSummary:  toProjectSummary(data),
		DatabaseHost:    s("db_host"),
		DatabaseVersion: s("db_version"),
	}
}

// --- Commands ---

func newProjectsListCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		RunE:  makeRunProjectsList(factory),
	}
}

func makeRunProjectsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, "/projects", nil)
		if err != nil {
			return fmt.Errorf("listing projects: %w", err)
		}

		var data []map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing projects response: %w", err)
		}

		summaries := make([]ProjectSummary, 0, len(data))
		for _, d := range data {
			summaries = append(summaries, toProjectSummary(d))
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
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunProjectsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, "/projects/"+ref, nil)
		if err != nil {
			return fmt.Errorf("getting project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing project response: %w", err)
		}

		detail := toProjectDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:               %s", detail.ID),
			fmt.Sprintf("Name:             %s", detail.Name),
			fmt.Sprintf("Organization ID:  %s", detail.OrganizationID),
			fmt.Sprintf("Region:           %s", detail.Region),
			fmt.Sprintf("Status:           %s", detail.Status),
			fmt.Sprintf("Created At:       %s", detail.CreatedAt),
			fmt.Sprintf("Database Host:    %s", detail.DatabaseHost),
			fmt.Sprintf("Database Version: %s", detail.DatabaseVersion),
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
	cmd.Flags().String("org-id", "", "Organization ID (required)")
	cmd.Flags().String("region", "", "Region (required)")
	cmd.Flags().String("plan", "free", "Billing plan (default: free)")
	cmd.Flags().String("db-pass", "", "Database password")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("org-id")
	_ = cmd.MarkFlagRequired("region")
	return cmd
}

func makeRunProjectsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		orgID, _ := cmd.Flags().GetString("org-id")
		region, _ := cmd.Flags().GetString("region")
		plan, _ := cmd.Flags().GetString("plan")
		dbPass, _ := cmd.Flags().GetString("db-pass")

		if dryRunResult(cmd, fmt.Sprintf("Would create project %q in org %s (%s)", name, orgID, region)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"name":            name,
			"organization_id": orgID,
			"region":          region,
			"plan":            plan,
		}
		if dbPass != "" {
			body["db_pass"] = dbPass
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}

		raw, err := doSupabase(client, http.MethodPost, "/projects", bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("creating project %q: %w", name, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing create response: %w", err)
		}

		detail := toProjectDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Created: %s (%s)\n", detail.Name, detail.ID)
		return nil
	}
}

func newProjectsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a project",
		RunE:  makeRunProjectsUpdate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("name", "", "New project name")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunProjectsUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		name, _ := cmd.Flags().GetString("name")

		if dryRunResult(cmd, fmt.Sprintf("Would update project %q", ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{}
		if name != "" {
			body["name"] = name
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}

		raw, err := doSupabase(client, http.MethodPatch, "/projects/"+ref, bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("updating project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing update response: %w", err)
		}

		detail := toProjectDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Updated: %s (%s)\n", detail.Name, detail.ID)
		return nil
	}
}

func newProjectsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a project (irreversible)",
		RunE:  makeRunProjectsDelete(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunProjectsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if dryRunResult(cmd, fmt.Sprintf("Would permanently delete project %q", ref)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("deleting project %q is irreversible", ref)); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := doSupabase(client, http.MethodDelete, "/projects/"+ref, nil); err != nil {
			return fmt.Errorf("deleting project %q: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "ref": ref})
		}
		fmt.Printf("Deleted: %s\n", ref)
		return nil
	}
}

func newProjectsPauseCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause",
		Short: "Pause a project",
		RunE:  makeRunProjectsPause(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunProjectsPause(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if dryRunResult(cmd, fmt.Sprintf("Would pause project %q", ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodPost, "/projects/"+ref+"/pause", nil)
		if err != nil {
			return fmt.Errorf("pausing project %q: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			var result any
			json.Unmarshal(raw, &result)
			return cli.PrintJSON(result)
		}
		fmt.Printf("Paused: %s\n", ref)
		return nil
	}
}

func newProjectsRestoreCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore a paused project",
		RunE:  makeRunProjectsRestore(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunProjectsRestore(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if dryRunResult(cmd, fmt.Sprintf("Would restore project %q", ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodPost, "/projects/"+ref+"/restore", nil)
		if err != nil {
			return fmt.Errorf("restoring project %q: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			var result any
			json.Unmarshal(raw, &result)
			return cli.PrintJSON(result)
		}
		fmt.Printf("Restored: %s\n", ref)
		return nil
	}
}

func newProjectsHealthCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Get project service health",
		RunE:  makeRunProjectsHealth(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunProjectsHealth(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, "/projects/"+ref+"/health?services=auth,rest,realtime,storage,db", nil)
		if err != nil {
			return fmt.Errorf("getting health for project %q: %w", ref, err)
		}

		var data []map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing health response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		if len(data) == 0 {
			fmt.Println("No health data available.")
			return nil
		}
		lines := make([]string, 0, len(data)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-15s  %s", "SERVICE", "STATUS", "ERROR"))
		for _, svc := range data {
			name, _ := svc["name"].(string)
			status, _ := svc["status"].(string)
			errMsg, _ := svc["error"].(string)
			lines = append(lines, fmt.Sprintf("%-20s  %-15s  %s", name, status, errMsg))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newProjectsRegionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regions",
		Short: "List available regions",
		RunE:  makeRunProjectsRegions(factory),
	}
	cmd.Flags().String("org-slug", "", "Organization slug (required)")
	_ = cmd.MarkFlagRequired("org-slug")
	return cmd
}

func makeRunProjectsRegions(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		orgSlug, _ := cmd.Flags().GetString("org-slug")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, "/projects/available-regions?organization_slug="+orgSlug, nil)
		if err != nil {
			return fmt.Errorf("listing available regions: %w", err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing regions response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		allGroup, _ := data["all"].(map[string]any)
		specific, _ := allGroup["specific"].([]any)
		if len(specific) == 0 {
			fmt.Println("No regions available.")
			return nil
		}
		lines := make([]string, 0, len(specific)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-40s  %s", "CODE", "NAME", "PROVIDER"))
		for _, r := range specific {
			region, ok := r.(map[string]any)
			if !ok {
				continue
			}
			code, _ := region["code"].(string)
			name, _ := region["name"].(string)
			provider, _ := region["provider"].(string)
			lines = append(lines, fmt.Sprintf("%-20s  %-40s  %s", code, name, provider))
		}
		cli.PrintText(lines)
		return nil
	}
}
