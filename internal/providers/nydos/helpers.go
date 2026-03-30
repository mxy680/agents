package nydos

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// entityTypeMap maps CLI shorthand values to Socrata entity_type strings.
var entityTypeMap = map[string]string{
	"llc":  "DOMESTIC LIMITED LIABILITY COMPANY",
	"corp": "DOMESTIC BUSINESS CORPORATION",
	"lp":   "DOMESTIC LIMITED PARTNERSHIP",
}

// EntitySummary is the normalized result for a single NY DOS entity.
type EntitySummary struct {
	DOSID          string `json:"dos_id"`
	Name           string `json:"name"`
	FilingDate     string `json:"filing_date"`
	EntityType     string `json:"entity_type"`
	ProcessName    string `json:"process_name,omitempty"`
	ProcessAddress string `json:"process_address,omitempty"`
	FilerName      string `json:"filer_name,omitempty"`
	FilerAddress   string `json:"filer_address,omitempty"`
	County         string `json:"county,omitempty"`
}

// DailyFilingRecord is the raw Socrata record from the daily filings endpoint.
type DailyFilingRecord struct {
	DOSID       string `json:"dos_id"`
	CorpName    string `json:"corp_name"`
	FilingDate  string `json:"filing_date"`
	EntityType  string `json:"entity_type"`
	SOPName     string `json:"sop_name"`
	SOPAddr1    string `json:"sop_addr1"`
	SOPCity     string `json:"sop_city"`
	SOPState    string `json:"sop_state"`
	SOPZip5     string `json:"sop_zip5"`
	FilerName   string `json:"filer_name"`
	FilerAddr1  string `json:"filer_addr1"`
	FilerCity   string `json:"filer_city"`
	FilerState  string `json:"filer_state"`
}

// ActiveCorpRecord is the raw Socrata record from the active corporations endpoint.
type ActiveCorpRecord struct {
	DOSID              string `json:"dos_id"`
	CurrentEntityName  string `json:"current_entity_name"`
	InitialDOSFilingDate string `json:"initial_dos_filing_date"`
	County             string `json:"county"`
	Jurisdiction       string `json:"jurisdiction"`
	EntityTypeCode     string `json:"entity_type_code"`
	DOSProcessName     string `json:"dos_process_name"`
	DOSProcessAddr1    string `json:"dos_process_addr_1"`
	DOSProcessCity     string `json:"dos_process_city"`
	DOSProcessState    string `json:"dos_process_state"`
	DOSProcessZip      string `json:"dos_process_zip"`
}

// dailyFilingToSummary converts a DailyFilingRecord to an EntitySummary.
func dailyFilingToSummary(r DailyFilingRecord) EntitySummary {
	return EntitySummary{
		DOSID:          r.DOSID,
		Name:           r.CorpName,
		FilingDate:     truncateDate(r.FilingDate),
		EntityType:     r.EntityType,
		ProcessName:    r.SOPName,
		ProcessAddress: buildAddress(r.SOPAddr1, r.SOPCity, r.SOPState, r.SOPZip5),
		FilerName:      r.FilerName,
		FilerAddress:   buildAddress(r.FilerAddr1, r.FilerCity, r.FilerState, ""),
	}
}

// activeCorpToSummary converts an ActiveCorpRecord to an EntitySummary.
func activeCorpToSummary(r ActiveCorpRecord) EntitySummary {
	return EntitySummary{
		DOSID:          r.DOSID,
		Name:           r.CurrentEntityName,
		FilingDate:     truncateDate(r.InitialDOSFilingDate),
		EntityType:     r.EntityTypeCode,
		County:         r.County,
		ProcessName:    r.DOSProcessName,
		ProcessAddress: buildAddress(r.DOSProcessAddr1, r.DOSProcessCity, r.DOSProcessState, r.DOSProcessZip),
	}
}

// buildAddress concatenates address fields, omitting empty parts.
func buildAddress(addr1, city, state, zip string) string {
	var parts []string
	if addr1 != "" {
		parts = append(parts, addr1)
	}
	if city != "" {
		parts = append(parts, city)
	}
	if state != "" {
		parts = append(parts, state)
	}
	if zip != "" {
		parts = append(parts, zip)
	}
	return strings.Join(parts, ", ")
}

// truncateDate trims Socrata ISO timestamps to date-only (YYYY-MM-DD).
func truncateDate(s string) string {
	if len(s) > 10 {
		return s[:10]
	}
	return s
}

// lookupEntityType maps a CLI shorthand to the full Socrata entity_type string.
func lookupEntityType(t string) (string, error) {
	full, ok := entityTypeMap[strings.ToLower(t)]
	if !ok {
		var valid []string
		for k := range entityTypeMap {
			valid = append(valid, k)
		}
		return "", fmt.Errorf("unknown entity type %q; valid options: %s", t, strings.Join(valid, ", "))
	}
	return full, nil
}

// printEntities outputs a slice of EntitySummary as JSON or a text table.
func printEntities(cmd *cobra.Command, entities []EntitySummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(entities)
	}

	if len(entities) == 0 {
		cli.PrintText([]string{"No entities found."})
		return nil
	}

	header := fmt.Sprintf("%-12s  %-10s  %-50s  %-40s",
		"DOS ID", "DATE", "NAME", "ENTITY TYPE")
	lines := []string{header}
	for _, e := range entities {
		name := e.Name
		if len(name) > 50 {
			name = name[:47] + "..."
		}
		etype := e.EntityType
		if len(etype) > 40 {
			etype = etype[:37] + "..."
		}
		lines = append(lines, fmt.Sprintf("%-12s  %-10s  %-50s  %-40s",
			e.DOSID, e.FilingDate, name, etype))
	}
	cli.PrintText(lines)
	return nil
}
