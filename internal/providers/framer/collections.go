package framer

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newCollectionsListCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List CMS collections",
		RunE:  makeRunCollectionsList(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunCollectionsList(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getCollections", nil)
		if err != nil {
			return fmt.Errorf("list collections: %w", err)
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

func newCollectionsGetCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a CMS collection by ID",
		RunE:  makeRunCollectionsGet(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunCollectionsGet(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getCollection", map[string]any{"id": id})
		if err != nil {
			return fmt.Errorf("get collection: %w", err)
		}

		var collection CollectionSummary
		if err := json.Unmarshal(result, &collection); err != nil {
			return fmt.Errorf("parse collection: %w", err)
		}

		lines := []string{fmt.Sprintf("%s  %s", collection.ID, collection.Name)}
		return cli.PrintResult(cmd, collection, lines)
	}
}

func newCollectionsCreateCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new CMS collection",
		RunE:  makeRunCollectionsCreate(factory),
	}
	cmd.Flags().String("name", "", "Collection name (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without making changes")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunCollectionsCreate(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		if ok, err := dryRunResult(cmd, fmt.Sprintf("create collection %q", name), map[string]any{"name": name}); ok {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createCollection", map[string]any{"name": name})
		if err != nil {
			return fmt.Errorf("create collection: %w", err)
		}

		var collection CollectionSummary
		if err := json.Unmarshal(result, &collection); err != nil {
			return fmt.Errorf("parse collection: %w", err)
		}

		lines := []string{fmt.Sprintf("Created collection %s  %s", collection.ID, collection.Name)}
		return cli.PrintResult(cmd, collection, lines)
	}
}

func newCollectionsFieldsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fields",
		Short: "List fields of a CMS collection",
		RunE:  makeRunCollectionsFields(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunCollectionsFields(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getCollectionFields", map[string]any{"id": id})
		if err != nil {
			return fmt.Errorf("get collection fields: %w", err)
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

func newCollectionsAddFieldsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-fields",
		Short: "Add fields to a CMS collection",
		RunE:  makeRunCollectionsAddFields(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().String("fields", "", "Fields as JSON array (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without making changes")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("fields")
	return cmd
}

func makeRunCollectionsAddFields(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		fieldsStr, _ := cmd.Flags().GetString("fields")

		fieldsJSON, err := parseJSONFlag(fieldsStr)
		if err != nil {
			return fmt.Errorf("parse fields: %w", err)
		}

		if ok, err := dryRunResult(cmd, fmt.Sprintf("add fields to collection %s", id), map[string]any{"id": id, "fields": fieldsJSON}); ok {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("addCollectionFields", map[string]any{
			"id":     id,
			"fields": json.RawMessage(fieldsJSON),
		})
		if err != nil {
			return fmt.Errorf("add collection fields: %w", err)
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

func newCollectionsRemoveFieldsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-fields",
		Short: "Remove fields from a CMS collection",
		RunE:  makeRunCollectionsRemoveFields(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().String("field-ids", "", "Comma-separated field IDs to remove (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive operation")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("field-ids")
	return cmd
}

func makeRunCollectionsRemoveFields(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		fieldIDsStr, _ := cmd.Flags().GetString("field-ids")

		fieldIDs := parseStringList(fieldIDsStr)
		if len(fieldIDs) == 0 {
			return fmt.Errorf("no field IDs provided")
		}

		if !confirmDestructive(cmd, fmt.Sprintf("remove %d field(s) from collection %s", len(fieldIDs), id)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("removeCollectionFields", map[string]any{
			"id":       id,
			"fieldIds": fieldIDs,
		})
		if err != nil {
			return fmt.Errorf("remove collection fields: %w", err)
		}

		var raw any
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse result: %w", err)
		}

		lines := []string{fmt.Sprintf("Removed %d field(s) from collection %s", len(fieldIDs), id)}
		return cli.PrintResult(cmd, raw, lines)
	}
}

func newCollectionsSetFieldOrderCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-field-order",
		Short: "Set the order of fields in a CMS collection",
		RunE:  makeRunCollectionsSetFieldOrder(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().String("field-ids", "", "Comma-separated field IDs in desired order (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("field-ids")
	return cmd
}

func makeRunCollectionsSetFieldOrder(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		fieldIDsStr, _ := cmd.Flags().GetString("field-ids")

		fieldIDs := parseStringList(fieldIDsStr)
		if len(fieldIDs) == 0 {
			return fmt.Errorf("no field IDs provided")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("setCollectionFieldOrder", map[string]any{
			"id":       id,
			"fieldIds": fieldIDs,
		})
		if err != nil {
			return fmt.Errorf("set collection field order: %w", err)
		}

		var raw any
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse result: %w", err)
		}

		lines := []string{fmt.Sprintf("Updated field order for collection %s", id)}
		return cli.PrintResult(cmd, raw, lines)
	}
}

func newCollectionsItemsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items",
		Short: "List items in a CMS collection",
		RunE:  makeRunCollectionsItems(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunCollectionsItems(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getCollectionItems", map[string]any{"id": id})
		if err != nil {
			return fmt.Errorf("get collection items: %w", err)
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

func newCollectionsAddItemsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-items",
		Short: "Add items to a CMS collection",
		RunE:  makeRunCollectionsAddItems(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().String("items", "", "Items as JSON array")
	cmd.Flags().String("items-file", "", "Path to JSON file containing items")
	cmd.Flags().Bool("dry-run", false, "Preview without making changes")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunCollectionsAddItems(factory BridgeClientFactory) func(*cobra.Command, []string) error {
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

		if ok, err := dryRunResult(cmd, fmt.Sprintf("add items to collection %s", id), map[string]any{"id": id, "items": itemsJSON}); ok {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("addCollectionItems", map[string]any{
			"id":    id,
			"items": json.RawMessage(itemsJSON),
		})
		if err != nil {
			return fmt.Errorf("add collection items: %w", err)
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

func newCollectionsRemoveItemsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-items",
		Short: "Remove items from a CMS collection",
		RunE:  makeRunCollectionsRemoveItems(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().String("item-ids", "", "Comma-separated item IDs to remove (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive operation")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("item-ids")
	return cmd
}

func makeRunCollectionsRemoveItems(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		itemIDsStr, _ := cmd.Flags().GetString("item-ids")

		itemIDs := parseStringList(itemIDsStr)
		if len(itemIDs) == 0 {
			return fmt.Errorf("no item IDs provided")
		}

		if !confirmDestructive(cmd, fmt.Sprintf("remove %d item(s) from collection %s", len(itemIDs), id)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("removeCollectionItems", map[string]any{
			"id":      id,
			"itemIds": itemIDs,
		})
		if err != nil {
			return fmt.Errorf("remove collection items: %w", err)
		}

		var raw any
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse result: %w", err)
		}

		lines := []string{fmt.Sprintf("Removed %d item(s) from collection %s", len(itemIDs), id)}
		return cli.PrintResult(cmd, raw, lines)
	}
}

func newCollectionsSetItemOrderCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-item-order",
		Short: "Set the order of items in a CMS collection",
		RunE:  makeRunCollectionsSetItemOrder(factory),
	}
	cmd.Flags().String("id", "", "Collection ID (required)")
	cmd.Flags().String("item-ids", "", "Comma-separated item IDs in desired order (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("item-ids")
	return cmd
}

func makeRunCollectionsSetItemOrder(factory BridgeClientFactory) func(*cobra.Command, []string) error {
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

		result, err := client.Call("setCollectionItemOrder", map[string]any{
			"id":      id,
			"itemIds": itemIDs,
		})
		if err != nil {
			return fmt.Errorf("set collection item order: %w", err)
		}

		var raw any
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse result: %w", err)
		}

		lines := []string{fmt.Sprintf("Updated item order for collection %s", id)}
		return cli.PrintResult(cmd, raw, lines)
	}
}
