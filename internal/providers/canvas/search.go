package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newSearchCmd returns the parent "search" command with all subcommands attached.
func newSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Search Canvas for recipients, courses, and other content",
		Aliases: []string{"find"},
	}

	cmd.AddCommand(newSearchRecipientsCmd(factory))
	cmd.AddCommand(newSearchCoursesCmd(factory))
	cmd.AddCommand(newSearchAllCmd(factory))

	return cmd
}

func newSearchRecipientsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recipients",
		Short: "Search for message recipients (users and contexts)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			search, _ := cmd.Flags().GetString("search")
			if search == "" {
				return fmt.Errorf("--search is required")
			}

			context, _ := cmd.Flags().GetString("context")
			searchType, _ := cmd.Flags().GetString("type")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			params.Set("search", search)
			if context != "" {
				params.Set("context", context)
			}
			if searchType != "" {
				params.Set("type", searchType)
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/search/recipients", params)
			if err != nil {
				return err
			}

			var results []map[string]any
			if err := json.Unmarshal(data, &results); err != nil {
				return fmt.Errorf("parse search results: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(results)
			}

			if len(results) == 0 {
				fmt.Println("No recipients found.")
				return nil
			}
			for _, r := range results {
				id, _ := r["id"]
				name, _ := r["name"].(string)
				resultType, _ := r["type"].(string)
				fmt.Printf("%-20v  %-10s  %s\n", id, resultType, truncate(name, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("search", "", "Search query (required)")
	cmd.Flags().String("context", "", "Limit search to a context (e.g. course_123)")
	cmd.Flags().String("type", "", "Filter by type: user or context")
	cmd.Flags().Int("limit", 0, "Maximum number of results to return")
	return cmd
}

func newSearchCoursesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "courses",
		Short: "Search all available courses",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			search, _ := cmd.Flags().GetString("search")
			if search == "" {
				return fmt.Errorf("--search is required")
			}

			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			params.Set("search", search)
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/search/all_courses", params)
			if err != nil {
				return err
			}

			var results []map[string]any
			if err := json.Unmarshal(data, &results); err != nil {
				return fmt.Errorf("parse course search results: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(results)
			}

			if len(results) == 0 {
				fmt.Println("No courses found.")
				return nil
			}
			for _, r := range results {
				id, _ := r["id"]
				name, _ := r["name"].(string)
				courseCode, _ := r["course_code"].(string)
				fmt.Printf("%-6v  %-12s  %s\n", id, courseCode, truncate(name, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("search", "", "Search query (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of results to return")
	return cmd
}

func newSearchAllCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Search across all Canvas content (falls back to recipients if unavailable)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			search, _ := cmd.Flags().GetString("search")
			if search == "" {
				return fmt.Errorf("--search is required")
			}

			context, _ := cmd.Flags().GetString("context")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			params.Set("search", search)
			if context != "" {
				params.Set("context", context)
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			// /search/all may not exist on all Canvas instances; fall back to /search/recipients.
			data, err := client.Get(ctx, "/search/all", params)
			if err != nil {
				data, err = client.Get(ctx, "/search/recipients", params)
				if err != nil {
					return err
				}
			}

			var results []map[string]any
			if err := json.Unmarshal(data, &results); err != nil {
				return fmt.Errorf("parse search results: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(results)
			}

			if len(results) == 0 {
				fmt.Println("No results found.")
				return nil
			}
			for _, r := range results {
				id, _ := r["id"]
				name, _ := r["name"].(string)
				resultType, _ := r["type"].(string)
				fmt.Printf("%-20v  %-10s  %s\n", id, resultType, truncate(name, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("search", "", "Search query (required)")
	cmd.Flags().String("context", "", "Limit search to a context (e.g. course_123)")
	cmd.Flags().Int("limit", 0, "Maximum number of results to return")
	return cmd
}
