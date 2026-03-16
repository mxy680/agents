package gmail

import "github.com/spf13/cobra"

// Provider implements the Gmail integration.
type Provider struct{}

// New creates a new Gmail provider.
func New() *Provider {
	return &Provider{}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "gmail"
}

// RegisterCommands adds all Gmail subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	gmailCmd := &cobra.Command{
		Use:   "gmail",
		Short: "Interact with Gmail",
		Long:  "List, read, send, and search emails via the Gmail API.",
	}

	gmailCmd.AddCommand(newListUnreadCmd())
	gmailCmd.AddCommand(newReadCmd())
	gmailCmd.AddCommand(newSendCmd())
	gmailCmd.AddCommand(newSearchCmd())

	parent.AddCommand(gmailCmd)
}
