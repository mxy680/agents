package imessage

import (
	"github.com/spf13/cobra"
)

// Provider implements the iMessage integration via BlueBubbles.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new iMessage provider using the BlueBubbles REST API.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "imessage"
}

// RegisterCommands adds all iMessage subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	imsgCmd := &cobra.Command{
		Use:     "imessage",
		Short:   "Interact with iMessage via BlueBubbles",
		Long:    "Send and receive messages, manage chats, attachments, contacts, scheduled messages, and more via the BlueBubbles REST API.",
		Aliases: []string{"imsg"},
	}

	imsgCmd.AddCommand(newChatsCmd(p.ClientFactory))
	imsgCmd.AddCommand(newParticipantsCmd(p.ClientFactory))
	imsgCmd.AddCommand(newMessagesCmd(p.ClientFactory))
	imsgCmd.AddCommand(newScheduledCmd(p.ClientFactory))
	imsgCmd.AddCommand(newAttachmentsCmd(p.ClientFactory))
	imsgCmd.AddCommand(newHandlesCmd(p.ClientFactory))
	imsgCmd.AddCommand(newContactsCmd(p.ClientFactory))
	imsgCmd.AddCommand(newFaceTimeCmd(p.ClientFactory))
	imsgCmd.AddCommand(newFindMyCmd(p.ClientFactory))
	imsgCmd.AddCommand(newICloudCmd(p.ClientFactory))
	imsgCmd.AddCommand(newServerCmd(p.ClientFactory))
	imsgCmd.AddCommand(newWebhooksCmd(p.ClientFactory))
	imsgCmd.AddCommand(newMacCmd(p.ClientFactory))

	parent.AddCommand(imsgCmd)
}
