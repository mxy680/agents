package gcp

import (
	"github.com/spf13/cobra"
)

// Provider implements the Google Cloud Platform integration.
type Provider struct {
	// ClientFactory creates the GCP API client. Defaults to NewClient.
	// Override in tests to inject a mock client pointing at a test server.
	ClientFactory ClientFactory
}

// New creates a new GCP provider using the real GCP APIs.
func New() *Provider {
	return &Provider{
		ClientFactory: NewClient,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "gcp"
}

// RegisterCommands adds all GCP subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	gcpCmd := &cobra.Command{
		Use:     "gcp",
		Short:   "Interact with Google Cloud Platform",
		Long:    "Manage GCP projects, services, OAuth clients, consent screens, and IAM via the GCP REST APIs.",
		Aliases: []string{"gcloud", "gc"},
	}

	projectsCmd := &cobra.Command{
		Use:     "projects",
		Short:   "Manage GCP projects",
		Aliases: []string{"proj"},
	}
	projectsCmd.AddCommand(newProjectsListCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsGetCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsCreateCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsDeleteCmd(p.ClientFactory))
	gcpCmd.AddCommand(projectsCmd)

	servicesCmd := &cobra.Command{
		Use:     "services",
		Short:   "Enable and disable GCP APIs",
		Aliases: []string{"svc", "api"},
	}
	servicesCmd.AddCommand(newServicesListCmd(p.ClientFactory))
	servicesCmd.AddCommand(newServicesEnableCmd(p.ClientFactory))
	servicesCmd.AddCommand(newServicesDisableCmd(p.ClientFactory))
	gcpCmd.AddCommand(servicesCmd)

	oauthCmd := &cobra.Command{
		Use:     "oauth",
		Short:   "Manage IAM Workforce OAuth clients and credentials",
		Aliases: []string{"creds"},
	}
	oauthCmd.AddCommand(newOAuthListCmd(p.ClientFactory))
	oauthCmd.AddCommand(newOAuthCreateCmd(p.ClientFactory))
	oauthCmd.AddCommand(newOAuthUpdateCmd(p.ClientFactory))
	oauthCmd.AddCommand(newOAuthDeleteCmd(p.ClientFactory))
	oauthCmd.AddCommand(newOAuthCreateCredentialsCmd(p.ClientFactory))
	oauthCmd.AddCommand(newOAuthListCredentialsCmd(p.ClientFactory))
	gcpCmd.AddCommand(oauthCmd)

	brandsCmd := &cobra.Command{
		Use:     "brands",
		Short:   "Manage OAuth consent screen (IAP brands)",
		Aliases: []string{"consent"},
	}
	brandsCmd.AddCommand(newBrandsListCmd(p.ClientFactory))
	brandsCmd.AddCommand(newBrandsCreateCmd(p.ClientFactory))
	brandsCmd.AddCommand(newBrandsGetCmd(p.ClientFactory))
	gcpCmd.AddCommand(brandsCmd)

	iamCmd := &cobra.Command{
		Use:     "iam",
		Short:   "Manage service accounts",
		Aliases: []string{"sa"},
	}
	iamCmd.AddCommand(newIAMListCmd(p.ClientFactory))
	iamCmd.AddCommand(newIAMCreateCmd(p.ClientFactory))
	iamCmd.AddCommand(newIAMCreateKeyCmd(p.ClientFactory))
	iamCmd.AddCommand(newIAMDeleteCmd(p.ClientFactory))
	gcpCmd.AddCommand(iamCmd)

	parent.AddCommand(gcpCmd)
}
