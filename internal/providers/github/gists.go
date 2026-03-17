package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// toGistSummary converts a GitHub API gist list response item to a GistSummary.
func toGistSummary(data map[string]any) GistSummary {
	files := extractGistFilenames(data["files"])
	return GistSummary{
		ID:          jsonString(data["id"]),
		Description: jsonString(data["description"]),
		Public:      jsonBool(data["public"]),
		Files:       files,
		CreatedAt:   jsonString(data["created_at"]),
		UpdatedAt:   jsonString(data["updated_at"]),
	}
}

// toGistDetail converts a GitHub API gist detail response to a GistDetail.
func toGistDetail(data map[string]any) GistDetail {
	files := extractGistFiles(data["files"])
	return GistDetail{
		ID:          jsonString(data["id"]),
		Description: jsonString(data["description"]),
		Public:      jsonBool(data["public"]),
		Files:       files,
		URL:         jsonString(data["html_url"]),
		CreatedAt:   jsonString(data["created_at"]),
		UpdatedAt:   jsonString(data["updated_at"]),
	}
}

// extractGistFilenames returns the list of filenames from the gist files map.
func extractGistFilenames(v any) []string {
	if v == nil {
		return nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	return names
}

// extractGistFiles converts the raw files map from the API into a map of GistFile.
func extractGistFiles(v any) map[string]GistFile {
	if v == nil {
		return nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]GistFile, len(m))
	for name, raw := range m {
		if fileMap, ok := raw.(map[string]any); ok {
			result[name] = GistFile{
				Filename: jsonString(fileMap["filename"]),
				Language: jsonString(fileMap["language"]),
				Size:     jsonInt(fileMap["size"]),
				Content:  jsonString(fileMap["content"]),
			}
		}
	}
	return result
}

// parseFilesJSON reads the files JSON from --files or --files-file.
// Returns a map of filename -> {"content": "..."} objects suitable for the GitHub API.
func parseFilesJSON(filesFlag, filesFileFlag string) (map[string]map[string]string, error) {
	var raw string
	if filesFlag != "" && filesFileFlag != "" {
		return nil, fmt.Errorf("--files and --files-file are mutually exclusive")
	}
	if filesFileFlag != "" {
		data, err := os.ReadFile(filesFileFlag)
		if err != nil {
			return nil, fmt.Errorf("reading files file %s: %w", filesFileFlag, err)
		}
		raw = string(data)
	} else {
		raw = filesFlag
	}
	if raw == "" {
		return nil, nil
	}
	var result map[string]map[string]string
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parsing files JSON: %w", err)
	}
	return result, nil
}

func newGistsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List gists for the authenticated user",
		RunE:  makeRunGistsList(factory),
	}
	cmd.Flags().Bool("public", false, "Show only public gists (informational; lists all authenticated user gists)")
	cmd.Flags().Int("limit", 20, "Maximum number of gists to return")
	cmd.Flags().String("page-token", "", "Page number for pagination")
	return cmd
}

func makeRunGistsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/gists?per_page=%d", limit)
		if pageToken != "" {
			path = fmt.Sprintf("%s&page=%s", path, pageToken)
		}

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing gists: %w", err)
		}

		summaries := make([]GistSummary, 0, len(data))
		for _, d := range data {
			summaries = append(summaries, toGistSummary(d))
		}
		return printGistSummaries(cmd, summaries)
	}
}

func newGistsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get gist details by ID",
		RunE:  makeRunGistsGet(factory),
	}
	cmd.Flags().String("gist-id", "", "Gist ID (required)")
	_ = cmd.MarkFlagRequired("gist-id")
	return cmd
}

func makeRunGistsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		gistID, _ := cmd.Flags().GetString("gist-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, fmt.Sprintf("/gists/%s", gistID), nil, &data); err != nil {
			return fmt.Errorf("getting gist %s: %w", gistID, err)
		}

		detail := toGistDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", detail.ID),
			fmt.Sprintf("Description: %s", detail.Description),
			fmt.Sprintf("Public:      %v", detail.Public),
			fmt.Sprintf("URL:         %s", detail.URL),
			fmt.Sprintf("Created:     %s", detail.CreatedAt),
			fmt.Sprintf("Updated:     %s", detail.UpdatedAt),
		}
		for name, f := range detail.Files {
			lines = append(lines, fmt.Sprintf("File:        %s (%s, %d bytes)", name, f.Language, f.Size))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newGistsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new gist",
		RunE:  makeRunGistsCreate(factory),
	}
	cmd.Flags().String("description", "", "Gist description")
	cmd.Flags().String("files", "", `Files as JSON object: {"file.txt":{"content":"hello"}}`)
	cmd.Flags().String("files-file", "", "Path to JSON file containing the files object")
	cmd.Flags().Bool("public", false, "Make the gist public (default: secret)")
	return cmd
}

func makeRunGistsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		description, _ := cmd.Flags().GetString("description")
		filesFlag, _ := cmd.Flags().GetString("files")
		filesFileFlag, _ := cmd.Flags().GetString("files-file")
		public, _ := cmd.Flags().GetBool("public")

		files, err := parseFilesJSON(filesFlag, filesFileFlag)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create gist %q", description), map[string]any{
				"action":      "create",
				"description": description,
				"public":      public,
				"files":       files,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"description": description,
			"public":      public,
			"files":       files,
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodPost, "/gists", body, &data); err != nil {
			return fmt.Errorf("creating gist: %w", err)
		}

		detail := toGistDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Created: %s (%s)\n", detail.ID, detail.URL)
		return nil
	}
}

func newGistsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing gist",
		RunE:  makeRunGistsUpdate(factory),
	}
	cmd.Flags().String("gist-id", "", "Gist ID (required)")
	cmd.Flags().String("description", "", "New gist description")
	cmd.Flags().String("files", "", `Files to update as JSON object: {"file.txt":{"content":"new content"}}`)
	cmd.Flags().String("files-file", "", "Path to JSON file containing the files object")
	_ = cmd.MarkFlagRequired("gist-id")
	return cmd
}

func makeRunGistsUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		gistID, _ := cmd.Flags().GetString("gist-id")
		description, _ := cmd.Flags().GetString("description")
		filesFlag, _ := cmd.Flags().GetString("files")
		filesFileFlag, _ := cmd.Flags().GetString("files-file")

		files, err := parseFilesJSON(filesFlag, filesFileFlag)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update gist %s", gistID), map[string]any{
				"action":      "update",
				"gistId":      gistID,
				"description": description,
				"files":       files,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{}
		if description != "" {
			body["description"] = description
		}
		if files != nil {
			body["files"] = files
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodPatch, fmt.Sprintf("/gists/%s", gistID), body, &data); err != nil {
			return fmt.Errorf("updating gist %s: %w", gistID, err)
		}

		detail := toGistDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Updated: %s (%s)\n", detail.ID, detail.URL)
		return nil
	}
}

func newGistsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a gist (irreversible)",
		RunE:  makeRunGistsDelete(factory),
	}
	cmd.Flags().String("gist-id", "", "Gist ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("gist-id")
	return cmd
}

func makeRunGistsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		gistID, _ := cmd.Flags().GetString("gist-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete gist %s", gistID), map[string]any{
				"action": "delete",
				"gistId": gistID,
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

		if _, err := doGitHub(client, http.MethodDelete, fmt.Sprintf("/gists/%s", gistID), nil, nil); err != nil {
			return fmt.Errorf("deleting gist %s: %w", gistID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "gistId": gistID})
		}
		fmt.Printf("Deleted: %s\n", gistID)
		return nil
	}
}
