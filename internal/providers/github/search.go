package github

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newSearchReposCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repos",
		Short: "Search GitHub repositories",
		RunE:  makeRunSearchRepos(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().String("sort", "", "Sort field: stars, forks, updated")
	cmd.Flags().String("order", "desc", "Sort order: asc, desc")
	cmd.Flags().Int("limit", 20, "Maximum number of results to return")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunSearchRepos(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		sort, _ := cmd.Flags().GetString("sort")
		order, _ := cmd.Flags().GetString("order")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", query)
		params.Set("order", order)
		params.Set("per_page", fmt.Sprintf("%d", limit))
		if sort != "" {
			params.Set("sort", sort)
		}
		path := "/search/repositories?" + params.Encode()

		var raw map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &raw); err != nil {
			return fmt.Errorf("searching repositories: %w", err)
		}

		totalCount := jsonInt(raw["total_count"])
		items, _ := raw["items"].([]any)

		summaries := make([]RepoSummary, 0, len(items))
		for _, item := range items {
			if d, ok := item.(map[string]any); ok {
				summaries = append(summaries, toRepoSummary(d))
			}
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(SearchResult{TotalCount: totalCount, Items: summaries})
		}

		lines := []string{fmt.Sprintf("Found %d results", totalCount)}
		if len(summaries) > 0 {
			lines = append(lines, fmt.Sprintf("%-40s  %-20s  %-7s  %s", "NAME", "OWNER", "PRIVATE", "UPDATED"))
			for _, r := range summaries {
				lines = append(lines, fmt.Sprintf("%-40s  %-20s  %-7v  %s", truncate(r.FullName, 40), truncate(r.Owner, 20), r.Private, r.UpdatedAt))
			}
		}
		cli.PrintText(lines)
		return nil
	}
}

func newSearchCodeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code",
		Short: "Search code on GitHub",
		RunE:  makeRunSearchCode(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().String("sort", "", "Sort field: indexed")
	cmd.Flags().String("order", "desc", "Sort order: asc, desc")
	cmd.Flags().Int("limit", 20, "Maximum number of results to return")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunSearchCode(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		sort, _ := cmd.Flags().GetString("sort")
		order, _ := cmd.Flags().GetString("order")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", query)
		params.Set("order", order)
		params.Set("per_page", fmt.Sprintf("%d", limit))
		if sort != "" {
			params.Set("sort", sort)
		}
		path := "/search/code?" + params.Encode()

		var raw map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &raw); err != nil {
			return fmt.Errorf("searching code: %w", err)
		}

		totalCount := jsonInt(raw["total_count"])
		items, _ := raw["items"].([]any)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(SearchResult{TotalCount: totalCount, Items: items})
		}

		lines := []string{fmt.Sprintf("Found %d results", totalCount)}
		if len(items) > 0 {
			lines = append(lines, fmt.Sprintf("%-50s  %-35s  %s", "PATH", "REPO", "SHA"))
			for _, item := range items {
				d, ok := item.(map[string]any)
				if !ok {
					continue
				}
				filePath := jsonString(d["path"])
				repoFullName := jsonNestedString(d["repository"], "full_name")
				sha := jsonString(d["sha"])
				lines = append(lines, fmt.Sprintf("%-50s  %-35s  %s", truncate(filePath, 50), truncate(repoFullName, 35), truncate(sha, 40)))
			}
		}
		cli.PrintText(lines)
		return nil
	}
}

func newSearchIssuesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issues",
		Short: "Search GitHub issues and pull requests",
		RunE:  makeRunSearchIssues(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().String("sort", "", "Sort field: created, updated, comments")
	cmd.Flags().String("order", "desc", "Sort order: asc, desc")
	cmd.Flags().Int("limit", 20, "Maximum number of results to return")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunSearchIssues(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		sort, _ := cmd.Flags().GetString("sort")
		order, _ := cmd.Flags().GetString("order")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", query)
		params.Set("order", order)
		params.Set("per_page", fmt.Sprintf("%d", limit))
		if sort != "" {
			params.Set("sort", sort)
		}
		path := "/search/issues?" + params.Encode()

		var raw map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &raw); err != nil {
			return fmt.Errorf("searching issues: %w", err)
		}

		totalCount := jsonInt(raw["total_count"])
		items, _ := raw["items"].([]any)

		summaries := make([]IssueSummary, 0, len(items))
		for _, item := range items {
			if d, ok := item.(map[string]any); ok {
				summaries = append(summaries, toIssueSummary(d))
			}
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(SearchResult{TotalCount: totalCount, Items: summaries})
		}

		lines := []string{fmt.Sprintf("Found %d results", totalCount)}
		if len(summaries) > 0 {
			lines = append(lines, fmt.Sprintf("%-6s  %-50s  %-8s  %-15s  %s", "NUM", "TITLE", "STATE", "USER", "UPDATED"))
			for _, i := range summaries {
				lines = append(lines, fmt.Sprintf("%-6d  %-50s  %-8s  %-15s  %s", i.Number, truncate(i.Title, 50), i.State, truncate(i.User, 15), i.UpdatedAt))
			}
		}
		cli.PrintText(lines)
		return nil
	}
}

func newSearchCommitsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commits",
		Short: "Search commits on GitHub",
		RunE:  makeRunSearchCommits(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().String("sort", "", "Sort field: author-date, committer-date")
	cmd.Flags().String("order", "desc", "Sort order: asc, desc")
	cmd.Flags().Int("limit", 20, "Maximum number of results to return")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunSearchCommits(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		sort, _ := cmd.Flags().GetString("sort")
		order, _ := cmd.Flags().GetString("order")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", query)
		params.Set("order", order)
		params.Set("per_page", fmt.Sprintf("%d", limit))
		if sort != "" {
			params.Set("sort", sort)
		}
		path := "/search/commits?" + params.Encode()

		var raw map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &raw); err != nil {
			return fmt.Errorf("searching commits: %w", err)
		}

		totalCount := jsonInt(raw["total_count"])
		items, _ := raw["items"].([]any)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(SearchResult{TotalCount: totalCount, Items: items})
		}

		lines := []string{fmt.Sprintf("Found %d results", totalCount)}
		if len(items) > 0 {
			lines = append(lines, fmt.Sprintf("%-7s  %-50s  %-20s  %s", "SHA", "MESSAGE", "AUTHOR", "DATE"))
			for _, item := range items {
				d, ok := item.(map[string]any)
				if !ok {
					continue
				}
				sha := jsonString(d["sha"])
				if len(sha) > 7 {
					sha = sha[:7]
				}
				commitData, _ := d["commit"].(map[string]any)
				message := jsonString(commitData["message"])
				// Use only the first line of the commit message
				if idx := strings.IndexByte(message, '\n'); idx >= 0 {
					message = message[:idx]
				}
				authorData, _ := commitData["author"].(map[string]any)
				authorName := jsonString(authorData["name"])
				authorDate := jsonString(authorData["date"])
				lines = append(lines, fmt.Sprintf("%-7s  %-50s  %-20s  %s", sha, truncate(message, 50), truncate(authorName, 20), authorDate))
			}
		}
		cli.PrintText(lines)
		return nil
	}
}

func newSearchUsersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Search GitHub users",
		RunE:  makeRunSearchUsers(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().String("sort", "", "Sort field: followers, repositories, joined")
	cmd.Flags().String("order", "desc", "Sort order: asc, desc")
	cmd.Flags().Int("limit", 20, "Maximum number of results to return")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunSearchUsers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		sort, _ := cmd.Flags().GetString("sort")
		order, _ := cmd.Flags().GetString("order")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", query)
		params.Set("order", order)
		params.Set("per_page", fmt.Sprintf("%d", limit))
		if sort != "" {
			params.Set("sort", sort)
		}
		path := "/search/users?" + params.Encode()

		var raw map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &raw); err != nil {
			return fmt.Errorf("searching users: %w", err)
		}

		totalCount := jsonInt(raw["total_count"])
		items, _ := raw["items"].([]any)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(SearchResult{TotalCount: totalCount, Items: items})
		}

		lines := []string{fmt.Sprintf("Found %d results", totalCount)}
		if len(items) > 0 {
			lines = append(lines, fmt.Sprintf("%-30s  %-12s  %s", "LOGIN", "ID", "TYPE"))
			for _, item := range items {
				d, ok := item.(map[string]any)
				if !ok {
					continue
				}
				login := jsonString(d["login"])
				id := jsonInt64(d["id"])
				userType := jsonString(d["type"])
				lines = append(lines, fmt.Sprintf("%-30s  %-12d  %s", truncate(login, 30), id, userType))
			}
		}
		cli.PrintText(lines)
		return nil
	}
}
