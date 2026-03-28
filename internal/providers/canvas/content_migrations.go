package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newContentMigrationsCmd returns the parent "content-migrations" command with all subcommands attached.
func newContentMigrationsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "content-migrations",
		Short:   "Manage Canvas content migrations",
		Aliases: []string{"migration", "migrate"},
	}

	cmd.AddCommand(newContentMigrationsListCmd(factory))
	cmd.AddCommand(newContentMigrationsGetCmd(factory))
	cmd.AddCommand(newContentMigrationsProgressCmd(factory))
	cmd.AddCommand(newContentMigrationsContentListCmd(factory))
	cmd.AddCommand(newContentMigrationsCreateCmd(factory))

	return cmd
}

func newContentMigrationsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new content migration for a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			migrationType, _ := cmd.Flags().GetString("type")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if migrationType == "" {
				return fmt.Errorf("--type is required")
			}

			body := map[string]any{"migration_type": migrationType}
			data, err := client.Post(ctx, "/courses/"+courseID+"/content_migrations", body)
			if err != nil {
				return err
			}

			var migration map[string]any
			if err := json.Unmarshal(data, &migration); err != nil {
				return fmt.Errorf("parse content migration: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(migration)
			}

			id, _ := migration["id"]
			workflowState, _ := migration["workflow_state"]
			fmt.Printf("Content migration %v created (type: %s, state: %v)\n", id, migrationType, workflowState)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("type", "", "Migration type (e.g. course_copy_importer) (required)")
	return cmd
}

func newContentMigrationsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List content migrations for a course",
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

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/content_migrations", params)
			if err != nil {
				return err
			}

			var migrations []map[string]any
			if err := json.Unmarshal(data, &migrations); err != nil {
				return fmt.Errorf("parse content migrations: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(migrations)
			}

			if len(migrations) == 0 {
				fmt.Println("No content migrations found.")
				return nil
			}
			for _, m := range migrations {
				id, _ := m["id"]
				migrationType, _ := m["migration_type"]
				workflowState, _ := m["workflow_state"]
				fmt.Printf("%-6v  %-30v  %v\n", id, migrationType, workflowState)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of migrations to return")
	return cmd
}

func newContentMigrationsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific content migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			migrationID, _ := cmd.Flags().GetString("migration-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if migrationID == "" {
				return fmt.Errorf("--migration-id is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/content_migrations/"+migrationID, nil)
			if err != nil {
				return err
			}

			var migration map[string]any
			if err := json.Unmarshal(data, &migration); err != nil {
				return fmt.Errorf("parse content migration: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(migration)
			}

			id, _ := migration["id"]
			migrationType, _ := migration["migration_type"]
			workflowState, _ := migration["workflow_state"]
			fmt.Printf("ID:             %v\n", id)
			fmt.Printf("Type:           %v\n", migrationType)
			fmt.Printf("Workflow State: %v\n", workflowState)
			if progressURL, ok := migration["progress_url"].(string); ok && progressURL != "" {
				fmt.Printf("Progress URL:   %s\n", progressURL)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("migration-id", "", "Canvas content migration ID (required)")
	return cmd
}

func newContentMigrationsProgressCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "progress",
		Short: "Get progress for a content migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			migrationID, _ := cmd.Flags().GetString("migration-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if migrationID == "" {
				return fmt.Errorf("--migration-id is required")
			}

			// Fetch the migration to get workflow_state and progress information.
			data, err := client.Get(ctx, "/courses/"+courseID+"/content_migrations/"+migrationID, nil)
			if err != nil {
				return err
			}

			var migration map[string]any
			if err := json.Unmarshal(data, &migration); err != nil {
				return fmt.Errorf("parse content migration: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(migration)
			}

			id, _ := migration["id"]
			workflowState, _ := migration["workflow_state"]
			fmt.Printf("Migration ID:   %v\n", id)
			fmt.Printf("State:          %v\n", workflowState)
			if progressURL, ok := migration["progress_url"].(string); ok && progressURL != "" {
				fmt.Printf("Progress URL:   %s\n", progressURL)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("migration-id", "", "Canvas content migration ID (required)")
	return cmd
}

func newContentMigrationsContentListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "content-list",
		Short: "List content items available in a content migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			migrationID, _ := cmd.Flags().GetString("migration-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if migrationID == "" {
				return fmt.Errorf("--migration-id is required")
			}

			path := "/courses/" + courseID + "/content_migrations/" + migrationID + "/content_list"
			data, err := client.Get(ctx, path, nil)
			if err != nil {
				return err
			}

			var items []map[string]any
			if err := json.Unmarshal(data, &items); err != nil {
				return fmt.Errorf("parse content list: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(items)
			}

			if len(items) == 0 {
				fmt.Println("No content items found.")
				return nil
			}
			for _, item := range items {
				itemType, _ := item["type"]
				title, _ := item["title"]
				fmt.Printf("%-20v  %v\n", itemType, title)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("migration-id", "", "Canvas content migration ID (required)")
	return cmd
}
