package places

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/places/v1"
)

// PlaceSummary is the JSON-serializable summary of a place.
type PlaceSummary struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Address          string   `json:"address"`
	Types            []string `json:"types,omitempty"`
	Rating           float64  `json:"rating,omitempty"`
	UserRatingCount  int64    `json:"userRatingCount,omitempty"`
	PriceLevel       string   `json:"priceLevel,omitempty"`
	BusinessStatus   string   `json:"businessStatus,omitempty"`
	OpenNow          *bool    `json:"openNow,omitempty"`
	GoogleMapsURI    string   `json:"googleMapsUri,omitempty"`
	PhoneNumber      string   `json:"phoneNumber,omitempty"`
	WebsiteURI       string   `json:"websiteUri,omitempty"`
	EditorialSummary string   `json:"editorialSummary,omitempty"`
}

// PlaceDetail is the JSON-serializable full detail of a place.
type PlaceDetail struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Address              string            `json:"address"`
	ShortAddress         string            `json:"shortAddress,omitempty"`
	Types                []string          `json:"types,omitempty"`
	PrimaryType          string            `json:"primaryType,omitempty"`
	Rating               float64           `json:"rating,omitempty"`
	UserRatingCount      int64             `json:"userRatingCount,omitempty"`
	PriceLevel           string            `json:"priceLevel,omitempty"`
	BusinessStatus       string            `json:"businessStatus,omitempty"`
	OpenNow              *bool             `json:"openNow,omitempty"`
	WeekdayHours         []string          `json:"weekdayHours,omitempty"`
	GoogleMapsURI        string            `json:"googleMapsUri,omitempty"`
	PhoneNumber          string            `json:"phoneNumber,omitempty"`
	WebsiteURI           string            `json:"websiteUri,omitempty"`
	EditorialSummary     string            `json:"editorialSummary,omitempty"`
	Location             *LatLng           `json:"location,omitempty"`
	Delivery             bool              `json:"delivery,omitempty"`
	DineIn               bool              `json:"dineIn,omitempty"`
	Takeout              bool              `json:"takeout,omitempty"`
	CurbsidePickup       bool              `json:"curbsidePickup,omitempty"`
	Reservable           bool              `json:"reservable,omitempty"`
	Reviews              []ReviewSummary   `json:"reviews,omitempty"`
	Photos               []PhotoReference  `json:"photos,omitempty"`
	AddressComponents    []AddressComponent `json:"addressComponents,omitempty"`
}

// LatLng represents a geographic coordinate.
type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ReviewSummary is a simplified review.
type ReviewSummary struct {
	Author        string  `json:"author"`
	Rating        float64 `json:"rating"`
	Text          string  `json:"text"`
	RelativeTime  string  `json:"relativeTime,omitempty"`
	PublishTime   string  `json:"publishTime,omitempty"`
}

// PhotoReference contains a photo resource name for fetching.
type PhotoReference struct {
	Name          string `json:"name"`
	WidthPx       int64  `json:"widthPx,omitempty"`
	HeightPx      int64  `json:"heightPx,omitempty"`
}

// AddressComponent is a component of an address.
type AddressComponent struct {
	LongText  string   `json:"longText"`
	ShortText string   `json:"shortText"`
	Types     []string `json:"types"`
}

// AutocompleteSuggestion is a single autocomplete suggestion.
type AutocompleteSuggestion struct {
	Type        string `json:"type"` // "place" or "query"
	Text        string `json:"text"`
	PlaceID     string `json:"placeId,omitempty"`
	Distance    int64  `json:"distanceMeters,omitempty"`
	PrimaryType string `json:"primaryType,omitempty"`
}

// PhotoMedia is the result of fetching a photo.
type PhotoMedia struct {
	Name     string `json:"name"`
	PhotoURI string `json:"photoUri"`
}

