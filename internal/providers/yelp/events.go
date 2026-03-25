package yelp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newEventsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "events",
		Short:   "Search and view Yelp events",
		Aliases: []string{"event"},
	}

	cmd.AddCommand(newEventSearchCmd(factory))
	cmd.AddCommand(newEventGetCmd(factory))
	cmd.AddCommand(newEventFeaturedCmd(factory))

	return cmd
}

func newEventSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search events by location and filters",
		RunE:  makeRunEventSearch(factory),
	}
	cmd.Flags().String("location", "", "Location text (e.g., 'San Francisco, CA')")
	cmd.Flags().Float64("latitude", 0, "Latitude for geo-based search")
	cmd.Flags().Float64("longitude", 0, "Longitude for geo-based search")
	cmd.Flags().Int("radius", 0, "Search radius in meters")
	cmd.Flags().String("categories", "", "Comma-separated event category filters")
	cmd.Flags().Int64("start-date", 0, "Unix timestamp for event start lower bound")
	cmd.Flags().Int64("end-date", 0, "Unix timestamp for event end upper bound")
	cmd.Flags().Bool("is-free", false, "Only return free events")
	cmd.Flags().String("sort-by", "", "Sort direction: desc or asc")
	cmd.Flags().String("sort-on", "", "Field to sort on: popularity or time_start")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().Int("offset", 0, "Offset for pagination")
	return cmd
}

func makeRunEventSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		location, _ := cmd.Flags().GetString("location")
		latitude, _ := cmd.Flags().GetFloat64("latitude")
		longitude, _ := cmd.Flags().GetFloat64("longitude")
		radius, _ := cmd.Flags().GetInt("radius")
		categories, _ := cmd.Flags().GetString("categories")
		startDate, _ := cmd.Flags().GetInt64("start-date")
		endDate, _ := cmd.Flags().GetInt64("end-date")
		isFree, _ := cmd.Flags().GetBool("is-free")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		sortOn, _ := cmd.Flags().GetString("sort-on")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

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
		if radius > 0 {
			params.Set("radius", strconv.Itoa(radius))
		}
		if categories != "" {
			params.Set("categories", categories)
		}
		if startDate > 0 {
			params.Set("start_date", strconv.FormatInt(startDate, 10))
		}
		if endDate > 0 {
			params.Set("end_date", strconv.FormatInt(endDate, 10))
		}
		if isFree {
			params.Set("is_free", "true")
		}
		if sortBy != "" {
			params.Set("sort_by", sortBy)
		}
		if sortOn != "" {
			params.Set("sort_on", sortOn)
		}
		if limit > 0 {
			params.Set("limit", strconv.Itoa(limit))
		}
		if offset > 0 {
			params.Set("offset", strconv.Itoa(offset))
		}

		body, err := client.doYelp(ctx, "GET", "/events", params)
		if err != nil {
			return fmt.Errorf("event search: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var resp struct {
			Events []EventSummary `json:"events"`
			Total  int            `json:"total"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		return printEventSummaries(cmd, resp.Events)
	}
}

func newEventGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific event by ID",
		RunE:  makeRunEventGet(factory),
	}
	cmd.Flags().String("event-id", "", "Yelp event ID")
	_ = cmd.MarkFlagRequired("event-id")
	return cmd
}

func makeRunEventGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		eventID, _ := cmd.Flags().GetString("event-id")

		body, err := client.doYelp(ctx, "GET", "/events/"+eventID, nil)
		if err != nil {
			return fmt.Errorf("get event: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var event EventSummary
		if err := json.Unmarshal(body, &event); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		lines := formatEventDetail(event)
		cli.PrintText(lines)
		return nil
	}
}

func newEventFeaturedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "featured",
		Short: "Get the featured event for a location",
		RunE:  makeRunEventFeatured(factory),
	}
	cmd.Flags().String("location", "", "Location text (e.g., 'San Francisco, CA')")
	cmd.Flags().Float64("latitude", 0, "Latitude")
	cmd.Flags().Float64("longitude", 0, "Longitude")
	return cmd
}

func makeRunEventFeatured(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

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

		body, err := client.doYelp(ctx, "GET", "/events/featured", params)
		if err != nil {
			return fmt.Errorf("featured event: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var event EventSummary
		if err := json.Unmarshal(body, &event); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		lines := formatEventDetail(event)
		cli.PrintText(lines)
		return nil
	}
}

// formatEventDetail formats an EventSummary for text output.
func formatEventDetail(e EventSummary) []string {
	free := "no"
	if e.IsFree {
		free = "yes"
	}
	cost := "-"
	if e.Cost > 0 {
		cost = fmt.Sprintf("$%.2f", e.Cost)
	}
	loc := fmt.Sprintf("%s, %s %s", e.Location.Address1, e.Location.City, e.Location.State)

	lines := []string{
		fmt.Sprintf("Name:        %s", e.Name),
		fmt.Sprintf("ID:          %s", e.ID),
		fmt.Sprintf("Free:        %s", free),
		fmt.Sprintf("Cost:        %s", cost),
		fmt.Sprintf("Start:       %s", e.TimeStart),
		fmt.Sprintf("End:         %s", orDash(e.TimeEnd)),
		fmt.Sprintf("Location:    %s", loc),
		fmt.Sprintf("Attending:   %d", e.AttendingCount),
		fmt.Sprintf("URL:         %s", orDash(e.EventURL)),
	}
	if e.Description != "" {
		lines = append(lines, fmt.Sprintf("Description: %s", truncate(e.Description, 200)))
	}
	return lines
}
