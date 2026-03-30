package vercel

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// TeamSummary is the JSON-serializable summary of a Vercel team.
type TeamSummary struct {
	ID        string `json:"id"`
	Slug      string `json:"slug,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

// TeamMember is the JSON-serializable representation of a Vercel team member.
type TeamMember struct {
	UID      string `json:"uid"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
	JoinedAt int64  `json:"joinedAt,omitempty"`
}

func toTeamSummary(data map[string]any) TeamSummary {
	return TeamSummary{
		ID:        jsonString(data["id"]),
		Slug:      jsonString(data["slug"]),
		Name:      jsonString(data["name"]),
		CreatedAt: jsonInt64(data["createdAt"]),
	}
}

func toTeamMember(data map[string]any) TeamMember {
	return TeamMember{
		UID:      jsonString(data["uid"]),
		Username: jsonString(data["username"]),
		Email:    jsonString(data["email"]),
		Role:     jsonString(data["role"]),
		JoinedAt: jsonInt64(data["joinedAt"]),
	}
}

func printTeamSummaries(cmd *cobra.Command, teams []TeamSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(teams)
	}
	if len(teams) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No teams found.")
		return nil
	}
	lines := make([]string, 0, len(teams)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-20s  %s", "ID", "SLUG", "NAME"))
	for _, t := range teams {
		lines = append(lines, fmt.Sprintf("%-28s  %-20s  %s", truncate(t.ID, 28), truncate(t.Slug, 20), t.Name))
	}
	cli.PrintText(lines)
	return nil
}

func printTeamMembers(cmd *cobra.Command, members []TeamMember) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(members)
	}
	if len(members) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No members found.")
		return nil
	}
	lines := make([]string, 0, len(members)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-20s  %-30s  %s", "UID", "USERNAME", "EMAIL", "ROLE"))
	for _, m := range members {
		lines = append(lines, fmt.Sprintf("%-28s  %-20s  %-30s  %s",
			truncate(m.UID, 28), truncate(m.Username, 20), truncate(m.Email, 30), m.Role))
	}
	cli.PrintText(lines)
	return nil
}

func newTeamsListCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List teams",
		RunE:  makeRunTeamsList(factory),
	}
}

func makeRunTeamsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp struct {
			Teams []map[string]any `json:"teams"`
		}
		if err := client.doJSON(ctx, http.MethodGet, "/v2/teams", nil, &resp); err != nil {
			return fmt.Errorf("listing teams: %w", err)
		}

		teams := make([]TeamSummary, 0, len(resp.Teams))
		for _, t := range resp.Teams {
			teams = append(teams, toTeamSummary(t))
		}

		return printTeamSummaries(cmd, teams)
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
		teamID, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v2/teams/%s", teamID), nil, &data); err != nil {
			return fmt.Errorf("getting team %q: %w", teamID, err)
		}

		t := toTeamSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(t)
		}

		lines := []string{
			fmt.Sprintf("ID:    %s", t.ID),
			fmt.Sprintf("Slug:  %s", t.Slug),
			fmt.Sprintf("Name:  %s", t.Name),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newTeamsMembersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List members of a team",
		RunE:  makeRunTeamsMembers(factory),
	}
	cmd.Flags().String("id", "", "Team ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunTeamsMembers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		teamID, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp struct {
			Members []map[string]any `json:"members"`
		}
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v2/teams/%s/members", teamID), nil, &resp); err != nil {
			return fmt.Errorf("listing members for team %q: %w", teamID, err)
		}

		members := make([]TeamMember, 0, len(resp.Members))
		for _, m := range resp.Members {
			members = append(members, toTeamMember(m))
		}

		return printTeamMembers(cmd, members)
	}
}
