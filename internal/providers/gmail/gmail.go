package gmail

import (
	"context"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// ServiceFactory is a function that creates a Gmail API service.
type ServiceFactory func(ctx context.Context) (*api.Service, error)

// Provider implements the Gmail integration.
type Provider struct {
	// ServiceFactory creates the Gmail API service. Defaults to auth.NewGmailService.
	// Override in tests to inject a mock service pointing at a test server.
	ServiceFactory ServiceFactory
}

// New creates a new Gmail provider using the real Gmail API.
func New() *Provider {
	return &Provider{
		ServiceFactory: auth.NewGmailService,
	}
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

	gmailCmd.AddCommand(newListUnreadCmd(p.ServiceFactory))
	gmailCmd.AddCommand(newReadCmd(p.ServiceFactory))
	gmailCmd.AddCommand(newSendCmd(p.ServiceFactory))
	gmailCmd.AddCommand(newSearchCmd(p.ServiceFactory))

	parent.AddCommand(gmailCmd)
}
