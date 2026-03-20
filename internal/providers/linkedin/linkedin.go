package linkedin

import (
	"github.com/spf13/cobra"
)

// Provider implements the LinkedIn integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new LinkedIn provider using the real LinkedIn Voyager API.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "linkedin"
}

// RegisterCommands adds all LinkedIn subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	liCmd := &cobra.Command{
		Use:     "linkedin",
		Short:   "Interact with LinkedIn",
		Long:    "View profiles, posts, connections, messages, and more via the LinkedIn Voyager API.",
		Aliases: []string{"li"},
	}

	liCmd.AddCommand(newProfileCmd(p.ClientFactory))
	liCmd.AddCommand(newConnectionsCmd(p.ClientFactory))
	liCmd.AddCommand(newInvitationsCmd(p.ClientFactory))
	liCmd.AddCommand(newPostsCmd(p.ClientFactory))
	liCmd.AddCommand(newCommentsCmd(p.ClientFactory))
	liCmd.AddCommand(newFeedCmd(p.ClientFactory))
	liCmd.AddCommand(newSearchCmd(p.ClientFactory))
	liCmd.AddCommand(newNetworkCmd(p.ClientFactory))
	liCmd.AddCommand(newNotificationsCmd(p.ClientFactory))
	liCmd.AddCommand(newAnalyticsCmd(p.ClientFactory))
	liCmd.AddCommand(newSettingsCmd(p.ClientFactory))

	parent.AddCommand(liCmd)
}
