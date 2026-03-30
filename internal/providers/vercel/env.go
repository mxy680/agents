package vercel

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// EnvSummary is the JSON-serializable summary of a Vercel environment variable.
type EnvSummary struct {
	ID        string   `json:"id"`
	Key       string   `json:"key"`
	Value     string   `json:"value,omitempty"`
	Type      string   `json:"type,omitempty"`
	Target    []string `json:"target,omitempty"`
	CreatedAt int64    `json:"createdAt,omitempty"`
	UpdatedAt int64    `json:"updatedAt,omitempty"`
}

func toEnvSummary(data map[string]any) EnvSummary {
	return EnvSummary{
		ID:        jsonString(data["id"]),
		Key:       jsonString(data["key"]),
		Value:     jsonString(data["value"]),
		Type:      jsonString(data["type"]),
		Target:    jsonStringSlice(data["target"]),
		CreatedAt: jsonInt64(data["createdAt"]),
		UpdatedAt: jsonInt64(data["updatedAt"]),
	}
}

func printEnvSummaries(cmd *cobra.Command, envs []EnvSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(envs)
	}
	if len(envs) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No environment variables found.")
		return nil
	}
	lines := make([]string, 0, len(envs)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-40s  %-10s  %s", "ID", "KEY", "TYPE", "TARGET"))
	for _, e := range envs {
		lines = append(lines, fmt.Sprintf("%-28s  %-40s  %-10s  %s",
			truncate(e.ID, 28), truncate(e.Key, 40), e.Type, strings.Join(e.Target, ",")))
	}
	cli.PrintText(lines)
	return nil
}

func newEnvListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List environment variables for a project",
		RunE:  makeRunEnvList(factory),
	}
	cmd.Flags().String("project", "", "Project name or ID (required)")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func makeRunEnvList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp struct {
			Envs []map[string]any `json:"envs"`
		}
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v10/projects/%s/env", project), nil, &resp); err != nil {
			return fmt.Errorf("listing env vars for project %q: %w", project, err)
		}

		summaries := make([]EnvSummary, 0, len(resp.Envs))
		for _, e := range resp.Envs {
			summaries = append(summaries, toEnvSummary(e))
		}

		return printEnvSummaries(cmd, summaries)
	}
}

func newEnvGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a specific environment variable",
		RunE:  makeRunEnvGet(factory),
	}
	cmd.Flags().String("project", "", "Project name or ID (required)")
	cmd.Flags().String("key", "", "Environment variable ID (required)")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func makeRunEnvGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")
		envID, _ := cmd.Flags().GetString("key")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v10/projects/%s/env/%s", project, envID), nil, &data); err != nil {
			return fmt.Errorf("getting env var %q for project %q: %w", envID, project, err)
		}

		e := toEnvSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(e)
		}

		lines := []string{
			fmt.Sprintf("ID:      %s", e.ID),
			fmt.Sprintf("Key:     %s", e.Key),
			fmt.Sprintf("Value:   %s", e.Value),
			fmt.Sprintf("Type:    %s", e.Type),
			fmt.Sprintf("Target:  %s", strings.Join(e.Target, ", ")),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newEnvSetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Create or upsert an environment variable",
		RunE:  makeRunEnvSet(factory),
	}
	cmd.Flags().String("project", "", "Project name or ID (required)")
	cmd.Flags().String("key", "", "Variable key (required)")
	cmd.Flags().String("value", "", "Variable value (required)")
	cmd.Flags().StringSlice("target", []string{"production", "preview", "development"}, "Deployment targets (comma-separated)")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("key")
	_ = cmd.MarkFlagRequired("value")
	return cmd
}

func makeRunEnvSet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		target, _ := cmd.Flags().GetStringSlice("target")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would set env var %q on project %q", key, project), map[string]any{
				"action":  "set",
				"project": project,
				"key":     key,
				"target":  target,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"key":    key,
			"value":  value,
			"type":   "plain",
			"target": target,
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v10/projects/%s/env", project), body, &data); err != nil {
			return fmt.Errorf("setting env var %q on project %q: %w", key, project, err)
		}

		e := toEnvSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(e)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Set env var: %s (ID: %s)\n", e.Key, e.ID)
		return nil
	}
}

func newEnvRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Delete an environment variable",
		RunE:  makeRunEnvRemove(factory),
	}
	cmd.Flags().String("project", "", "Project name or ID (required)")
	cmd.Flags().String("key", "", "Environment variable ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func makeRunEnvRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		project, _ := cmd.Flags().GetString("project")
		envID, _ := cmd.Flags().GetString("key")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete env var %q from project %q", envID, project), map[string]any{
				"action":  "remove",
				"project": project,
				"envId":   envID,
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/v10/projects/%s/env/%s", project, envID), nil); err != nil {
			return fmt.Errorf("deleting env var %q from project %q: %w", envID, project, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "envId": envID})
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted env var: %s\n", envID)
		return nil
	}
}
