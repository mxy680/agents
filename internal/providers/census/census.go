package census

import (
	"github.com/spf13/cobra"
)

// Provider implements the US Census Bureau ACS 5-Year Estimates integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new Census provider using real HTTP requests.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "census"
}

// RegisterCommands adds all Census subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	censusCmd := &cobra.Command{
		Use:     "census",
		Short:   "Query US Census Bureau ACS 5-Year demographic estimates",
		Long:    "Query the US Census Bureau American Community Survey (ACS) 5-Year Estimates for NYC census tract demographics including population, income, rent, and housing data.",
		Aliases: []string{"acs"},
	}

	censusCmd.AddCommand(newTractsCmd(p.ClientFactory))

	parent.AddCommand(censusCmd)
}