// toPlaceSummary converts an API place to a PlaceSummary.
func toPlaceSummary(p *api.GoogleMapsPlacesV1Place) PlaceSummary {
	s := PlaceSummary{
		ID:              extractPlaceID(p.Name),
		Rating:          p.Rating,
		UserRatingCount: p.UserRatingCount,
		PriceLevel:      formatPriceLevel(p.PriceLevel),
		BusinessStatus:  p.BusinessStatus,
		GoogleMapsURI:   p.GoogleMapsUri,
		PhoneNumber:     p.InternationalPhoneNumber,
		WebsiteURI:      p.WebsiteUri,
	}
	if p.DisplayName != nil {
		s.Name = p.DisplayName.Text
	}
	if p.FormattedAddress != "" {
		s.Address = p.FormattedAddress
	}
	if len(p.Types) > 0 {
		s.Types = p.Types
	}
	if p.EditorialSummary != nil {
		s.EditorialSummary = p.EditorialSummary.Text
	}
	if p.CurrentOpeningHours != nil {
		s.OpenNow = &p.CurrentOpeningHours.OpenNow
	} else if p.RegularOpeningHours != nil {
		s.OpenNow = &p.RegularOpeningHours.OpenNow
	}
	return s
}

// toPlaceDetail converts an API place to a PlaceDetail.
func toPlaceDetail(p *api.GoogleMapsPlacesV1Place) PlaceDetail {
	d := PlaceDetail{
		ID:              extractPlaceID(p.Name),
		Rating:          p.Rating,
		UserRatingCount: p.UserRatingCount,
		PriceLevel:      formatPriceLevel(p.PriceLevel),
		BusinessStatus:  p.BusinessStatus,
		GoogleMapsURI:   p.GoogleMapsUri,
		PhoneNumber:     p.InternationalPhoneNumber,
		WebsiteURI:      p.WebsiteUri,
		Delivery:        p.Delivery,
		DineIn:          p.DineIn,
		Takeout:         p.Takeout,
		CurbsidePickup:  p.CurbsidePickup,
		Reservable:      p.Reservable,
	}
	if p.DisplayName != nil {
		d.Name = p.DisplayName.Text
	}
	if p.FormattedAddress != "" {
		d.Address = p.FormattedAddress
	}
	if p.ShortFormattedAddress != "" {
		d.ShortAddress = p.ShortFormattedAddress
	}
	if len(p.Types) > 0 {
		d.Types = p.Types
	}
	if p.PrimaryType != "" {
		d.PrimaryType = p.PrimaryType
	}
	if p.EditorialSummary != nil {
		d.EditorialSummary = p.EditorialSummary.Text
	}
	if p.CurrentOpeningHours != nil {
		d.OpenNow = &p.CurrentOpeningHours.OpenNow
		d.WeekdayHours = p.CurrentOpeningHours.WeekdayDescriptions
	} else if p.RegularOpeningHours != nil {
		d.OpenNow = &p.RegularOpeningHours.OpenNow
		d.WeekdayHours = p.RegularOpeningHours.WeekdayDescriptions
	}
	if p.Location != nil {
		d.Location = &LatLng{
			Latitude:  p.Location.Latitude,
			Longitude: p.Location.Longitude,
		}
	}
	for _, r := range p.Reviews {
		rs := ReviewSummary{
			Rating:      r.Rating,
			PublishTime: r.PublishTime,
		}
		if r.AuthorAttribution != nil {
			rs.Author = r.AuthorAttribution.DisplayName
		}
		if r.Text != nil {
			rs.Text = r.Text.Text
		}
		if r.RelativePublishTimeDescription != "" {
			rs.RelativeTime = r.RelativePublishTimeDescription
		}
		d.Reviews = append(d.Reviews, rs)
	}
	for _, ph := range p.Photos {
		d.Photos = append(d.Photos, PhotoReference{
			Name:     ph.Name,
			WidthPx:  ph.WidthPx,
			HeightPx: ph.HeightPx,
		})
	}
	for _, ac := range p.AddressComponents {
		d.AddressComponents = append(d.AddressComponents, AddressComponent{
			LongText:  ac.LongText,
			ShortText: ac.ShortText,
			Types:     ac.Types,
		})
	}
	return d
}

