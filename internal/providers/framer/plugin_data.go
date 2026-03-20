package framer

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newPluginDataCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin-data",
		Short: "Plugin data storage commands",
	}
	cmd.AddCommand(
		newPluginDataGetCmd(factory),
		newPluginDataSetCmd(factory),
		newPluginDataKeysCmd(factory),
	)
	return cmd
}

func newPluginDataGetCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a plugin data value by key",
		RunE:  makeRunPluginDataGet(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("key", "", "Plugin data key (required)")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func makeRunPluginDataGet(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		key, _ := cmd.Flags().GetString("key")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getPluginData", map[string]any{
			"key": key,
		})
		if err != nil {
			return fmt.Errorf("get plugin data: %w", err)
		}

		var value string
		if err := json.Unmarshal(result, &value); err != nil {
			return fmt.Errorf("parse plugin data value: %w", err)
		}

		return cli.PrintResult(cmd, value, []string{
			fmt.Sprintf("%s = %s", key, value),
		})
	}
}

func newPluginDataSetCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set a plugin data value",
		RunE:  makeRunPluginDataSet(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("dry-run", false, "Preview without setting data")
	cmd.Flags().String("key", "", "Plugin data key (required)")
	cmd.Flags().String("value", "", "Plugin data value (required)")
	_ = cmd.MarkFlagRequired("key")
	_ = cmd.MarkFlagRequired("value")
	return cmd
}

func makeRunPluginDataSet(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		params := map[string]any{
			"key":   key,
			"value": value,
		}

		if ok, err := dryRunResult(cmd, fmt.Sprintf("set plugin data key %q", key), params); ok {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("setPluginData", params)
		if err != nil {
			return fmt.Errorf("set plugin data: %w", err)
		}

		var raw json.RawMessage
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse set result: %w", err)
		}

		return cli.PrintResult(cmd, raw, []string{
			fmt.Sprintf("Set %s = %s", key, value),
		})
	}
}

func newPluginDataKeysCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "List all plugin data keys",
		RunE:  makeRunPluginDataKeys(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunPluginDataKeys(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getPluginDataKeys", nil)
		if err != nil {
			return fmt.Errorf("get plugin data keys: %w", err)
		}

		var keys []string
		if err := json.Unmarshal(result, &keys); err != nil {
			return fmt.Errorf("parse plugin data keys: %w", err)
		}

		lines := make([]string, 0, len(keys)+1)
		lines = append(lines, fmt.Sprintf("Keys (%d):", len(keys)))
		for _, k := range keys {
			lines = append(lines, "  "+k)
		}

		return cli.PrintResult(cmd, keys, lines)
	}
}
