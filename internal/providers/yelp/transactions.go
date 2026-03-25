package yelp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newTransactionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transactions",
		Short:   "Search businesses by transaction type (delivery, pickup)",
		Aliases: []string{"transaction", "tx"},
	}

	cmd.AddCommand(newTransactionSearchCmd(factory))

	return cmd
}

func newTransactionSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search businesses that support a specific transaction type",
		RunE:  makeRunTransactionSearch(factory),
	}
	cmd.Flags().String("type", "", "Transaction type: delivery or pickup")
	cmd.Flags().String("location", "", "Location text (e.g., 'San Francisco, CA')")
	cmd.Flags().Float64("latitude", 0, "Latitude for geo-based search")
	cmd.Flags().Float64("longitude", 0, "Longitude for geo-based search")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func makeRunTransactionSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		txType, _ := cmd.Flags().GetString("type")
		location, _ := cmd.Flags().GetString("location")
		latitude, _ := cmd.Flags().GetFloat64("latitude")
		longitude, _ := cmd.Flags().GetFloat64("longitude")

		if location == "" && (latitude == 0 || longitude == 0) {
			return fmt.Errorf("either --location or both --latitude and --longitude are required")
		}

		params := url.Values{}
		if location != "" {
			params.Set("location", location)
		}
		if latitude != 0 {
			params.Set("latitude", strconv.FormatFloat(latitude, 'f', -1, 64))
		}
		if longitude != 0 {
			params.Set("longitude", strconv.FormatFloat(longitude, 'f', -1, 64))
		}

		body, err := client.doYelp(ctx, "GET", "/transactions/"+txType+"/search", params)
		if err != nil {
			return fmt.Errorf("transaction search: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var resp struct {
			Businesses []BusinessSummary `json:"businesses"`
			Total      int               `json:"total"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		return printBusinessSummaries(cmd, resp.Businesses)
	}
}
