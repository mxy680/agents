package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newExternalToolsCmd returns the parent "external-tools" command with all subcommands attached.
func newExternalToolsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "external-tools",
		Short:   "Manage Canvas external tools (LTI)",
		Aliases: []string{"lti", "tool"},
	}

	cmd.AddCommand(newExternalToolsListCmd(factory))
	cmd.AddCommand(newExternalToolsGetCmd(factory))
	cmd.AddCommand(newExternalToolsSessionlessLaunchCmd(factory))

	return cmd
}

func newExternalToolsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List external tools for a course",
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

			search, _ := cmd.Flags().GetString("search")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			if search != "" {
				params.Set("search_term", search)
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/external_tools", params)
			if err != nil {
				return err
			}

			var tools []ExternalToolSummary
			if err := json.Unmarshal(data, &tools); err != nil {
				return fmt.Errorf("parse external tools: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(tools)
			}

			if len(tools) == 0 {
				fmt.Println("No external tools found.")
				return nil
			}
			for _, t := range tools {
				fmt.Printf("%-6d  %-12s  %s\n", t.ID, t.PrivacyLevel, truncate(t.Name, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("search", "", "Search term to filter tools")
	cmd.Flags().Int("limit", 0, "Maximum number of tools to return")
	return cmd
}

func newExternalToolsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific external tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			toolID, _ := cmd.Flags().GetString("tool-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if toolID == "" {
				return fmt.Errorf("--tool-id is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/external_tools/"+toolID, nil)
			if err != nil {
				return err
			}

			var tool ExternalToolSummary
			if err := json.Unmarshal(data, &tool); err != nil {
				return fmt.Errorf("parse external tool: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(tool)
			}

			fmt.Printf("ID:            %d\n", tool.ID)
			fmt.Printf("Name:          %s\n", tool.Name)
			if tool.URL != "" {
				fmt.Printf("URL:           %s\n", tool.URL)
			}
			if tool.Domain != "" {
				fmt.Printf("Domain:        %s\n", tool.Domain)
			}
			if tool.PrivacyLevel != "" {
				fmt.Printf("Privacy Level: %s\n", tool.PrivacyLevel)
			}
			if tool.Description != "" {
				fmt.Printf("Description:   %s\n", truncate(tool.Description, 80))
			}
			if tool.CreatedAt != "" {
				fmt.Printf("Created:       %s\n", tool.CreatedAt)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("tool-id", "", "Canvas external tool ID (required)")
	return cmd
}

func newExternalToolsSessionlessLaunchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessionless-launch",
		Short: "Get a sessionless launch URL for an external tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			toolID, _ := cmd.Flags().GetString("tool-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if toolID == "" {
				return fmt.Errorf("--tool-id is required")
			}

			params := url.Values{}
			params.Set("id", toolID)

			data, err := client.Get(ctx, "/courses/"+courseID+"/external_tools/sessionless_launch", params)
			if err != nil {
				return err
			}

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("parse sessionless launch: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}

			if launchURL, ok := result["url"].(string); ok {
				fmt.Printf("Launch URL: %s\n", launchURL)
			} else {
				fmt.Println("Sessionless launch URL retrieved.")
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("tool-id", "", "Canvas external tool ID (required)")
	return cmd
}
