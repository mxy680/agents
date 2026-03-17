package sheets

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	sheetsapi "google.golang.org/api/sheets/v4"
)

// newTabsListCmd returns the `tabs list` command.
func newTabsListCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sheet tabs in a spreadsheet",
		RunE:  makeRunTabsList(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunTabsList(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Spreadsheets.Get(spreadsheetID).Do()
		if err != nil {
			return fmt.Errorf("getting spreadsheet: %w", err)
		}

		tabs := make([]SheetInfo, 0, len(resp.Sheets))
		for _, s := range resp.Sheets {
			tabs = append(tabs, SheetInfo{
				SheetID: s.Properties.SheetId,
				Title:   s.Properties.Title,
				Index:   s.Properties.Index,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(tabs)
		}

		if len(tabs) == 0 {
			fmt.Println("No sheets found.")
			return nil
		}

		lines := make([]string, 0, len(tabs)+1)
		lines = append(lines, fmt.Sprintf("%-10s  %-5s  %s", "SHEET_ID", "INDEX", "TITLE"))
		for _, t := range tabs {
			lines = append(lines, fmt.Sprintf("%-10d  %-5d  %s", t.SheetID, t.Index, t.Title))
		}
		cli.PrintText(lines)
		return nil
	}
}

// newTabsCreateCmd returns the `tabs create` command.
func newTabsCreateCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new sheet tab",
		RunE:  makeRunTabsCreate(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	cmd.Flags().String("title", "", "Title for the new sheet tab (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func makeRunTabsCreate(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")
		title, _ := cmd.Flags().GetString("title")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create tab %q in spreadsheet %s", title, spreadsheetID), map[string]string{
				"status":        "dry-run",
				"spreadsheetId": spreadsheetID,
				"title":         title,
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := &sheetsapi.BatchUpdateSpreadsheetRequest{
			Requests: []*sheetsapi.Request{
				{
					AddSheet: &sheetsapi.AddSheetRequest{
						Properties: &sheetsapi.SheetProperties{
							Title: title,
						},
					},
				},
			},
		}

		resp, err := svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
		if err != nil {
			return fmt.Errorf("creating tab: %w", err)
		}

		var sheetID int64
		if len(resp.Replies) > 0 && resp.Replies[0].AddSheet != nil {
			sheetID = resp.Replies[0].AddSheet.Properties.SheetId
		}

		result := SheetInfo{
			SheetID: sheetID,
			Title:   title,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Created tab %q (sheetId: %d)\n", title, sheetID)
		return nil
	}
}

// newTabsDeleteCmd returns the `tabs delete` command.
func newTabsDeleteCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a sheet tab (IRREVERSIBLE)",
		RunE:  makeRunTabsDelete(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	cmd.Flags().Int64("sheet-id", 0, "Sheet ID of the tab to delete (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("sheet-id")
	return cmd
}

func makeRunTabsDelete(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")
		sheetID, _ := cmd.Flags().GetInt64("sheet-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete tab %d from spreadsheet %s", sheetID, spreadsheetID), map[string]any{
				"status":        "dry-run",
				"spreadsheetId": spreadsheetID,
				"sheetId":       sheetID,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := &sheetsapi.BatchUpdateSpreadsheetRequest{
			Requests: []*sheetsapi.Request{
				{
					DeleteSheet: &sheetsapi.DeleteSheetRequest{
						SheetId: sheetID,
					},
				},
			},
		}

		_, err = svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
		if err != nil {
			return fmt.Errorf("deleting tab %d: %w", sheetID, err)
		}

		result := map[string]any{"spreadsheetId": spreadsheetID, "sheetId": sheetID, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Tab %d deleted from spreadsheet %s\n", sheetID, spreadsheetID)
		return nil
	}
}

// newTabsRenameCmd returns the `tabs rename` command.
func newTabsRenameCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename a sheet tab",
		RunE:  makeRunTabsRename(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	cmd.Flags().Int64("sheet-id", 0, "Sheet ID of the tab to rename (required)")
	cmd.Flags().String("title", "", "New title for the tab (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("sheet-id")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func makeRunTabsRename(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")
		sheetID, _ := cmd.Flags().GetInt64("sheet-id")
		title, _ := cmd.Flags().GetString("title")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would rename tab %d to %q in spreadsheet %s", sheetID, title, spreadsheetID), map[string]any{
				"status":        "dry-run",
				"spreadsheetId": spreadsheetID,
				"sheetId":       sheetID,
				"title":         title,
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := &sheetsapi.BatchUpdateSpreadsheetRequest{
			Requests: []*sheetsapi.Request{
				{
					UpdateSheetProperties: &sheetsapi.UpdateSheetPropertiesRequest{
						Properties: &sheetsapi.SheetProperties{
							SheetId: sheetID,
							Title:   title,
						},
						Fields: "title",
					},
				},
			},
		}

		_, err = svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
		if err != nil {
			return fmt.Errorf("renaming tab %d: %w", sheetID, err)
		}

		result := map[string]any{"spreadsheetId": spreadsheetID, "sheetId": sheetID, "title": title, "status": "renamed"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Tab %d renamed to %q\n", sheetID, title)
		return nil
	}
}
