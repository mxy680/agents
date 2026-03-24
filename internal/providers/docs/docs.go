package docs

import (
	"context"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
	docsapi "google.golang.org/api/docs/v1"
)

// DocsServiceFactory is a function that creates a Google Docs API service.
type DocsServiceFactory func(ctx context.Context) (*docsapi.Service, error)

// Provider implements the Google Docs integration.
type Provider struct {
	DocsServiceFactory DocsServiceFactory
}

// New creates a new Docs provider using the real Google APIs.
func New() *Provider {
	return &Provider{
		DocsServiceFactory: auth.NewDocsService,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "docs"
}

// RegisterCommands adds all Docs subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	docsCmd := &cobra.Command{
		Use:   "docs",
		Short: "Interact with Google Docs",
		Long:  "Create, read, and update Google Docs documents.",
	}

	documentsCmd := &cobra.Command{
		Use:     "documents",
		Short:   "Manage documents",
		Aliases: []string{"doc"},
	}
	documentsCmd.AddCommand(newDocumentsCreateCmd(p.DocsServiceFactory))
	documentsCmd.AddCommand(newDocumentsGetCmd(p.DocsServiceFactory))
	documentsCmd.AddCommand(newDocumentsAppendCmd(p.DocsServiceFactory))
	documentsCmd.AddCommand(newDocumentsBatchUpdateCmd(p.DocsServiceFactory))
	docsCmd.AddCommand(documentsCmd)

	parent.AddCommand(docsCmd)
}
