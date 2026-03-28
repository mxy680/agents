package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// IPRanges holds the Cloudflare IP ranges response.
type IPRanges struct {
	IPv4CIDRs    []string `json:"ipv4_cidrs"`
	IPv6CIDRs    []string `json:"ipv6_cidrs"`
	EthereumIPs  []string `json:"ethereum_ips,omitempty"`
}

func newIPsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Cloudflare IP ranges",
		RunE:  makeRunIPsList(factory),
	}
	return cmd
}

func makeRunIPsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, "/ips", nil, &data); err != nil {
			return fmt.Errorf("listing Cloudflare IPs: %w", err)
		}

		ranges := IPRanges{
			IPv4CIDRs: jsonStringSlice(data["ipv4_cidrs"]),
			IPv6CIDRs: jsonStringSlice(data["ipv6_cidrs"]),
		}
		if eth, ok := data["ethereum_ips"]; ok {
			ranges.EthereumIPs = jsonStringSlice(eth)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(ranges)
		}

		lines := make([]string, 0, len(ranges.IPv4CIDRs)+len(ranges.IPv6CIDRs)+2)
		lines = append(lines, "=== IPv4 CIDRs ===")
		lines = append(lines, ranges.IPv4CIDRs...)
		lines = append(lines, "=== IPv6 CIDRs ===")
		lines = append(lines, ranges.IPv6CIDRs...)
		if len(ranges.EthereumIPs) > 0 {
			lines = append(lines, "=== Ethereum IPs ===")
			lines = append(lines, ranges.EthereumIPs...)
		}
		cli.PrintText(lines)
		return nil
	}
}
