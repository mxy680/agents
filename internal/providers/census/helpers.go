package census

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// Core ACS variable codes for demographic data.
const (
	varPopulation     = "B01001_001E" // Total population
	varMedianIncome   = "B19013_001E" // Median household income
	varMedianRent     = "B25064_001E" // Median gross rent
	varMedianHomeVal  = "B25077_001E" // Median home value (owner-occupied)
	varTotalHousing   = "B25002_001E" // Total housing units
	varVacantUnits    = "B25002_003E" // Vacant housing units
	varOwnerOccupied  = "B25003_002E" // Owner-occupied units
	varRenterOccupied = "B25003_003E" // Renter-occupied units
)

// allVars is the comma-separated list of all variables to request.
const allVars = "NAME," + varPopulation + "," + varMedianIncome + "," + varMedianRent + "," +
	varMedianHomeVal + "," + varTotalHousing + "," + varVacantUnits + "," +
	varOwnerOccupied + "," + varRenterOccupied

// boroughToFIPS maps borough names to their NYC county FIPS codes (state 36).
var boroughToFIPS = map[string]string{
	"bronx":         "005",
	"brooklyn":      "047",
	"manhattan":     "061",
	"queens":        "081",
	"staten-island": "085",
}

// boroughNames maps county FIPS to human-readable borough names.
var boroughNames = map[string]string{
	"005": "Bronx",
	"047": "Brooklyn",
	"061": "Manhattan",
	"081": "Queens",
	"085": "Staten Island",
}

// TractProfile is the full demographic profile for a single census tract.
type TractProfile struct {
	Tract             string  `json:"tract"`
	Name              string  `json:"name"`
	Population        int     `json:"population"`
	MedianIncome      int     `json:"median_income"`
	MedianRent        int     `json:"median_rent"`
	MedianHomeValue   int     `json:"median_home_value"`
	VacancyRate       float64 `json:"vacancy_rate"`
	OwnerOccupiedPct  float64 `json:"owner_occupied_pct"`
	RenterOccupiedPct float64 `json:"renter_occupied_pct"`
}

// BoroughSummary is the aggregate statistics for an entire borough.
type BoroughSummary struct {
	Borough              string  `json:"borough"`
	TotalPopulation      int     `json:"total_population"`
	AvgMedianIncome      int     `json:"avg_median_income"`
	AvgMedianRent        int     `json:"avg_median_rent"`
	AvgVacancyRate       float64 `json:"avg_vacancy_rate"`
	TotalTracts          int     `json:"total_tracts"`
	HighVacancyTracts    int     `json:"high_vacancy_tracts"`
	HighRentBurdenTracts int     `json:"high_rent_burden_tracts"`
}

// lookupBoroughFIPS returns the county FIPS for a borough name, or an error.
func lookupBoroughFIPS(borough string) (string, error) {
	fips, ok := boroughToFIPS[strings.ToLower(borough)]
	if !ok {
		var valid []string
		for k := range boroughToFIPS {
			valid = append(valid, k)
		}
		return "", fmt.Errorf("unknown borough %q; valid options: %s", borough, strings.Join(valid, ", "))
	}
	return fips, nil
}

// parseTractFIPS splits "36005000100" into state="36", county="005", tract="000100".
func parseTractFIPS(fips string) (state, county, tract string, err error) {
	if len(fips) != 11 {
		return "", "", "", fmt.Errorf("tract FIPS must be 11 digits, got %d", len(fips))
	}
	return fips[:2], fips[2:5], fips[5:], nil
}

// rowsToMaps converts Census API [][]string rows (row 0 = headers) into a slice of maps.
func rowsToMaps(data [][]string) []map[string]string {
	if len(data) < 2 {
		return nil
	}
	headers := data[0]
	var results []map[string]string
	for _, row := range data[1:] {
		m := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(row) {
				m[h] = row[i]
			}
		}
		results = append(results, m)
	}
	return results
}

// parseInt parses a Census API string value into an int.
// Census API returns "-" for suppressed/unavailable data; those parse as 0.
func parseInt(s string) int {
	if s == "-" || s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}

// computeVacancyRate returns the vacancy rate as a percentage (0–100).
func computeVacancyRate(totalHousing, vacantUnits int) float64 {
	if totalHousing == 0 {
		return 0
	}
	return float64(vacantUnits) / float64(totalHousing) * 100
}

// computeOwnerPct returns the owner-occupied percentage of occupied units.
func computeOwnerPct(ownerOccupied, renterOccupied int) float64 {
	total := ownerOccupied + renterOccupied
	if total == 0 {
		return 0
	}
	return float64(ownerOccupied) / float64(total) * 100
}

