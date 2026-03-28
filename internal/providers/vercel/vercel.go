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

	parent.AddCommand(vercelCmd)
}
