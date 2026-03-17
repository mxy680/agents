package sheets

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	sheetsapi "google.golang.org/api/sheets/v4"
)

// newSpreadsheetsListCmd returns the `spreadsheets list` command.
// Uses the Drive API to list spreadsheet files.
func newSpreadsheetsListCmd(factory DriveServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List spreadsheets",
		Long:  "List Google Sheets spreadsheets accessible to the authenticated user.",
		RunE:  makeRunSpreadsheetsList(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of spreadsheets to return")
	cmd.Flags().String("page-token", "", "Page token for pagination")
	return cmd
}

func makeRunSpreadsheetsList(factory DriveServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := svc.Files.List().
			Q("mimeType='application/vnd.google-apps.spreadsheet'").
			Fields("files(id,name,webViewLink),nextPageToken").
			PageSize(int64(limit)).
			OrderBy("modifiedTime desc")

		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		resp, err := req.Do()
		if err != nil {
			return fmt.Errorf("listing spreadsheets: %w", err)
		}

		summaries := make([]SpreadsheetSummary, 0, len(resp.Files))
		for _, f := range resp.Files {
			summaries = append(summaries, SpreadsheetSummary{
				ID:    f.Id,
				Title: f.Name,
				URL:   f.WebViewLink,
			})
		}

		if cli.IsJSONOutput(cmd) {
			result := map[string]any{
				"spreadsheets": summaries,
			}
			if resp.NextPageToken != "" {
				result["nextPageToken"] = resp.NextPageToken
			}
			return cli.PrintJSON(result)
		}

		if len(summaries) == 0 {
			fmt.Println("No spreadsheets found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-44s  %s", "ID", "TITLE"))
		for _, s := range summaries {
			lines = append(lines, fmt.Sprintf("%-44s  %s", s.ID, s.Title))
		}
		cli.PrintText(lines)
		return nil
	}
}

// newSpreadsheetsGetCmd returns the `spreadsheets get` command.
func newSpreadsheetsGetCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get spreadsheet details",
		Long:  "Get metadata about a spreadsheet including its title, locale, and sheet tabs.",
		RunE:  makeRunSpreadsheetsGet(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunSpreadsheetsGet(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
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

		detail := SpreadsheetDetail{
			ID:     resp.SpreadsheetId,
			Title:  resp.Properties.Title,
			URL:    resp.SpreadsheetUrl,
			Locale: resp.Properties.Locale,
		}

		for _, s := range resp.Sheets {
			detail.Sheets = append(detail.Sheets, SheetInfo{
				SheetID: s.Properties.SheetId,
				Title:   s.Properties.Title,
				Index:   s.Properties.Index,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("Title:  %s", detail.Title),
			fmt.Sprintf("ID:     %s", detail.ID),
			fmt.Sprintf("URL:    %s", detail.URL),
			fmt.Sprintf("Locale: %s", detail.Locale),
			"",
			"Sheets:",
		}
		for _, s := range detail.Sheets {
			lines = append(lines, fmt.Sprintf("  [%d] %s", s.SheetID, s.Title))
		}
		cli.PrintText(lines)
		return nil
	}
}

// newSpreadsheetsCreateCmd returns the `spreadsheets create` command.
func newSpreadsheetsCreateCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new spreadsheet",
		RunE:  makeRunSpreadsheetsCreate(factory),
	}
	cmd.Flags().String("title", "", "Title for the new spreadsheet (required)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func makeRunSpreadsheetsCreate(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		title, _ := cmd.Flags().GetString("title")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create spreadsheet %q", title), map[string]string{
				"status": "dry-run",
				"title":  title,
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		ss := &sheetsapi.Spreadsheet{
			Properties: &sheetsapi.SpreadsheetProperties{
				Title: title,
			},
		}

		resp, err := svc.Spreadsheets.Create(ss).Do()
		if err != nil {
			return fmt.Errorf("creating spreadsheet: %w", err)
		}

		result := SpreadsheetSummary{
			ID:    resp.SpreadsheetId,
			Title: resp.Properties.Title,
			URL:   resp.SpreadsheetUrl,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Created spreadsheet %q (id: %s)\n", result.Title, result.ID)
		return nil
	}
}

// newSpreadsheetsDeleteCmd returns the `spreadsheets delete` command.
// Uses the Drive API to delete the spreadsheet file.
func newSpreadsheetsDeleteCmd(factory DriveServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a spreadsheet (IRREVERSIBLE)",
		RunE:  makeRunSpreadsheetsDelete(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunSpreadsheetsDelete(factory DriveServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would delete spreadsheet "+spreadsheetID, map[string]string{
				"id":     spreadsheetID,
				"status": "deleted",
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Files.Delete(spreadsheetID).Do()
		if err != nil {
			return fmt.Errorf("deleting spreadsheet %s: %w", spreadsheetID, err)
		}

		result := map[string]string{"id": spreadsheetID, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Spreadsheet %s deleted\n", spreadsheetID)
		return nil
	}
}
