package places

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/places/v1"
)

func newAutocompleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autocomplete",
		Short: "Get place predictions for a text input",
		Long:  "Return autocomplete suggestions as the user types a place name or address.",
		RunE:  makeRunAutocomplete(factory),
	}
	cmd.Flags().String("input", "", "Text to autocomplete (required)")
	cmd.Flags().Int("input-offset", -1, "Cursor position in input (0-based)")
	cmd.Flags().StringSlice("types", nil, "Filter by primary place types (up to 5)")
	cmd.Flags().StringSlice("regions", nil, "Restrict to CLDR region codes (up to 15)")
	cmd.Flags().String("location-bias", "", "Bias results toward lat,lng,radiusM")
	cmd.Flags().String("location-restrict", "", "Restrict results to lat,lng,radiusM")
	cmd.Flags().String("origin", "", "Origin lat,lng for distance calculation")
	cmd.Flags().String("lang", "", "Language code (e.g. en)")
	cmd.Flags().String("region", "", "CLDR region code for formatting")
	cmd.Flags().Bool("include-queries", false, "Include query predictions alongside place predictions")
	cmd.Flags().String("session-token", "", "Session token for billing (UUID)")
	_ = cmd.MarkFlagRequired("input")
	return cmd
}

func makeRunAutocomplete(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}

		input, _ := cmd.Flags().GetString("input")
		req := &api.GoogleMapsPlacesV1AutocompletePlacesRequest{
			Input: input,
		}

		if offset, _ := cmd.Flags().GetInt("input-offset"); offset >= 0 {
			req.InputOffset = int64(offset)
		}
		if types, _ := cmd.Flags().GetStringSlice("types"); len(types) > 0 {
			req.IncludedPrimaryTypes = types
		}
		if regions, _ := cmd.Flags().GetStringSlice("regions"); len(regions) > 0 {
			req.IncludedRegionCodes = regions
		}
		if bias, _ := cmd.Flags().GetString("location-bias"); bias != "" {
			lat, lng, radius, err := parseLocationBias(bias)
			if err != nil {
				return err
			}
			req.LocationBias = &api.GoogleMapsPlacesV1AutocompletePlacesRequestLocationBias{
				Circle: &api.GoogleMapsPlacesV1Circle{
					Center: &api.GoogleTypeLatLng{Latitude: lat, Longitude: lng},
					Radius: radius,
				},
			}
		}
		if restrict, _ := cmd.Flags().GetString("location-restrict"); restrict != "" {
			lat, lng, radius, err := parseLocationBias(restrict)
			if err != nil {
				return err
			}
			req.LocationRestriction = &api.GoogleMapsPlacesV1AutocompletePlacesRequestLocationRestriction{
				Circle: &api.GoogleMapsPlacesV1Circle{
					Center: &api.GoogleTypeLatLng{Latitude: lat, Longitude: lng},
					Radius: radius,
				},
			}
		}
		if origin, _ := cmd.Flags().GetString("origin"); origin != "" {
			lat, lng, err := parseLatLng(origin)
			if err != nil {
				return err
			}
			req.Origin = &api.GoogleTypeLatLng{Latitude: lat, Longitude: lng}
		}
		if lang, _ := cmd.Flags().GetString("lang"); lang != "" {
			req.LanguageCode = lang
		}
		if region, _ := cmd.Flags().GetString("region"); region != "" {
			req.RegionCode = region
		}
		if iq, _ := cmd.Flags().GetBool("include-queries"); iq {
			req.IncludeQueryPredictions = true
		}
		if st, _ := cmd.Flags().GetString("session-token"); st != "" {
			req.SessionToken = st
		}

		resp, err := svc.Places.Autocomplete(req).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("autocomplete: %w", err)
		}

		suggestions := make([]AutocompleteSuggestion, 0, len(resp.Suggestions))
		for _, s := range resp.Suggestions {
			if s.PlacePrediction != nil {
				pp := s.PlacePrediction
				as := AutocompleteSuggestion{
					Type: "place",
				}
				if pp.Text != nil {
					as.Text = pp.Text.Text
				}
				as.PlaceID = extractPlaceID(pp.Place)
				as.Distance = pp.DistanceMeters
				if pp.StructuredFormat != nil && pp.StructuredFormat.MainText != nil {
					as.PrimaryType = pp.StructuredFormat.MainText.Text
				}
				suggestions = append(suggestions, as)
			}
			if s.QueryPrediction != nil {
				qp := s.QueryPrediction
				as := AutocompleteSuggestion{
					Type: "query",
				}
				if qp.Text != nil {
					as.Text = qp.Text.Text
				}
				suggestions = append(suggestions, as)
			}
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(suggestions)
		}

		if len(suggestions) == 0 {
			fmt.Println("No suggestions found.")
			return nil
		}

		lines := make([]string, 0, len(suggestions)+1)
		lines = append(lines, fmt.Sprintf("%-6s  %-50s  %-20s", "TYPE", "TEXT", "PLACE ID"))
		for _, s := range suggestions {
			text := truncate(s.Text, 50)
			placeID := truncate(s.PlaceID, 20)
			lines = append(lines, fmt.Sprintf("%-6s  %-50s  %-20s", s.Type, text, placeID))
		}
		cli.PrintText(lines)
		return nil
	}
}