// computeRenterPct returns the renter-occupied percentage of occupied units.
func computeRenterPct(ownerOccupied, renterOccupied int) float64 {
	total := ownerOccupied + renterOccupied
	if total == 0 {
		return 0
	}
	return float64(renterOccupied) / float64(total) * 100
}

// rowToProfile converts a census API row map to a TractProfile.
// fullFIPS is the 11-digit FIPS (state+county+tract).
func rowToProfile(m map[string]string, fullFIPS string) TractProfile {
	totalHousing := parseInt(m[varTotalHousing])
	vacantUnits := parseInt(m[varVacantUnits])
	ownerOccupied := parseInt(m[varOwnerOccupied])
	renterOccupied := parseInt(m[varRenterOccupied])

	return TractProfile{
		Tract:             fullFIPS,
		Name:              m["NAME"],
		Population:        parseInt(m[varPopulation]),
		MedianIncome:      parseInt(m[varMedianIncome]),
		MedianRent:        parseInt(m[varMedianRent]),
		MedianHomeValue:   parseInt(m[varMedianHomeVal]),
		VacancyRate:       roundTwo(computeVacancyRate(totalHousing, vacantUnits)),
		OwnerOccupiedPct:  roundTwo(computeOwnerPct(ownerOccupied, renterOccupied)),
		RenterOccupiedPct: roundTwo(computeRenterPct(ownerOccupied, renterOccupied)),
	}
}

// roundTwo rounds a float64 to two decimal places.
func roundTwo(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// printTractProfile outputs a TractProfile as JSON or formatted text.
func printTractProfile(cmd *cobra.Command, p TractProfile) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(p)
	}

	lines := []string{
		fmt.Sprintf("Tract:               %s", p.Tract),
		fmt.Sprintf("Name:                %s", p.Name),
		fmt.Sprintf("Population:          %d", p.Population),
		fmt.Sprintf("Median Income:       %s", formatDollars(p.MedianIncome)),
		fmt.Sprintf("Median Rent:         %s", formatDollars(p.MedianRent)),
		fmt.Sprintf("Median Home Value:   %s", formatDollars(p.MedianHomeValue)),
		fmt.Sprintf("Vacancy Rate:        %.1f%%", p.VacancyRate),
		fmt.Sprintf("Owner Occupied:      %.1f%%", p.OwnerOccupiedPct),
		fmt.Sprintf("Renter Occupied:     %.1f%%", p.RenterOccupiedPct),
	}
	cli.PrintText(lines)
	return nil
}

// printTractProfiles outputs a slice of TractProfiles as JSON or a text table.
func printTractProfiles(cmd *cobra.Command, profiles []TractProfile) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(profiles)
	}

	if len(profiles) == 0 {
		fmt.Println("No tracts found.")
		return nil
	}

	lines := make([]string, 0, len(profiles)+1)
	lines = append(lines, fmt.Sprintf("%-14s  %-8s  %-10s  %-10s  %-10s",
		"TRACT", "POP", "INCOME", "RENT", "VACANCY%"))
	for _, p := range profiles {
		lines = append(lines, fmt.Sprintf("%-14s  %-8d  %-10s  %-10s  %-10.1f",
			p.Tract, p.Population, formatDollars(p.MedianIncome),
			formatDollars(p.MedianRent), p.VacancyRate))
	}
	cli.PrintText(lines)
	return nil
}

// printBoroughSummary outputs a BoroughSummary as JSON or formatted text.
func printBoroughSummary(cmd *cobra.Command, s BoroughSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(s)
	}

	lines := []string{
		fmt.Sprintf("Borough:               %s", s.Borough),
		fmt.Sprintf("Total Population:      %d", s.TotalPopulation),
		fmt.Sprintf("Avg Median Income:     %s", formatDollars(s.AvgMedianIncome)),
		fmt.Sprintf("Avg Median Rent:       %s", formatDollars(s.AvgMedianRent)),
		fmt.Sprintf("Avg Vacancy Rate:      %.1f%%", s.AvgVacancyRate),
		fmt.Sprintf("Total Tracts:          %d", s.TotalTracts),
		fmt.Sprintf("High Vacancy Tracts:   %d", s.HighVacancyTracts),
		fmt.Sprintf("High Rent Burden:      %d", s.HighRentBurdenTracts),
	}
	cli.PrintText(lines)
	return nil
}

// formatDollars formats an int as "$1,234,567". Returns "$0" for zero.
func formatDollars(amount int) string {
	if amount == 0 {
		return "$0"
	}
	s := strconv.Itoa(amount)
	n := len(s)
	if n <= 3 {
		return "$" + s
	}
	var parts []string
	for n > 3 {
		parts = append([]string{s[n-3 : n]}, parts...)
		n -= 3
	}
	parts = append([]string{s[:n]}, parts...)
	return "$" + strings.Join(parts, ",")
}
