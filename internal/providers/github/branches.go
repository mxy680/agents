package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- Conversion functions ---

// toBranchSummary converts a GitHub API branch response to a BranchSummary.
func toBranchSummary(data map[string]any) BranchSummary {
	sha := ""
	if commit, ok := data["commit"].(map[string]any); ok {
		sha = jsonString(commit["sha"])
	}
	return BranchSummary{
		Name:      jsonString(data["name"]),
		SHA:       sha,
		Protected: jsonBool(data["protected"]),
	}
}

// toBranchProtectionInfo converts a GitHub API branch protection response to a BranchProtectionInfo.
func toBranchProtectionInfo(data map[string]any) BranchProtectionInfo {
	info := BranchProtectionInfo{
		URL: jsonString(data["url"]),
	}

	if ea, ok := data["enforce_admins"].(map[string]any); ok {
		info.EnforceAdmins = jsonBool(ea["enabled"])
	}

	if rpr, ok := data["required_pull_request_reviews"].(map[string]any); ok {
		info.RequiredPullReviewCount = jsonInt(rpr["required_approving_review_count"])
		info.RequireCodeOwnerReviews = jsonBool(rpr["require_code_owner_reviews"])
	}

	if rsc, ok := data["required_status_checks"].(map[string]any); ok {
		if contexts, ok := rsc["contexts"].([]any); ok {
			checks := make([]string, 0, len(contexts))
			for _, c := range contexts {
				if s, ok := c.(string); ok {
					checks = append(checks, s)
				}
			}
			info.RequiredStatusChecks = checks
		}
	}

	return info
}

// --- Command constructors ---

func newBranchesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List branches for a repository",
		RunE:  makeRunBranchesList(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Bool("protected", false, "Filter to protected branches only")
	cmd.Flags().Int("limit", 20, "Maximum number of branches to return")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunBranchesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		protected, _ := cmd.Flags().GetBool("protected")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/branches?per_page=%d", owner, repo, limit)
		if protected {
			path = fmt.Sprintf("%s&protected=%s", path, strconv.FormatBool(protected))
		}

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing branches for %s/%s: %w", owner, repo, err)
		}

		summaries := make([]BranchSummary, 0, len(data))
		for _, d := range data {
			summaries = append(summaries, toBranchSummary(d))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No branches found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-40s  %-40s  %s", "NAME", "SHA", "PROTECTED"))
		for _, b := range summaries {
			lines = append(lines, fmt.Sprintf("%-40s  %-40s  %v", truncate(b.Name, 40), truncate(b.SHA, 40), b.Protected))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newBranchesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get branch details",
		RunE:  makeRunBranchesGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("branch", "", "Branch name (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("branch")
	return cmd
}

func makeRunBranchesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		branch, _ := cmd.Flags().GetString("branch")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/branches/%s", owner, repo, branch)

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("getting branch %s in %s/%s: %w", branch, owner, repo, err)
		}

		summary := toBranchSummary(data)
		branchURL := ""
		if links, ok := data["_links"].(map[string]any); ok {
			branchURL = jsonString(links["html"])
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summary)
		}

		lines := []string{
			fmt.Sprintf("Name:       %s", summary.Name),
			fmt.Sprintf("SHA:        %s", summary.SHA),
			fmt.Sprintf("Protected:  %v", summary.Protected),
		}
		if branchURL != "" {
			lines = append(lines, fmt.Sprintf("URL:        %s", branchURL))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newProtectionGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get branch protection rules",
		RunE:  makeRunProtectionGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("branch", "", "Branch name (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("branch")
	return cmd
}

func makeRunProtectionGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		branch, _ := cmd.Flags().GetString("branch")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/branches/%s/protection", owner, repo, branch)

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("getting branch protection for %s in %s/%s: %w", branch, owner, repo, err)
		}

		info := toBranchProtectionInfo(data)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("URL:                      %s", info.URL),
			fmt.Sprintf("Enforce Admins:           %v", info.EnforceAdmins),
			fmt.Sprintf("Required Review Count:    %d", info.RequiredPullReviewCount),
			fmt.Sprintf("Require Code Owner:       %v", info.RequireCodeOwnerReviews),
			fmt.Sprintf("Required Status Checks:   %s", strings.Join(info.RequiredStatusChecks, ", ")),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newProtectionUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update branch protection rules",
		RunE:  makeRunProtectionUpdate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("branch", "", "Branch name (required)")
	cmd.Flags().String("settings", "", "Protection settings as a JSON string")
	cmd.Flags().String("settings-file", "", "Path to a JSON file containing protection settings")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("branch")
	return cmd
}

func makeRunProtectionUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		branch, _ := cmd.Flags().GetString("branch")
		settings, _ := cmd.Flags().GetString("settings")
		settingsFile, _ := cmd.Flags().GetString("settings-file")

		if settings != "" && settingsFile != "" {
			return fmt.Errorf("--settings and --settings-file are mutually exclusive")
		}
		if settings == "" && settingsFile == "" {
			return fmt.Errorf("one of --settings or --settings-file is required")
		}

		var rawJSON string
		if settingsFile != "" {
			data, err := os.ReadFile(settingsFile)
			if err != nil {
				return fmt.Errorf("reading settings file %s: %w", settingsFile, err)
			}
			rawJSON = string(data)
		} else {
			rawJSON = settings
		}

		var body map[string]any
		if err := json.Unmarshal([]byte(rawJSON), &body); err != nil {
			return fmt.Errorf("parsing protection settings JSON: %w", err)
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update branch protection for %s in %s/%s", branch, owner, repo), map[string]any{
				"action":   "update_protection",
				"owner":    owner,
				"repo":     repo,
				"branch":   branch,
				"settings": body,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/branches/%s/protection", owner, repo, branch)

		var data map[string]any
		if _, err := doGitHub(client, http.MethodPut, path, body, &data); err != nil {
			return fmt.Errorf("updating branch protection for %s in %s/%s: %w", branch, owner, repo, err)
		}

		info := toBranchProtectionInfo(data)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		fmt.Printf("Branch protection updated for %s in %s/%s\n", branch, owner, repo)
		lines := []string{
			fmt.Sprintf("URL:                      %s", info.URL),
			fmt.Sprintf("Enforce Admins:           %v", info.EnforceAdmins),
			fmt.Sprintf("Required Review Count:    %d", info.RequiredPullReviewCount),
			fmt.Sprintf("Require Code Owner:       %v", info.RequireCodeOwnerReviews),
			fmt.Sprintf("Required Status Checks:   %s", strings.Join(info.RequiredStatusChecks, ", ")),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newProtectionDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete branch protection rules",
		RunE:  makeRunProtectionDelete(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("branch", "", "Branch name (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("branch")
	return cmd
}

func makeRunProtectionDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		branch, _ := cmd.Flags().GetString("branch")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete branch protection for %s in %s/%s", branch, owner, repo), map[string]any{
				"action": "delete_protection",
				"owner":  owner,
				"repo":   repo,
				"branch": branch,
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

		path := fmt.Sprintf("/repos/%s/%s/branches/%s/protection", owner, repo, branch)

		if _, err := doGitHub(client, http.MethodDelete, path, nil, nil); err != nil {
			return fmt.Errorf("deleting branch protection for %s in %s/%s: %w", branch, owner, repo, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{
				"status": "deleted",
				"owner":  owner,
				"repo":   repo,
				"branch": branch,
			})
		}
		fmt.Printf("Deleted branch protection for %s in %s/%s\n", branch, owner, repo)
		return nil
	}
}
