package framer

import (
	"github.com/spf13/cobra"
)

// Provider implements the Framer integration.
type Provider struct {
	BridgeClientFactory BridgeClientFactory
}

// New creates a new Framer provider.
func New() *Provider {
	return &Provider{
		BridgeClientFactory: DefaultBridgeClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "framer"
}

// RegisterCommands adds the Framer root subcommand to the parent command.
// Individual resource subcommands will be added incrementally.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	framerCmd := &cobra.Command{
		Use:     "framer",
		Short:   "Interact with Framer",
		Long:    "Manage Framer projects, pages, CMS collections, styles, and deployments.",
		Aliases: []string{"fr"},
	}

	parent.AddCommand(framerCmd)
}
