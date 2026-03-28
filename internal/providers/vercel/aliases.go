package vercel

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// AliasSummary is the JSON-serializable representation of a deployment alias.
type AliasSummary struct {
	UID          string `json:"uid,omitempty"`
	Alias        string `json:"alias"`
	DeploymentID string `json:"deploymentId,omitempty"`
	CreatedAt    int64  `json:"createdAt,omitempty"`
}

func toAliasSummary(data map[string]any) AliasSummary {
	return AliasSummary{
		UID:          jsonString(data["uid"]),
		Alias:        jsonString(data["alias"]),
		DeploymentID: jsonString(data["deploymentId"]),
		CreatedAt:    jsonInt64(data["createdAt"]),
	}
}

func printAliasSummaries(cmd *cobra.Command, aliases []AliasSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(aliases)
	}
	if len(aliases) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No aliases found.")
		return nil
	}
	lines := make([]string, 0, len(aliases)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %s", "UID", "ALIAS"))
	for _, a := range aliases {
		lines = append(lines, fmt.Sprintf("%-28s  %s", truncate(a.UID, 28), a.Alias))
	}
	cli.PrintText(lines)
	return nil
}

func newAliasesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List aliases for a deployment",
		RunE:  makeRunAliasesList(factory),
	}
	cmd.Flags().String("deployment-id", "", "Deployment ID (required)")
	_ = cmd.MarkFlagRequired("deployment-id")
	return cmd
}

func makeRunAliasesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		deploymentID, _ := cmd.Flags().GetString("deployment-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp struct {
			Aliases []map[string]any `json:"aliases"`
		}
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v2/deployments/%s/aliases", deploymentID), nil, &resp); err != nil {
			return fmt.Errorf("listing aliases for deployment %q: %w", deploymentID, err)
		}

		aliases := make([]AliasSummary, 0, len(resp.Aliases))
		for _, a := range resp.Aliases {
			aliases = append(aliases, toAliasSummary(a))
		}

		return printAliasSummaries(cmd, aliases)
	}
}

func newAliasesAssignCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign",
		Short: "Assign an alias to a deployment",
		RunE:  makeRunAliasesAssign(factory),
	}
	cmd.Flags().String("deployment-id", "", "Deployment ID (required)")
	cmd.Flags().String("alias", "", "Alias hostname to assign (required)")
	_ = cmd.MarkFlagRequired("deployment-id")
	_ = cmd.MarkFlagRequired("alias")
	return cmd
}

func makeRunAliasesAssign(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		deploymentID, _ := cmd.Flags().GetString("deployment-id")
		alias, _ := cmd.Flags().GetString("alias")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would assign alias %q to deployment %q", alias, deploymentID), map[string]any{
				"action":       "assign",
				"deploymentId": deploymentID,
				"alias":        alias,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{"alias": alias}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v2/deployments/%s/aliases", deploymentID), body, &data); err != nil {
			return fmt.Errorf("assigning alias %q to deployment %q: %w", alias, deploymentID, err)
		}

		a := toAliasSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(a)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Assigned alias: %s\n", a.Alias)
		return nil
	}
}
