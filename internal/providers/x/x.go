package x

import (
	"github.com/spf13/cobra"
)

// Provider implements the X (Twitter) integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new X provider using the real X internal API.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "x"
}

// RegisterCommands adds all X subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	xCmd := &cobra.Command{
		Use:     "x",
		Short:   "Interact with X (Twitter)",
		Long:    "View and manage tweets, users, timelines, and more via X's internal GraphQL API.",
		Aliases: []string{"twitter"},
	}

	xCmd.AddCommand(newPostsCmd(p.ClientFactory))
	xCmd.AddCommand(newUsersCmd(p.ClientFactory))
	xCmd.AddCommand(newFollowsCmd(p.ClientFactory))
	xCmd.AddCommand(newBlocksCmd(p.ClientFactory))
	xCmd.AddCommand(newMutesCmd(p.ClientFactory))
	xCmd.AddCommand(newLikesCmd(p.ClientFactory))
	xCmd.AddCommand(newRetweetsCmd(p.ClientFactory))
	xCmd.AddCommand(newBookmarksCmd(p.ClientFactory))
	xCmd.AddCommand(newDMCmd(p.ClientFactory))
	xCmd.AddCommand(newListsCmd(p.ClientFactory))

	parent.AddCommand(xCmd)
}
