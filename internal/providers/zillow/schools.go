package zillow

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newSchoolsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schools",
		Short:   "View nearby schools",
		Aliases: []string{"school"},
	}

	cmd.AddCommand(newSchoolsNearbyCmd(factory))

	return cmd
}

func newSchoolsNearbyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nearby",
		Short: "Get schools near a property",
		RunE:  makeRunSchoolsNearby(factory),
	}
	cmd.Flags().String("zpid", "", "Zillow property ID")
	cmd.Flags().Int("limit", 10, "Maximum results")
	cmd.MarkFlagRequired("zpid")
	return cmd
}

func makeRunSchoolsNearby(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		zpid, _ := cmd.Flags().GetString("zpid")
		limit, _ := cmd.Flags().GetInt("limit")

		schools, err := fetchNearbySchools(ctx, client, zpid)
		if err != nil {
			return fmt.Errorf("get schools: %w", err)
		}

		if limit > 0 && len(schools) > limit {
			schools = schools[:limit]
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(schools)
		}

		if len(schools) == 0 {
			fmt.Println("No schools found.")
			return nil
		}
		lines := []string{fmt.Sprintf("Schools near %s:", zpid)}
		lines = append(lines, fmt.Sprintf("  %-30s  %-6s  %-12s  %-8s  %-10s  %-5s",
			"NAME", "RATING", "LEVEL", "TYPE", "GRADES", "DIST"))
		for _, s := range schools {
			rating := "-"
			if s.Rating > 0 {
				rating = fmt.Sprintf("%d/10", s.Rating)
			}
			dist := "-"
			if s.Distance > 0 {
				dist = fmt.Sprintf("%.1fmi", s.Distance)
			}
			lines = append(lines, fmt.Sprintf("  %-30s  %-6s  %-12s  %-8s  %-10s  %-5s",
				truncate(s.Name, 30), rating, s.Level, s.Type, s.Grades, dist))
		}
		cli.PrintText(lines)
		return nil
	}
}

// fetchNearbySchools gets schools from the property detail response.
func fetchNearbySchools(ctx context.Context, client *Client, zpid string) ([]SchoolSummary, error) {
	detail, err := fetchPropertyDetail(ctx, client, zpid)
	if err != nil {
		return nil, err
	}
	return detail.Schools, nil
}
