package dof

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// boroughCodes maps human-readable borough names to their numeric codes.
var boroughCodes = map[string]string{
	"manhattan":    "1",
	"bronx":        "2",
	"brooklyn":     "3",
	"queens":       "4",
	"staten-island": "5",
}

// boroughNames maps numeric codes to human-readable borough names.
var boroughNames = map[string]string{
	"1": "Manhattan",
	"2": "Bronx",
	"3": "Brooklyn",
	"4": "Queens",
	"5": "Staten Island",
}

// rawOwnerRecord is the raw struct mapping Socrata field names exactly as returned by the API.
type rawOwnerRecord struct {
	BBLE          string `json:"bble"`
	OwnerName     string `json:"owner"`
	TaxClass      string `json:"taxclass"`
	AssessedValue string `json:"avtot"`
	Borough       string `json:"boro"`
	Block         string `json:"block"`
	Lot           string `json:"lot"`
	Address       string `json:"address"`
}

// OwnerRecord is the normalized output type for property owner data.
type OwnerRecord struct {
	BBL           string `json:"bbl"`
	OwnerName     string `json:"owner_name"`
	TaxClass      string `json:"tax_class"`
	AssessedValue string `json:"assessed_value,omitempty"`
	Borough       string `json:"borough,omitempty"`
	Block         string `json:"block,omitempty"`
	Lot           string `json:"lot,omitempty"`
	Address       string `json:"address,omitempty"`
}

// toOwnerRecord converts a rawOwnerRecord to an OwnerRecord.
func toOwnerRecord(r rawOwnerRecord) OwnerRecord {
	return OwnerRecord{
		BBL:           r.BBLE,
		OwnerName:     r.OwnerName,
		TaxClass:      r.TaxClass,
		AssessedValue: r.AssessedValue,
		Borough:       r.Borough,
		Block:         r.Block,
		Lot:           r.Lot,
		Address:       r.Address,
	}
}

// toOwnerRecords converts a slice of rawOwnerRecords to OwnerRecords.
func toOwnerRecords(raw []rawOwnerRecord) []OwnerRecord {
	out := make([]OwnerRecord, len(raw))
	for i, r := range raw {
		out[i] = toOwnerRecord(r)
	}
	return out
}

// printOwnerRecord outputs a single OwnerRecord as JSON or formatted text.
func printOwnerRecord(cmd *cobra.Command, rec OwnerRecord) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(rec)
	}

	lines := []string{
		fmt.Sprintf("BBL:            %s", rec.BBL),
		fmt.Sprintf("Owner:          %s", rec.OwnerName),
		fmt.Sprintf("Tax Class:      %s", rec.TaxClass),
		fmt.Sprintf("Assessed Value: %s", rec.AssessedValue),
	}
	if rec.Borough != "" {
		lines = append(lines, fmt.Sprintf("Borough:        %s", boroughLabel(rec.Borough)))
	}
	if rec.Block != "" {
		lines = append(lines, fmt.Sprintf("Block:          %s", rec.Block))
	}
	if rec.Lot != "" {
		lines = append(lines, fmt.Sprintf("Lot:            %s", rec.Lot))
	}
	if rec.Address != "" {
		lines = append(lines, fmt.Sprintf("Address:        %s", rec.Address))
	}
	cli.PrintText(lines)
	return nil
}

// printOwnerRecords outputs a slice of OwnerRecords as JSON or formatted text table.
func printOwnerRecords(cmd *cobra.Command, recs []OwnerRecord) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(recs)
	}

	if len(recs) == 0 {
		cli.PrintText([]string{"No results found."})
		return nil
	}

	header := fmt.Sprintf("%-12s  %-40s  %-10s  %-14s  %s",
		"BBL", "OWNER", "TAX CLASS", "ASSESSED VALUE", "ADDRESS")
	lines := []string{header}
	for _, r := range recs {
		lines = append(lines, fmt.Sprintf("%-12s  %-40s  %-10s  %-14s  %s",
			r.BBL,
			truncateStr(r.OwnerName, 40),
			r.TaxClass,
			r.AssessedValue,
			r.Address,
		))
	}
	cli.PrintText(lines)
	return nil
}

// boroughLabel returns a human-readable label for a borough code.
func boroughLabel(code string) string {
	if name, ok := boroughNames[code]; ok {
		return name
	}
	return code
}

// lookupBoroughCode returns the numeric code for a borough name, or an error if not found.
func lookupBoroughCode(borough string) (string, error) {
	code, ok := boroughCodes[strings.ToLower(borough)]
	if !ok {
		var valid []string
		for k := range boroughCodes {
			valid = append(valid, k)
		}
		return "", fmt.Errorf("unknown borough %q; valid options: %s", borough, strings.Join(valid, ", "))
	}
	return code, nil
}

// truncateStr truncates s to at most n characters, appending "..." if truncated.
func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
