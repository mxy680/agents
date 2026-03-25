package hmda

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

func newLoansCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "loans",
		Short:   "Query HMDA loan origination data",
		Aliases: []string{"loan", "ln"},
	}

	cmd.AddCommand(newLoansSummaryCmd(factory))
	cmd.AddCommand(newLoansTractCmd(factory))

	return cmd
}

// newLoansSummaryCmd returns the `loans summary` subcommand.
func newLoansSummaryCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Get loan origination summary by county",
		RunE:  makeRunLoansSummary(factory),
	}
	cmd.Flags().String("county", "", "NYC county: bronx, brooklyn, manhattan, queens, staten-island")
	cmd.Flags().Int("year", 2023, "Filing year")
	cmd.Flags().String("purpose", "", "Loan purpose: purchase, refinance (default: all)")
	if err := cmd.MarkFlagRequired("county"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunLoansSummary(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		county, _ := cmd.Flags().GetString("county")
		year, _ := cmd.Flags().GetInt("year")
		purpose, _ := cmd.Flags().GetString("purpose")

		fips, err := lookupFIPS(county)
		if err != nil {
			return err
		}

		params := url.Values{
			"counties":      []string{fips},
			"actions_taken": []string{"1"},
			"years":         []string{fmt.Sprintf("%d", year)},
		}

		if purpose != "" {
			lp, err := purposeCode(purpose)
			if err != nil {
				return err
			}
			params.Set("loan_purposes", lp)
		}

		body, err := client.Get(ctx, "aggregations", params)
		if err != nil {
			return fmt.Errorf("fetch aggregations: %w", err)
		}

		var resp AggregationResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		summary := buildCountySummary(countyNameForFIPS(fips), fips, year, resp.Aggregations)
		return printCountySummary(cmd, summary)
	}
}

// newLoansTractCmd returns the `loans tract` subcommand.
func newLoansTractCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tract",
		Short: "Get loan origination data for a specific census tract",
		RunE:  makeRunLoansTract(factory),
	}
	cmd.Flags().String("tract", "", "11-digit FIPS census tract code (e.g. 36005000100)")
	cmd.Flags().Int("year", 2023, "Filing year")
	if err := cmd.MarkFlagRequired("tract"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunLoansTract(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		tract, _ := cmd.Flags().GetString("tract")
		year, _ := cmd.Flags().GetInt("year")

		if len(tract) != 11 {
			return fmt.Errorf("tract must be an 11-digit FIPS code (e.g. 36005000100), got %q", tract)
		}

		// Derive the 5-digit county FIPS from the first 5 digits of the tract code.
		countyFIPSCode := tract[:5]

		params := url.Values{
			"counties":      []string{countyFIPSCode},
			"actions_taken": []string{"1"},
			"years":         []string{fmt.Sprintf("%d", year)},
		}

		body, err := client.Get(ctx, "aggregations", params)
		if err != nil {
			return fmt.Errorf("fetch aggregations: %w", err)
		}

		var resp AggregationResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		// Find the specific tract in the response.
		for _, item := range resp.Aggregations {
			if item.CensusTract == tract {
				avg := int64(0)
				if item.Count > 0 {
					avg = item.Sum / int64(item.Count)
				}
				summary := TractSummary{
					Tract:       item.CensusTract,
					Count:       item.Count,
					Volume:      item.Sum,
					AvgLoanSize: avg,
				}
				return printTractSummary(cmd, summary)
			}
		}

		// Tract exists but had no originations.
		summary := TractSummary{
			Tract:       tract,
			Count:       0,
			Volume:      0,
			AvgLoanSize: 0,
		}
		return printTractSummary(cmd, summary)
	}
}

// purposeCode maps human-readable loan purpose names to HMDA API codes.
func purposeCode(purpose string) (string, error) {
	switch purpose {
	case "purchase":
		return "1", nil
	case "refinance":
		return "31", nil
	default:
		return "", fmt.Errorf("unknown loan purpose %q; valid options: purchase, refinance", purpose)
	}
}
