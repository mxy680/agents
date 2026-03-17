package github

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- Conversion functions ---

// toOrgSummary converts a GitHub API org list response item to an OrgSummary.
func toOrgSummary(data map[string]any) OrgSummary {
	return OrgSummary{
		Login:       jsonString(data["login"]),
		ID:          jsonInt64(data["id"]),
		Description: jsonString(data["description"]),
		URL:         jsonString(data["url"]),
	}
}

// toOrgDetail converts a GitHub API org detail response to an OrgDetail.
func toOrgDetail(data map[string]any) OrgDetail {
	return OrgDetail{
		Login:       jsonString(data["login"]),
		ID:          jsonInt64(data["id"]),
		Name:        jsonString(data["name"]),
		Description: jsonString(data["description"]),
		URL:         jsonString(data["url"]),
		Blog:        jsonString(data["blog"]),
		Location:    jsonString(data["location"]),
		Email:       jsonString(data["email"]),
		PublicRepos: jsonInt(data["public_repos"]),
		CreatedAt:   jsonString(data["created_at"]),
	}
}

// toMemberInfo converts a GitHub API member response item to a MemberInfo.
func toMemberInfo(data map[string]any) MemberInfo {
	return MemberInfo{
		Login:     jsonString(data["login"]),
		ID:        jsonInt64(data["id"]),
		AvatarURL: jsonString(data["avatar_url"]),
		Role:      jsonString(data["role"]),
	}
}

// --- Commands ---

func newOrgsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organizations for the authenticated user",
		RunE:  makeRunOrgsList(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of organizations to return")
	cmd.Flags().String("page-token", "", "Page number for pagination")
	return cmd
}

func makeRunOrgsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/user/orgs?per_page=%d", limit)
		if pageToken != "" {
			path = fmt.Sprintf("%s&page=%s", path, pageToken)
		}

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing organizations: %w", err)
		}

		summaries := make([]OrgSummary, 0, len(data))
		for _, d := range data {
			summaries = append(summaries, toOrgSummary(d))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}
		if len(summaries) == 0 {
			fmt.Println("No organizations found.")
			return nil
		}
		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-25s  %-12s  %s", "LOGIN", "ID", "DESCRIPTION"))
		for _, o := range summaries {
			lines = append(lines, fmt.Sprintf("%-25s  %-12s  %s", truncate(o.Login, 25), strconv.FormatInt(o.ID, 10), truncate(o.Description, 50)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newOrgsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get organization details",
		RunE:  makeRunOrgsGet(factory),
	}
	cmd.Flags().String("org", "", "Organization login (required)")
	_ = cmd.MarkFlagRequired("org")
	return cmd
}

func makeRunOrgsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		org, _ := cmd.Flags().GetString("org")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, fmt.Sprintf("/orgs/%s", org), nil, &data); err != nil {
			return fmt.Errorf("getting organization %s: %w", org, err)
		}

		detail := toOrgDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("Login:        %s", detail.Login),
			fmt.Sprintf("ID:           %d", detail.ID),
			fmt.Sprintf("Name:         %s", detail.Name),
			fmt.Sprintf("Description:  %s", detail.Description),
			fmt.Sprintf("URL:          %s", detail.URL),
			fmt.Sprintf("Blog:         %s", detail.Blog),
			fmt.Sprintf("Location:     %s", detail.Location),
			fmt.Sprintf("Email:        %s", detail.Email),
			fmt.Sprintf("Public Repos: %d", detail.PublicRepos),
			fmt.Sprintf("Created:      %s", detail.CreatedAt),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newOrgsMembersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List members of an organization",
		RunE:  makeRunOrgsMembers(factory),
	}
	cmd.Flags().String("org", "", "Organization login (required)")
	cmd.Flags().String("role", "all", "Filter by role: all, admin, member")
	cmd.Flags().Int("limit", 20, "Maximum number of members to return")
	_ = cmd.MarkFlagRequired("org")
	return cmd
}

func makeRunOrgsMembers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		org, _ := cmd.Flags().GetString("org")
		role, _ := cmd.Flags().GetString("role")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/orgs/%s/members?role=%s&per_page=%d", org, role, limit)

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing members of organization %s: %w", org, err)
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
		lines = append(lines, fmt.Sprintf("%-25s  %-12s  %s", "LOGIN", "ID", "ROLE"))
		for _, m := range members {
			lines = append(lines, fmt.Sprintf("%-25s  %-12s  %s", truncate(m.Login, 25), strconv.FormatInt(m.ID, 10), m.Role))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newOrgsReposCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repos",
		Short: "List repositories of an organization",
		RunE:  makeRunOrgsRepos(factory),
	}
	cmd.Flags().String("org", "", "Organization login (required)")
	cmd.Flags().String("type", "all", "Filter by type: all, public, private, forks, sources, member")
	cmd.Flags().Int("limit", 20, "Maximum number of repositories to return")
	_ = cmd.MarkFlagRequired("org")
	return cmd
}

func makeRunOrgsRepos(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		org, _ := cmd.Flags().GetString("org")
		repoType, _ := cmd.Flags().GetString("type")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/orgs/%s/repos?type=%s&per_page=%d", org, repoType, limit)

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing repositories of organization %s: %w", org, err)
		}

		summaries := make([]RepoSummary, 0, len(data))
		for _, d := range data {
			summaries = append(summaries, toRepoSummary(d))
		}
		return printRepoSummaries(cmd, summaries)
	}
}
