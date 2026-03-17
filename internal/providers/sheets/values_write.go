package sheets

import (
	"context"
	"fmt"
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	sheetsapi "google.golang.org/api/sheets/v4"
)

// readValuesInput reads values from --values or --values-file flags.
func readValuesInput(cmd *cobra.Command) ([][]interface{}, error) {
	valuesStr, _ := cmd.Flags().GetString("values")
	valuesFile, _ := cmd.Flags().GetString("values-file")

	if valuesStr == "" && valuesFile == "" {
		return nil, fmt.Errorf("either --values or --values-file is required")
	}

	if valuesFile != "" {
		data, err := os.ReadFile(valuesFile)
		if err != nil {
			return nil, fmt.Errorf("reading values file: %w", err)
		}
		valuesStr = string(data)
	}

	return parseValuesJSON(valuesStr)
}

// newValuesUpdateCmd returns the `values update` command.
func newValuesUpdateCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Write cell values to a range",
		Long:  "Write cell values to a spreadsheet range. Values are specified as a JSON 2D array.",
		RunE:  makeRunValuesUpdate(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	cmd.Flags().String("range", "", "Cell range in A1 notation (required)")
	cmd.Flags().String("values", "", "JSON 2D array of values (e.g. [[\"a\",\"b\"],[\"c\",\"d\"]])")
	cmd.Flags().String("values-file", "", "Path to file containing JSON 2D array of values")
	cmd.Flags().String("value-input", "USER_ENTERED", "How to interpret input: RAW or USER_ENTERED")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("range")
	return cmd
}

func makeRunValuesUpdate(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")
		cellRange, _ := cmd.Flags().GetString("range")
		valueInput, _ := cmd.Flags().GetString("value-input")

		if err := validateValueInput(valueInput); err != nil {
			return err
		}

		values, err := readValuesInput(cmd)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update %s in spreadsheet %s", cellRange, spreadsheetID), map[string]any{
				"status":        "dry-run",
				"spreadsheetId": spreadsheetID,
				"range":         cellRange,
				"valueInput":    valueInput,
				"values":        values,
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		vr := &sheetsapi.ValueRange{Values: values}
		resp, err := svc.Spreadsheets.Values.Update(spreadsheetID, cellRange, vr).
			ValueInputOption(valueInput).Do()
		if err != nil {
			return fmt.Errorf("updating values: %w", err)
		}

		result := UpdateResult{
			SpreadsheetID:  resp.SpreadsheetId,
			UpdatedRange:   resp.UpdatedRange,
			UpdatedRows:    resp.UpdatedRows,
			UpdatedColumns: resp.UpdatedColumns,
			UpdatedCells:   resp.UpdatedCells,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Updated %s: %d cells (%d rows x %d columns)\n",
			resp.UpdatedRange, resp.UpdatedCells, resp.UpdatedRows, resp.UpdatedColumns)
		return nil
	}
}

// newValuesAppendCmd returns the `values append` command.
func newValuesAppendCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "append",
		Short: "Append rows to a sheet",
		Long:  "Append rows after the last row with data in the specified range.",
		RunE:  makeRunValuesAppend(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	cmd.Flags().String("range", "", "Cell range in A1 notation (required)")
	cmd.Flags().String("values", "", "JSON 2D array of values to append")
	cmd.Flags().String("values-file", "", "Path to file containing JSON 2D array of values")
	cmd.Flags().String("value-input", "USER_ENTERED", "How to interpret input: RAW or USER_ENTERED")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("range")
	return cmd
}

func makeRunValuesAppend(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")
		cellRange, _ := cmd.Flags().GetString("range")
		valueInput, _ := cmd.Flags().GetString("value-input")

		if err := validateValueInput(valueInput); err != nil {
			return err
		}

		values, err := readValuesInput(cmd)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would append %d rows to %s in spreadsheet %s", len(values), cellRange, spreadsheetID), map[string]any{
				"status":        "dry-run",
				"spreadsheetId": spreadsheetID,
				"range":         cellRange,
				"rows":          len(values),
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		vr := &sheetsapi.ValueRange{Values: values}
		resp, err := svc.Spreadsheets.Values.Append(spreadsheetID, cellRange, vr).
			ValueInputOption(valueInput).Do()
		if err != nil {
			return fmt.Errorf("appending values: %w", err)
		}

		result := AppendResult{
			SpreadsheetID: resp.SpreadsheetId,
			TableRange:    resp.TableRange,
		}
		if resp.Updates != nil {
			result.UpdatedRange = resp.Updates.UpdatedRange
			result.UpdatedRows = resp.Updates.UpdatedRows
			result.UpdatedCells = resp.Updates.UpdatedCells
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Appended %d rows to %s\n", result.UpdatedRows, result.UpdatedRange)
		return nil
	}
}

