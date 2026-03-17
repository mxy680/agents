package sheets

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newValuesGetCmd returns the `values get` command.
func newValuesGetCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read cell values from a range",
		Long:  "Read cell values from a spreadsheet range (e.g. Sheet1!A1:D10).",
		RunE:  makeRunValuesGet(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	cmd.Flags().String("range", "", "Cell range in A1 notation (required)")
	cmd.Flags().String("major-dimension", "ROWS", "Major dimension: ROWS or COLUMNS")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("range")
	return cmd
}

func makeRunValuesGet(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")
		cellRange, _ := cmd.Flags().GetString("range")
		majorDimension, _ := cmd.Flags().GetString("major-dimension")

		if err := validateMajorDimension(majorDimension); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Spreadsheets.Values.Get(spreadsheetID, cellRange).
			MajorDimension(majorDimension).Do()
		if err != nil {
			return fmt.Errorf("reading values: %w", err)
		}

		result := CellData{
			Range:  resp.Range,
			Values: resp.Values,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Range: %s\n", resp.Range)
		lines := formatCellsTable(resp.Values)
		cli.PrintText(lines)
		return nil
	}
}

// newValuesBatchGetCmd returns the `values batch-get` command.
func newValuesBatchGetCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-get",
		Short: "Read multiple ranges at once",
		Long:  "Read cell values from multiple ranges in a single request.",
		RunE:  makeRunValuesBatchGet(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	cmd.Flags().String("ranges", "", "Comma-separated ranges in A1 notation (required)")
	cmd.Flags().String("major-dimension", "ROWS", "Major dimension: ROWS or COLUMNS")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("ranges")
	return cmd
}

func makeRunValuesBatchGet(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")
		rangesStr, _ := cmd.Flags().GetString("ranges")
		majorDimension, _ := cmd.Flags().GetString("major-dimension")

		if err := validateMajorDimension(majorDimension); err != nil {
			return err
		}

		ranges := splitAndTrimRanges(rangesStr)

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Spreadsheets.Values.BatchGet(spreadsheetID).
			Ranges(ranges...).MajorDimension(majorDimension).Do()
		if err != nil {
			return fmt.Errorf("batch reading values: %w", err)
		}

		results := make([]CellData, 0, len(resp.ValueRanges))
		for _, vr := range resp.ValueRanges {
			results = append(results, CellData{
				Range:  vr.Range,
				Values: vr.Values,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(results)
		}

		for _, r := range results {
			fmt.Printf("\nRange: %s\n", r.Range)
			lines := formatCellsTable(r.Values)
			cli.PrintText(lines)
		}
		return nil
	}
}
