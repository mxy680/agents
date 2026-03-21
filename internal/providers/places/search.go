package places

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/places/v1"
	"google.golang.org/api/googleapi"
)

func newSearchTextCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "text",
		Short: "Search for places by text query",
		Long:  "Full-text search for places (e.g. \"coffee shops in Cleveland\").",
		RunE:  makeRunSearchText(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().String("type", "", "Filter to a specific place type (e.g. restaurant, cafe)")
	cmd.Flags().String("location-bias", "", "Bias results toward lat,lng,radiusM")
	cmd.Flags().String("location-restrict", "", "Restrict results to south,west,north,east")
	cmd.Flags().Float64("min-rating", 0, "Minimum rating (0-5, steps of 0.5)")
	cmd.Flags().Bool("open-now", false, "Only show currently open places")
	cmd.Flags().StringSlice("price-levels", nil, "Filter by price levels (1-4)")
	cmd.Flags().String("rank", "", "Rank preference: RELEVANCE or DISTANCE")
	cmd.Flags().String("region", "", "CLDR region code (e.g. us)")
	cmd.Flags().String("lang", "", "Language code (e.g. en)")
	cmd.Flags().String("fields", "advanced", "Field tier: basic, advanced, preferred, or all")
	cmd.Flags().Int("limit", 20, "Maximum results (1-20)")
	cmd.Flags().String("page-token", "", "Pagination token")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunSearchText(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}

		query, _ := cmd.Flags().GetString("query")
		req := &api.GoogleMapsPlacesV1SearchTextRequest{
			TextQuery: query,
		}

		if t, _ := cmd.Flags().GetString("type"); t != "" {
			req.IncludedType = t
		}
		if bias, _ := cmd.Flags().GetString("location-bias"); bias != "" {
			lat, lng, radius, err := parseLocationBias(bias)
			if err != nil {
				return err
			}
			req.LocationBias = &api.GoogleMapsPlacesV1SearchTextRequestLocationBias{
				Circle: &api.GoogleMapsPlacesV1Circle{
					Center: &api.GoogleTypeLatLng{Latitude: lat, Longitude: lng},
					Radius: radius,
				},
			}
		}
		if restrict, _ := cmd.Flags().GetString("location-restrict"); restrict != "" {
			s, w, n, e, err := parseLocationRestrict(restrict)
			if err != nil {
				return err
			}
			req.LocationRestriction = &api.GoogleMapsPlacesV1SearchTextRequestLocationRestriction{
				Rectangle: &api.GoogleGeoTypeViewport{
					Low:  &api.GoogleTypeLatLng{Latitude: s, Longitude: w},
					High: &api.GoogleTypeLatLng{Latitude: n, Longitude: e},
				},
			}
		}
		if mr, _ := cmd.Flags().GetFloat64("min-rating"); mr > 0 {
			req.MinRating = mr
		}
		if on, _ := cmd.Flags().GetBool("open-now"); on {
			req.OpenNow = true
		}
		if pl, _ := cmd.Flags().GetStringSlice("price-levels"); len(pl) > 0 {
			for _, p := range pl {
				req.PriceLevels = append(req.PriceLevels, priceLevelToEnum(p))
			}
		}
		if rank, _ := cmd.Flags().GetString("rank"); rank != "" {
			req.RankPreference = strings.ToUpper(rank)
		}
		if region, _ := cmd.Flags().GetString("region"); region != "" {
			req.RegionCode = region
		}
		if lang, _ := cmd.Flags().GetString("lang"); lang != "" {
			req.LanguageCode = lang
		}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 && limit <= 20 {
			req.PageSize = int64(limit)
		}
		if pt, _ := cmd.Flags().GetString("page-token"); pt != "" {
			req.PageToken = pt
		}

		tier, _ := cmd.Flags().GetString("fields")
		call := svc.Places.SearchText(req).Fields(googleapi.Field(fieldMaskForTier(tier)))

		resp, err := call.Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("search text: %w", err)
		}

		summaries := make([]PlaceSummary, 0, len(resp.Places))
		for _, p := range resp.Places {
			summaries = append(summaries, toPlaceSummary(p))
		}

		if cli.IsJSONOutput(cmd) {
			result := map[string]any{
				"places": summaries,
			}
			if resp.NextPageToken != "" {
				result["nextPageToken"] = resp.NextPageToken
			}
			return cli.PrintJSON(result)
		}

		if err := printPlaceSummaries(cmd, summaries); err != nil {
			return err
		}
		if resp.NextPageToken != "" {
			fmt.Printf("\nNext page: --page-token=%s\n", resp.NextPageToken)
		}
		return nil
	}
}

func newSearchNearbyCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nearby",
		Short: "Search for places near a location",
		Long:  "Find places within a radius of a geographic point.",
		RunE:  makeRunSearchNearby(factory),
	}
	cmd.Flags().Float64("lat", 0, "Center latitude (required)")
	cmd.Flags().Float64("lng", 0, "Center longitude (required)")
	cmd.Flags().Float64("radius", 5000, "Search radius in meters (max 50000)")
	cmd.Flags().StringSlice("types", nil, "Include place types (e.g. restaurant,cafe)")
	cmd.Flags().StringSlice("exclude-types", nil, "Exclude place types")
	cmd.Flags().StringSlice("primary-types", nil, "Include primary place types")
	cmd.Flags().String("rank", "", "Rank preference: POPULARITY or DISTANCE")
	cmd.Flags().String("region", "", "CLDR region code (e.g. us)")
	cmd.Flags().String("lang", "", "Language code (e.g. en)")
	cmd.Flags().String("fields", "advanced", "Field tier: basic, advanced, preferred, or all")
	cmd.Flags().Int("limit", 20, "Maximum results (1-20)")
	_ = cmd.MarkFlagRequired("lat")
	_ = cmd.MarkFlagRequired("lng")
	return cmd
}

func makeRunSearchNearby(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}

		lat, _ := cmd.Flags().GetFloat64("lat")
		lng, _ := cmd.Flags().GetFloat64("lng")
		radius, _ := cmd.Flags().GetFloat64("radius")

		req := &api.GoogleMapsPlacesV1SearchNearbyRequest{
			LocationRestriction: &api.GoogleMapsPlacesV1SearchNearbyRequestLocationRestriction{
				Circle: &api.GoogleMapsPlacesV1Circle{
					Center: &api.GoogleTypeLatLng{Latitude: lat, Longitude: lng},
					Radius: radius,
				},
			},
		}

		if types, _ := cmd.Flags().GetStringSlice("types"); len(types) > 0 {
			req.IncludedTypes = types
		}
		if excludeTypes, _ := cmd.Flags().GetStringSlice("exclude-types"); len(excludeTypes) > 0 {
			req.ExcludedTypes = excludeTypes
		}
		if primaryTypes, _ := cmd.Flags().GetStringSlice("primary-types"); len(primaryTypes) > 0 {
			req.IncludedPrimaryTypes = primaryTypes
		}
		if rank, _ := cmd.Flags().GetString("rank"); rank != "" {
			req.RankPreference = strings.ToUpper(rank)
		}
		if region, _ := cmd.Flags().GetString("region"); region != "" {
			req.RegionCode = region
		}
		if lang, _ := cmd.Flags().GetString("lang"); lang != "" {
			req.LanguageCode = lang
		}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 && limit <= 20 {
			req.MaxResultCount = int64(limit)
		}

		tier, _ := cmd.Flags().GetString("fields")
		call := svc.Places.SearchNearby(req).Fields(googleapi.Field(fieldMaskForTier(tier)))

		resp, err := call.Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("search nearby: %w", err)
		}

		summaries := make([]PlaceSummary, 0, len(resp.Places))
		for _, p := range resp.Places {
			summaries = append(summaries, toPlaceSummary(p))
		}

		return printPlaceSummaries(cmd, summaries)
	}
}

// priceLevelToEnum converts a user-facing price level string to the API enum.
func priceLevelToEnum(level string) string {
	switch level {
	case "0":
		return "PRICE_LEVEL_FREE"
	case "1":
		return "PRICE_LEVEL_INEXPENSIVE"
	case "2":
		return "PRICE_LEVEL_MODERATE"
	case "3":
		return "PRICE_LEVEL_EXPENSIVE"
	case "4":
		return "PRICE_LEVEL_VERY_EXPENSIVE"
	default:
		return level
	}
}
