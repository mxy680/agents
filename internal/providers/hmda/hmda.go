package hmda

import (
	"github.com/spf13/cobra"
)

// Provider implements the HMDA (Home Mortgage Disclosure Act) integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new HMDA provider using real HTTP requests.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "hmda"
}

// RegisterCommands adds all HMDA subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	hmdaCmd := &cobra.Command{
		Use:   "hmda",
		Short: "Query CFPB HMDA mortgage origination data for NYC counties",
		Long:  "Query the CFPB Home Mortgage Disclosure Act (HMDA) public API for loan origination data by county and census tract.",
	}

	hmdaCmd.AddCommand(newLoansCmd(p.ClientFactory))

	parent.AddCommand(hmdaCmd)
}
