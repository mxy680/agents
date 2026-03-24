package streeteasy

import (
	"github.com/spf13/cobra"
)

// Provider implements the StreetEasy real estate integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new StreetEasy provider using real HTTP requests.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "streeteasy"
}

// RegisterCommands adds all StreetEasy subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	seCmd := &cobra.Command{
		Use:     "streeteasy",
		Short:   "Search StreetEasy for NYC real estate listings and price history",
		Long:    "Search StreetEasy's NYC real estate listings and price history — uses cookies from Playwright session capture.",
		Aliases: []string{"se"},
	}

	seCmd.AddCommand(newListingsCmd(p.ClientFactory))

	parent.AddCommand(seCmd)
}
