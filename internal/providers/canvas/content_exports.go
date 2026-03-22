package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newContentExportsCmd returns the parent "content-exports" command with all subcommands attached.
func newContentExportsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "content-exports",
		Short:   "Manage Canvas content exports",
		Aliases: []string{"export"},
	}

	cmd.AddCommand(newContentExportsListCmd(factory))
	cmd.AddCommand(newContentExportsGetCmd(factory))
	cmd.AddCommand(newContentExportsCreateCmd(factory))

	return cmd
}

func newContentExportsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List content exports for a course",
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

			data, err := client.Get(ctx, "/courses/"+courseID+"/content_exports", params)
			if err != nil {
				return err
			}

			var exports []map[string]any
			if err := json.Unmarshal(data, &exports); err != nil {
				return fmt.Errorf("parse content exports: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(exports)
			}

			if len(exports) == 0 {
				fmt.Println("No content exports found.")
				return nil
			}
			for _, e := range exports {
				id, _ := e["id"]
				exportType, _ := e["export_type"]
				workflowState, _ := e["workflow_state"]
				fmt.Printf("%-6v  %-20v  %v\n", id, exportType, workflowState)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of exports to return")
	return cmd
}

func newContentExportsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific content export",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			exportID, _ := cmd.Flags().GetString("export-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if exportID == "" {
				return fmt.Errorf("--export-id is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/content_exports/"+exportID, nil)
			if err != nil {
				return err
			}

			var export map[string]any
			if err := json.Unmarshal(data, &export); err != nil {
				return fmt.Errorf("parse content export: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(export)
			}

			id, _ := export["id"]
			exportType, _ := export["export_type"]
			workflowState, _ := export["workflow_state"]
			fmt.Printf("ID:             %v\n", id)
			fmt.Printf("Type:           %v\n", exportType)
			fmt.Printf("Workflow State: %v\n", workflowState)
			if downloadURL, ok := export["attachment"].(map[string]any); ok {
				if u, ok := downloadURL["url"].(string); ok && u != "" {
					fmt.Printf("Download URL:   %s\n", u)
				}
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("export-id", "", "Canvas content export ID (required)")
	return cmd
}

func newContentExportsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new content export for a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			exportType, _ := cmd.Flags().GetString("type")
			if exportType == "" {
				return fmt.Errorf("--type is required")
			}

			body := map[string]any{
				"export_type": exportType,
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create content export type %q for course %s", exportType, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/courses/"+courseID+"/content_exports", body)
			if err != nil {
				return err
			}

			var export map[string]any
			if err := json.Unmarshal(data, &export); err != nil {
				return fmt.Errorf("parse created content export: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(export)
			}
			id, _ := export["id"]
			workflowState, _ := export["workflow_state"]
			fmt.Printf("Content export created: id=%v  state=%v\n", id, workflowState)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("type", "", "Export type: common_cartridge, qti, or zip (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}
