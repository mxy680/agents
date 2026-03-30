package nydos

import (
	"github.com/spf13/cobra"
)

// Provider implements the NY Department of State Socrata Open Data integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new NY DOS provider using real HTTP requests.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "nydos"
}

// RegisterCommands adds all NY DOS subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	dosCmd := &cobra.Command{
		Use:     "nydos",
		Short:   "Query NY Department of State entity formation data",
		Long:    "Query the NY Department of State Socrata Open Data API for corporation and LLC formation data.",
		Aliases: []string{"dos"},
	}

	dosCmd.AddCommand(newEntitiesCmd(p.ClientFactory))

	parent.AddCommand(dosCmd)
}
