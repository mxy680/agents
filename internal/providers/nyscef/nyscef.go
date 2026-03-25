package nyscef

import (
	"github.com/spf13/cobra"
)

// Provider implements the NYSCEF court filing integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new NYSCEF provider using real HTTP requests.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "nyscef"
}

// RegisterCommands adds all NYSCEF subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	nyscefCmd := &cobra.Command{
		Use:   "nyscef",
		Short: "Search and retrieve NY State Courts Electronic Filing (NYSCEF) cases",
		Long:  "Search and retrieve NY State Courts Electronic Filing (NYSCEF) cases — uses Chrome TLS fingerprint to bypass Cloudflare protection.",
	}

	nyscefCmd.AddCommand(newCasesCmd(p.ClientFactory))

	parent.AddCommand(nyscefCmd)
}
