package linear

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newLabelsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issue labels",
		RunE:  makeRunLabelsList(factory),
	}
	cmd.Flags().Int("limit", 50, "Maximum number of labels to return")
	return cmd
}

func makeRunLabelsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($first: Int!) {
  issueLabels(first: $first) {
    nodes {
      id
      name
      color
    }
  }
}`

		var resp struct {
			IssueLabels struct {
				Nodes []struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Color string `json:"color"`
				} `json:"nodes"`
			} `json:"issueLabels"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"first": limit}, &resp); err != nil {
			return fmt.Errorf("listing labels: %w", err)
		}

		labels := make([]LabelSummary, 0, len(resp.IssueLabels.Nodes))
		for _, n := range resp.IssueLabels.Nodes {
			labels = append(labels, LabelSummary{
				ID:    n.ID,
				Name:  n.Name,
				Color: n.Color,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(labels)
		}
		if len(labels) == 0 {
			fmt.Println("No labels found.")
			return nil
		}
		lines := make([]string, 0, len(labels)+1)
		lines = append(lines, fmt.Sprintf("%-28s  %-30s  %s", "ID", "NAME", "COLOR"))
		for _, l := range labels {
			lines = append(lines, fmt.Sprintf("%-28s  %-30s  %s", truncate(l.ID, 28), truncate(l.Name, 30), l.Color))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newLabelsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an issue label",
		RunE:  makeRunLabelsCreate(factory),
	}
	cmd.Flags().String("name", "", "Label name (required)")
	cmd.Flags().String("color", "", "Label color as hex (e.g. #ff0000, required)")
	cmd.Flags().String("team", "", "Team ID to scope the label (optional)")
	cmd.Flags().Bool("dry-run", false, "Print what would be created without making changes")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("color")
	return cmd
}

func makeRunLabelsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		color, _ := cmd.Flags().GetString("color")
		teamID, _ := cmd.Flags().GetString("team")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create label %q (%s)", name, color), map[string]any{
				"action": "create",
				"name":   name,
				"color":  color,
				"teamId": teamID,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
mutation($input: IssueLabelCreateInput!) {
  issueLabelCreate(input: $input) {
    issueLabel {
      id
      name
      color
    }
  }
}`

		input := map[string]any{
			"name":  name,
			"color": color,
		}
		if teamID != "" {
			input["teamId"] = teamID
		}

		var resp struct {
			IssueLabelCreate struct {
				IssueLabel struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Color string `json:"color"`
				} `json:"issueLabel"`
			} `json:"issueLabelCreate"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"input": input}, &resp); err != nil {
			return fmt.Errorf("creating label: %w", err)
		}

		label := resp.IssueLabelCreate.IssueLabel
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(label)
		}
		fmt.Printf("Created label: %s (ID: %s, Color: %s)\n", label.Name, label.ID, label.Color)
		return nil
	}
}

func newLabelsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an issue label (irreversible)",
		RunE:  makeRunLabelsDelete(factory),
	}
	cmd.Flags().String("id", "", "Label ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without making changes")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunLabelsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete label %q", id), map[string]any{
				"action": "delete",
				"id":     id,
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

		const q = `
mutation($id: String!) {
  issueLabelDelete(id: $id) {
    success
  }
}`

		var resp struct {
			IssueLabelDelete struct {
				Success bool `json:"success"`
			} `json:"issueLabelDelete"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("deleting label %q: %w", id, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"success": resp.IssueLabelDelete.Success, "id": id})
		}
		fmt.Printf("Deleted label: %s\n", id)
		return nil
	}
}
