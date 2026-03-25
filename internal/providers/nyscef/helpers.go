package nyscef

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// CaseSummary holds a brief overview of a NYSCEF case from search results.
type CaseSummary struct {
	DocketID    string `json:"docketId"`
	IndexNumber string `json:"indexNumber"`
	CaseType    string `json:"caseType"`
	Caption     string `json:"caption"`
	FilingDate  string `json:"filingDate"`
	Court       string `json:"court"`
	Status      string `json:"status"`
	URL         string `json:"url"`
}

// countyCode maps a county name to its NYSCEF numeric code.
var countyCode = map[string]string{
	"bronx":     "2",
	"kings":     "24",
	"brooklyn":  "24",
	"new york":  "31",
	"manhattan": "31",
	"queens":    "41",
	"richmond":  "43",
	"staten island": "43",
}

// resolveCountyCode converts a county name or code string to a NYSCEF county code.
// Returns the code and whether it was resolved successfully.
func resolveCountyCode(county string) (string, bool) {
	// If it's already numeric, pass through.
	if isNumeric(county) {
		return county, true
	}
	lower := strings.ToLower(strings.TrimSpace(county))
	if code, ok := countyCode[lower]; ok {
		return code, true
	}
	return "", false
}

// isNumeric returns true if s contains only digit characters.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// printCaseSummaries outputs case summaries as JSON or a formatted text table.
func printCaseSummaries(cmd *cobra.Command, cases []CaseSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(cases)
	}

	if len(cases) == 0 {
		fmt.Println("No cases found.")
		return nil
	}

	lines := make([]string, 0, len(cases)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-15s  %-30s  %-12s  %-10s",
		"INDEX NUMBER", "CASE TYPE", "CAPTION", "FILING DATE", "STATUS"))
	for _, c := range cases {
		caption := truncate(c.Caption, 30)
		lines = append(lines, fmt.Sprintf("%-20s  %-15s  %-30s  %-12s  %-10s",
			c.IndexNumber, truncate(c.CaseType, 15), caption, c.FilingDate, c.Status))
	}
	cli.PrintText(lines)
	return nil
}

// printCaseDetail outputs a single case summary as JSON or text.
func printCaseDetail(cmd *cobra.Command, c CaseSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(c)
	}

	lines := []string{
		fmt.Sprintf("Docket ID:    %s", c.DocketID),
		fmt.Sprintf("Index Number: %s", c.IndexNumber),
		fmt.Sprintf("Case Type:    %s", c.CaseType),
		fmt.Sprintf("Caption:      %s", c.Caption),
		fmt.Sprintf("Filing Date:  %s", c.FilingDate),
		fmt.Sprintf("Court:        %s", c.Court),
		fmt.Sprintf("Status:       %s", c.Status),
		fmt.Sprintf("URL:          %s", c.URL),
	}
	cli.PrintText(lines)
	return nil
}
