package linear

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newCyclesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cycles for a team",
		RunE:  makeRunCyclesList(factory),
	}
	cmd.Flags().String("team", "", "Team ID (required)")
	_ = cmd.MarkFlagRequired("team")
	return cmd
}

func makeRunCyclesList(factory ClientFactory) func(*cobra.Command, []string) error {
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
    cycles {
      nodes {
        id
        number
        startsAt
        endsAt
      }
    }
  }
}`

		var resp struct {
			Team struct {
				Cycles struct {
					Nodes []struct {
						ID       string `json:"id"`
						Number   int    `json:"number"`
						StartsAt string `json:"startsAt"`
						EndsAt   string `json:"endsAt"`
					} `json:"nodes"`
				} `json:"cycles"`
			} `json:"team"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"teamId": teamID}, &resp); err != nil {
			return fmt.Errorf("listing cycles: %w", err)
		}

		summaries := make([]CycleSummary, 0, len(resp.Team.Cycles.Nodes))
		for _, n := range resp.Team.Cycles.Nodes {
			summaries = append(summaries, CycleSummary{
				ID:       n.ID,
				Number:   n.Number,
				StartsAt: n.StartsAt,
				EndsAt:   n.EndsAt,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}
		if len(summaries) == 0 {
			fmt.Println("No cycles found.")
			return nil
		}
		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-28s  %-8s  %-24s  %s", "ID", "NUMBER", "STARTS AT", "ENDS AT"))
		for _, c := range summaries {
			lines = append(lines, fmt.Sprintf("%-28s  %-8d  %-24s  %s", truncate(c.ID, 28), c.Number, c.StartsAt, c.EndsAt))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newCyclesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get cycle details",
		RunE:  makeRunCyclesGet(factory),
	}
	cmd.Flags().String("id", "", "Cycle ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunCyclesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($id: String!) {
  cycle(id: $id) {
    id
    number
    startsAt
    endsAt
    issues {
      nodes {
        id
        title
      }
    }
  }
}`

		var resp struct {
			Cycle struct {
				ID       string `json:"id"`
				Number   int    `json:"number"`
				StartsAt string `json:"startsAt"`
				EndsAt   string `json:"endsAt"`
				Issues   struct {
					Nodes []struct {
						ID    string `json:"id"`
						Title string `json:"title"`
					} `json:"nodes"`
				} `json:"issues"`
			} `json:"cycle"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("getting cycle %q: %w", id, err)
		}

		n := resp.Cycle
		detail := CycleDetail{
			ID:       n.ID,
			Number:   n.Number,
			StartsAt: n.StartsAt,
			EndsAt:   n.EndsAt,
		}
		for _, i := range n.Issues.Nodes {
			detail.Issues = append(detail.Issues, IssueSummary{ID: i.ID, Title: i.Title})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:       %s", detail.ID),
			fmt.Sprintf("Number:   %d", detail.Number),
			fmt.Sprintf("Starts:   %s", detail.StartsAt),
			fmt.Sprintf("Ends:     %s", detail.EndsAt),
			fmt.Sprintf("Issues:   %d", len(detail.Issues)),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newCyclesCurrentCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Get the active cycle for a team",
		RunE:  makeRunCyclesCurrent(factory),
	}
	cmd.Flags().String("team", "", "Team ID (required)")
	_ = cmd.MarkFlagRequired("team")
	return cmd
}

func makeRunCyclesCurrent(factory ClientFactory) func(*cobra.Command, []string) error {
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
    activeCycle {
      id
      number
      startsAt
      endsAt
    }
  }
}`

		var resp struct {
			Team struct {
				ActiveCycle *struct {
					ID       string `json:"id"`
					Number   int    `json:"number"`
					StartsAt string `json:"startsAt"`
					EndsAt   string `json:"endsAt"`
				} `json:"activeCycle"`
			} `json:"team"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"teamId": teamID}, &resp); err != nil {
			return fmt.Errorf("getting active cycle: %w", err)
		}

		if resp.Team.ActiveCycle == nil {
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(nil)
			}
			fmt.Println("No active cycle.")
			return nil
		}

		c := resp.Team.ActiveCycle
		summary := CycleSummary{
			ID:       c.ID,
			Number:   c.Number,
			StartsAt: c.StartsAt,
			EndsAt:   c.EndsAt,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summary)
		}

		lines := []string{
			fmt.Sprintf("ID:     %s", summary.ID),
			fmt.Sprintf("Number: %d", summary.Number),
			fmt.Sprintf("Starts: %s", summary.StartsAt),
			fmt.Sprintf("Ends:   %s", summary.EndsAt),
		}
		cli.PrintText(lines)
		return nil
	}
}
