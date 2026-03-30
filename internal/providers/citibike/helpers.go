package citibike

import (
	"fmt"
	"math"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// StationSummary is a simplified station view including haversine distance.
type StationSummary struct {
	Name      string  `json:"name"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Capacity  int     `json:"capacity"`
	DistanceM float64 `json:"distance_m"`
}

// DensitySummary summarises station density within a radius.
type DensitySummary struct {
	Count         int     `json:"count"`
	AvgCapacity   float64 `json:"avg_capacity"`
	TotalCapacity int     `json:"total_capacity"`
	RadiusM       float64 `json:"radius_m"`
}

// haversineMeters returns the great-circle distance in metres between two
// latitude/longitude coordinates.
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in metres
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// printStationSummaries outputs station summaries as JSON or a text table.
func printStationSummaries(cmd *cobra.Command, summaries []StationSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No stations found within the specified radius.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-10s  %-10s  %-10s  %-10s",
		"NAME", "LAT", "LNG", "CAPACITY", "DIST (m)"))
	for _, s := range summaries {
		name := truncateName(s.Name, 40)
		lines = append(lines, fmt.Sprintf("%-40s  %-10.6f  %-10.6f  %-10d  %-10.0f",
			name, s.Lat, s.Lng, s.Capacity, s.DistanceM))
	}
	cli.PrintText(lines)
	return nil
}

// printDensitySummary outputs a density summary as JSON or a text table.
func printDensitySummary(cmd *cobra.Command, d DensitySummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(d)
	}

	lines := []string{
		fmt.Sprintf("%-20s  %d", "Station count:", d.Count),
		fmt.Sprintf("%-20s  %.1f", "Avg capacity:", d.AvgCapacity),
		fmt.Sprintf("%-20s  %d", "Total capacity:", d.TotalCapacity),
		fmt.Sprintf("%-20s  %.0f m", "Search radius:", d.RadiusM),
	}
	cli.PrintText(lines)
	return nil
}

// truncateName shortens s to at most max runes, appending "..." if truncated.
func truncateName(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}
