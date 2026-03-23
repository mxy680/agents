package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newPagesCmd returns the parent "pages" command with all subcommands attached.
func newPagesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pages",
		Short:   "Manage Canvas wiki pages",
		Aliases: []string{"page"},
	}

	cmd.AddCommand(newPagesListCmd(factory))
	cmd.AddCommand(newPagesGetCmd(factory))
	cmd.AddCommand(newPagesCreateCmd(factory))
	cmd.AddCommand(newPagesUpdateCmd(factory))
	cmd.AddCommand(newPagesDeleteCmd(factory))
	cmd.AddCommand(newPagesRevisionsCmd(factory))

	return cmd
}

// PageRevisionSummary is a condensed Canvas page revision.
type PageRevisionSummary struct {
	RevisionID int    `json:"revision_id"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	Latest     bool   `json:"latest,omitempty"`
	EditedBy   string `json:"edited_by,omitempty"`
}

func newPagesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List wiki pages in a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}

			sort, _ := cmd.Flags().GetString("sort")
			search, _ := cmd.Flags().GetString("search")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			if sort != "" {
				params.Set("sort", sort)
			}
			if search != "" {
				params.Set("search_term", search)
			}
			if cmd.Flags().Changed("published") {
				published, _ := cmd.Flags().GetBool("published")
				if published {
					params.Set("published", "true")
				} else {
					params.Set("published", "false")
				}
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/pages", params)
			if err != nil {
				return err
			}

			var pages []PageSummary
			if err := json.Unmarshal(data, &pages); err != nil {
				return fmt.Errorf("parse pages: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(pages)
			}

			if len(pages) == 0 {
				fmt.Println("No pages found.")
				return nil
			}
			for _, p := range pages {
				front := ""
				if p.FrontPage {
					front = " [front]"
				}
				pub := "draft"
				if p.Published {
					pub = "published"
				}
				fmt.Printf("%-10s  %-12s  %s%s\n", pub, p.URL, truncate(p.Title, 50), front)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("sort", "", "Sort by: title, created_at, updated_at")
	cmd.Flags().String("search", "", "Search term to filter pages")
	cmd.Flags().Bool("published", false, "Filter by published status")
	cmd.Flags().Int("limit", 0, "Maximum number of pages to return")
	return cmd
}

func newPagesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific wiki page",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			pageURL, _ := cmd.Flags().GetString("url")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if pageURL == "" {
				return fmt.Errorf("--url is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/pages/"+pageURL, nil)
			if err != nil {
				return err
			}

			var page PageSummary
			if err := json.Unmarshal(data, &page); err != nil {
				return fmt.Errorf("parse page: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(page)
			}

			fmt.Printf("URL:          %s\n", page.URL)
			fmt.Printf("Title:        %s\n", page.Title)
			fmt.Printf("Published:    %v\n", page.Published)
			fmt.Printf("Front Page:   %v\n", page.FrontPage)
			if page.EditingRoles != "" {
				fmt.Printf("Editing:      %s\n", page.EditingRoles)
			}
			if page.CreatedAt != "" {
				fmt.Printf("Created:      %s\n", page.CreatedAt)
			}
			if page.UpdatedAt != "" {
				fmt.Printf("Updated:      %s\n", page.UpdatedAt)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("url", "", "Page URL slug (required)")
	return cmd
}

func newPagesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new wiki page",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			title, _ := cmd.Flags().GetString("title")
			if title == "" {
				return fmt.Errorf("--title is required")
			}

			body, _ := cmd.Flags().GetString("body")
			published, _ := cmd.Flags().GetBool("published")
			frontPage, _ := cmd.Flags().GetBool("front-page")
			editingRoles, _ := cmd.Flags().GetString("editing-roles")

			wikiPage := map[string]any{
				"title":     title,
				"published": published,
			}
			if body != "" {
				wikiPage["body"] = body
			}
			if frontPage {
				wikiPage["front_page"] = true
			}
			if editingRoles != "" {
				wikiPage["editing_roles"] = editingRoles
			}

			requestBody := map[string]any{"wiki_page": wikiPage}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create page %q in course %s", title, courseID), requestBody)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/courses/"+courseID+"/pages", requestBody)
			if err != nil {
				return err
			}

			var page PageSummary
			if err := json.Unmarshal(data, &page); err != nil {
				return fmt.Errorf("parse created page: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(page)
			}
			fmt.Printf("Page created: %s — %s\n", page.URL, page.Title)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("title", "", "Page title (required)")
	cmd.Flags().String("body", "", "Page HTML body content")
	cmd.Flags().Bool("published", false, "Publish the page immediately")
	cmd.Flags().Bool("front-page", false, "Set as the course front page")
	cmd.Flags().String("editing-roles", "", "Who can edit: teachers, students, members, public")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newPagesUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing wiki page",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			pageURL, _ := cmd.Flags().GetString("url")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if pageURL == "" {
				return fmt.Errorf("--url is required")
			}

			wikiPage := map[string]any{}
			if cmd.Flags().Changed("title") {
				v, _ := cmd.Flags().GetString("title")
				wikiPage["title"] = v
			}
			if cmd.Flags().Changed("body") {
				v, _ := cmd.Flags().GetString("body")
				wikiPage["body"] = v
			}
			if cmd.Flags().Changed("published") {
				v, _ := cmd.Flags().GetBool("published")
				wikiPage["published"] = v
			}

			requestBody := map[string]any{"wiki_page": wikiPage}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update page %s in course %s", pageURL, courseID), requestBody)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/courses/"+courseID+"/pages/"+pageURL, requestBody)
			if err != nil {
				return err
			}

			var page PageSummary
			if err := json.Unmarshal(data, &page); err != nil {
				return fmt.Errorf("parse updated page: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(page)
			}
			fmt.Printf("Page %s updated.\n", page.URL)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("url", "", "Page URL slug (required)")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("body", "", "New HTML body content")
	cmd.Flags().Bool("published", false, "Set published state")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newPagesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a wiki page",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			pageURL, _ := cmd.Flags().GetString("url")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if pageURL == "" {
				return fmt.Errorf("--url is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the page"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete page %s in course %s", pageURL, courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/courses/"+courseID+"/pages/"+pageURL)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "url": pageURL})
			}
			fmt.Printf("Page %s deleted.\n", pageURL)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("url", "", "Page URL slug (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newPagesRevisionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revisions",
		Short: "List revision history for a wiki page",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			pageURL, _ := cmd.Flags().GetString("url")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if pageURL == "" {
				return fmt.Errorf("--url is required")
			}

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			path := "/courses/" + courseID + "/pages/" + pageURL + "/revisions"
			data, err := client.Get(ctx, path, params)
			if err != nil {
				return err
			}

			var revisions []PageRevisionSummary
			if err := json.Unmarshal(data, &revisions); err != nil {
				return fmt.Errorf("parse page revisions: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(revisions)
			}

			if len(revisions) == 0 {
				fmt.Println("No revisions found.")
				return nil
			}
			for _, r := range revisions {
				latest := ""
				if r.Latest {
					latest = " [latest]"
				}
				fmt.Printf("rev:%-4d  %-30s  %s%s\n", r.RevisionID, r.UpdatedAt, r.EditedBy, latest)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("url", "", "Page URL slug (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of revisions to return")
	return cmd
}
