package github

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- list ---

func newTeamsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List teams in an organization",
		RunE:  makeRunTeamsList(factory),
	}
	cmd.Flags().String("org", "", "Organization name (required)")
	cmd.Flags().Int("limit", 20, "Maximum number of teams to return")
	_ = cmd.MarkFlagRequired("org")
	return cmd
}

func makeRunTeamsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		org, _ := cmd.Flags().GetString("org")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/orgs/%s/teams?per_page=%d", org, limit)

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing teams for org %s: %w", org, err)
		}

		summaries := make([]TeamSummary, 0, len(data))
		for _, d := range data {
			summaries = append(summaries, toTeamSummary(d))
		}
		return printTeamSummaries(cmd, summaries)
	}
}

// --- get ---

func newTeamsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get team details",
		RunE:  makeRunTeamsGet(factory),
	}
	cmd.Flags().String("org", "", "Organization name (required)")
	cmd.Flags().String("team-slug", "", "Team slug (required)")
	_ = cmd.MarkFlagRequired("org")
	_ = cmd.MarkFlagRequired("team-slug")
	return cmd
}

func makeRunTeamsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		org, _ := cmd.Flags().GetString("org")
		teamSlug, _ := cmd.Flags().GetString("team-slug")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, fmt.Sprintf("/orgs/%s/teams/%s", org, teamSlug), nil, &data); err != nil {
			return fmt.Errorf("getting team %s/%s: %w", org, teamSlug, err)
		}

		team := toTeamSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(team)
		}

		lines := []string{
			fmt.Sprintf("ID:          %d", team.ID),
			fmt.Sprintf("Name:        %s", team.Name),
			fmt.Sprintf("Slug:        %s", team.Slug),
			fmt.Sprintf("Description: %s", team.Description),
			fmt.Sprintf("Permission:  %s", team.Permission),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- members ---

func newTeamsMembersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List members of a team",
		RunE:  makeRunTeamsMembers(factory),
	}
	cmd.Flags().String("org", "", "Organization name (required)")
	cmd.Flags().String("team-slug", "", "Team slug (required)")
	cmd.Flags().String("role", "all", "Filter by role: all, member, maintainer")
	cmd.Flags().Int("limit", 20, "Maximum number of members to return")
	_ = cmd.MarkFlagRequired("org")
	_ = cmd.MarkFlagRequired("team-slug")
	return cmd
}

func makeRunTeamsMembers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		org, _ := cmd.Flags().GetString("org")
		teamSlug, _ := cmd.Flags().GetString("team-slug")
		role, _ := cmd.Flags().GetString("role")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/orgs/%s/teams/%s/members?role=%s&per_page=%d", org, teamSlug, role, limit)

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing members of team %s/%s: %w", org, teamSlug, err)
		}

		members := make([]MemberInfo, 0, len(data))
		for _, d := range data {
			members = append(members, toMemberInfo(d))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(members)
		}
		if len(members) == 0 {
			fmt.Println("No members found.")
			return nil
		}
		lines := make([]string, 0, len(members)+1)
		lines = append(lines, fmt.Sprintf("%-30s  %-12s  %s", "LOGIN", "ID", "ROLE"))
		for _, m := range members {
			lines = append(lines, fmt.Sprintf("%-30s  %-12s  %s", truncate(m.Login, 30), strconv.FormatInt(m.ID, 10), m.Role))
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- repos ---

func newTeamsReposCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repos",
		Short: "List repositories accessible to a team",
		RunE:  makeRunTeamsRepos(factory),
	}
	cmd.Flags().String("org", "", "Organization name (required)")
	cmd.Flags().String("team-slug", "", "Team slug (required)")
	cmd.Flags().Int("limit", 20, "Maximum number of repositories to return")
	_ = cmd.MarkFlagRequired("org")
	_ = cmd.MarkFlagRequired("team-slug")
	return cmd
}

func makeRunTeamsRepos(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		org, _ := cmd.Flags().GetString("org")
		teamSlug, _ := cmd.Flags().GetString("team-slug")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/orgs/%s/teams/%s/repos?per_page=%d", org, teamSlug, limit)

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing repos for team %s/%s: %w", org, teamSlug, err)
		}

		summaries := make([]RepoSummary, 0, len(data))
		for _, d := range data {
			summaries = append(summaries, toRepoSummary(d))
		}
		return printRepoSummaries(cmd, summaries)
	}
}

