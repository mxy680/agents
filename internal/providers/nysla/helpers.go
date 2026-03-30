package nysla

import (
	"fmt"
	"sort"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// boroughToCounty maps CLI borough names to the Socrata county_name values.
var boroughToCounty = map[string]string{
	"bronx":         "BRONX",
	"brooklyn":      "KINGS",
	"manhattan":     "NEW YORK",
	"queens":        "QUEENS",
	"staten-island": "RICHMOND",
}

// licenseTypeFilter maps human-readable type flags to partial license_type_name matches.
var licenseTypeFilter = map[string]string{
	"restaurant":   "RESTAURANT",
	"bar":          "BAR",
	"tavern":       "TAVERN",
	"liquor-store": "LIQUOR STORE",
	"wine":         "WINE",
}

// rawLicense is the raw Socrata response record.
type rawLicense struct {
	SerialNumber    string `json:"serial_number"`
	LicenseTypeName string `json:"license_type_name"`
	LicenseTypeCode string `json:"license_type_code"`
	PremisesName    string `json:"premises_name"`
	PremisesAddress string `json:"premises_address"`
	City            string `json:"city"`
	CountyName      string `json:"county_name"`
	ZIP             string `json:"zip"`
	EffectiveDate   string `json:"effective_date"`
	ExpirationDate  string `json:"expiration_date"`
	LicenseStatus   string `json:"license_status"`
}

// LicenseSummary is a simplified view of a single liquor license.
type LicenseSummary struct {
	SerialNumber   string `json:"serial_number"`
	PremisesName   string `json:"premises_name"`
	Address        string `json:"address"`
	City           string `json:"city"`
	ZIP            string `json:"zip"`
	LicenseType    string `json:"license_type"`
	EffectiveDate  string `json:"effective_date"`
	ExpirationDate string `json:"expiration_date"`
}

// CountResult holds the output of the `licenses count` command.
type CountResult struct {
	Borough     string      `json:"borough"`
	ZIP         string      `json:"zip,omitempty"`
	Since       string      `json:"since"`
	NewLicenses int         `json:"new_licenses"`
	Breakdown   []TypeCount `json:"breakdown"`
}

// TypeCount holds a license type name and its count.
type TypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// DensityResult holds the output of the `licenses density` command.
type DensityResult struct {
	ZIP         string      `json:"zip"`
	TotalActive int         `json:"total_active"`
	ByType      []TypeCount `json:"by_type"`
}

// lookupCounty returns the Socrata county_name for a borough flag value,
// or an error if the borough is not recognised.
func lookupCounty(borough string) (string, error) {
	county, ok := boroughToCounty[strings.ToLower(borough)]
	if !ok {
		var valid []string
		for k := range boroughToCounty {
			valid = append(valid, k)
		}
		sort.Strings(valid)
		return "", fmt.Errorf("unknown borough %q; valid options: %s", borough, strings.Join(valid, ", "))
	}
	return county, nil
}

// toSummary converts a rawLicense to a LicenseSummary.
func toSummary(r rawLicense) LicenseSummary {
	return LicenseSummary{
		SerialNumber:   r.SerialNumber,
		PremisesName:   r.PremisesName,
		Address:        r.PremisesAddress,
		City:           r.City,
		ZIP:            r.ZIP,
		LicenseType:    r.LicenseTypeName,
		EffectiveDate:  r.EffectiveDate,
		ExpirationDate: r.ExpirationDate,
	}
}

// buildTypeBreakdown counts rawLicenses by license_type_name and returns
// a breakdown slice sorted by count descending.
func buildTypeBreakdown(licenses []rawLicense) []TypeCount {
	counts := make(map[string]int)
	for _, l := range licenses {
		counts[l.LicenseTypeName]++
	}
	breakdown := make([]TypeCount, 0, len(counts))
	for t, c := range counts {
		breakdown = append(breakdown, TypeCount{Type: t, Count: c})
	}
	sort.Slice(breakdown, func(i, j int) bool {
		if breakdown[i].Count != breakdown[j].Count {
			return breakdown[i].Count > breakdown[j].Count
		}
		return breakdown[i].Type < breakdown[j].Type
	})
	return breakdown
}

// printLicenseSummaries outputs license summaries as JSON or a text table.
func printLicenseSummaries(cmd *cobra.Command, summaries []LicenseSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No licenses found matching the specified criteria.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-35s  %-30s  %-8s  %-25s  %-10s  %-10s",
		"NAME", "ADDRESS", "ZIP", "TYPE", "EFFECTIVE", "EXPIRES"))
	for _, s := range summaries {
		lines = append(lines, fmt.Sprintf("%-35s  %-30s  %-8s  %-25s  %-10s  %-10s",
			truncateStr(s.PremisesName, 35),
			truncateStr(s.Address, 30),
			s.ZIP,
			truncateStr(s.LicenseType, 25),
			shortDate(s.EffectiveDate),
			shortDate(s.ExpirationDate),
		))
	}
	cli.PrintText(lines)
	return nil
}

// printCountResult outputs a CountResult as JSON or formatted text.
func printCountResult(cmd *cobra.Command, result CountResult) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(result)
	}

	lines := []string{
		fmt.Sprintf("Borough:       %s", result.Borough),
		fmt.Sprintf("Since:         %s", result.Since),
		fmt.Sprintf("New licenses:  %d", result.NewLicenses),
	}
	if result.ZIP != "" {
		lines = append([]string{lines[0], fmt.Sprintf("ZIP:           %s", result.ZIP)}, lines[1:]...)
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%-30s  %-6s", "TYPE", "COUNT"))
	for _, tc := range result.Breakdown {
		lines = append(lines, fmt.Sprintf("%-30s  %-6d", tc.Type, tc.Count))
	}
	cli.PrintText(lines)
	return nil
}

// printDensityResult outputs a DensityResult as JSON or formatted text.
func printDensityResult(cmd *cobra.Command, result DensityResult) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(result)
	}

	lines := []string{
		fmt.Sprintf("ZIP:           %s", result.ZIP),
		fmt.Sprintf("Total active:  %d", result.TotalActive),
		"",
		fmt.Sprintf("%-30s  %-6s", "TYPE", "COUNT"),
	}
	for _, tc := range result.ByType {
		lines = append(lines, fmt.Sprintf("%-30s  %-6d", tc.Type, tc.Count))
	}
	cli.PrintText(lines)
	return nil
}

// truncateStr shortens s to at most max runes, appending "..." if truncated.
func truncateStr(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// shortDate returns the first 10 characters of an ISO 8601 date string (YYYY-MM-DD).
func shortDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}
