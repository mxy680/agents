package linear

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newWorkflowsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflow states for a team",
		RunE:  makeRunWorkflowsList(factory),
	}
	cmd.Flags().String("team", "", "Team ID (required)")
	_ = cmd.MarkFlagRequired("team")
	return cmd
}

func makeRunWorkflowsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		teamID, _ := cmd.Flags().GetString("team")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($teamId: String!) {
  team(id: $teamId) {
    states {
      nodes {
        id
        name
        color
        type
        position
      }
    }
  }
}`

		var resp struct {
			Team struct {
				States struct {
					Nodes []struct {
						ID       string  `json:"id"`
						Name     string  `json:"name"`
						Color    string  `json:"color"`
						Type     string  `json:"type"`
						Position float64 `json:"position"`
					} `json:"nodes"`
				} `json:"states"`
			} `json:"team"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"teamId": teamID}, &resp); err != nil {
			return fmt.Errorf("listing workflow states: %w", err)
		}

		states := make([]WorkflowState, 0, len(resp.Team.States.Nodes))
		for _, n := range resp.Team.States.Nodes {
			states = append(states, WorkflowState{
				ID:       n.ID,
				Name:     n.Name,
				Color:    n.Color,
				Type:     n.Type,
				Position: n.Position,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(states)
		}
		if len(states) == 0 {
			fmt.Println("No workflow states found.")
			return nil
		}
		lines := make([]string, 0, len(states)+1)
		lines = append(lines, fmt.Sprintf("%-28s  %-25s  %-12s  %-8s  %s", "ID", "NAME", "TYPE", "COLOR", "POSITION"))
		for _, s := range states {
			lines = append(lines, fmt.Sprintf("%-28s  %-25s  %-12s  %-8s  %.0f",
				truncate(s.ID, 28), truncate(s.Name, 25), truncate(s.Type, 12), s.Color, s.Position))
		}
		cli.PrintText(lines)
		return nil
	}
}
