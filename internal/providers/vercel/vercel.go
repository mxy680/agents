package vercel

import (
	"github.com/spf13/cobra"
)

// Provider implements the Vercel integration.
type Provider struct {
	// ClientFactory creates the Vercel API client. Defaults to NewClient.
	// Override in tests to inject a mock client pointing at a test server.
	ClientFactory ClientFactory
}

// New creates a new Vercel provider using the real Vercel API.
func New() *Provider {
	return &Provider{
		ClientFactory: NewClient,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "vercel"
}

// RegisterCommands adds all Vercel subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	vercelCmd := &cobra.Command{
		Use:     "vercel",
		Short:   "Interact with Vercel",
		Long:    "Manage projects, deployments, domains, and more via the Vercel API.",
		Aliases: []string{"vc"},
	}

	projectsCmd := &cobra.Command{
		Use:     "projects",
		Short:   "Manage Vercel projects",
		Aliases: []string{"proj"},
	}
	projectsCmd.AddCommand(newProjectsListCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsGetCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsCreateCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsUpdateCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsDeleteCmd(p.ClientFactory))
	vercelCmd.AddCommand(projectsCmd)

	deploymentsCmd := &cobra.Command{
		Use:     "deployments",
		Short:   "Manage Vercel deployments",
		Aliases: []string{"deploy", "dep"},
	}
	deploymentsCmd.AddCommand(newDeploymentsListCmd(p.ClientFactory))
	deploymentsCmd.AddCommand(newDeploymentsGetCmd(p.ClientFactory))
	deploymentsCmd.AddCommand(newDeploymentsCreateCmd(p.ClientFactory))
	deploymentsCmd.AddCommand(newDeploymentsCancelCmd(p.ClientFactory))
	deploymentsCmd.AddCommand(newDeploymentsDeleteCmd(p.ClientFactory))
	vercelCmd.AddCommand(deploymentsCmd)

	domainsCmd := &cobra.Command{
		Use:     "domains",
		Short:   "Manage Vercel domains",
		Aliases: []string{"domain", "dns"},
	}
	domainsCmd.AddCommand(newDomainsListCmd(p.ClientFactory))
	domainsCmd.AddCommand(newDomainsGetCmd(p.ClientFactory))
	domainsCmd.AddCommand(newDomainsAddCmd(p.ClientFactory))
	domainsCmd.AddCommand(newDomainsVerifyCmd(p.ClientFactory))
	domainsCmd.AddCommand(newDomainsRemoveCmd(p.ClientFactory))
	vercelCmd.AddCommand(domainsCmd)

	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage project environment variables",
	}
	envCmd.AddCommand(newEnvListCmd(p.ClientFactory))
	envCmd.AddCommand(newEnvGetCmd(p.ClientFactory))
	envCmd.AddCommand(newEnvSetCmd(p.ClientFactory))
	envCmd.AddCommand(newEnvRemoveCmd(p.ClientFactory))
	vercelCmd.AddCommand(envCmd)

	dnsCmd := &cobra.Command{
		Use:   "dns",
		Short: "Manage DNS records",
	}
	dnsCmd.AddCommand(newDNSListCmd(p.ClientFactory))
	dnsCmd.AddCommand(newDNSAddCmd(p.ClientFactory))
	dnsCmd.AddCommand(newDNSRemoveCmd(p.ClientFactory))
	vercelCmd.AddCommand(dnsCmd)

	certsCmd := &cobra.Command{
		Use:     "certs",
		Short:   "Manage TLS certificates",
		Aliases: []string{"cert"},
	}
	certsCmd.AddCommand(newCertsListCmd(p.ClientFactory))
	certsCmd.AddCommand(newCertsGetCmd(p.ClientFactory))
	vercelCmd.AddCommand(certsCmd)

	teamsCmd := &cobra.Command{
		Use:     "teams",
		Short:   "Manage Vercel teams",
		Aliases: []string{"team"},
	}
	teamsCmd.AddCommand(newTeamsListCmd(p.ClientFactory))
	teamsCmd.AddCommand(newTeamsGetCmd(p.ClientFactory))
	teamsCmd.AddCommand(newTeamsMembersCmd(p.ClientFactory))
	vercelCmd.AddCommand(teamsCmd)

	aliasesCmd := &cobra.Command{
		Use:     "aliases",
		Short:   "Manage deployment aliases",
		Aliases: []string{"alias"},
	}
	aliasesCmd.AddCommand(newAliasesListCmd(p.ClientFactory))
	aliasesCmd.AddCommand(newAliasesAssignCmd(p.ClientFactory))
	vercelCmd.AddCommand(aliasesCmd)

	logsCmd := &cobra.Command{
		Use:     "logs",
		Short:   "View deployment logs",
		Aliases: []string{"log"},
	}
	logsCmd.AddCommand(newLogsGetCmd(p.ClientFactory))
	vercelCmd.AddCommand(logsCmd)

	webhooksCmd := &cobra.Command{
		Use:     "webhooks",
		Short:   "Manage webhooks",
		Aliases: []string{"wh"},
	}
	webhooksCmd.AddCommand(newWebhooksListCmd(p.ClientFactory))
	webhooksCmd.AddCommand(newWebhooksCreateCmd(p.ClientFactory))
	webhooksCmd.AddCommand(newWebhooksDeleteCmd(p.ClientFactory))
	vercelCmd.AddCommand(webhooksCmd)

	parent.AddCommand(vercelCmd)
}
