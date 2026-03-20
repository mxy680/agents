package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newKeysCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keys",
		Aliases: []string{"key"},
		Short:   "Manage project API keys",
	}
	cmd.AddCommand(
		newKeysListCmd(factory),
		newKeysGetCmd(factory),
		newKeysCreateCmd(factory),
		newKeysUpdateCmd(factory),
		newKeysDeleteCmd(factory),
	)
	return cmd
}

// toAPIKeySummary converts a raw API response map to an APIKeySummary.
func toAPIKeySummary(data map[string]any) APIKeySummary {
	id, _ := data["id"].(string)
	name, _ := data["name"].(string)
	apiKey, _ := data["api_key"].(string)
	keyType, _ := data["type"].(string)
	return APIKeySummary{
		ID:     id,
		Name:   name,
		APIKey: apiKey,
		Type:   keyType,
	}
}

// printAPIKeySummaries outputs API key summaries as JSON or a formatted text table.
// The API key is masked in table output for security.
func printAPIKeySummaries(cmd *cobra.Command, keys []APIKeySummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(keys)
	}
	if len(keys) == 0 {
		fmt.Println("No API keys found.")
		return nil
	}
	lines := make([]string, 0, len(keys)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-20s  %-10s  %s", "ID", "NAME", "TYPE", "API_KEY"))
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%-20s  %-20s  %-10s  %s",
			truncate(k.ID, 20), truncate(k.Name, 20), truncate(k.Type, 10), maskKey(k.APIKey)))
	}
	cli.PrintText(lines)
	return nil
}

// --- list ---

func newKeysListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List API keys for a project",
		RunE:  makeRunKeysList(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunKeysList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/api-keys", ref), nil)
		if err != nil {
			return fmt.Errorf("listing API keys: %w", err)
		}

		var raw []map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		keys := make([]APIKeySummary, 0, len(raw))
		for _, r := range raw {
			keys = append(keys, toAPIKeySummary(r))
		}
		return printAPIKeySummaries(cmd, keys)
	}
}

// --- get ---

func newKeysGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of an API key",
		RunE:  makeRunKeysGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("key-id", "", "API key ID (required)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("key-id")
	return cmd
}

func makeRunKeysGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		keyID, _ := cmd.Flags().GetString("key-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/api-keys/%s", ref, keyID), nil)
		if err != nil {
			return fmt.Errorf("getting API key %s: %w", keyID, err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		k := toAPIKeySummary(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(k)
		}

		lines := []string{
			fmt.Sprintf("ID:      %s", k.ID),
			fmt.Sprintf("Name:    %s", k.Name),
			fmt.Sprintf("Type:    %s", k.Type),
			fmt.Sprintf("API Key: %s", k.APIKey),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- create ---

func newKeysCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API key",
		RunE:  makeRunKeysCreate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("name", "", "Key name (required)")
	cmd.Flags().String("type", "anon", "Key type (default: anon)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunKeysCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		name, _ := cmd.Flags().GetString("name")
		keyType, _ := cmd.Flags().GetString("type")

		if dryRunResult(cmd, fmt.Sprintf("Would create API key %q (type: %s) in project %s", name, keyType, ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{
			"name": name,
			"type": keyType,
		}
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}

		data, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/api-keys", ref), bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("creating API key: %w", err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		k := toAPIKeySummary(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(k)
		}
		fmt.Printf("Created API key: %s (ID: %s)\n", k.Name, k.ID)
		return nil
	}
}

// --- update ---

func newKeysUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an API key",
		RunE:  makeRunKeysUpdate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("key-id", "", "API key ID (required)")
	cmd.Flags().String("name", "", "New key name")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("key-id")
	return cmd
}

func makeRunKeysUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		keyID, _ := cmd.Flags().GetString("key-id")
		name, _ := cmd.Flags().GetString("name")

		if dryRunResult(cmd, fmt.Sprintf("Would update API key %s in project %s", keyID, ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{}
		if name != "" {
			bodyMap["name"] = name
		}
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}

		data, err := doSupabase(client, http.MethodPatch, fmt.Sprintf("/projects/%s/api-keys/%s", ref, keyID), bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("updating API key %s: %w", keyID, err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		k := toAPIKeySummary(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(k)
		}
		fmt.Printf("Updated API key: %s (ID: %s)\n", k.Name, k.ID)
		return nil
	}
}

// --- delete ---

func newKeysDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an API key (irreversible)",
		RunE:  makeRunKeysDelete(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("key-id", "", "API key ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("key-id")
	return cmd
}

func makeRunKeysDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		keyID, _ := cmd.Flags().GetString("key-id")

		if dryRunResult(cmd, fmt.Sprintf("Would permanently delete API key %s from project %s", keyID, ref)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("deleting API key %s is irreversible", keyID)); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := doSupabase(client, http.MethodDelete, fmt.Sprintf("/projects/%s/api-keys/%s", ref, keyID), nil); err != nil {
			return fmt.Errorf("deleting API key %s: %w", keyID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "keyId": keyID})
		}
		fmt.Printf("Deleted API key: %s\n", keyID)
		return nil
	}
}
