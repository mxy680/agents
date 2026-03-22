package x

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GeoPlace is a representation of a geographic place returned by the X API.
type GeoPlace struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name,omitempty"`
	PlaceType   string `json:"place_type,omitempty"`
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
}

// newGeoCmd builds the "geo" subcommand group.
func newGeoCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "geo",
		Short:   "Look up geographic places",
		Aliases: []string{"location"},
	}
	cmd.AddCommand(newGeoReverseCmd(factory))
	cmd.AddCommand(newGeoSearchCmd(factory))
	cmd.AddCommand(newGeoGetCmd(factory))
	return cmd
}

func newGeoReverseCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reverse",
		Short: "Reverse-geocode a lat/lng to a place",
		RunE:  makeRunGeoReverse(factory),
	}
	cmd.Flags().Float64("lat", 0, "Latitude (required)")
	_ = cmd.MarkFlagRequired("lat")
	cmd.Flags().Float64("lng", 0, "Longitude (required)")
	_ = cmd.MarkFlagRequired("lng")
	return cmd
}

func newGeoSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for places by query",
		RunE:  makeRunGeoSearch(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Float64("lat", 0, "Latitude hint")
	cmd.Flags().Float64("lng", 0, "Longitude hint")
	return cmd
}

func newGeoGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a place by ID",
		RunE:  makeRunGeoGet(factory),
	}
	cmd.Flags().String("place-id", "", "Place ID (required)")
	_ = cmd.MarkFlagRequired("place-id")
	return cmd
}

// --- RunE implementations ---

func makeRunGeoReverse(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		lat, _ := cmd.Flags().GetFloat64("lat")
		lng, _ := cmd.Flags().GetFloat64("lng")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("lat", fmt.Sprintf("%f", lat))
		params.Set("long", fmt.Sprintf("%f", lng))

		resp, err := client.Get(ctx, "/i/api/1.1/geo/reverse_geocode.json", params)
		if err != nil {
			return fmt.Errorf("reverse geocoding: %w", err)
		}

		var result struct {
			Result struct {
				Places []json.RawMessage `json:"places"`
			} `json:"result"`
		}
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decode reverse geocode response: %w", err)
		}

		places := parsePlaces(result.Result.Places)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(places)
		}

		return printGeoPlaces(cmd, places)
	}
}

func makeRunGeoSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		lat, _ := cmd.Flags().GetFloat64("lat")
		lng, _ := cmd.Flags().GetFloat64("lng")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("query", query)
		if lat != 0 {
			params.Set("lat", fmt.Sprintf("%f", lat))
		}
		if lng != 0 {
			params.Set("long", fmt.Sprintf("%f", lng))
		}

		resp, err := client.Get(ctx, "/i/api/1.1/geo/search.json", params)
		if err != nil {
			return fmt.Errorf("searching places: %w", err)
		}

		var result struct {
			Result struct {
				Places []json.RawMessage `json:"places"`
			} `json:"result"`
		}
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decode geo search response: %w", err)
		}

		places := parsePlaces(result.Result.Places)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(places)
		}

		return printGeoPlaces(cmd, places)
	}
}

func makeRunGeoGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		placeID, _ := cmd.Flags().GetString("place-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/i/api/1.1/geo/id/"+placeID+".json", nil)
		if err != nil {
			return fmt.Errorf("getting place %s: %w", placeID, err)
		}

		var raw json.RawMessage
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decode geo get response: %w", err)
		}

		place := parseSinglePlace(raw)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(place)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", place.ID),
			fmt.Sprintf("Name:        %s", place.Name),
			fmt.Sprintf("Full Name:   %s", place.FullName),
			fmt.Sprintf("Type:        %s", place.PlaceType),
			fmt.Sprintf("Country:     %s (%s)", place.Country, place.CountryCode),
		}
		cli.PrintText(lines)
		return nil
	}
}

// parsePlaces parses a list of raw place JSON objects.
func parsePlaces(raws []json.RawMessage) []GeoPlace {
	places := make([]GeoPlace, 0, len(raws))
	for _, raw := range raws {
		p := parseSinglePlace(raw)
		if p.ID != "" {
			places = append(places, p)
		}
	}
	return places
}

// parseSinglePlace parses one raw place JSON object.
func parseSinglePlace(raw json.RawMessage) GeoPlace {
	var p struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		PlaceType   string `json:"place_type"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
	}
	_ = json.Unmarshal(raw, &p)
	return GeoPlace{
		ID:          p.ID,
		Name:        p.Name,
		FullName:    p.FullName,
		PlaceType:   p.PlaceType,
		Country:     p.Country,
		CountryCode: p.CountryCode,
	}
}

// printGeoPlaces outputs place summaries as text.
func printGeoPlaces(cmd *cobra.Command, places []GeoPlace) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(places)
	}
	if len(places) == 0 {
		fmt.Println("No places found.")
		return nil
	}
	lines := make([]string, 0, len(places)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-15s  %-20s", "ID", "FULL NAME", "TYPE", "COUNTRY"))
	for _, p := range places {
		lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-15s  %-20s",
			truncate(p.ID, 20),
			truncate(p.FullName, 30),
			truncate(p.PlaceType, 15),
			truncate(p.Country, 20),
		))
	}
	cli.PrintText(lines)
	return nil
}
