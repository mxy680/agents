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