// newValuesClearCmd returns the `values clear` command.
func newValuesClearCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear cell values in a range",
		Long:  "Clear all values in the specified range without removing formatting.",
		RunE:  makeRunValuesClear(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	cmd.Flags().String("range", "", "Cell range in A1 notation (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive clear operation")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("range")
	return cmd
}

func makeRunValuesClear(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")
		cellRange, _ := cmd.Flags().GetString("range")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would clear %s in spreadsheet %s", cellRange, spreadsheetID), map[string]any{
				"status":        "dry-run",
				"spreadsheetId": spreadsheetID,
				"range":         cellRange,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Spreadsheets.Values.Clear(spreadsheetID, cellRange,
			&sheetsapi.ClearValuesRequest{}).Do()
		if err != nil {
			return fmt.Errorf("clearing values: %w", err)
		}

		result := ClearResult{
			SpreadsheetID: resp.SpreadsheetId,
			ClearedRange:  resp.ClearedRange,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Cleared %s\n", resp.ClearedRange)
		return nil
	}
}

// newValuesBatchUpdateCmd returns the `values batch-update` command.
func newValuesBatchUpdateCmd(factory SheetsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-update",
		Short: "Write to multiple ranges at once",
		Long:  "Write cell values to multiple ranges in a single request.",
		RunE:  makeRunValuesBatchUpdate(factory),
	}
	cmd.Flags().String("id", "", "Spreadsheet ID (required)")
	cmd.Flags().String("data", "", `JSON array of {range, values} objects`)
	cmd.Flags().String("data-file", "", "Path to file containing JSON data array")
	cmd.Flags().String("value-input", "USER_ENTERED", "How to interpret input: RAW or USER_ENTERED")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunValuesBatchUpdate(factory SheetsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		spreadsheetID, _ := cmd.Flags().GetString("id")
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")
		valueInput, _ := cmd.Flags().GetString("value-input")

		if err := validateValueInput(valueInput); err != nil {
			return err
		}

		if dataStr == "" && dataFile == "" {
			return fmt.Errorf("either --data or --data-file is required")
		}

		if dataFile != "" {
			data, err := os.ReadFile(dataFile)
			if err != nil {
				return fmt.Errorf("reading data file: %w", err)
			}
			dataStr = string(data)
		}

		entries, err := parseBatchData(dataStr)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			ranges := make([]string, len(entries))
			for i, e := range entries {
				ranges[i] = e.Range
			}
			return dryRunResult(cmd, fmt.Sprintf("Would batch-update %d ranges in spreadsheet %s", len(entries), spreadsheetID), map[string]any{
				"status":        "dry-run",
				"spreadsheetId": spreadsheetID,
				"ranges":        ranges,
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		valueRanges := make([]*sheetsapi.ValueRange, len(entries))
		for i, e := range entries {
			valueRanges[i] = &sheetsapi.ValueRange{
				Range:  e.Range,
				Values: e.Values,
			}
		}

		req := &sheetsapi.BatchUpdateValuesRequest{
			ValueInputOption: valueInput,
			Data:             valueRanges,
		}

		resp, err := svc.Spreadsheets.Values.BatchUpdate(spreadsheetID, req).Do()
		if err != nil {
			return fmt.Errorf("batch updating values: %w", err)
		}

		result := map[string]any{
			"spreadsheetId":       resp.SpreadsheetId,
			"totalUpdatedRows":    resp.TotalUpdatedRows,
			"totalUpdatedColumns": resp.TotalUpdatedColumns,
			"totalUpdatedCells":   resp.TotalUpdatedCells,
			"totalUpdatedSheets":  resp.TotalUpdatedSheets,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Batch updated %d cells across %d sheets\n",
			resp.TotalUpdatedCells, resp.TotalUpdatedSheets)
		return nil
	}
}
