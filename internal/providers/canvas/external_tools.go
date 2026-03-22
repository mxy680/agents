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
	cmd.AddCommand(newExternalToolsCreateCmd(factory))
	cmd.AddCommand(newExternalToolsUpdateCmd(factory))
	cmd.AddCommand(newExternalToolsDeleteCmd(factory))
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

func newExternalToolsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new external tool in a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			toolURL, _ := cmd.Flags().GetString("url")
			if toolURL == "" {
				return fmt.Errorf("--url is required")
			}
			consumerKey, _ := cmd.Flags().GetString("consumer-key")
			if consumerKey == "" {
				return fmt.Errorf("--consumer-key is required")
			}
			sharedSecret, _ := cmd.Flags().GetString("shared-secret")
			if sharedSecret == "" {
				return fmt.Errorf("--shared-secret is required")
			}

			privacyLevel, _ := cmd.Flags().GetString("privacy-level")

			body := map[string]any{
				"name":          name,
				"url":           toolURL,
				"consumer_key":  consumerKey,
				"shared_secret": sharedSecret,
			}
			if privacyLevel != "" {
				body["privacy_level"] = privacyLevel
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create external tool %q in course %s", name, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/courses/"+courseID+"/external_tools", body)
			if err != nil {
				return err
			}

			var tool ExternalToolSummary
			if err := json.Unmarshal(data, &tool); err != nil {
				return fmt.Errorf("parse created external tool: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(tool)
			}
			fmt.Printf("External tool created: %d — %s\n", tool.ID, tool.Name)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("name", "", "Tool name (required)")
	cmd.Flags().String("url", "", "Tool launch URL (required)")
	cmd.Flags().String("consumer-key", "", "LTI consumer key (required)")
	cmd.Flags().String("shared-secret", "", "LTI shared secret (required)")
	cmd.Flags().String("privacy-level", "", "Privacy level: anonymous, name_only, or public")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newExternalToolsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing external tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			toolID, _ := cmd.Flags().GetString("tool-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if toolID == "" {
				return fmt.Errorf("--tool-id is required")
			}

			body := map[string]any{}
			if cmd.Flags().Changed("name") {
				v, _ := cmd.Flags().GetString("name")
				body["name"] = v
			}
			if cmd.Flags().Changed("url") {
				v, _ := cmd.Flags().GetString("url")
				body["url"] = v
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update external tool %s in course %s", toolID, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/courses/"+courseID+"/external_tools/"+toolID, body)
			if err != nil {
				return err
			}

			var tool ExternalToolSummary
			if err := json.Unmarshal(data, &tool); err != nil {
				return fmt.Errorf("parse updated external tool: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(tool)
			}
			fmt.Printf("External tool %d updated.\n", tool.ID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("tool-id", "", "Canvas external tool ID (required)")
	cmd.Flags().String("name", "", "New tool name")
	cmd.Flags().String("url", "", "New tool launch URL")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newExternalToolsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an external tool from a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			toolID, _ := cmd.Flags().GetString("tool-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if toolID == "" {
				return fmt.Errorf("--tool-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the external tool"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete external tool %s from course %s", toolID, courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/courses/"+courseID+"/external_tools/"+toolID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "tool_id": toolID})
			}
			fmt.Printf("External tool %s deleted.\n", toolID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("tool-id", "", "Canvas external tool ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
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
