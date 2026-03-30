package citibike

import (
	"github.com/spf13/cobra"
)

// Provider implements the Citi Bike GBFS integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new Citi Bike provider using real HTTP requests.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "citibike"
}

// RegisterCommands adds all Citi Bike subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	cbCmd := &cobra.Command{
		Use:     "citibike",
		Short:   "Query Citi Bike station availability and density",
		Long:    "Query Citi Bike station information and real-time availability via the GBFS public feed.",
		Aliases: []string{"cb"},
	}

	cbCmd.AddCommand(newStationsCmd(p.ClientFactory))

	parent.AddCommand(cbCmd)
}