// extractPlaceID extracts the place ID from a resource name like "places/ChIJ...".
func extractPlaceID(name string) string {
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

// formatPriceLevel converts API price level enum to a human-readable string.
func formatPriceLevel(level string) string {
	switch level {
	case "PRICE_LEVEL_FREE":
		return "Free"
	case "PRICE_LEVEL_INEXPENSIVE":
		return "$"
	case "PRICE_LEVEL_MODERATE":
		return "$$"
	case "PRICE_LEVEL_EXPENSIVE":
		return "$$$"
	case "PRICE_LEVEL_VERY_EXPENSIVE":
		return "$$$$"
	default:
		return ""
	}
}

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// openNowStr returns "Yes", "No", or "-" for an *bool open-now value.
func openNowStr(b *bool) string {
	if b == nil {
		return "-"
	}
	if *b {
		return "Yes"
	}
	return "No"
}

// printPlaceSummaries outputs place summaries as JSON or a formatted text table.
func printPlaceSummaries(cmd *cobra.Command, summaries []PlaceSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No places found.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-35s  %-40s  %-6s  %-5s  %-4s", "NAME", "ADDRESS", "RATING", "OPEN", "PRICE"))
	for _, s := range summaries {
		name := truncate(s.Name, 35)
		addr := truncate(s.Address, 40)
		rating := "-"
		if s.Rating > 0 {
			rating = fmt.Sprintf("%.1f", s.Rating)
		}
		lines = append(lines, fmt.Sprintf("%-35s  %-40s  %-6s  %-5s  %-4s",
			name, addr, rating, openNowStr(s.OpenNow), s.PriceLevel))
	}
	cli.PrintText(lines)
	return nil
}

// confirmDestructive returns an error if the --confirm flag is absent or false.
func confirmDestructive(cmd *cobra.Command) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("this action is irreversible; re-run with --confirm to proceed")
	}
	return nil
}

// dryRunResult prints a standardised dry-run response and returns nil.
func dryRunResult(cmd *cobra.Command, description string, data any) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(data)
	}
	fmt.Printf("[DRY RUN] %s\n", description)
	return nil
}

// parseLatLng parses a "lat,lng" string into latitude and longitude.
func parseLatLng(s string) (float64, float64, error) {
	var lat, lng float64
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid lat,lng format: %q (expected lat,lng)", s)
	}
	if _, err := fmt.Sscanf(parts[0], "%f", &lat); err != nil {
		return 0, 0, fmt.Errorf("invalid latitude: %w", err)
	}
	if _, err := fmt.Sscanf(parts[1], "%f", &lng); err != nil {
		return 0, 0, fmt.Errorf("invalid longitude: %w", err)
	}
	return lat, lng, nil
}

// parseLocationBias parses "lat,lng,radiusM" into lat, lng, and radius.
func parseLocationBias(s string) (float64, float64, float64, error) {
	var lat, lng, radius float64
	parts := strings.SplitN(s, ",", 3)
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid location-bias format: %q (expected lat,lng,radiusM)", s)
	}
	if _, err := fmt.Sscanf(parts[0], "%f", &lat); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid latitude: %w", err)
	}
	if _, err := fmt.Sscanf(parts[1], "%f", &lng); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid longitude: %w", err)
	}
	if _, err := fmt.Sscanf(parts[2], "%f", &radius); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid radius: %w", err)
	}
	return lat, lng, radius, nil
}

