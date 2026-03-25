package trends

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// newInterestCmd returns the `interest` subcommand group.
func newInterestCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "interest",
		Short:   "Query Google Trends interest data",
		Aliases: []string{"int"},
	}

	cmd.AddCommand(newInterestSearchCmd(factory))
	cmd.AddCommand(newInterestCompareCmd(factory))
	cmd.AddCommand(newInterestMomentumCmd(factory))

	return cmd
}

// newInterestSearchCmd returns the `interest search` subcommand.
func newInterestSearchCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Get interest-over-time for a keyword",
		RunE:  makeRunInterestSearch(factory),
	}
	cmd.Flags().String("keyword", "", "Keyword to search (e.g. \"mott haven apartments\")")
	cmd.Flags().String("geo", "US-NY", "Geographic region (e.g. \"US-NY\", \"US\")")
	cmd.Flags().String("time", "today 12-m", "Time range: today 1-m, today 3-m, today 12-m, today 5-y")
	_ = cmd.MarkFlagRequired("keyword")
	return cmd
}

func makeRunInterestSearch(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}

		keyword, _ := cmd.Flags().GetString("keyword")
		geo, _ := cmd.Flags().GetString("geo")
		timeRange, _ := cmd.Flags().GetString("time")

		points, err := svc.InterestOverTime(ctx, keyword, geo, timeRange)
		if err != nil {
			return fmt.Errorf("get interest over time: %w", err)
		}

		return printTimePoints(cmd, points)
	}
}

// newInterestCompareCmd returns the `interest compare` subcommand.
func newInterestCompareCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare interest across multiple keywords",
		RunE:  makeRunInterestCompare(factory),
	}
	cmd.Flags().String("keywords", "", "Comma-separated keywords (e.g. \"mott haven,east new york,bed stuy\")")
	cmd.Flags().String("geo", "US-NY", "Geographic region (e.g. \"US-NY\", \"US\")")
	cmd.Flags().String("time", "today 12-m", "Time range: today 1-m, today 3-m, today 12-m, today 5-y")
	_ = cmd.MarkFlagRequired("keywords")
	return cmd
}

func makeRunInterestCompare(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}

		keywordsRaw, _ := cmd.Flags().GetString("keywords")
		geo, _ := cmd.Flags().GetString("geo")
		timeRange, _ := cmd.Flags().GetString("time")

		keywords := splitKeywords(keywordsRaw)
		if len(keywords) == 0 {
			return fmt.Errorf("--keywords must contain at least one keyword")
		}

		results, err := svc.Compare(ctx, keywords, geo, timeRange)
		if err != nil {
			return fmt.Errorf("compare keywords: %w", err)
		}

		return printCompareResults(cmd, results)
	}
}

// newInterestMomentumCmd returns the `interest momentum` subcommand.
func newInterestMomentumCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "momentum",
		Short: "Calculate momentum score for a keyword (rising/stable/declining)",
		RunE:  makeRunInterestMomentum(factory),
	}
	cmd.Flags().String("keyword", "", "Keyword to analyze (e.g. \"mott haven\")")
	cmd.Flags().String("geo", "US-NY", "Geographic region (e.g. \"US-NY\", \"US\")")
	_ = cmd.MarkFlagRequired("keyword")
	return cmd
}

func makeRunInterestMomentum(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}

		keyword, _ := cmd.Flags().GetString("keyword")
		geo, _ := cmd.Flags().GetString("geo")

		// Always fetch 12-month data for momentum calculation.
		points, err := svc.InterestOverTime(ctx, keyword, geo, "today 12-m")
		if err != nil {
			return fmt.Errorf("get interest over time: %w", err)
		}

		result, err := calculateMomentum(keyword, points)
		if err != nil {
			return err
		}

		return printMomentumResult(cmd, result)
	}
}

// calculateMomentum computes the momentum result from a time series.
// It requires at least 6 data points to compare the first and last 3 months.
func calculateMomentum(keyword string, points []TimePoint) (MomentumResult, error) {
	if len(points) < 6 {
		return MomentumResult{}, fmt.Errorf("insufficient data: need at least 6 data points, got %d", len(points))
	}

	earlierAvg := avgValues(points[:3])
	recentAvg := avgValues(points[len(points)-3:])

	var momentumPct float64
	if earlierAvg == 0 {
		momentumPct = 0
	} else {
		momentumPct = ((recentAvg - earlierAvg) / earlierAvg) * 100
	}

	trend := classifyTrend(momentumPct)

	return MomentumResult{
		Keyword:     keyword,
		RecentAvg:   recentAvg,
		EarlierAvg:  earlierAvg,
		MomentumPct: momentumPct,
		Trend:       trend,
	}, nil
}

// avgValues returns the average of TimePoint values in the slice.
func avgValues(points []TimePoint) float64 {
	if len(points) == 0 {
		return 0
	}
	sum := 0
	for _, p := range points {
		sum += p.Value
	}
	return float64(sum) / float64(len(points))
}

// classifyTrend returns "rising", "declining", or "stable" based on momentum percentage.
func classifyTrend(momentumPct float64) string {
	switch {
	case momentumPct > 15:
		return "rising"
	case momentumPct < -15:
		return "declining"
	default:
		return "stable"
	}
}

// splitKeywords splits a comma-separated keyword string into trimmed, non-empty parts.
func splitKeywords(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
