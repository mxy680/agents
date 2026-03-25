package nysla

import (
	"github.com/spf13/cobra"
)

// Provider implements the NY State Liquor Authority integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new NYSLA provider using real HTTP requests.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "nysla"
}

// RegisterCommands adds all NYSLA subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	nyslaCmd := &cobra.Command{
		Use:     "nysla",
		Short:   "Query NY State Liquor Authority active license data",
		Long:    "Query NY SLA active liquor license data via the Socrata Open Data API (data.ny.gov). Useful for identifying gentrification signals and commercial density in NYC boroughs.",
		Aliases: []string{"liquor"},
	}

	nyslaCmd.AddCommand(newLicensesCmd(p.ClientFactory))

	parent.AddCommand(nyslaCmd)
}
