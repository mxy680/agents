package linear

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newTeamsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all teams",
		RunE:  makeRunTeamsList(factory),
	}
	return cmd
}

func makeRunTeamsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query {
  teams {
    nodes {
      id
      name
      key
    }
  }
}`

		var resp struct {
			Teams struct {
				Nodes []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Key  string `json:"key"`
				} `json:"nodes"`
			} `json:"teams"`
		}

		if err := client.graphQL(ctx, q, nil, &resp); err != nil {
			return fmt.Errorf("listing teams: %w", err)
		}

		summaries := make([]TeamSummary, 0, len(resp.Teams.Nodes))
		for _, n := range resp.Teams.Nodes {
			summaries = append(summaries, TeamSummary{
				ID:   n.ID,
				Name: n.Name,
				Key:  n.Key,
			})
		}

		return printTeamSummaries(cmd, summaries)
	}
}

func newTeamsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get team details",
		RunE:  makeRunTeamsGet(factory),
	}
	cmd.Flags().String("id", "", "Team ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunTeamsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($id: String!) {
  team(id: $id) {
    id
    name
    key
    description
    members {
      nodes {
        id
        name
        email
      }
    }
  }
}`

		var resp struct {
			Team struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Key         string `json:"key"`
				Description string `json:"description"`
				Members     struct {
					Nodes []struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						Email string `json:"email"`
					} `json:"nodes"`
				} `json:"members"`
			} `json:"team"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("getting team %q: %w", id, err)
		}

		n := resp.Team
		detail := TeamDetail{
			ID:          n.ID,
			Name:        n.Name,
			Key:         n.Key,
			Description: n.Description,
		}
		for _, m := range n.Members.Nodes {
			detail.Members = append(detail.Members, UserSummary{
				ID:    m.ID,
				Name:  m.Name,
				Email: m.Email,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", detail.ID),
			fmt.Sprintf("Name:        %s", detail.Name),
			fmt.Sprintf("Key:         %s", detail.Key),
			fmt.Sprintf("Description: %s", detail.Description),
			fmt.Sprintf("Members:     %d", len(detail.Members)),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newTeamsMembersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List team members",
		RunE:  makeRunTeamsMembers(factory),
	}
	cmd.Flags().String("id", "", "Team ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunTeamsMembers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($id: String!) {
  team(id: $id) {
    members {
      nodes {
        id
        name
        email
      }
    }
  }
}`

		var resp struct {
			Team struct {
				Members struct {
					Nodes []struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						Email string `json:"email"`
					} `json:"nodes"`
				} `json:"members"`
			} `json:"team"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("listing team members: %w", err)
		}

		members := make([]UserSummary, 0, len(resp.Team.Members.Nodes))
		for _, m := range resp.Team.Members.Nodes {
			members = append(members, UserSummary{
				ID:    m.ID,
				Name:  m.Name,
				Email: m.Email,
			})
		}

		return printUserSummaries(cmd, members)
	}
}
