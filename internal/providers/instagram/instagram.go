package instagram

import (
	"github.com/spf13/cobra"
)

// Provider implements the Instagram integration.
type Provider struct {
	// ClientFactory creates the Instagram API client. Defaults to DefaultClientFactory.
	// Override in tests to inject a mock client pointing at a test server.
	ClientFactory ClientFactory
}

// New creates a new Instagram provider using the real Instagram web API.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "instagram"
}

// RegisterCommands adds all Instagram subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	igCmd := &cobra.Command{
		Use:     "instagram",
		Short:   "Interact with Instagram",
		Long:    "View profiles, media, stories, and more via the Instagram web API.",
		Aliases: []string{"ig"},
	}

	profileCmd := &cobra.Command{
		Use:     "profile",
		Short:   "View and edit profiles",
		Aliases: []string{"prof"},
	}
	profileCmd.AddCommand(newProfileGetCmd(p.ClientFactory))
	igCmd.AddCommand(profileCmd)

	igCmd.AddCommand(newMediaCmd(p.ClientFactory))
	igCmd.AddCommand(newStoriesCmd(p.ClientFactory))
	igCmd.AddCommand(newReelsCmd(p.ClientFactory))
	igCmd.AddCommand(newCommentsCmd(p.ClientFactory))
	igCmd.AddCommand(newLikesCmd(p.ClientFactory))
	igCmd.AddCommand(newRelationshipsCmd(p.ClientFactory))
	igCmd.AddCommand(newSearchCmd(p.ClientFactory))
	igCmd.AddCommand(newCollectionsCmd(p.ClientFactory))
	igCmd.AddCommand(newTagsCmd(p.ClientFactory))
	igCmd.AddCommand(newLocationsCmd(p.ClientFactory))
	igCmd.AddCommand(newActivityCmd(p.ClientFactory))
	igCmd.AddCommand(newLiveCmd(p.ClientFactory))
	igCmd.AddCommand(newHighlightsCmd(p.ClientFactory))
	igCmd.AddCommand(newCloseFriendsCmd(p.ClientFactory))
	igCmd.AddCommand(newSettingsCmd(p.ClientFactory))

	parent.AddCommand(igCmd)
}
