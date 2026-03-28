package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newFirewallListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List firewall rules for a zone",
		RunE:  makeRunFirewallList(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	_ = cmd.MarkFlagRequired("zone")
	return cmd
}

func makeRunFirewallList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/zones/%s/firewall/rules", zoneID), nil, &resp); err != nil {
			return fmt.Errorf("listing firewall rules for zone %q: %w", zoneID, err)
		}

		rules := make([]FirewallRuleSummary, 0, len(resp))
		for _, r := range resp {
			rules = append(rules, toFirewallRuleSummary(r))
		}

		return printFirewallRules(cmd, rules)
	}
}

func newFirewallCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a firewall rule",
		RunE:  makeRunFirewallCreate(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	cmd.Flags().String("action", "", "Rule action (e.g. block, challenge, allow) (required)")
	cmd.Flags().String("expression", "", "Filter expression (required)")
	cmd.Flags().String("description", "", "Rule description")
	_ = cmd.MarkFlagRequired("zone")
	_ = cmd.MarkFlagRequired("action")
	_ = cmd.MarkFlagRequired("expression")
	return cmd
}

func makeRunFirewallCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")
		action, _ := cmd.Flags().GetString("action")
		expression, _ := cmd.Flags().GetString("expression")
		description, _ := cmd.Flags().GetString("description")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create firewall rule (action=%q) in zone %q", action, zoneID), map[string]any{
				"action":      action,
				"expression":  expression,
				"description": description,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := []map[string]any{
			{
				"action": action,
				"filter": map[string]any{
					"expression": expression,
				},
				"description": description,
			},
		}

		var data []map[string]any
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/zones/%s/firewall/rules", zoneID), body, &data); err != nil {
			return fmt.Errorf("creating firewall rule: %w", err)
		}

		if len(data) == 0 {
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]string{"status": "created"})
			}
			fmt.Println("Created firewall rule.")
			return nil
		}

		rule := toFirewallRuleSummary(data[0])
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(rule)
		}
		fmt.Printf("Created firewall rule: %s (action: %s)\n", rule.ID, rule.Action)
		return nil
	}
}

func newFirewallDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a firewall rule (irreversible)",
		RunE:  makeRunFirewallDelete(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	cmd.Flags().String("rule", "", "Firewall rule ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("zone")
	_ = cmd.MarkFlagRequired("rule")
	return cmd
}

func makeRunFirewallDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")
		ruleID, _ := cmd.Flags().GetString("rule")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete firewall rule %q from zone %q", ruleID, zoneID), map[string]any{
				"action":  "delete",
				"zone_id": zoneID,
				"rule_id": ruleID,
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/zones/%s/firewall/rules/%s", zoneID, ruleID), nil); err != nil {
			return fmt.Errorf("deleting firewall rule %q: %w", ruleID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "rule_id": ruleID})
		}
		fmt.Printf("Deleted firewall rule: %s\n", ruleID)
		return nil
	}
}
