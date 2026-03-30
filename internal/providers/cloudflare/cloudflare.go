package cloudflare

import (
	"github.com/spf13/cobra"
)

// Provider implements the Cloudflare integration.
type Provider struct {
	// ClientFactory creates the Cloudflare API client. Defaults to NewClient.
	// Override in tests to inject a mock client pointing at a test server.
	ClientFactory ClientFactory
}

// New creates a new Cloudflare provider using the real Cloudflare API.
func New() *Provider {
	return &Provider{
		ClientFactory: NewClient,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "cloudflare"
}

// RegisterCommands adds all Cloudflare subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	cfCmd := &cobra.Command{
		Use:     "cloudflare",
		Short:   "Interact with Cloudflare",
		Long:    "Manage zones, DNS, Workers, Pages, R2, KV, firewall rules, and more via the Cloudflare API.",
		Aliases: []string{"cf"},
	}

	zonesCmd := &cobra.Command{
		Use:     "zones",
		Short:   "Manage Cloudflare zones",
		Aliases: []string{"zone", "z"},
	}
	zonesCmd.AddCommand(newZonesListCmd(p.ClientFactory))
	zonesCmd.AddCommand(newZonesGetCmd(p.ClientFactory))
	zonesCmd.AddCommand(newZonesPurgeCacheCmd(p.ClientFactory))
	cfCmd.AddCommand(zonesCmd)

	dnsCmd := &cobra.Command{
		Use:     "dns",
		Short:   "Manage DNS records",
		Aliases: []string{"d"},
	}
	dnsCmd.AddCommand(newDNSListCmd(p.ClientFactory))
	dnsCmd.AddCommand(newDNSGetCmd(p.ClientFactory))
	dnsCmd.AddCommand(newDNSCreateCmd(p.ClientFactory))
	dnsCmd.AddCommand(newDNSUpdateCmd(p.ClientFactory))
	dnsCmd.AddCommand(newDNSDeleteCmd(p.ClientFactory))
	cfCmd.AddCommand(dnsCmd)

	workersCmd := &cobra.Command{
		Use:     "workers",
		Short:   "Manage Cloudflare Workers",
		Aliases: []string{"worker", "w"},
	}
	workersCmd.AddCommand(newWorkersListCmd(p.ClientFactory))
	workersCmd.AddCommand(newWorkersGetCmd(p.ClientFactory))
	workersCmd.AddCommand(newWorkersDeployCmd(p.ClientFactory))
	workersCmd.AddCommand(newWorkersDeleteCmd(p.ClientFactory))
	cfCmd.AddCommand(workersCmd)

	pagesCmd := &cobra.Command{
		Use:     "pages",
		Short:   "Manage Cloudflare Pages",
		Aliases: []string{"pg"},
	}
	pagesCmd.AddCommand(newPagesListCmd(p.ClientFactory))
	pagesCmd.AddCommand(newPagesGetCmd(p.ClientFactory))
	deploymentsCmd := &cobra.Command{
		Use:   "deployments",
		Short: "Manage Pages deployments",
	}
	deploymentsCmd.AddCommand(newPagesDeploymentsListCmd(p.ClientFactory))
	deploymentsCmd.AddCommand(newPagesDeploymentsGetCmd(p.ClientFactory))
	pagesCmd.AddCommand(deploymentsCmd)
	cfCmd.AddCommand(pagesCmd)

	r2Cmd := &cobra.Command{
		Use:     "r2",
		Short:   "Manage R2 object storage buckets",
		Aliases: []string{"storage"},
	}
	r2Cmd.AddCommand(newR2ListCmd(p.ClientFactory))
	r2Cmd.AddCommand(newR2CreateCmd(p.ClientFactory))
	r2Cmd.AddCommand(newR2DeleteCmd(p.ClientFactory))
	cfCmd.AddCommand(r2Cmd)

	kvCmd := &cobra.Command{
		Use:     "kv",
		Short:   "Manage KV namespaces and keys",
		Aliases: []string{"kvs"},
	}
	kvCmd.AddCommand(newKVNamespacesListCmd(p.ClientFactory))
	kvCmd.AddCommand(newKVNamespacesCreateCmd(p.ClientFactory))
	kvCmd.AddCommand(newKVKeysListCmd(p.ClientFactory))
	kvCmd.AddCommand(newKVGetCmd(p.ClientFactory))
	kvCmd.AddCommand(newKVPutCmd(p.ClientFactory))
	kvCmd.AddCommand(newKVDeleteCmd(p.ClientFactory))
	cfCmd.AddCommand(kvCmd)

	firewallCmd := &cobra.Command{
		Use:     "firewall",
		Short:   "Manage firewall rules",
		Aliases: []string{"fw"},
	}
	firewallCmd.AddCommand(newFirewallListCmd(p.ClientFactory))
	firewallCmd.AddCommand(newFirewallCreateCmd(p.ClientFactory))
	firewallCmd.AddCommand(newFirewallDeleteCmd(p.ClientFactory))
	cfCmd.AddCommand(firewallCmd)

	certsCmd := &cobra.Command{
		Use:     "certs",
		Short:   "Manage SSL certificate packs",
		Aliases: []string{"cert"},
	}
	certsCmd.AddCommand(newCertsListCmd(p.ClientFactory))
	certsCmd.AddCommand(newCertsGetCmd(p.ClientFactory))
	cfCmd.AddCommand(certsCmd)

	accountsCmd := &cobra.Command{
		Use:     "accounts",
		Short:   "List and inspect Cloudflare accounts",
		Aliases: []string{"acct"},
	}
	accountsCmd.AddCommand(newAccountsListCmd(p.ClientFactory))
	accountsCmd.AddCommand(newAccountsGetCmd(p.ClientFactory))
	cfCmd.AddCommand(accountsCmd)

	ipsCmd := &cobra.Command{
		Use:   "ips",
		Short: "List Cloudflare IP ranges",
	}
	ipsCmd.AddCommand(newIPsListCmd(p.ClientFactory))
	cfCmd.AddCommand(ipsCmd)

	parent.AddCommand(cfCmd)
}
