package supabase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newNetworkCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "network",
		Aliases: []string{"net"},
		Short:   "Network restrictions and bans",
	}

	restrictionsCmd := &cobra.Command{
		Use:     "restrictions",
		Aliases: []string{"restrict"},
		Short:   "Network restrictions",
	}
	restrictionsCmd.AddCommand(
		newNetRestrictionsGetCmd(factory),
		newNetRestrictionsUpdateCmd(factory),
		newNetRestrictionsApplyCmd(factory),
	)

	bansCmd := &cobra.Command{
		Use:     "bans",
		Aliases: []string{"ban"},
		Short:   "Network bans",
	}
	bansCmd.AddCommand(
		newNetBansListCmd(factory),
		newNetBansRemoveCmd(factory),
	)

	cmd.AddCommand(restrictionsCmd, bansCmd)
	return cmd
}

// --- restrictions get ---

func newNetRestrictionsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get network restrictions",
		RunE:  makeRunNetRestrictionsGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunNetRestrictionsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/network-restrictions", ref), nil)
		if err != nil {
			return fmt.Errorf("getting network restrictions for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing network restrictions response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		pretty, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(pretty))
		return nil
	}
}

// --- restrictions update ---

func newNetRestrictionsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update network restrictions",
		RunE:  makeRunNetRestrictionsUpdate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("config", "", "Network restrictions config as JSON")
	cmd.Flags().String("config-file", "", "Path to network restrictions config JSON file")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunNetRestrictionsUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		configStr, _ := cmd.Flags().GetString("config")
		configFile, _ := cmd.Flags().GetString("config-file")

		if dryRunResult(cmd, fmt.Sprintf("Would update network restrictions for project %q", ref)) {
			return nil
		}

		var bodyReader *strings.Reader
		if configFile != "" {
			data, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("read config file: %w", err)
			}
			bodyReader = strings.NewReader(string(data))
		} else if configStr != "" {
			bodyReader = strings.NewReader(configStr)
		} else {
			return fmt.Errorf("either --config or --config-file is required")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodPatch, fmt.Sprintf("/projects/%s/network-restrictions", ref), bodyReader)
		if err != nil {
			return fmt.Errorf("updating network restrictions for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing network restrictions response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		pretty, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(pretty))
		return nil
	}
}

// --- restrictions apply ---

func newNetRestrictionsApplyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply pending network restrictions",
		RunE:  makeRunNetRestrictionsApply(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunNetRestrictionsApply(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if dryRunResult(cmd, fmt.Sprintf("Would apply network restrictions for project %q", ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/network-restrictions/apply", ref), nil)
		if err != nil {
			return fmt.Errorf("applying network restrictions for project %q: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			var result any
			json.Unmarshal(raw, &result)
			return cli.PrintJSON(result)
		}

		fmt.Printf("Network restrictions applied for project %q\n", ref)
		return nil
	}
}

// --- bans list ---

func newNetBansListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List network bans",
		RunE:  makeRunNetBansList(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunNetBansList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/network-bans/retrieve", ref), nil)
		if err != nil {
			return fmt.Errorf("listing network bans for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing network bans response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		bans, _ := data["banned_ipv4_addresses"].([]any)
		if len(bans) == 0 {
			fmt.Println("No network bans found.")
			return nil
		}
		lines := make([]string, 0, len(bans)+1)
		lines = append(lines, "BANNED IP ADDRESSES")
		for _, b := range bans {
			ip, _ := b.(string)
			lines = append(lines, ip)
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- bans remove ---

func newNetBansRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove network bans",
		RunE:  makeRunNetBansRemove(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("ips", "", "Comma-separated list of IPs to unban (required)")
	cmd.Flags().Bool("confirm", false, "Confirm removal of bans")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("ips")
	return cmd
}

func makeRunNetBansRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		ipsStr, _ := cmd.Flags().GetString("ips")

		if dryRunResult(cmd, fmt.Sprintf("Would remove network bans for IPs %q in project %q", ipsStr, ref)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("removing network bans for IPs %q", ipsStr)); err != nil {
			return err
		}

		ips := strings.Split(ipsStr, ",")
		for i, ip := range ips {
			ips[i] = strings.TrimSpace(ip)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{"ipv4_addresses": ips}
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}

		if _, err := doSupabase(client, http.MethodDelete, fmt.Sprintf("/projects/%s/network-bans", ref), strings.NewReader(string(bodyBytes))); err != nil {
			return fmt.Errorf("removing network bans for project %q: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "removed", "ips": ipsStr})
		}
		fmt.Printf("Removed network bans for IPs: %s\n", ipsStr)
		return nil
	}
}
