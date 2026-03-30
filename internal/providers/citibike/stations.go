package citibike

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

// gbfsStationInformation maps to the station_information.json GBFS feed.
type gbfsStationInformation struct {
	Data struct {
		Stations []gbfsStationInfo `json:"stations"`
	} `json:"data"`
}

// gbfsStationInfo represents a single station's static information.
type gbfsStationInfo struct {
	StationID string  `json:"station_id"`
	Name      string  `json:"name"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Capacity  int     `json:"capacity"`
}

func newStationsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stations",
		Short:   "Search and analyse Citi Bike stations",
		Aliases: []string{"station", "st"},
	}

	cmd.AddCommand(newStationsSearchCmd(factory))
	cmd.AddCommand(newStationsDensityCmd(factory))

	return cmd
}

// newStationsSearchCmd returns the `stations search` subcommand.
func newStationsSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Search for Citi Bike stations near a location",
		Aliases: []string{"nearby"},
		RunE:    makeRunStationsSearch(factory),
	}
	cmd.Flags().Float64("lat", 0, "Latitude of the search centre (required)")
	cmd.Flags().Float64("lng", 0, "Longitude of the search centre (required)")
	cmd.Flags().Float64("radius", 500, "Search radius in metres (default 500)")
	cmd.MarkFlagRequired("lat")
	cmd.MarkFlagRequired("lng")
	return cmd
}

func makeRunStationsSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		lat, _ := cmd.Flags().GetFloat64("lat")
		lng, _ := cmd.Flags().GetFloat64("lng")
		radius, _ := cmd.Flags().GetFloat64("radius")

		var feed gbfsStationInformation
		if err := client.GetJSON(ctx, "station_information.json", &feed); err != nil {
			return fmt.Errorf("fetch station information: %w", err)
		}

		summaries := filterStationsByRadius(feed.Data.Stations, lat, lng, radius)
		return printStationSummaries(cmd, summaries)
	}
}

// newStationsDensityCmd returns the `stations density` subcommand.
func newStationsDensityCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "density",
		Short: "Count Citi Bike stations within a radius (transit accessibility signal)",
		RunE:  makeRunStationsDensity(factory),
	}
	cmd.Flags().Float64("lat", 0, "Latitude of the search centre (required)")
	cmd.Flags().Float64("lng", 0, "Longitude of the search centre (required)")
	cmd.Flags().Float64("radius", 1000, "Search radius in metres (default 1000)")
	cmd.MarkFlagRequired("lat")
	cmd.MarkFlagRequired("lng")
	return cmd
}

func makeRunStationsDensity(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		lat, _ := cmd.Flags().GetFloat64("lat")
		lng, _ := cmd.Flags().GetFloat64("lng")
		radius, _ := cmd.Flags().GetFloat64("radius")

		var feed gbfsStationInformation
		if err := client.GetJSON(ctx, "station_information.json", &feed); err != nil {
			return fmt.Errorf("fetch station information: %w", err)
		}

		density := computeDensity(feed.Data.Stations, lat, lng, radius)
		return printDensitySummary(cmd, density)
	}
}

// filterStationsByRadius returns stations within radius metres of (lat, lng),
// sorted by ascending distance.
func filterStationsByRadius(stations []gbfsStationInfo, lat, lng, radius float64) []StationSummary {
	var result []StationSummary
	for _, s := range stations {
		d := haversineMeters(lat, lng, s.Lat, s.Lon)
		if d <= radius {
			result = append(result, StationSummary{
				Name:      s.Name,
				Lat:       s.Lat,
				Lng:       s.Lon,
				Capacity:  s.Capacity,
				DistanceM: d,
			})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].DistanceM < result[j].DistanceM
	})
	return result
}

// computeDensity counts stations within radius metres and aggregates capacity.
func computeDensity(stations []gbfsStationInfo, lat, lng, radius float64) DensitySummary {
	var count, total int
	for _, s := range stations {
		d := haversineMeters(lat, lng, s.Lat, s.Lon)
		if d <= radius {
			count++
			total += s.Capacity
		}
	}
	avg := 0.0
	if count > 0 {
		avg = float64(total) / float64(count)
	}
	return DensitySummary{
		Count:         count,
		AvgCapacity:   avg,
		TotalCapacity: total,
		RadiusM:       radius,
	}
}
