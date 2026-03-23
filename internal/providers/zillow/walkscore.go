package zillow

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newWalkScoreCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "walkscore",
		Short:   "View walk, transit, and bike scores",
		Aliases: []string{"ws"},
	}

	cmd.AddCommand(newWalkScoreGetCmd(factory))

	return cmd
}

func newWalkScoreGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get walk/transit/bike scores for a property",
		RunE:  makeRunWalkScoreGet(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunWalkScoreGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")

		reqURL := client.baseURL + "/graphql/?zpid=" + zpid
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get walk score: %w", err)
		}

		result, err := parseWalkScore(body, zpid)
		if err != nil {
			return fmt.Errorf("parse walk score: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		lines := []string{fmt.Sprintf("Scores for %s:", zpid)}
		if result.WalkScore > 0 {
			lines = append(lines, fmt.Sprintf("  Walk Score:    %d — %s", result.WalkScore, result.WalkDesc))
		}
		if result.TransitScore > 0 {
			lines = append(lines, fmt.Sprintf("  Transit Score: %d — %s", result.TransitScore, result.TransitDesc))
		}
		if result.BikeScore > 0 {
			lines = append(lines, fmt.Sprintf("  Bike Score:    %d — %s", result.BikeScore, result.BikeDesc))
		}
		cli.PrintText(lines)
		return nil
	}
}

// parseWalkScore extracts walk/transit/bike scores from the GraphQL response.
func parseWalkScore(body []byte, zpid string) (WalkScoreResult, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return WalkScoreResult{}, err
	}

	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return WalkScoreResult{ZPID: zpid}, nil
	}
	prop, _ := data["property"].(map[string]any)
	if prop == nil {
		return WalkScoreResult{ZPID: zpid}, nil
	}

	result := WalkScoreResult{ZPID: zpid}

	if ws, ok := prop["walkScore"].(map[string]any); ok {
		if score, ok := ws["walkscore"].(float64); ok {
			result.WalkScore = int(score)
		}
		result.WalkDesc = jsonStr(ws, "description")
	}
	if ts, ok := prop["transitScore"].(map[string]any); ok {
		if score, ok := ts["transit_score"].(float64); ok {
			result.TransitScore = int(score)
		}
		result.TransitDesc = jsonStr(ts, "description")
	}
	if bs, ok := prop["bikeScore"].(map[string]any); ok {
		if score, ok := bs["bike_score"].(float64); ok {
			result.BikeScore = int(score)
		}
		result.BikeDesc = jsonStr(bs, "description")
	}

	return result, nil
}
