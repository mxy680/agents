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

	messagesCmd := &cobra.Command{
		Use:     "messages",
		Short:   "Manage Gmail messages",
		Aliases: []string{"msg"},
	}
	messagesCmd.AddCommand(newMessagesListCmd(p.ServiceFactory))
	messagesCmd.AddCommand(newMessagesGetCmd(p.ServiceFactory))
	messagesCmd.AddCommand(newMessagesSendCmd(p.ServiceFactory))
	messagesCmd.AddCommand(newMessagesTrashCmd(p.ServiceFactory))
	messagesCmd.AddCommand(newMessagesUntrashCmd(p.ServiceFactory))
	messagesCmd.AddCommand(newMessagesDeleteCmd(p.ServiceFactory))
	messagesCmd.AddCommand(newMessagesModifyCmd(p.ServiceFactory))
	messagesCmd.AddCommand(newMessagesImportCmd(p.ServiceFactory))
	messagesCmd.AddCommand(newMessagesInsertCmd(p.ServiceFactory))
	messagesCmd.AddCommand(newMessagesBatchModifyCmd(p.ServiceFactory))
	messagesCmd.AddCommand(newMessagesBatchDeleteCmd(p.ServiceFactory))
	gmailCmd.AddCommand(messagesCmd)

	threadsCmd := &cobra.Command{
		Use:     "threads",
		Short:   "Manage Gmail threads",
		Aliases: []string{"thread"},
	}
	threadsCmd.AddCommand(newThreadsListCmd(p.ServiceFactory))
	threadsCmd.AddCommand(newThreadsGetCmd(p.ServiceFactory))
	threadsCmd.AddCommand(newThreadsTrashCmd(p.ServiceFactory))
	threadsCmd.AddCommand(newThreadsUntrashCmd(p.ServiceFactory))
	threadsCmd.AddCommand(newThreadsDeleteCmd(p.ServiceFactory))
	threadsCmd.AddCommand(newThreadsModifyCmd(p.ServiceFactory))
	gmailCmd.AddCommand(threadsCmd)

	labelsCmd := &cobra.Command{
		Use:     "labels",
		Short:   "Manage Gmail labels",
		Aliases: []string{"label"},
	}
	labelsCmd.AddCommand(newLabelsListCmd(p.ServiceFactory))
	labelsCmd.AddCommand(newLabelsGetCmd(p.ServiceFactory))
	labelsCmd.AddCommand(newLabelsCreateCmd(p.ServiceFactory))
	labelsCmd.AddCommand(newLabelsUpdateCmd(p.ServiceFactory))
	labelsCmd.AddCommand(newLabelsPatchCmd(p.ServiceFactory))
	labelsCmd.AddCommand(newLabelsDeleteCmd(p.ServiceFactory))
	gmailCmd.AddCommand(labelsCmd)

	parent.AddCommand(gmailCmd)
}
