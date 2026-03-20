package framer

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newManagedCollectionsListCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List managed CMS collections",
		RunE:  makeRunManagedCollectionsList(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunManagedCollectionsList(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getManagedCollections", nil)
		if err != nil {
			return fmt.Errorf("list managed collections: %w", err)
		}

		var collections []CollectionSummary
		if err := json.Unmarshal(result, &collections); err != nil {
			return fmt.Errorf("parse collections: %w", err)
		}

		lines := make([]string, 0, len(collections))
		for _, c := range collections {
			lines = append(lines, fmt.Sprintf("%s  %s", c.ID, c.Name))
		}
		return cli.PrintResult(cmd, collections, lines)
	}
}

func newManagedCollectionsCreateCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new managed CMS collection",
		RunE:  makeRunManagedCollectionsCreate(factory),
	}
	cmd.Flags().String("name", "", "Collection name (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without making changes")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunManagedCollectionsCreate(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		if ok, err := dryRunResult(cmd, fmt.Sprintf("create managed collection %q", name), map[string]any{"name": name}); ok {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createManagedCollection", map[string]any{"name": name})
		if err != nil {
			return fmt.Errorf("create managed collection: %w", err)
		}

		var collection CollectionSummary
		if err := json.Unmarshal(result, &collection); err != nil {
			return fmt.Errorf("parse collection: %w", err)
		}

		lines := []string{fmt.Sprintf("Created managed collection %s  %s", collection.ID, collection.Name)}
		return cli.PrintResult(cmd, collection, lines)
	}
}

func newManagedCollectionsFieldsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fields",
		Short: "List fields of a managed CMS collection",
		RunE:  makeRunManagedCollectionsFields(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunManagedCollectionsFields(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getManagedCollectionFields", map[string]any{"id": id})
		if err != nil {
			return fmt.Errorf("get managed collection fields: %w", err)
		}

		var fields []Field
		if err := json.Unmarshal(result, &fields); err != nil {
			return fmt.Errorf("parse fields: %w", err)
		}

		lines := make([]string, 0, len(fields))
		for _, f := range fields {
			lines = append(lines, fmt.Sprintf("%s  %-20s  %s", f.ID, f.Name, f.Type))
		}
		return cli.PrintResult(cmd, fields, lines)
	}
}

func newManagedCollectionsSetFieldsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-fields",
		Short: "Set the fields of a managed CMS collection",
		RunE:  makeRunManagedCollectionsSetFields(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().String("fields", "", "Fields as JSON array")
	cmd.Flags().String("fields-file", "", "Path to JSON file containing fields")
	cmd.Flags().Bool("dry-run", false, "Preview without making changes")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunManagedCollectionsSetFields(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		fieldsStr, _ := cmd.Flags().GetString("fields")
		fieldsFile, _ := cmd.Flags().GetString("fields-file")

		if fieldsStr == "" && fieldsFile == "" {
			return fmt.Errorf("--fields or --fields-file is required")
		}

		fieldsJSON, err := parseJSONFlagOrFile(fieldsStr, fieldsFile)
		if err != nil {
			return fmt.Errorf("parse fields: %w", err)
		}

		if ok, err := dryRunResult(cmd, fmt.Sprintf("set fields on managed collection %s", id), map[string]any{"id": id, "fields": fieldsJSON}); ok {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("setManagedCollectionFields", map[string]any{
			"id":     id,
			"fields": json.RawMessage(fieldsJSON),
		})
		if err != nil {
			return fmt.Errorf("set managed collection fields: %w", err)
		}

		var fields []Field
		if err := json.Unmarshal(result, &fields); err != nil {
			return fmt.Errorf("parse fields: %w", err)
		}

		lines := make([]string, 0, len(fields))
		for _, f := range fields {
			lines = append(lines, fmt.Sprintf("%s  %-20s  %s", f.ID, f.Name, f.Type))
		}
		return cli.PrintResult(cmd, fields, lines)
	}
}

func newManagedCollectionsItemsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items",
		Short: "List item IDs in a managed CMS collection",
		RunE:  makeRunManagedCollectionsItems(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunManagedCollectionsItems(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getManagedCollectionItemIds", map[string]any{"id": id})
		if err != nil {
			return fmt.Errorf("get managed collection items: %w", err)
		}

		var itemIDs []string
		if err := json.Unmarshal(result, &itemIDs); err != nil {
			return fmt.Errorf("parse item IDs: %w", err)
		}

		lines := make([]string, 0, len(itemIDs))
		for _, itemID := range itemIDs {
			lines = append(lines, itemID)
		}
		return cli.PrintResult(cmd, itemIDs, lines)
	}
}

func newManagedCollectionsAddItemsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-items",
		Short: "Add items to a managed CMS collection",
		RunE:  makeRunManagedCollectionsAddItems(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().String("items", "", "Items as JSON array")
	cmd.Flags().String("items-file", "", "Path to JSON file containing items")
	cmd.Flags().Bool("dry-run", false, "Preview without making changes")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunManagedCollectionsAddItems(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		itemsStr, _ := cmd.Flags().GetString("items")
		itemsFile, _ := cmd.Flags().GetString("items-file")

		if itemsStr == "" && itemsFile == "" {
			return fmt.Errorf("--items or --items-file is required")
		}

		itemsJSON, err := parseJSONFlagOrFile(itemsStr, itemsFile)
		if err != nil {
			return fmt.Errorf("parse items: %w", err)
		}

		if ok, err := dryRunResult(cmd, fmt.Sprintf("add items to managed collection %s", id), map[string]any{"id": id, "items": itemsJSON}); ok {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("addManagedCollectionItems", map[string]any{
			"id":    id,
			"items": json.RawMessage(itemsJSON),
		})
		if err != nil {
			return fmt.Errorf("add managed collection items: %w", err)
		}

		var items []CollectionItem
		if err := json.Unmarshal(result, &items); err != nil {
			return fmt.Errorf("parse items: %w", err)
		}

		lines := make([]string, 0, len(items))
		for _, item := range items {
			lines = append(lines, fmt.Sprintf("%s  %s", item.ID, item.Slug))
		}
		return cli.PrintResult(cmd, items, lines)
	}
}

func newManagedCollectionsRemoveItemsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-items",
		Short: "Remove items from a managed CMS collection",
		RunE:  makeRunManagedCollectionsRemoveItems(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().String("item-ids", "", "Comma-separated item IDs to remove (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive operation")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("item-ids")
	return cmd
}

func makeRunManagedCollectionsRemoveItems(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		itemIDsStr, _ := cmd.Flags().GetString("item-ids")

		itemIDs := parseStringList(itemIDsStr)
		if len(itemIDs) == 0 {
			return fmt.Errorf("no item IDs provided")
		}

		if !confirmDestructive(cmd, fmt.Sprintf("remove %d item(s) from managed collection %s", len(itemIDs), id)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("removeManagedCollectionItems", map[string]any{
			"id":      id,
			"itemIds": itemIDs,
		})
		if err != nil {
			return fmt.Errorf("remove managed collection items: %w", err)
		}

		var raw any
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse result: %w", err)
		}

		lines := []string{fmt.Sprintf("Removed %d item(s) from managed collection %s", len(itemIDs), id)}
		return cli.PrintResult(cmd, raw, lines)
	}
}

func newManagedCollectionsSetItemOrderCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-item-order",
		Short: "Set the order of items in a managed CMS collection",
		RunE:  makeRunManagedCollectionsSetItemOrder(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().String("item-ids", "", "Comma-separated item IDs in desired order (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("item-ids")
	return cmd
}

func makeRunManagedCollectionsSetItemOrder(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		itemIDsStr, _ := cmd.Flags().GetString("item-ids")

		itemIDs := parseStringList(itemIDsStr)
		if len(itemIDs) == 0 {
			return fmt.Errorf("no item IDs provided")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("setManagedCollectionItemOrder", map[string]any{
			"id":      id,
			"itemIds": itemIDs,
		})
		if err != nil {
			return fmt.Errorf("set managed collection item order: %w", err)
		}

		var raw any
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse result: %w", err)
		}

		lines := []string{fmt.Sprintf("Updated item order for managed collection %s", id)}
		return cli.PrintResult(cmd, raw, lines)
	}
}
