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
	cmd.AddCommand(newModulesCreateCmd(factory))
	cmd.AddCommand(newModulesUpdateCmd(factory))
	cmd.AddCommand(newModulesDeleteCmd(factory))
	cmd.AddCommand(newModulesAddItemCmd(factory))
	cmd.AddCommand(newModulesRemoveItemCmd(factory))

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

func newModulesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new module in a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			name, _ := cmd.Flags().GetString("name")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "create module: "+name, map[string]any{"course_id": courseID, "name": name})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			body := map[string]any{"name": name}
			if position, _ := cmd.Flags().GetInt("position"); position > 0 {
				body["position"] = position
			}

			data, err := client.Post(ctx, "/courses/"+courseID+"/modules", body)
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
			fmt.Printf("Module %d created: %s\n", module.ID, module.Name)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("name", "", "Module name (required)")
	cmd.Flags().Int("position", 0, "Module position")
	return cmd
}

func newModulesUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a module",
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

			body := map[string]any{}
			if name, _ := cmd.Flags().GetString("name"); name != "" {
				body["name"] = name
			}

			data, err := client.Put(ctx, "/courses/"+courseID+"/modules/"+moduleID, body)
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
			fmt.Printf("Module %s updated\n", moduleID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("module-id", "", "Canvas module ID (required)")
	cmd.Flags().String("name", "", "New name")
	return cmd
}

func newModulesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a module",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			moduleID, _ := cmd.Flags().GetString("module-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if moduleID == "" {
				return fmt.Errorf("--module-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the module"); err != nil {
				return err
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			if _, err := client.Delete(ctx, "/courses/"+courseID+"/modules/"+moduleID); err != nil {
				return err
			}

			fmt.Printf("Module %s deleted\n", moduleID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("module-id", "", "Canvas module ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
	return cmd
}

func newModulesAddItemCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-item",
		Short: "Add an item to a module",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			moduleID, _ := cmd.Flags().GetString("module-id")
			itemType, _ := cmd.Flags().GetString("type")
			contentID, _ := cmd.Flags().GetString("content-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if moduleID == "" {
				return fmt.Errorf("--module-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "add item (type: "+itemType+")", map[string]any{
					"course_id": courseID, "module_id": moduleID, "type": itemType, "content_id": contentID,
				})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			body := map[string]any{"type": itemType}
			if contentID != "" {
				body["content_id"] = contentID
			}

			path := "/courses/" + courseID + "/modules/" + moduleID + "/items"
			data, err := client.Post(ctx, path, body)
			if err != nil {
				return err
			}

			var item ModuleItemSummary
			if err := json.Unmarshal(data, &item); err != nil {
				return fmt.Errorf("parse module item: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(item)
			}
			fmt.Printf("Item %d added to module %s: %s\n", item.ID, moduleID, item.Title)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("module-id", "", "Canvas module ID (required)")
	cmd.Flags().String("type", "", "Item type (e.g. Assignment, Page, File)")
	cmd.Flags().String("content-id", "", "Content ID to add")
	return cmd
}

func newModulesRemoveItemCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-item",
		Short: "Remove an item from a module",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			moduleID, _ := cmd.Flags().GetString("module-id")
			itemID, _ := cmd.Flags().GetString("item-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if moduleID == "" {
				return fmt.Errorf("--module-id is required")
			}
			if itemID == "" {
				return fmt.Errorf("--item-id is required")
			}

			if err := confirmDestructive(cmd, "this will remove the item from the module"); err != nil {
				return err
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			path := "/courses/" + courseID + "/modules/" + moduleID + "/items/" + itemID
			if _, err := client.Delete(ctx, path); err != nil {
				return err
			}

			fmt.Printf("Item %s removed from module %s\n", itemID, moduleID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("module-id", "", "Canvas module ID (required)")
	cmd.Flags().String("item-id", "", "Canvas module item ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
	return cmd
}
