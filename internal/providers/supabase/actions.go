package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newActionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "actions",
		Aliases: []string{"action"},
		Short:   "CI/CD action runs",
	}
	cmd.AddCommand(
		newActionsListCmd(factory),
		newActionsGetCmd(factory),
		newActionsLogsCmd(factory),
		newActionsUpdateStatusCmd(factory),
	)
	return cmd
}

// ActionSummary is a lightweight representation of a CI/CD action run.
type ActionSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Type   string `json:"type,omitempty"`
}

// toActionSummary converts a raw API response map to an ActionSummary.
func toActionSummary(data map[string]any) ActionSummary {
	id, _ := data["id"].(string)
	status, _ := data["status"].(string)
	actionType, _ := data["type"].(string)
	return ActionSummary{ID: id, Status: status, Type: actionType}
}

// printActionSummaries outputs action summaries as JSON or a formatted text table.
func printActionSummaries(cmd *cobra.Command, actions []ActionSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(actions)
	}
	if len(actions) == 0 {
		fmt.Println("No action runs found.")
		return nil
	}
	lines := make([]string, 0, len(actions)+1)
	lines = append(lines, fmt.Sprintf("%-36s  %-20s  %s", "ID", "STATUS", "TYPE"))
	for _, a := range actions {
		lines = append(lines, fmt.Sprintf("%-36s  %-20s  %s",
			truncate(a.ID, 36), truncate(a.Status, 20), a.Type))
	}
	cli.PrintText(lines)
	return nil
}

// --- list ---

func newActionsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List CI/CD action runs for a project",
		RunE:  makeRunActionsList(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunActionsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/actions", ref), nil)
		if err != nil {
			return fmt.Errorf("listing actions: %w", err)
		}

		var raw []map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		actions := make([]ActionSummary, 0, len(raw))
		for _, r := range raw {
			actions = append(actions, toActionSummary(r))
		}
		return printActionSummaries(cmd, actions)
	}
}

// --- get ---

func newActionsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a specific CI/CD action run",
		RunE:  makeRunActionsGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("run-id", "", "Action run ID (required)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("run-id")
	return cmd
}

func makeRunActionsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		runID, _ := cmd.Flags().GetString("run-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet,
			fmt.Sprintf("/projects/%s/actions/%s", ref, runID), nil)
		if err != nil {
			return fmt.Errorf("getting action run %s: %w", runID, err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		a := toActionSummary(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(a)
		}
		lines := []string{
			fmt.Sprintf("ID:     %s", a.ID),
			fmt.Sprintf("Status: %s", a.Status),
			fmt.Sprintf("Type:   %s", a.Type),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- logs ---

func newActionsLogsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Get logs for a CI/CD action run",
		RunE:  makeRunActionsLogs(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("run-id", "", "Action run ID (required)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("run-id")
	return cmd
}

func makeRunActionsLogs(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		runID, _ := cmd.Flags().GetString("run-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet,
			fmt.Sprintf("/projects/%s/actions/%s/logs", ref, runID), nil)
		if err != nil {
			return fmt.Errorf("getting action logs for %s: %w", runID, err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, data, "", "  "); err != nil {
			fmt.Println(string(data))
			return nil
		}
		fmt.Println(pretty.String())
		return nil
	}
}

// --- update-status ---

func newActionsUpdateStatusCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-status",
		Short: "Update the status of a CI/CD action run",
		RunE:  makeRunActionsUpdateStatus(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("run-id", "", "Action run ID (required)")
	cmd.Flags().String("status", "", "New status (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without executing")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("run-id")
	_ = cmd.MarkFlagRequired("status")
	return cmd
}

func makeRunActionsUpdateStatus(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		runID, _ := cmd.Flags().GetString("run-id")
		status, _ := cmd.Flags().GetString("status")

		if dryRunResult(cmd, fmt.Sprintf("Would update action run %s status to %q in project %s", runID, status, ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{"status": status}
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}

		data, err := doSupabase(client, http.MethodPatch,
			fmt.Sprintf("/projects/%s/actions/%s/status", ref, runID), bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("updating action run status: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			return cli.PrintJSON(raw)
		}
		fmt.Printf("Updated action run %s status to: %s\n", runID, status)
		return nil
	}
}
