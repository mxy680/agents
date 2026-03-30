package trends

import (
	"github.com/spf13/cobra"
)

// Provider implements the Google Trends integration.
type Provider struct {
	ServiceFactory ServiceFactory
}

// New creates a new Google Trends provider using the real gogtrends library.
func New() *Provider {
	return &Provider{
		ServiceFactory: DefaultServiceFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "trends"
}

// RegisterCommands adds all Google Trends subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	gtCmd := &cobra.Command{
		Use:     "trends",
		Short:   "Query Google Trends interest data for keywords and neighborhoods",
		Long:    "Query Google Trends interest-over-time, compare multiple keywords, and calculate momentum scores.",
		Aliases: []string{"gt"},
	}

	gtCmd.AddCommand(newInterestCmd(p.ServiceFactory))

	parent.AddCommand(gtCmd)
}
