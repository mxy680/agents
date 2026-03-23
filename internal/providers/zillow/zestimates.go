package zillow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newZestimatesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "zestimates",
		Short:   "View Zestimate home value estimates",
		Aliases: []string{"zestimate", "zest"},
	}

	cmd.AddCommand(newZestimateGetCmd(factory))
	cmd.AddCommand(newZestimateRentCmd(factory))
	cmd.AddCommand(newZestimateChartCmd(factory))

	return cmd
}

func newZestimateGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get Zestimate for a property",
		RunE:  makeRunZestimateGet(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunZestimateGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")

		zest, err := fetchZestimate(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get zestimate: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(zest)
		}

		lines := []string{
			fmt.Sprintf("ZPID:          %s", zest.ZPID),
			fmt.Sprintf("Address:       %s", zest.Address),
			fmt.Sprintf("Zestimate:     %s", formatPrice(zest.Zestimate)),
			fmt.Sprintf("Rent Zest:     %s/mo", formatPrice(zest.RentZestimate)),
		}
		if zest.ValueLow > 0 {
			lines = append(lines, fmt.Sprintf("Range:         %s - %s", formatPrice(zest.ValueLow), formatPrice(zest.ValueHigh)))
		}
		if zest.ValueChange != 0 {
			lines = append(lines, fmt.Sprintf("30-Day Change: %s", formatPrice(zest.ValueChange)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newZestimateRentCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rent",
		Short: "Get Rent Zestimate for a property",
		RunE:  makeRunZestimateRent(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunZestimateRent(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")

		zest, err := fetchZestimate(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get rent zestimate: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"zpid":          zest.ZPID,
				"address":       zest.Address,
				"rentZestimate": zest.RentZestimate,
			})
		}

		lines := []string{
			fmt.Sprintf("ZPID:          %s", zest.ZPID),
			fmt.Sprintf("Address:       %s", zest.Address),
			fmt.Sprintf("Rent Zestimate: %s/mo", formatPrice(zest.RentZestimate)),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newZestimateChartCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chart",
		Short: "Get Zestimate value chart data",
		RunE:  makeRunZestimateChart(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.Flags().String("duration", "1y", "Chart duration: 1y, 5y, 10y")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunZestimateChart(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")
		duration, _ := cmd.Flags().GetString("duration")

		points, err := fetchZestimateChart(ctx, client, zpid, duration)
		if err != nil {
			return fmt.Errorf("get zestimate chart: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"zpid":     zpid,
				"duration": duration,
				"points":   points,
			})
		}

		if len(points) == 0 {
			fmt.Println("No chart data found.")
			return nil
		}
		lines := []string{fmt.Sprintf("Zestimate Chart for %s (%s):", zpid, duration)}
		lines = append(lines, fmt.Sprintf("  %-12s  %s", "DATE", "VALUE"))
		for _, p := range points {
			lines = append(lines, fmt.Sprintf("  %-12s  %s", p.Date, formatPrice(p.Value)))
		}
		cli.PrintText(lines)
		return nil
	}
}

// fetchZestimate gets Zestimate data from the property detail endpoint.
func fetchZestimate(ctx context.Context, client *Client, zpid string) (ZestimateSummary, error) {
	detail, err := fetchPropertyDetail(ctx, client, zpid)
	if err != nil {
		return ZestimateSummary{}, err
	}
	return ZestimateSummary{
		ZPID:          detail.ZPID,
		Address:       detail.Address,
		Zestimate:     detail.Zestimate,
		RentZestimate: detail.RentZestimate,
	}, nil
}

// fetchZestimateChart gets chart data. Uses the property detail endpoint
// since the chart API endpoint is not publicly documented.
func fetchZestimateChart(ctx context.Context, client *Client, zpid string, duration string) ([]ZestimateChartPoint, error) {
	reqURL := client.baseURL + "/graphql/?zpid=" + zpid
	body, err := client.Get(ctx, reqURL)
	if err != nil {
		return nil, err
	}

	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return nil, nil
	}
	prop, _ := data["property"].(map[string]any)
	if prop == nil {
		return nil, nil
	}

	// Try to get chart data from the homeValueChartData field
	chartData, _ := prop["homeValueChartData"].([]any)
	var points []ZestimateChartPoint
	for _, item := range chartData {
		if m, ok := item.(map[string]any); ok {
			p := ZestimateChartPoint{
				Date: jsonStr(m, "date"),
			}
			if v, ok := m["value"].(float64); ok {
				p.Value = int64(v)
			}
			points = append(points, p)
		}
	}
	return points, nil
}
