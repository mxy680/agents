package framer

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newRedirectsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redirects",
		Short: "URL redirect management commands",
	}
	cmd.AddCommand(
		newRedirectsListCmd(factory),
		newRedirectsAddCmd(factory),
		newRedirectsRemoveCmd(factory),
		newRedirectsSetOrderCmd(factory),
	)
	return cmd
}

func newRedirectsListCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List URL redirects",
		RunE:  makeRunRedirectsList(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunRedirectsList(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getRedirects", nil)
		if err != nil {
			return fmt.Errorf("list redirects: %w", err)
		}

		var redirects []Redirect
		if err := json.Unmarshal(result, &redirects); err != nil {
			return fmt.Errorf("parse redirects: %w", err)
		}

		lines := make([]string, 0, len(redirects))
		for _, r := range redirects {
			lines = append(lines, fmt.Sprintf("%s  %s → %s", r.ID, r.From, r.To))
		}
		return cli.PrintResult(cmd, redirects, lines)
	}
}

func newRedirectsAddCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add URL redirects",
		RunE:  makeRunRedirectsAdd(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("dry-run", false, "Preview without making changes")
	cmd.Flags().String("redirects", "", `JSON array of redirects, e.g. [{"from":"/old","to":"/new"}]`)
	cmd.Flags().String("redirects-file", "", "Path to JSON file containing redirects array")
	return cmd
}

func makeRunRedirectsAdd(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		redirectsVal, _ := cmd.Flags().GetString("redirects")
		redirectsFile, _ := cmd.Flags().GetString("redirects-file")

		raw, err := parseJSONFlagOrFile(redirectsVal, redirectsFile)
		if err != nil {
			return fmt.Errorf("--redirects: %w", err)
		}

		var items []map[string]any
		if err := json.Unmarshal(raw, &items); err != nil {
			return fmt.Errorf("--redirects must be a JSON array: %w", err)
		}

		if isDry, result := dryRunResult(cmd, "add redirects", items); isDry {
			return result
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("addRedirects", map[string]any{
			"redirects": items,
		})
		if err != nil {
			return fmt.Errorf("add redirects: %w", err)
		}

		var added []Redirect
		if err := json.Unmarshal(result, &added); err != nil {
			return fmt.Errorf("parse added redirects: %w", err)
		}

		lines := make([]string, 0, len(added))
		for _, r := range added {
			lines = append(lines, fmt.Sprintf("%s  %s → %s", r.ID, r.From, r.To))
		}
		return cli.PrintResult(cmd, added, lines)
	}
}

func newRedirectsRemoveCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove URL redirects by ID",
		RunE:  makeRunRedirectsRemove(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
	cmd.Flags().String("ids", "", "Comma-separated redirect IDs to remove")
	_ = cmd.MarkFlagRequired("ids")
	return cmd
}

func makeRunRedirectsRemove(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if !confirmDestructive(cmd, "remove redirects") {
			return nil
		}

		idsVal, _ := cmd.Flags().GetString("ids")
		ids := parseStringList(idsVal)
		if len(ids) == 0 {
			return fmt.Errorf("--ids must specify at least one ID")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("removeRedirects", map[string]any{
			"ids": ids,
		})
		if err != nil {
			return fmt.Errorf("remove redirects: %w", err)
		}

		var raw json.RawMessage = result
		lines := []string{fmt.Sprintf("Removed %d redirect(s)", len(ids))}
		return cli.PrintResult(cmd, raw, lines)
	}
}

func newRedirectsSetOrderCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-order",
		Short: "Set the order of URL redirects",
		RunE:  makeRunRedirectsSetOrder(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("ids", "", "Comma-separated redirect IDs in desired order")
	_ = cmd.MarkFlagRequired("ids")
	return cmd
}

func makeRunRedirectsSetOrder(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		idsVal, _ := cmd.Flags().GetString("ids")
		ids := parseStringList(idsVal)
		if len(ids) == 0 {
			return fmt.Errorf("--ids must specify at least one ID")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("setRedirectOrder", map[string]any{
			"ids": ids,
		})
		if err != nil {
			return fmt.Errorf("set redirect order: %w", err)
		}

		var raw json.RawMessage = result
		lines := []string{fmt.Sprintf("Reordered %d redirect(s)", len(ids))}
		return cli.PrintResult(cmd, raw, lines)
	}
}
