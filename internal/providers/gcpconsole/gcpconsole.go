package gcpconsole

import (
	"github.com/spf13/cobra"
)

// Provider implements the GCP Console integration.
type Provider struct {
	// ClientFactory creates the GCP Console API client. Defaults to NewClient.
	// Override in tests to inject a mock client pointing at a test server.
	ClientFactory ClientFactory
}

// New creates a new GCP Console provider using the real API.
func New() *Provider {
	return &Provider{
		ClientFactory: NewClient,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "gcp-console"
}

// RegisterCommands adds all GCP Console subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	gcpConsoleCmd := &cobra.Command{
		Use:     "gcp-console",
		Short:   "Interact with GCP Console",
		Long:    "Manage OAuth clients via the GCP Console internal API (SAPISIDHASH authentication).",
		Aliases: []string{"gcc"},
	}

	oauthCmd := &cobra.Command{
		Use:   "oauth",
		Short: "Manage OAuth 2.0 clients",
	}
	oauthCmd.AddCommand(newOAuthListCmd(p.ClientFactory))
	oauthCmd.AddCommand(newOAuthGetCmd(p.ClientFactory))
	oauthCmd.AddCommand(newOAuthCreateCmd(p.ClientFactory))
	oauthCmd.AddCommand(newOAuthUpdateCmd(p.ClientFactory))
	oauthCmd.AddCommand(newOAuthDeleteCmd(p.ClientFactory))
	gcpConsoleCmd.AddCommand(oauthCmd)

	parent.AddCommand(gcpConsoleCmd)
}
