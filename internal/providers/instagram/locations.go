package instagram

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// locationInfoResponse is the response for GET /api/v1/locations/{id}/info/.
type locationInfoResponse struct {
	Location rawLocationDetail `json:"location"`
	Status   string            `json:"status"`
}

// rawLocationDetail is the full location object from the info endpoint.
type rawLocationDetail struct {
	PK         int64   `json:"pk"`
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	City       string  `json:"city"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	MediaCount int64   `json:"media_count"`
}

// locationSectionsResponse is the response for GET /api/v1/locations/{id}/sections/.
type locationSectionsResponse struct {
	Sections      []rawLocationSection `json:"sections"`
	NextMaxID     string               `json:"next_max_id"`
	MoreAvailable bool                 `json:"more_available"`
	Status        string               `json:"status"`
}

// rawLocationSection represents a section in the location feed.
type rawLocationSection struct {
	FeedType      string `json:"feed_type"`
	LayoutContent struct {
		Medias []struct {
			Media rawMediaItem `json:"media"`
		} `json:"medias"`
	} `json:"layout_content"`
}

// locationStoryResponse is the response for GET /api/v1/locations/{id}/story/.
type locationStoryResponse struct {
	Story  map[string]any `json:"story"`
	Status string         `json:"status"`
}

// newLocationsCmd builds the `locations` subcommand group.
func newLocationsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "locations",
		Short:   "Browse location pages",
		Aliases: []string{"location", "loc"},
	}
	cmd.AddCommand(newLocationsGetCmd(factory))
	cmd.AddCommand(newLocationsFeedCmd(factory))
	cmd.AddCommand(newLocationsSearchCmd(factory))
	cmd.AddCommand(newLocationsStoriesCmd(factory))
	return cmd
}

func newLocationsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get location info",
		RunE:  makeRunLocationsGet(factory),
	}
	cmd.Flags().String("location-id", "", "Location ID")
	_ = cmd.MarkFlagRequired("location-id")
	return cmd
}

func makeRunLocationsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		locationID, _ := cmd.Flags().GetString("location-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/locations/"+url.PathEscape(locationID)+"/info/", nil)
		if err != nil {
			return fmt.Errorf("getting location %s: %w", locationID, err)
		}

		var result locationInfoResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding location info: %w", err)
		}

		loc := LocationSummary{
			PK:         result.Location.PK,
			Name:       result.Location.Name,
			Address:    result.Location.Address,
			City:       result.Location.City,
			Lat:        result.Location.Lat,
			Lng:        result.Location.Lng,
			MediaCount: result.Location.MediaCount,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(loc)
		}

		lines := []string{
			fmt.Sprintf("ID:      %d", loc.PK),
			fmt.Sprintf("Name:    %s", loc.Name),
			fmt.Sprintf("Address: %s", loc.Address),
			fmt.Sprintf("City:    %s", loc.City),
			fmt.Sprintf("Lat:     %g", loc.Lat),
			fmt.Sprintf("Lng:     %g", loc.Lng),
			fmt.Sprintf("Posts:   %s", formatCount(loc.MediaCount)),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newLocationsFeedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feed",
		Short: "Get posts for a location",
		RunE:  makeRunLocationsFeed(factory),
	}
	cmd.Flags().String("location-id", "", "Location ID")
	_ = cmd.MarkFlagRequired("location-id")
	cmd.Flags().String("tab", "ranked", "Feed tab: ranked or recent")
	cmd.Flags().Int("limit", 20, "Maximum number of items")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunLocationsFeed(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		locationID, _ := cmd.Flags().GetString("location-id")
		tab, _ := cmd.Flags().GetString("tab")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("tab", tab)
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}

		resp, err := client.MobileGet(ctx, "/api/v1/locations/"+url.PathEscape(locationID)+"/sections/", params)
		if err != nil {
			return fmt.Errorf("getting location feed for %s: %w", locationID, err)
		}

		var result locationSectionsResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding location feed: %w", err)
		}

		var summaries []MediaSummary
		for _, section := range result.Sections {
			for _, m := range section.LayoutContent.Medias {
				summaries = append(summaries, toMediaSummary(m.Media))
			}
		}

		if err := printMediaSummaries(cmd, summaries); err != nil {
			return err
		}
		if result.MoreAvailable && result.NextMaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.NextMaxID)
		}
		return nil
	}
}

func newLocationsSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for locations by name",
		RunE:  makeRunLocationsSearch(factory),
	}
	cmd.Flags().String("query", "", "Search query")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Float64("lat", 0, "Latitude hint")
	cmd.Flags().Float64("lng", 0, "Longitude hint")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	return cmd
}

func makeRunLocationsSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		lat, _ := cmd.Flags().GetFloat64("lat")
		lng, _ := cmd.Flags().GetFloat64("lng")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("search_query", query)
		params.Set("count", strconv.Itoa(limit))
		if lat != 0 {
			params.Set("latitude", strconv.FormatFloat(lat, 'f', 6, 64))
		}
		if lng != 0 {
			params.Set("longitude", strconv.FormatFloat(lng, 'f', 6, 64))
		}

		resp, err := client.MobileGet(ctx, "/api/v1/location_search/", params)
		if err != nil {
			return fmt.Errorf("searching locations: %w", err)
		}

		var result searchLocationsResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding location search response: %w", err)
		}

		venues := result.Venues
		if len(venues) > limit {
			venues = venues[:limit]
		}

		summaries := make([]LocationSummary, 0, len(venues))
		for _, v := range venues {
			summaries = append(summaries, LocationSummary{
				PK:         v.PK,
				Name:       v.Name,
				Address:    v.Address,
				City:       v.City,
				Lat:        v.Lat,
				Lng:        v.Lng,
				MediaCount: v.MediaCount,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No locations found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-12s  %-30s  %-30s  %-12s", "ID", "NAME", "CITY", "POSTS"))
		for _, loc := range summaries {
			lines = append(lines, fmt.Sprintf("%-12d  %-30s  %-30s  %-12s",
				loc.PK,
				truncate(loc.Name, 30),
				truncate(loc.City, 30),
				formatCount(loc.MediaCount),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newLocationsStoriesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stories",
		Short: "Get stories for a location",
		RunE:  makeRunLocationsStories(factory),
	}
	cmd.Flags().String("location-id", "", "Location ID")
	_ = cmd.MarkFlagRequired("location-id")
	return cmd
}

func makeRunLocationsStories(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		locationID, _ := cmd.Flags().GetString("location-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/locations/"+url.PathEscape(locationID)+"/story/", nil)
		if err != nil {
			return fmt.Errorf("getting location stories for %s: %w", locationID, err)
		}

		var result locationStoryResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding location story response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Location story for %s retrieved.\n", locationID)
		return nil
	}
}
