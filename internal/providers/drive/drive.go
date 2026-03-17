package drive

import (
	"context"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
	api "google.golang.org/api/drive/v3"
)

// ServiceFactory is a function that creates a Drive API service.
type ServiceFactory func(ctx context.Context) (*api.Service, error)

// Provider implements the Google Drive integration.
type Provider struct {
	// ServiceFactory creates the Drive API service. Defaults to auth.NewDriveService.
	// Override in tests to inject a mock service pointing at a test server.
	ServiceFactory ServiceFactory
}

// New creates a new Drive provider using the real Drive API.
func New() *Provider {
	return &Provider{
		ServiceFactory: auth.NewDriveService,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "drive"
}

// RegisterCommands adds all Drive subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	driveCmd := &cobra.Command{
		Use:   "drive",
		Short: "Interact with Google Drive",
		Long:  "List, upload, download, and manage files and permissions via the Google Drive API.",
	}

	filesCmd := &cobra.Command{
		Use:     "files",
		Short:   "Manage files and folders",
		Aliases: []string{"file", "f"},
	}
	filesCmd.AddCommand(newFilesListCmd(p.ServiceFactory))
	filesCmd.AddCommand(newFilesGetCmd(p.ServiceFactory))
	filesCmd.AddCommand(newFilesDownloadCmd(p.ServiceFactory))
	filesCmd.AddCommand(newFilesUploadCmd(p.ServiceFactory))
	filesCmd.AddCommand(newFilesCopyCmd(p.ServiceFactory))
	filesCmd.AddCommand(newFilesMoveCmd(p.ServiceFactory))
	filesCmd.AddCommand(newFilesTrashCmd(p.ServiceFactory))
	filesCmd.AddCommand(newFilesUntrashCmd(p.ServiceFactory))
	filesCmd.AddCommand(newFilesDeleteCmd(p.ServiceFactory))
	driveCmd.AddCommand(filesCmd)

	permissionsCmd := &cobra.Command{
		Use:     "permissions",
		Short:   "Manage file permissions",
		Aliases: []string{"permission", "perm"},
	}
	permissionsCmd.AddCommand(newPermissionsListCmd(p.ServiceFactory))
	permissionsCmd.AddCommand(newPermissionsGetCmd(p.ServiceFactory))
	permissionsCmd.AddCommand(newPermissionsCreateCmd(p.ServiceFactory))
	permissionsCmd.AddCommand(newPermissionsDeleteCmd(p.ServiceFactory))
	driveCmd.AddCommand(permissionsCmd)

	parent.AddCommand(driveCmd)
}
