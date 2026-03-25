package nysla

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newLicensesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "licenses",
		Short:   "Query NY SLA liquor licenses",
		Aliases: []string{"license", "lic"},
	}

	cmd.AddCommand(newLicensesSearchCmd(factory))
	cmd.AddCommand(newLicensesCountCmd(factory))
	cmd.AddCommand(newLicensesDensityCmd(factory))

	return cmd
}

// newLicensesSearchCmd returns the `licenses search` subcommand.
func newLicensesSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search active liquor licenses by location",
		RunE:  makeRunLicensesSearch(factory),
	}
	cmd.Flags().String("borough", "", "NYC borough: bronx, brooklyn, manhattan, queens, staten-island (required)")
	cmd.Flags().String("zip", "", "Filter by ZIP code")
	cmd.Flags().String("type", "", "License type filter: restaurant, bar, tavern, liquor-store, wine")
	cmd.Flags().String("since", "", "Only licenses effective after this date (YYYY-MM-DD)")
	cmd.Flags().Int("limit", 50, "Maximum number of results (default 50)")
	if err := cmd.MarkFlagRequired("borough"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunLicensesSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		borough, _ := cmd.Flags().GetString("borough")
		zip, _ := cmd.Flags().GetString("zip")
		licType, _ := cmd.Flags().GetString("type")
		since, _ := cmd.Flags().GetString("since")
		limit, _ := cmd.Flags().GetInt("limit")

		county, err := lookupCounty(borough)
		if err != nil {
			return err
		}

		where := fmt.Sprintf("county_name='%s'", county)
		if zip != "" {
			where += fmt.Sprintf(" AND zip='%s'", zip)
		}
		if since != "" {
			where += fmt.Sprintf(" AND effective_date > '%s'", since)
		}
		if licType != "" {
			fragment, ok := licenseTypeFilter[strings.ToLower(licType)]
			if !ok {
				return fmt.Errorf("unknown license type %q; valid options: restaurant, bar, tavern, liquor-store, wine", licType)
			}
			where += fmt.Sprintf(" AND upper(license_type_name) LIKE '%%%s%%'", fragment)
		}

		params := url.Values{}
		params.Set("$where", where)
		params.Set("$limit", fmt.Sprintf("%d", limit))
		params.Set("$order", "effective_date DESC")

		body, err := client.Get(ctx, params)
		if err != nil {
			return fmt.Errorf("fetch licenses: %w", err)
		}

		var records []rawLicense
		if err := json.Unmarshal(body, &records); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		summaries := make([]LicenseSummary, 0, len(records))
		for _, r := range records {
			summaries = append(summaries, toSummary(r))
		}

		return printLicenseSummaries(cmd, summaries)
	}
}

// newLicensesCountCmd returns the `licenses count` subcommand.
func newLicensesCountCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Count new liquor licenses issued since a date (gentrification signal)",
		RunE:  makeRunLicensesCount(factory),
	}
	cmd.Flags().String("borough", "", "NYC borough: bronx, brooklyn, manhattan, queens, staten-island (required)")
	cmd.Flags().String("zip", "", "Filter by ZIP code")
	cmd.Flags().String("since", "", "Count licenses issued after this date, e.g. 2025-01-01 (required)")
	if err := cmd.MarkFlagRequired("borough"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("since"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunLicensesCount(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		borough, _ := cmd.Flags().GetString("borough")
		zip, _ := cmd.Flags().GetString("zip")
		since, _ := cmd.Flags().GetString("since")

		county, err := lookupCounty(borough)
		if err != nil {
			return err
		}

		where := fmt.Sprintf("county_name='%s' AND effective_date > '%s'", county, since)
		if zip != "" {
			where += fmt.Sprintf(" AND zip='%s'", zip)
		}

		params := url.Values{}
		params.Set("$where", where)
		params.Set("$limit", "5000")
		params.Set("$order", "effective_date DESC")

		body, err := client.Get(ctx, params)
		if err != nil {
			return fmt.Errorf("fetch licenses: %w", err)
		}

		var records []rawLicense
		if err := json.Unmarshal(body, &records); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		result := CountResult{
			Borough:     borough,
			ZIP:         zip,
			Since:       since,
			NewLicenses: len(records),
			Breakdown:   buildTypeBreakdown(records),
		}

		return printCountResult(cmd, result)
	}
}

// newLicensesDensityCmd returns the `licenses density` subcommand.
func newLicensesDensityCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "density",
		Short: "Count active liquor licenses in a ZIP code (commercial density signal)",
		RunE:  makeRunLicensesDensity(factory),
	}
	cmd.Flags().String("zip", "", "ZIP code to query (required)")
	cmd.Flags().String("borough", "", "NYC borough: bronx, brooklyn, manhattan, queens, staten-island (required)")
	if err := cmd.MarkFlagRequired("zip"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("borough"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunLicensesDensity(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zip, _ := cmd.Flags().GetString("zip")
		borough, _ := cmd.Flags().GetString("borough")

		county, err := lookupCounty(borough)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("$where", fmt.Sprintf("county_name='%s' AND zip='%s'", county, zip))
		params.Set("$limit", "5000")

		body, err := client.Get(ctx, params)
		if err != nil {
			return fmt.Errorf("fetch licenses: %w", err)
		}

		var records []rawLicense
		if err := json.Unmarshal(body, &records); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		result := DensityResult{
			ZIP:         zip,
			TotalActive: len(records),
			ByType:      buildTypeBreakdown(records),
		}

		return printDensityResult(cmd, result)
	}
}
