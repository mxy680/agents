package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newKVNamespacesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespaces-list",
		Short: "List KV namespaces",
		RunE:  makeRunKVNamespacesList(factory),
	}
	return cmd
}

func makeRunKVNamespacesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath("/storage/kv/namespaces")
		if err != nil {
			return err
		}

		var resp []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing KV namespaces: %w", err)
		}

		namespaces := make([]KVNamespaceSummary, 0, len(resp))
		for _, n := range resp {
			namespaces = append(namespaces, toKVNamespaceSummary(n))
		}

		return printKVNamespaces(cmd, namespaces)
	}
}

func newKVNamespacesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespaces-create",
		Short: "Create a KV namespace",
		RunE:  makeRunKVNamespacesCreate(factory),
	}
	cmd.Flags().String("title", "", "Namespace title (required)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func makeRunKVNamespacesCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		title, _ := cmd.Flags().GetString("title")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create KV namespace %q", title), map[string]any{
				"action": "create",
				"title":  title,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath("/storage/kv/namespaces")
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, path, map[string]any{"title": title}, &data); err != nil {
			return fmt.Errorf("creating KV namespace %q: %w", title, err)
		}

		ns := toKVNamespaceSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(ns)
		}
		fmt.Printf("Created KV namespace: %s (ID: %s)\n", ns.Title, ns.ID)
		return nil
	}
}

func newKVKeysListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys-list",
		Short: "List keys in a KV namespace",
		RunE:  makeRunKVKeysList(factory),
	}
	cmd.Flags().String("namespace", "", "KV namespace ID (required)")
	cmd.Flags().Int("limit", 1000, "Maximum number of keys to return")
	_ = cmd.MarkFlagRequired("namespace")
	return cmd
}

func makeRunKVKeysList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		namespaceID, _ := cmd.Flags().GetString("namespace")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath(fmt.Sprintf("/storage/kv/namespaces/%s/keys?limit=%d", namespaceID, limit))
		if err != nil {
			return err
		}

		var resp []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing KV keys in namespace %q: %w", namespaceID, err)
		}

		keys := make([]KVKeySummary, 0, len(resp))
		for _, k := range resp {
			keys = append(keys, toKVKeySummary(k))
		}

		return printKVKeys(cmd, keys)
	}
}

func newKVGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a KV value",
		RunE:  makeRunKVGet(factory),
	}
	cmd.Flags().String("namespace", "", "KV namespace ID (required)")
	cmd.Flags().String("key", "", "Key to retrieve (required)")
	_ = cmd.MarkFlagRequired("namespace")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func makeRunKVGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		namespaceID, _ := cmd.Flags().GetString("namespace")
		key, _ := cmd.Flags().GetString("key")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath(fmt.Sprintf("/storage/kv/namespaces/%s/values/%s", namespaceID, key))
		if err != nil {
			return err
		}

		// KV value endpoint returns the raw value, not a JSON envelope
		raw, err := client.do(ctx, http.MethodGet, path, nil)
		if err != nil {
			return fmt.Errorf("getting KV value for key %q: %w", key, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"key": key, "value": string(raw)})
		}
		fmt.Printf("%s\n", raw)
		return nil
	}
}

func newKVPutCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put",
		Short: "Put a KV value",
		RunE:  makeRunKVPut(factory),
	}
	cmd.Flags().String("namespace", "", "KV namespace ID (required)")
	cmd.Flags().String("key", "", "Key to set (required)")
	cmd.Flags().String("value", "", "Value to store (required)")
	_ = cmd.MarkFlagRequired("namespace")
	_ = cmd.MarkFlagRequired("key")
	_ = cmd.MarkFlagRequired("value")
	return cmd
}

func makeRunKVPut(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		namespaceID, _ := cmd.Flags().GetString("namespace")
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would put key %q in namespace %q", key, namespaceID), map[string]any{
				"action":       "put",
				"namespace_id": namespaceID,
				"key":          key,
				"value":        truncate(value, 50),
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath(fmt.Sprintf("/storage/kv/namespaces/%s/values/%s", namespaceID, key))
		if err != nil {
			return err
		}

		body := rawBody{data: []byte(value), contentType: "text/plain"}
		if _, err := client.do(ctx, http.MethodPut, path, body); err != nil {
			return fmt.Errorf("putting KV value for key %q: %w", key, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "ok", "key": key})
		}
		fmt.Printf("Stored key: %s\n", key)
		return nil
	}
}

func newKVDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a KV key (irreversible)",
		RunE:  makeRunKVDelete(factory),
	}
	cmd.Flags().String("namespace", "", "KV namespace ID (required)")
	cmd.Flags().String("key", "", "Key to delete (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("namespace")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func makeRunKVDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		namespaceID, _ := cmd.Flags().GetString("namespace")
		key, _ := cmd.Flags().GetString("key")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete key %q from namespace %q", key, namespaceID), map[string]any{
				"action":       "delete",
				"namespace_id": namespaceID,
				"key":          key,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath(fmt.Sprintf("/storage/kv/namespaces/%s/values/%s", namespaceID, key))
		if err != nil {
			return err
		}

		if _, err := client.do(ctx, http.MethodDelete, path, nil); err != nil {
			return fmt.Errorf("deleting KV key %q: %w", key, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "key": key})
		}
		fmt.Printf("Deleted key: %s\n", key)
		return nil
	}
}
