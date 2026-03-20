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
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	framerCmd := &cobra.Command{
		Use:     "framer",
		Short:   "Interact with Framer",
		Long:    "Manage Framer projects, pages, CMS collections, styles, and deployments.",
		Aliases: []string{"fr"},
	}

	framerCmd.AddCommand(
		newProjectCmd(p.BridgeClientFactory),
		newPublishCmd(p.BridgeClientFactory),
		newAgentCmd(p.BridgeClientFactory),
		newScreenshotCmd(p.BridgeClientFactory),
		newPluginDataCmd(p.BridgeClientFactory),
	)

	collectionsCmd := &cobra.Command{
		Use:     "collections",
		Short:   "CMS collections",
		Aliases: []string{"col"},
	}
	collectionsCmd.AddCommand(newCollectionsListCmd(p.BridgeClientFactory))
	collectionsCmd.AddCommand(newCollectionsGetCmd(p.BridgeClientFactory))
	collectionsCmd.AddCommand(newCollectionsCreateCmd(p.BridgeClientFactory))
	collectionsCmd.AddCommand(newCollectionsFieldsCmd(p.BridgeClientFactory))
	collectionsCmd.AddCommand(newCollectionsAddFieldsCmd(p.BridgeClientFactory))
	collectionsCmd.AddCommand(newCollectionsRemoveFieldsCmd(p.BridgeClientFactory))
	collectionsCmd.AddCommand(newCollectionsSetFieldOrderCmd(p.BridgeClientFactory))
	collectionsCmd.AddCommand(newCollectionsItemsCmd(p.BridgeClientFactory))
	collectionsCmd.AddCommand(newCollectionsAddItemsCmd(p.BridgeClientFactory))
	collectionsCmd.AddCommand(newCollectionsRemoveItemsCmd(p.BridgeClientFactory))
	collectionsCmd.AddCommand(newCollectionsSetItemOrderCmd(p.BridgeClientFactory))
	framerCmd.AddCommand(collectionsCmd)

	managedColCmd := &cobra.Command{
		Use:     "managed-collections",
		Short:   "Managed CMS collections",
		Aliases: []string{"mcol"},
	}
	managedColCmd.AddCommand(newManagedCollectionsListCmd(p.BridgeClientFactory))
	managedColCmd.AddCommand(newManagedCollectionsCreateCmd(p.BridgeClientFactory))
	managedColCmd.AddCommand(newManagedCollectionsFieldsCmd(p.BridgeClientFactory))
	managedColCmd.AddCommand(newManagedCollectionsSetFieldsCmd(p.BridgeClientFactory))
	managedColCmd.AddCommand(newManagedCollectionsItemsCmd(p.BridgeClientFactory))
	managedColCmd.AddCommand(newManagedCollectionsAddItemsCmd(p.BridgeClientFactory))
	managedColCmd.AddCommand(newManagedCollectionsRemoveItemsCmd(p.BridgeClientFactory))
	managedColCmd.AddCommand(newManagedCollectionsSetItemOrderCmd(p.BridgeClientFactory))
	framerCmd.AddCommand(managedColCmd)

	framerCmd.AddCommand(newRedirectsCmd(p.BridgeClientFactory))
	framerCmd.AddCommand(newCodeCmd(p.BridgeClientFactory))
	framerCmd.AddCommand(newImagesCmd(p.BridgeClientFactory))
	framerCmd.AddCommand(newFilesCmd(p.BridgeClientFactory))
	framerCmd.AddCommand(newSVGCmd(p.BridgeClientFactory))

	framerCmd.AddCommand(newNodesCmd(p.BridgeClientFactory))
	framerCmd.AddCommand(newStylesCmd(p.BridgeClientFactory))
	framerCmd.AddCommand(newFontsCmd(p.BridgeClientFactory))
	framerCmd.AddCommand(newLocalesCmd(p.BridgeClientFactory))

	parent.AddCommand(framerCmd)
}
