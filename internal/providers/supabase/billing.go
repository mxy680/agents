package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newBillingCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "billing",
		Aliases: []string{"bill"},
		Short:   "Billing and addons",
	}
	addonsCmd := &cobra.Command{
		Use:     "addons",
		Aliases: []string{"addon"},
		Short:   "Manage addons",
	}
	addonsCmd.AddCommand(
		newBillingAddonsListCmd(factory),
		newBillingAddonsApplyCmd(factory),
		newBillingAddonsRemoveCmd(factory),
	)
	cmd.AddCommand(addonsCmd)
	return cmd
}

// --- addons list ---

func newBillingAddonsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List billing addons for a project",
		RunE:  makeRunBillingAddonsList(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunBillingAddonsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/billing/addons", ref), nil)
		if err != nil {
			return fmt.Errorf("listing billing addons: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, data, "", "  "); err != nil {
			fmt.Println(string(data))
			return nil
		}
		fmt.Println(pretty.String())
		return nil
	}
}

// --- addons apply ---

func newBillingAddonsApplyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a billing addon to a project",
		RunE:  makeRunBillingAddonsApply(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("addon", "", "Addon variant string (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without executing")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("addon")
	return cmd
}

func makeRunBillingAddonsApply(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		addon, _ := cmd.Flags().GetString("addon")

		if dryRunResult(cmd, fmt.Sprintf("Would apply addon %q to project %s", addon, ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{"addon_variant": addon}
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}

		data, err := doSupabase(client, http.MethodPatch,
			fmt.Sprintf("/projects/%s/billing/addons", ref), bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("applying addon: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			return cli.PrintJSON(raw)
		}
		fmt.Printf("Applied addon: %s\n", addon)
		return nil
	}
}

// --- addons remove ---

func newBillingAddonsRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a billing addon from a project",
		RunE:  makeRunBillingAddonsRemove(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("addon", "", "Addon variant string (required)")
	cmd.Flags().Bool("confirm", false, "Confirm removal")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without executing")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("addon")
	return cmd
}

func makeRunBillingAddonsRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		addon, _ := cmd.Flags().GetString("addon")

		if dryRunResult(cmd, fmt.Sprintf("Would remove addon %q from project %s", addon, ref)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("removing addon %q is irreversible", addon)); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := doSupabase(client, http.MethodDelete,
			fmt.Sprintf("/projects/%s/billing/addons/%s", ref, addon), nil); err != nil {
			return fmt.Errorf("removing addon: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "removed", "addon": addon})
		}
		fmt.Printf("Removed addon: %s\n", addon)
		return nil
	}
}
