package obituaries

import (
	"github.com/spf13/cobra"
)

// Provider implements the Legacy.com obituary search integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new obituaries provider using real HTTP requests.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "obituaries"
}

// RegisterCommands adds all obituaries subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	obitCmd := &cobra.Command{
		Use:     "obituaries",
		Short:   "Search obituaries via Legacy.com for ACRIS cross-referencing",
		Long:    "Search Legacy.com obituaries by location. Useful for identifying recently deceased property owners for ACRIS estate/probate research.",
		Aliases: []string{"obit"},
	}

	obitCmd.AddCommand(newSearchCmd(p.ClientFactory))
	obitCmd.AddCommand(newNamesCmd(p.ClientFactory))

	parent.AddCommand(obitCmd)
}