// parseLocationRestrict parses "south,west,north,east" into a rectangle.
func parseLocationRestrict(s string) (south, west, north, east float64, err error) {
	parts := strings.SplitN(s, ",", 4)
	if len(parts) != 4 {
		return 0, 0, 0, 0, fmt.Errorf("invalid location-restrict format: %q (expected south,west,north,east)", s)
	}
	if _, err = fmt.Sscanf(parts[0], "%f", &south); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid south: %w", err)
	}
	if _, err = fmt.Sscanf(parts[1], "%f", &west); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid west: %w", err)
	}
	if _, err = fmt.Sscanf(parts[2], "%f", &north); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid north: %w", err)
	}
	if _, err = fmt.Sscanf(parts[3], "%f", &east); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid east: %w", err)
	}
	return south, west, north, east, nil
}

// fieldMaskForTier returns the Places API field mask for a given tier.
func fieldMaskForTier(tier string) string {
	basic := []string{
		"places.id",
		"places.displayName",
		"places.types",
		"places.primaryType",
		"places.formattedAddress",
		"places.shortFormattedAddress",
		"places.location",
		"places.businessStatus",
		"places.googleMapsUri",
		"places.photos",
	}
	advanced := append(basic, []string{
		"places.rating",
		"places.userRatingCount",
		"places.priceLevel",
		"places.currentOpeningHours",
		"places.regularOpeningHours",
		"places.internationalPhoneNumber",
		"places.websiteUri",
		"places.editorialSummary",
		"places.delivery",
		"places.dineIn",
		"places.takeout",
		"places.curbsidePickup",
		"places.reservable",
		"places.reviews",
	}...)
	preferred := append(advanced, []string{
		"places.addressComponents",
		"places.parkingOptions",
		"places.paymentOptions",
		"places.accessibilityOptions",
		"places.allowsDogs",
		"places.goodForChildren",
		"places.goodForGroups",
		"places.outdoorSeating",
		"places.restroom",
		"places.liveMusic",
		"places.servesBeer",
		"places.servesWine",
		"places.servesCocktails",
		"places.servesCoffee",
		"places.servesBreakfast",
		"places.servesBrunch",
		"places.servesLunch",
		"places.servesDinner",
		"places.servesDessert",
		"places.servesVegetarianFood",
	}...)

	switch tier {
	case "basic":
		return strings.Join(basic, ",")
	case "advanced":
		return strings.Join(advanced, ",")
	case "preferred":
		return strings.Join(preferred, ",")
	case "all":
		return "*"
	default:
		return strings.Join(advanced, ",")
	}
}

// detailFieldMask returns the field mask for a single place detail request.
// Same fields but without the "places." prefix.
func detailFieldMask(tier string) string {
	basic := []string{
		"id",
		"displayName",
		"types",
		"primaryType",
		"formattedAddress",
		"shortFormattedAddress",
		"location",
		"businessStatus",
		"googleMapsUri",
		"photos",
	}
	advanced := append(basic, []string{
		"rating",
		"userRatingCount",
		"priceLevel",
		"currentOpeningHours",
		"regularOpeningHours",
		"internationalPhoneNumber",
		"websiteUri",
		"editorialSummary",
		"delivery",
		"dineIn",
		"takeout",
		"curbsidePickup",
		"reservable",
		"reviews",
	}...)
	preferred := append(advanced, []string{
		"addressComponents",
		"parkingOptions",
		"paymentOptions",
		"accessibilityOptions",
		"allowsDogs",
		"goodForChildren",
		"goodForGroups",
		"outdoorSeating",
		"restroom",
		"liveMusic",
		"servesBeer",
		"servesWine",
		"servesCocktails",
		"servesCoffee",
		"servesBreakfast",
		"servesBrunch",
		"servesLunch",
		"servesDinner",
		"servesDessert",
		"servesVegetarianFood",
	}...)

	switch tier {
	case "basic":
		return strings.Join(basic, ",")
	case "advanced":
		return strings.Join(advanced, ",")
	case "preferred":
		return strings.Join(preferred, ",")
	case "all":
		return "*"
	default:
		return strings.Join(advanced, ",")
	}
}
