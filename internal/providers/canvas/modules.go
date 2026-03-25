package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newModulesCmd returns the parent "modules" command with all subcommands attached.
func newModulesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "modules",
		Short:   "Manage Canvas course modules",
		Aliases: []string{"mod"},
	}

	cmd.AddCommand(newModulesListCmd(factory))
	cmd.AddCommand(newModulesGetCmd(factory))
	cmd.AddCommand(newModulesItemsCmd(factory))

	return cmd
}

func newModulesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List modules for a course",
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

			data, err := client.Get(ctx, "/courses/"+courseID+"/modules", params)
			if err != nil {
				return err
			}

			var modules []ModuleSummary
			if err := json.Unmarshal(data, &modules); err != nil {
				return fmt.Errorf("parse modules: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(modules)
			}

			if len(modules) == 0 {
				fmt.Println("No modules found.")
				return nil
			}
			for _, m := range modules {
				published := " "
				if m.Published {
					published = "✓"
				}
				fmt.Printf("%-6d  [%s]  %-4d items  %s\n", m.ID, published, m.ItemsCount, m.Name)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("search", "", "Search modules by name")
	cmd.Flags().Int("limit", 0, "Maximum number of modules to return")
	return cmd
}

func newModulesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific module",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			moduleID, _ := cmd.Flags().GetString("module-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if moduleID == "" {
				return fmt.Errorf("--module-id is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/modules/"+moduleID, nil)
			if err != nil {
				return err
			}

			var module ModuleSummary
			if err := json.Unmarshal(data, &module); err != nil {
				return fmt.Errorf("parse module: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(module)
			}

			fmt.Printf("ID:         %d\n", module.ID)
			fmt.Printf("Name:       %s\n", module.Name)
			fmt.Printf("Position:   %d\n", module.Position)
			fmt.Printf("Items:      %d\n", module.ItemsCount)
			fmt.Printf("Published:  %v\n", module.Published)
			if module.UnlockAt != "" {
				fmt.Printf("Unlock At:  %s\n", module.UnlockAt)
			}
			if module.State != "" {
				fmt.Printf("State:      %s\n", module.State)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("module-id", "", "Canvas module ID (required)")
	return cmd
}

func newModulesItemsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items",
		Short: "List items in a module",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			moduleID, _ := cmd.Flags().GetString("module-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if moduleID == "" {
				return fmt.Errorf("--module-id is required")
			}

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			path := "/courses/" + courseID + "/modules/" + moduleID + "/items"
			data, err := client.Get(ctx, path, params)
			if err != nil {
				return err
			}

			var items []ModuleItemSummary
			if err := json.Unmarshal(data, &items); err != nil {
				return fmt.Errorf("parse module items: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(items)
			}

			if len(items) == 0 {
				fmt.Println("No items found.")
				return nil
			}
			for _, item := range items {
				fmt.Printf("%-6d  %-20s  %s\n", item.ID, item.Type, truncate(item.Title, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("module-id", "", "Canvas module ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of items to return")
	return cmd
}
