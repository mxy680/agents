package github

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newReleasesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List releases for a repository",
		RunE:  makeRunReleasesList(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int("limit", 20, "Maximum number of releases to return")
	cmd.Flags().String("page-token", "", "Page number for pagination")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunReleasesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/releases?per_page=%d", owner, repo, limit)
		if pageToken != "" {
			path = fmt.Sprintf("%s&page=%s", path, pageToken)
		}

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing releases for %s/%s: %w", owner, repo, err)
		}

		summaries := make([]ReleaseSummary, 0, len(data))
		for _, d := range data {
			summaries = append(summaries, toReleaseSummary(d))
		}
		return printReleaseSummaries(cmd, summaries)
	}
}

func newReleasesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a release by ID, tag, or latest",
		RunE:  makeRunReleasesGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int64("release-id", 0, "Release ID")
	cmd.Flags().String("tag", "", "Tag name")
	cmd.Flags().Bool("latest", false, "Get the latest release")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunReleasesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		releaseID, _ := cmd.Flags().GetInt64("release-id")
		tag, _ := cmd.Flags().GetString("tag")
		latest, _ := cmd.Flags().GetBool("latest")

		// Validate exactly one selector is provided
		set := 0
		if releaseID != 0 {
			set++
		}
		if tag != "" {
			set++
		}
		if latest {
			set++
		}
		if set == 0 {
			return fmt.Errorf("exactly one of --release-id, --tag, or --latest must be provided")
		}
		if set > 1 {
			return fmt.Errorf("exactly one of --release-id, --tag, or --latest must be provided")
		}

		var path string
		switch {
		case releaseID != 0:
			path = fmt.Sprintf("/repos/%s/%s/releases/%s", owner, repo, strconv.FormatInt(releaseID, 10))
		case tag != "":
			path = fmt.Sprintf("/repos/%s/%s/releases/tags/%s", owner, repo, tag)
		default:
			path = fmt.Sprintf("/repos/%s/%s/releases/latest", owner, repo)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("getting release: %w", err)
		}

		detail := toReleaseDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:           %d", detail.ID),
			fmt.Sprintf("Tag:          %s", detail.TagName),
			fmt.Sprintf("Name:         %s", detail.Name),
			fmt.Sprintf("Target:       %s", detail.Target),
			fmt.Sprintf("Draft:        %v", detail.Draft),
			fmt.Sprintf("Prerelease:   %v", detail.Prerelease),
			fmt.Sprintf("Created:      %s", detail.CreatedAt),
			fmt.Sprintf("Published:    %s", detail.PublishedAt),
		}
		if detail.URL != "" {
			lines = append(lines, fmt.Sprintf("URL:          %s", detail.URL))
		}
		if detail.Body != "" {
			lines = append(lines, fmt.Sprintf("Body:\n%s", detail.Body))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newReleasesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new release",
		RunE:  makeRunReleasesCreate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("tag", "", "Tag name for the release (required)")
	cmd.Flags().String("name", "", "Release name")
	cmd.Flags().String("body", "", "Release description / notes")
	cmd.Flags().String("target", "", "Commitish value (branch, SHA) the tag is created from")
	cmd.Flags().Bool("draft", false, "Create as a draft release")
	cmd.Flags().Bool("prerelease", false, "Mark as a pre-release")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

func makeRunReleasesCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		tag, _ := cmd.Flags().GetString("tag")
		name, _ := cmd.Flags().GetString("name")
		body, _ := cmd.Flags().GetString("body")
		target, _ := cmd.Flags().GetString("target")
		draft, _ := cmd.Flags().GetBool("draft")
		prerelease, _ := cmd.Flags().GetBool("prerelease")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create release %q for %s/%s", tag, owner, repo), map[string]any{
				"action":     "create",
				"owner":      owner,
				"repo":       repo,
				"tagName":    tag,
				"name":       name,
				"body":       body,
				"target":     target,
				"draft":      draft,
				"prerelease": prerelease,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"tag_name":   tag,
			"name":       name,
			"body":       body,
			"draft":      draft,
			"prerelease": prerelease,
		}
		if target != "" {
			reqBody["target_commitish"] = target
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodPost, fmt.Sprintf("/repos/%s/%s/releases", owner, repo), reqBody, &data); err != nil {
			return fmt.Errorf("creating release %q for %s/%s: %w", tag, owner, repo, err)
		}

		detail := toReleaseDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Created: %s (%d)\n", detail.TagName, detail.ID)
		return nil
	}
}

func newReleasesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a release (irreversible)",
		RunE:  makeRunReleasesDelete(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int64("release-id", 0, "Release ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("release-id")
	return cmd
}

func makeRunReleasesDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		releaseID, _ := cmd.Flags().GetInt64("release-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete release %d from %s/%s", releaseID, owner, repo), map[string]any{
				"action":    "delete",
				"owner":     owner,
				"repo":      repo,
				"releaseId": releaseID,
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

		path := fmt.Sprintf("/repos/%s/%s/releases/%s", owner, repo, strconv.FormatInt(releaseID, 10))
		if _, err := doGitHub(client, http.MethodDelete, path, nil, nil); err != nil {
			return fmt.Errorf("deleting release %d from %s/%s: %w", releaseID, owner, repo, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"status": "deleted", "owner": owner, "repo": repo, "releaseId": releaseID})
		}
		fmt.Printf("Deleted: release %d\n", releaseID)
		return nil
	}
}
