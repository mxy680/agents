package census

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newTractsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tracts",
		Short:   "Query census tract demographic data",
		Aliases: []string{"tract", "tr"},
	}

	cmd.AddCommand(newTractsProfileCmd(factory))
	cmd.AddCommand(newTractsCompareCmd(factory))
	cmd.AddCommand(newTractsSummaryCmd(factory))

	return cmd
}

// newTractsProfileCmd returns the `tracts profile` subcommand.
func newTractsProfileCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Get demographic profile for a specific census tract",
		RunE:  makeRunTractsProfile(factory),
	}
	cmd.Flags().String("tract", "", "11-digit FIPS census tract code (e.g. 36005000100)")
	cmd.Flags().Int("year", 2023, "ACS survey year (default 2023)")
	if err := cmd.MarkFlagRequired("tract"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunTractsProfile(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		tractFIPS, _ := cmd.Flags().GetString("tract")
		year, _ := cmd.Flags().GetInt("year")

		state, county, tract, err := parseTractFIPS(tractFIPS)
		if err != nil {
			return err
		}

		baseURL := buildYearURL(client.baseURL, year)
		yearClient := &Client{
			httpClient: client.httpClient,
			baseURL:    baseURL,
			apiKey:     client.apiKey,
		}

		data, err := yearClient.Query(ctx, allVars,
			"tract:"+tract,
			[]string{"state:" + state, "county:" + county},
		)
		if err != nil {
			return fmt.Errorf("query census API: %w", err)
		}

		rows := rowsToMaps(data)
		if len(rows) == 0 {
			return fmt.Errorf("no data found for tract %s", tractFIPS)
		}

		profile := rowToProfile(rows[0], tractFIPS)
		return printTractProfile(cmd, profile)
	}
}

// newTractsCompareCmd returns the `tracts compare` subcommand.
func newTractsCompareCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare demographics across all tracts in a borough",
		RunE:  makeRunTractsCompare(factory),
	}
	cmd.Flags().String("borough", "", "NYC borough: bronx, brooklyn, manhattan, queens, staten-island")
	cmd.Flags().Int("year", 2023, "ACS survey year (default 2023)")
	cmd.Flags().String("sort", "income", "Sort field: income, rent, vacancy, population")
	cmd.Flags().Int("limit", 20, "Number of results to return (default 20)")
	if err := cmd.MarkFlagRequired("borough"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunTractsCompare(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		borough, _ := cmd.Flags().GetString("borough")
		year, _ := cmd.Flags().GetInt("year")
		sortBy, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")

		countyFIPS, err := lookupBoroughFIPS(borough)
		if err != nil {
			return err
		}

		if !isValidSortField(sortBy) {
			return fmt.Errorf("invalid sort field %q; valid options: income, rent, vacancy, population", sortBy)
		}

		baseURL := buildYearURL(client.baseURL, year)
		yearClient := &Client{
			httpClient: client.httpClient,
			baseURL:    baseURL,
			apiKey:     client.apiKey,
		}

		data, err := yearClient.Query(ctx, allVars,
			"tract:*",
			[]string{"state:36", "county:" + countyFIPS},
		)
		if err != nil {
			return fmt.Errorf("query census API: %w", err)
		}

		rows := rowsToMaps(data)
		profiles := make([]TractProfile, 0, len(rows))
		for _, row := range rows {
			fullFIPS := "36" + countyFIPS + row["tract"]
			profiles = append(profiles, rowToProfile(row, fullFIPS))
		}

		sortProfiles(profiles, sortBy)

		if limit > 0 && len(profiles) > limit {
			profiles = profiles[:limit]
		}

		return printTractProfiles(cmd, profiles)
	}
}

// newTractsSummaryCmd returns the `tracts summary` subcommand.
func newTractsSummaryCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Get borough-wide aggregate statistics",
		RunE:  makeRunTractsSummary(factory),
	}
	cmd.Flags().String("borough", "", "NYC borough: bronx, brooklyn, manhattan, queens, staten-island")
	cmd.Flags().Int("year", 2023, "ACS survey year (default 2023)")
	if err := cmd.MarkFlagRequired("borough"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunTractsSummary(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		borough, _ := cmd.Flags().GetString("borough")
		year, _ := cmd.Flags().GetInt("year")

		countyFIPS, err := lookupBoroughFIPS(borough)
		if err != nil {
			return err
		}

		baseURL := buildYearURL(client.baseURL, year)
		yearClient := &Client{
			httpClient: client.httpClient,
			baseURL:    baseURL,
			apiKey:     client.apiKey,
		}

		data, err := yearClient.Query(ctx, allVars,
			"tract:*",
			[]string{"state:36", "county:" + countyFIPS},
		)
		if err != nil {
			return fmt.Errorf("query census API: %w", err)
		}

		rows := rowsToMaps(data)
		profiles := make([]TractProfile, 0, len(rows))
		for _, row := range rows {
			fullFIPS := "36" + countyFIPS + row["tract"]
			profiles = append(profiles, rowToProfile(row, fullFIPS))
		}

		summary := buildBoroughSummary(strings.ToLower(borough), profiles)
		return printBoroughSummary(cmd, summary)
	}
}

// buildBoroughSummary aggregates tract profiles into a BoroughSummary.
func buildBoroughSummary(borough string, profiles []TractProfile) BoroughSummary {
	if len(profiles) == 0 {
		return BoroughSummary{Borough: borough}
	}

	var totalPop, sumIncome, sumRent int
	var sumVacancy float64
	var highVacancy, highRentBurden int

	for _, p := range profiles {
		totalPop += p.Population
		sumIncome += p.MedianIncome
		sumRent += p.MedianRent
		sumVacancy += p.VacancyRate

		if p.VacancyRate > 10.0 {
			highVacancy++
		}
		// High rent burden: monthly rent > 30% of monthly income
		if p.MedianIncome > 0 && p.MedianRent > 0 {
			monthlyIncome := float64(p.MedianIncome) / 12
			rentBurden := float64(p.MedianRent) / monthlyIncome
			if rentBurden > 0.30 {
				highRentBurden++
			}
		}
	}

	n := len(profiles)
	return BoroughSummary{
		Borough:              borough,
		TotalPopulation:      totalPop,
		AvgMedianIncome:      sumIncome / n,
		AvgMedianRent:        sumRent / n,
		AvgVacancyRate:       roundTwo(sumVacancy / float64(n)),
		TotalTracts:          n,
		HighVacancyTracts:    highVacancy,
		HighRentBurdenTracts: highRentBurden,
	}
}

// sortProfiles sorts profiles by the given field in descending order.
func sortProfiles(profiles []TractProfile, field string) {
	sort.Slice(profiles, func(i, j int) bool {
		switch field {
		case "rent":
			return profiles[i].MedianRent > profiles[j].MedianRent
		case "vacancy":
			return profiles[i].VacancyRate > profiles[j].VacancyRate
		case "population":
			return profiles[i].Population > profiles[j].Population
		default: // "income"
			return profiles[i].MedianIncome > profiles[j].MedianIncome
		}
	})
}

// isValidSortField returns true if the field name is a valid sort option.
func isValidSortField(field string) bool {
	switch field {
	case "income", "rent", "vacancy", "population":
		return true
	}
	return false
}

// buildYearURL replaces the year component in the base URL if year != 2023.
// The default base URL already contains "2023"; for other years we swap it.
func buildYearURL(baseURL string, year int) string {
	if year == 2023 {
		return baseURL
	}
	yearStr := fmt.Sprintf("%d", year)
	return strings.Replace(baseURL, "2023", yearStr, 1)
}
