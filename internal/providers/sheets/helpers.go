package sheets

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// SpreadsheetSummary is the JSON-serializable summary of a spreadsheet.
type SpreadsheetSummary struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	SheetCount int    `json:"sheetCount,omitempty"`
}

// SpreadsheetDetail is the JSON-serializable detail of a spreadsheet.
type SpreadsheetDetail struct {
	ID     string      `json:"id"`
	Title  string      `json:"title"`
	URL    string      `json:"url"`
	Locale string      `json:"locale,omitempty"`
	Sheets []SheetInfo `json:"sheets"`
}

// SheetInfo describes a single sheet/tab within a spreadsheet.
type SheetInfo struct {
	SheetID int64  `json:"sheetId"`
	Title   string `json:"title"`
	Index   int64  `json:"index"`
}

// CellData holds a range and its values.
type CellData struct {
	Range  string          `json:"range"`
	Values [][]interface{} `json:"values"`
}

// UpdateResult is the JSON-serializable result of a values update.
type UpdateResult struct {
	SpreadsheetID  string `json:"spreadsheetId"`
	UpdatedRange   string `json:"updatedRange"`
	UpdatedRows    int64  `json:"updatedRows"`
	UpdatedColumns int64  `json:"updatedColumns"`
	UpdatedCells   int64  `json:"updatedCells"`
}

// AppendResult is the JSON-serializable result of a values append.
type AppendResult struct {
	SpreadsheetID string `json:"spreadsheetId"`
	TableRange    string `json:"tableRange"`
	UpdatedRange  string `json:"updatedRange"`
	UpdatedRows   int64  `json:"updatedRows"`
	UpdatedCells  int64  `json:"updatedCells"`
}

// ClearResult is the JSON-serializable result of a values clear.
type ClearResult struct {
	SpreadsheetID string `json:"spreadsheetId"`
	ClearedRange  string `json:"clearedRange"`
}

// parseValuesJSON parses a JSON string representing a 2D array of values.
func parseValuesJSON(s string) ([][]interface{}, error) {
	var values [][]interface{}
	if err := json.Unmarshal([]byte(s), &values); err != nil {
		return nil, fmt.Errorf("invalid JSON for --values: %w (expected 2D array like [[\"a\",\"b\"],[\"c\",\"d\"]])", err)
	}
	return values, nil
}

// batchDataEntry represents a single range+values pair for batch operations.
type batchDataEntry struct {
	Range  string          `json:"range"`
	Values [][]interface{} `json:"values"`
}

// parseBatchData parses JSON for batch-update --data flag.
// Expected format: [{"range":"Sheet1!A1:B2","values":[["a","b"],["c","d"]]}]
func parseBatchData(s string) ([]batchDataEntry, error) {
	var entries []batchDataEntry
	if err := json.Unmarshal([]byte(s), &entries); err != nil {
		return nil, fmt.Errorf("invalid JSON for --data: %w (expected array of {range, values} objects)", err)
	}
	return entries, nil
}

// validateValueInput validates the --value-input flag.
func validateValueInput(v string) error {
	switch v {
	case "RAW", "USER_ENTERED":
		return nil
	default:
		return fmt.Errorf("--value-input must be RAW or USER_ENTERED, got %q", v)
	}
}

// validateMajorDimension validates the --major-dimension flag.
func validateMajorDimension(v string) error {
	switch v {
	case "ROWS", "COLUMNS":
		return nil
	default:
		return fmt.Errorf("--major-dimension must be ROWS or COLUMNS, got %q", v)
	}
}

// splitAndTrimRanges splits a comma-separated ranges string and trims whitespace.
func splitAndTrimRanges(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// formatCellsTable formats a 2D array of values as aligned text columns.
func formatCellsTable(values [][]interface{}) []string {
	if len(values) == 0 {
		return []string{"(empty)"}
	}

	// Convert all values to strings
	rows := make([][]string, len(values))
	maxCols := 0
	for i, row := range values {
		rows[i] = make([]string, len(row))
		for j, cell := range row {
			rows[i][j] = fmt.Sprintf("%v", cell)
		}
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	// Compute column widths
	widths := make([]int, maxCols)
	for _, row := range rows {
		for j, cell := range row {
			if len(cell) > widths[j] {
				widths[j] = len(cell)
			}
		}
	}

	// Format rows
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		parts := make([]string, maxCols)
		for j := range maxCols {
			cell := ""
			if j < len(row) {
				cell = row[j]
			}
			parts[j] = fmt.Sprintf("%-*s", widths[j], cell)
		}
		lines = append(lines, strings.Join(parts, "  "))
	}
	return lines
}

// confirmDestructive returns an error if the --confirm flag is absent or false.
func confirmDestructive(cmd *cobra.Command) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("this action is irreversible; re-run with --confirm to proceed")
	}
	return nil
}

// dryRunResult prints a standardised dry-run response and returns nil.
func dryRunResult(cmd *cobra.Command, description string, data any) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(data)
	}
	fmt.Printf("[DRY RUN] %s\n", description)
	return nil
}
