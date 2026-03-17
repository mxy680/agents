package github

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newReposListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories",
		RunE:  makeRunReposList(factory),
	}
	cmd.Flags().String("owner", "", "GitHub username to list repos for (default: authenticated user)")
	cmd.Flags().String("type", "all", "Type filter: all, owner, member, public, private")
	cmd.Flags().String("sort", "updated", "Sort order: created, updated, pushed, full_name")
	cmd.Flags().Int("limit", 20, "Maximum number of repositories to return")
	cmd.Flags().String("page-token", "", "Page number for pagination")
	return cmd
}

func makeRunReposList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repoType, _ := cmd.Flags().GetString("type")
		sort, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var path string
		if owner != "" {
			path = fmt.Sprintf("/users/%s/repos", owner)
		} else {
			path = "/user/repos"
		}

		path = fmt.Sprintf("%s?type=%s&sort=%s&per_page=%d", path, repoType, sort, limit)
		if pageToken != "" {
			path = fmt.Sprintf("%s&page=%s", path, pageToken)
		}

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing repositories: %w", err)
		}

		summaries := make([]RepoSummary, 0, len(data))
		for _, d := range data {
			summaries = append(summaries, toRepoSummary(d))
		}
		return printRepoSummaries(cmd, summaries)
	}
}

func newReposGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get repository details",
		RunE:  makeRunReposGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunReposGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, fmt.Sprintf("/repos/%s/%s", owner, repo), nil, &data); err != nil {
			return fmt.Errorf("getting repository %s/%s: %w", owner, repo, err)
		}

		detail := toRepoDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:             %d", detail.ID),
			fmt.Sprintf("Name:           %s", detail.Name),
			fmt.Sprintf("Full Name:      %s", detail.FullName),
			fmt.Sprintf("Owner:          %s", detail.Owner),
			fmt.Sprintf("Private:        %v", detail.Private),
			fmt.Sprintf("Description:    %s", detail.Description),
			fmt.Sprintf("URL:            %s", detail.URL),
			fmt.Sprintf("Clone URL:      %s", detail.CloneURL),
			fmt.Sprintf("Default Branch: %s", detail.DefaultBranch),
			fmt.Sprintf("Language:       %s", detail.Language),
			fmt.Sprintf("Stars:          %d", detail.Stars),
			fmt.Sprintf("Forks:          %d", detail.Forks),
			fmt.Sprintf("Open Issues:    %d", detail.OpenIssues),
			fmt.Sprintf("Created:        %s", detail.CreatedAt),
			fmt.Sprintf("Updated:        %s", detail.UpdatedAt),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newReposCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new repository",
		RunE:  makeRunReposCreate(factory),
	}
	cmd.Flags().String("name", "", "Repository name (required)")
	cmd.Flags().String("description", "", "Repository description")
	cmd.Flags().Bool("private", false, "Make the repository private")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunReposCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		private, _ := cmd.Flags().GetBool("private")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create repository %q", name), map[string]any{
				"action":      "create",
				"name":        name,
				"description": description,
				"private":     private,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"name":        name,
			"description": description,
			"private":     private,
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodPost, "/user/repos", body, &data); err != nil {
			return fmt.Errorf("creating repository %q: %w", name, err)
		}

		detail := toRepoDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Created: %s (%s)\n", detail.FullName, detail.URL)
		return nil
	}
}

func newReposForkCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fork",
		Short: "Fork a repository",
		RunE:  makeRunReposFork(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("org", "", "Organization to fork into (default: authenticated user)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunReposFork(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		org, _ := cmd.Flags().GetString("org")

		if cli.IsDryRun(cmd) {
			desc := fmt.Sprintf("Would fork repository %s/%s", owner, repo)
			if org != "" {
				desc = fmt.Sprintf("%s into organization %s", desc, org)
			}
			return dryRunResult(cmd, desc, map[string]any{
				"action": "fork",
				"owner":  owner,
				"repo":   repo,
				"org":    org,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{}
		if org != "" {
			body["organization"] = org
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodPost, fmt.Sprintf("/repos/%s/%s/forks", owner, repo), body, &data); err != nil {
			return fmt.Errorf("forking repository %s/%s: %w", owner, repo, err)
		}

		detail := toRepoDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Forked: %s/%s → %s (%s)\n", owner, repo, detail.FullName, detail.URL)
		return nil
	}
}

func newReposDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a repository (irreversible)",
		RunE:  makeRunReposDelete(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunReposDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete repository %s/%s", owner, repo), map[string]any{
				"action": "delete",
				"owner":  owner,
				"repo":   repo,
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

		if _, err := doGitHub(client, http.MethodDelete, fmt.Sprintf("/repos/%s/%s", owner, repo), nil, nil); err != nil {
			return fmt.Errorf("deleting repository %s/%s: %w", owner, repo, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "owner": owner, "repo": repo})
		}
		fmt.Printf("Deleted: %s/%s\n", owner, repo)
		return nil
	}
}
