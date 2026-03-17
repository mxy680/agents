package sheets

import (
	"context"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
	driveapi "google.golang.org/api/drive/v3"
	sheetsapi "google.golang.org/api/sheets/v4"
)

// SheetsServiceFactory is a function that creates a Google Sheets API service.
type SheetsServiceFactory func(ctx context.Context) (*sheetsapi.Service, error)

// DriveServiceFactory is a function that creates a Google Drive API service.
type DriveServiceFactory func(ctx context.Context) (*driveapi.Service, error)

// Provider implements the Google Sheets integration.
type Provider struct {
	SheetsServiceFactory SheetsServiceFactory
	DriveServiceFactory  DriveServiceFactory
}

// New creates a new Sheets provider using the real Google APIs.
func New() *Provider {
	return &Provider{
		SheetsServiceFactory: auth.NewSheetsService,
		DriveServiceFactory:  auth.NewDriveService,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "sheets"
}

// RegisterCommands adds all Sheets subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	sheetsCmd := &cobra.Command{
		Use:   "sheets",
		Short: "Interact with Google Sheets",
		Long:  "Read, write, and manage Google Sheets spreadsheets.",
	}

	spreadsheetsCmd := &cobra.Command{
		Use:     "spreadsheets",
		Short:   "Manage spreadsheets",
		Aliases: []string{"ss"},
	}
	spreadsheetsCmd.AddCommand(newSpreadsheetsListCmd(p.DriveServiceFactory))
	spreadsheetsCmd.AddCommand(newSpreadsheetsGetCmd(p.SheetsServiceFactory))
	spreadsheetsCmd.AddCommand(newSpreadsheetsCreateCmd(p.SheetsServiceFactory))
	spreadsheetsCmd.AddCommand(newSpreadsheetsDeleteCmd(p.DriveServiceFactory))
	sheetsCmd.AddCommand(spreadsheetsCmd)

	valuesCmd := &cobra.Command{
		Use:     "values",
		Short:   "Read and write cell values",
		Aliases: []string{"val"},
	}
	valuesCmd.AddCommand(newValuesGetCmd(p.SheetsServiceFactory))
	valuesCmd.AddCommand(newValuesUpdateCmd(p.SheetsServiceFactory))
	valuesCmd.AddCommand(newValuesAppendCmd(p.SheetsServiceFactory))
	valuesCmd.AddCommand(newValuesClearCmd(p.SheetsServiceFactory))
	valuesCmd.AddCommand(newValuesBatchGetCmd(p.SheetsServiceFactory))
	valuesCmd.AddCommand(newValuesBatchUpdateCmd(p.SheetsServiceFactory))
	sheetsCmd.AddCommand(valuesCmd)

	tabsCmd := &cobra.Command{
		Use:     "tabs",
		Short:   "Manage sheet tabs",
		Aliases: []string{"tab"},
	}
	tabsCmd.AddCommand(newTabsListCmd(p.SheetsServiceFactory))
	tabsCmd.AddCommand(newTabsCreateCmd(p.SheetsServiceFactory))
	tabsCmd.AddCommand(newTabsDeleteCmd(p.SheetsServiceFactory))
	tabsCmd.AddCommand(newTabsRenameCmd(p.SheetsServiceFactory))
	sheetsCmd.AddCommand(tabsCmd)

	parent.AddCommand(sheetsCmd)
}