// --- add-repo ---

func newTeamsAddRepoCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-repo",
		Short: "Add or update a repository for a team",
		RunE:  makeRunTeamsAddRepo(factory),
	}
	cmd.Flags().String("org", "", "Organization name (required)")
	cmd.Flags().String("team-slug", "", "Team slug (required)")
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("permission", "push", "Permission level: pull, push, admin")
	_ = cmd.MarkFlagRequired("org")
	_ = cmd.MarkFlagRequired("team-slug")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunTeamsAddRepo(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		org, _ := cmd.Flags().GetString("org")
		teamSlug, _ := cmd.Flags().GetString("team-slug")
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		permission, _ := cmd.Flags().GetString("permission")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would add repo %s/%s to team %s/%s with permission %q", owner, repo, org, teamSlug, permission), map[string]any{
				"action":     "add-repo",
				"org":        org,
				"teamSlug":   teamSlug,
				"owner":      owner,
				"repo":       repo,
				"permission": permission,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"permission": permission,
		}
		path := fmt.Sprintf("/orgs/%s/teams/%s/repos/%s/%s", org, teamSlug, owner, repo)
		if _, err := doGitHub(client, http.MethodPut, path, body, nil); err != nil {
			return fmt.Errorf("adding repo %s/%s to team %s/%s: %w", owner, repo, org, teamSlug, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{
				"status":     "added",
				"org":        org,
				"teamSlug":   teamSlug,
				"owner":      owner,
				"repo":       repo,
				"permission": permission,
			})
		}
		fmt.Printf("Added: %s/%s to team %s/%s (permission: %s)\n", owner, repo, org, teamSlug, permission)
		return nil
	}
}

// --- remove-repo ---

func newTeamsRemoveRepoCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-repo",
		Short: "Remove a repository from a team",
		RunE:  makeRunTeamsRemoveRepo(factory),
	}
	cmd.Flags().String("org", "", "Organization name (required)")
	cmd.Flags().String("team-slug", "", "Team slug (required)")
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Bool("confirm", false, "Confirm removal of repository from team")
	_ = cmd.MarkFlagRequired("org")
	_ = cmd.MarkFlagRequired("team-slug")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunTeamsRemoveRepo(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		org, _ := cmd.Flags().GetString("org")
		teamSlug, _ := cmd.Flags().GetString("team-slug")
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/orgs/%s/teams/%s/repos/%s/%s", org, teamSlug, owner, repo)
		if _, err := doGitHub(client, http.MethodDelete, path, nil, nil); err != nil {
			return fmt.Errorf("removing repo %s/%s from team %s/%s: %w", owner, repo, org, teamSlug, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{
				"status":   "removed",
				"org":      org,
				"teamSlug": teamSlug,
				"owner":    owner,
				"repo":     repo,
			})
		}
		fmt.Printf("Removed: %s/%s from team %s/%s\n", owner, repo, org, teamSlug)
		return nil
	}
}

// --- conversion helpers ---

// toTeamSummary converts a GitHub API team response to a TeamSummary.
func toTeamSummary(data map[string]any) TeamSummary {
	return TeamSummary{
		ID:          jsonInt64(data["id"]),
		Name:        jsonString(data["name"]),
		Slug:        jsonString(data["slug"]),
		Description: jsonString(data["description"]),
		Permission:  jsonString(data["permission"]),
	}
}

// --- output helpers ---

// printTeamSummaries outputs team summaries as JSON or a formatted text table.
func printTeamSummaries(cmd *cobra.Command, teams []TeamSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(teams)
	}
	if len(teams) == 0 {
		fmt.Println("No teams found.")
		return nil
	}
	lines := make([]string, 0, len(teams)+1)
	lines = append(lines, fmt.Sprintf("%-12s  %-30s  %-30s  %s", "ID", "NAME", "SLUG", "PERMISSION"))
	for _, t := range teams {
		lines = append(lines, fmt.Sprintf("%-12d  %-30s  %-30s  %s", t.ID, truncate(t.Name, 30), truncate(t.Slug, 30), t.Permission))
	}
	cli.PrintText(lines)
	return nil
}
