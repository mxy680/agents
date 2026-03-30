package dof

import (
	"github.com/spf13/cobra"
)

// Provider implements the NYC Department of Finance tax assessment integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new DOF provider using real HTTP requests.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "dof"
}

// RegisterCommands adds all DOF subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	dofCmd := &cobra.Command{
		Use:     "dof",
		Short:   "Query NYC Department of Finance real estate tax assessment data",
		Long:    "Query the NYC Department of Finance property tax assessment dataset via the NYC Open Data Socrata API.",
		Aliases: []string{"tax"},
	}

	dofCmd.AddCommand(newOwnersCmd(p.ClientFactory))

	parent.AddCommand(dofCmd)
}
