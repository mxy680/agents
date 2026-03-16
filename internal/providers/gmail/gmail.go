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

	draftsCmd := &cobra.Command{
		Use:     "drafts",
		Short:   "Manage Gmail drafts",
		Aliases: []string{"draft"},
	}
	draftsCmd.AddCommand(newDraftsListCmd(p.ServiceFactory))
	draftsCmd.AddCommand(newDraftsGetCmd(p.ServiceFactory))
	draftsCmd.AddCommand(newDraftsCreateCmd(p.ServiceFactory))
	draftsCmd.AddCommand(newDraftsUpdateCmd(p.ServiceFactory))
	draftsCmd.AddCommand(newDraftsSendCmd(p.ServiceFactory))
	draftsCmd.AddCommand(newDraftsDeleteCmd(p.ServiceFactory))
	gmailCmd.AddCommand(draftsCmd)

	attachmentsCmd := &cobra.Command{
		Use:     "attachments",
		Short:   "Manage Gmail message attachments",
		Aliases: []string{"attachment", "att"},
	}
	attachmentsCmd.AddCommand(newAttachmentsGetCmd(p.ServiceFactory))
	gmailCmd.AddCommand(attachmentsCmd)

	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "Track Gmail mailbox changes",
	}
	historyCmd.AddCommand(newHistoryListCmd(p.ServiceFactory))
	gmailCmd.AddCommand(historyCmd)

	settingsCmd := &cobra.Command{
		Use:   "settings",
		Short: "Manage Gmail account settings",
	}
	settingsCmd.AddCommand(newSettingsGetVacationCmd(p.ServiceFactory))
	settingsCmd.AddCommand(newSettingsSetVacationCmd(p.ServiceFactory))
	settingsCmd.AddCommand(newSettingsGetAutoForwardingCmd(p.ServiceFactory))
	settingsCmd.AddCommand(newSettingsSetAutoForwardingCmd(p.ServiceFactory))
	settingsCmd.AddCommand(newSettingsGetImapCmd(p.ServiceFactory))
	settingsCmd.AddCommand(newSettingsSetImapCmd(p.ServiceFactory))
	settingsCmd.AddCommand(newSettingsGetPopCmd(p.ServiceFactory))
	settingsCmd.AddCommand(newSettingsSetPopCmd(p.ServiceFactory))
	settingsCmd.AddCommand(newSettingsGetLanguageCmd(p.ServiceFactory))
	settingsCmd.AddCommand(newSettingsSetLanguageCmd(p.ServiceFactory))

	filtersCmd := &cobra.Command{
		Use:     "filters",
		Short:   "Manage Gmail email filters",
		Aliases: []string{"filter"},
	}
	filtersCmd.AddCommand(newSettingsFiltersListCmd(p.ServiceFactory))
	filtersCmd.AddCommand(newSettingsFiltersGetCmd(p.ServiceFactory))
	filtersCmd.AddCommand(newSettingsFiltersCreateCmd(p.ServiceFactory))
	filtersCmd.AddCommand(newSettingsFiltersDeleteCmd(p.ServiceFactory))
	settingsCmd.AddCommand(filtersCmd)

	forwardingCmd := &cobra.Command{
		Use:     "forwarding-addresses",
		Short:   "Manage Gmail forwarding addresses",
		Aliases: []string{"forwarding"},
	}
	forwardingCmd.AddCommand(newSettingsForwardingListCmd(p.ServiceFactory))
	forwardingCmd.AddCommand(newSettingsForwardingGetCmd(p.ServiceFactory))
	forwardingCmd.AddCommand(newSettingsForwardingCreateCmd(p.ServiceFactory))
	forwardingCmd.AddCommand(newSettingsForwardingDeleteCmd(p.ServiceFactory))
	settingsCmd.AddCommand(forwardingCmd)

	gmailCmd.AddCommand(settingsCmd)

	parent.AddCommand(gmailCmd)
}
