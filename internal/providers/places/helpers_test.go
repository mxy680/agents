package places

import (
	"testing"

	api "google.golang.org/api/places/v1"
)

func TestExtractPlaceID(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"places/ChIJ123", "ChIJ123"},
		{"ChIJ123", "ChIJ123"},
		{"places/abc/def", "def"},
		{"", ""},
	}
	for _, tt := range tests {
		got := extractPlaceID(tt.name)
		if got != tt.want {
			t.Errorf("extractPlaceID(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestFormatPriceLevel(t *testing.T) {
	tests := []struct {
		level string
		want  string
	}{
		{"PRICE_LEVEL_FREE", "Free"},
		{"PRICE_LEVEL_INEXPENSIVE", "$"},
		{"PRICE_LEVEL_MODERATE", "$$"},
		{"PRICE_LEVEL_EXPENSIVE", "$$$"},
		{"PRICE_LEVEL_VERY_EXPENSIVE", "$$$$"},
		{"PRICE_LEVEL_UNSPECIFIED", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := formatPriceLevel(tt.level)
		if got != tt.want {
			t.Errorf("formatPriceLevel(%q) = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"", 5, ""},
	}
	for _, tt := range tests {
		got := truncate(tt.s, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
		}
	}
}

func TestOpenNowStr(t *testing.T) {
	tr := true
	fa := false
	tests := []struct {
		b    *bool
		want string
	}{
		{nil, "-"},
		{&tr, "Yes"},
		{&fa, "No"},
	}
	for _, tt := range tests {
		got := openNowStr(tt.b)
		if got != tt.want {
			t.Errorf("openNowStr(%v) = %q, want %q", tt.b, got, tt.want)
		}
	}
}

func TestPriceLevelToEnum(t *testing.T) {
	tests := []struct {
		level string
		want  string
	}{
		{"0", "PRICE_LEVEL_FREE"},
		{"1", "PRICE_LEVEL_INEXPENSIVE"},
		{"2", "PRICE_LEVEL_MODERATE"},
		{"3", "PRICE_LEVEL_EXPENSIVE"},
		{"4", "PRICE_LEVEL_VERY_EXPENSIVE"},
		{"PRICE_LEVEL_MODERATE", "PRICE_LEVEL_MODERATE"},
	}
	for _, tt := range tests {
		got := priceLevelToEnum(tt.level)
		if got != tt.want {
			t.Errorf("priceLevelToEnum(%q) = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestParseLatLng(t *testing.T) {
	lat, lng, err := parseLatLng("41.4993,-81.6944")
	if err != nil {
		t.Fatal(err)
	}
	if lat != 41.4993 || lng != -81.6944 {
		t.Errorf("got (%f, %f), want (41.4993, -81.6944)", lat, lng)
	}

	_, _, err = parseLatLng("invalid")
	if err == nil {
		t.Fatal("expected error for invalid lat,lng")
	}
}

func TestParseLocationBias(t *testing.T) {
	lat, lng, radius, err := parseLocationBias("41.4993,-81.6944,5000")
	if err != nil {
		t.Fatal(err)
	}
	if lat != 41.4993 || lng != -81.6944 || radius != 5000 {
		t.Errorf("got (%f, %f, %f), want (41.4993, -81.6944, 5000)", lat, lng, radius)
	}

	_, _, _, err = parseLocationBias("41.4993,-81.6944")
	if err == nil {
		t.Fatal("expected error for missing radius")
	}
}

func TestParseLocationRestrict(t *testing.T) {
	s, w, n, e, err := parseLocationRestrict("41.0,-82.0,42.0,-81.0")
	if err != nil {
		t.Fatal(err)
	}
	if s != 41.0 || w != -82.0 || n != 42.0 || e != -81.0 {
		t.Errorf("got (%f, %f, %f, %f), want (41, -82, 42, -81)", s, w, n, e)
	}

	_, _, _, _, err = parseLocationRestrict("41.0,-82.0")
	if err == nil {
		t.Fatal("expected error for insufficient values")
	}
}

func TestFieldMaskForTier(t *testing.T) {
	basic := fieldMaskForTier("basic")
	if basic == "" {
		t.Fatal("basic tier returned empty")
	}

	advanced := fieldMaskForTier("advanced")
	if len(advanced) <= len(basic) {
		t.Error("advanced should have more fields than basic")
	}

	preferred := fieldMaskForTier("preferred")
	if len(preferred) <= len(advanced) {
		t.Error("preferred should have more fields than advanced")
	}

	all := fieldMaskForTier("all")
	if all != "*" {
		t.Errorf("all tier = %q, want *", all)
	}

	// Default falls through to advanced
	def := fieldMaskForTier("unknown")
	if def != advanced {
		t.Error("unknown tier should default to advanced")
	}
}

func TestDetailFieldMask(t *testing.T) {
	basic := detailFieldMask("basic")
	if basic == "" {
		t.Fatal("basic tier returned empty")
	}
	// Detail fields should NOT have "places." prefix
	if len(basic) > 7 && basic[:7] == "places." {
		t.Error("detail field mask should not have places. prefix")
	}

	all := detailFieldMask("all")
	if all != "*" {
		t.Errorf("all tier = %q, want *", all)
	}
}

func TestToPlaceSummary(t *testing.T) {
	openNow := true
	p := &api.GoogleMapsPlacesV1Place{
		Name:                     "places/ChIJ1",
		DisplayName:              &api.GoogleTypeLocalizedText{Text: "Test Place"},
		FormattedAddress:         "123 Test St",
		Types:                    []string{"cafe"},
		Rating:                   4.5,
		UserRatingCount:          100,
		PriceLevel:               "PRICE_LEVEL_MODERATE",
		BusinessStatus:           "OPERATIONAL",
		GoogleMapsUri:            "https://maps.google.com",
		InternationalPhoneNumber: "+1 555-0100",
		WebsiteUri:               "https://example.com",
		EditorialSummary:         &api.GoogleTypeLocalizedText{Text: "Great place"},
		RegularOpeningHours: &api.GoogleMapsPlacesV1PlaceOpeningHours{
			OpenNow: openNow,
		},
	}

	s := toPlaceSummary(p)
	if s.ID != "ChIJ1" {
		t.Errorf("ID = %q, want ChIJ1", s.ID)
	}
	if s.Name != "Test Place" {
		t.Errorf("Name = %q, want Test Place", s.Name)
	}
	if s.Rating != 4.5 {
		t.Errorf("Rating = %f, want 4.5", s.Rating)
	}
	if s.PriceLevel != "$$" {
		t.Errorf("PriceLevel = %q, want $$", s.PriceLevel)
	}
	if s.OpenNow == nil || !*s.OpenNow {
		t.Error("OpenNow should be true")
	}
	if s.EditorialSummary != "Great place" {
		t.Errorf("EditorialSummary = %q, want Great place", s.EditorialSummary)
	}
}

func TestToPlaceDetail(t *testing.T) {
	p := &api.GoogleMapsPlacesV1Place{
		Name:                     "places/ChIJ1",
		DisplayName:              &api.GoogleTypeLocalizedText{Text: "Test Place"},
		FormattedAddress:         "123 Test St",
		ShortFormattedAddress:    "123 Test St",
		Types:                    []string{"cafe"},
		PrimaryType:              "cafe",
		Rating:                   4.5,
		UserRatingCount:          100,
		Delivery:                 true,
		DineIn:                   true,
		Takeout:                  true,
		Location:                 &api.GoogleTypeLatLng{Latitude: 41.5, Longitude: -81.7},
		Reviews: []*api.GoogleMapsPlacesV1Review{
			{
				Rating:           5,
				AuthorAttribution: &api.GoogleMapsPlacesV1AuthorAttribution{DisplayName: "Bob"},
				Text:             &api.GoogleTypeLocalizedText{Text: "Excellent!"},
			},
		},
		Photos: []*api.GoogleMapsPlacesV1Photo{
			{Name: "places/ChIJ1/photos/abc", WidthPx: 1000, HeightPx: 800},
		},
		AddressComponents: []*api.GoogleMapsPlacesV1PlaceAddressComponent{
			{LongText: "123", ShortText: "123", Types: []string{"street_number"}},
		},
	}

	d := toPlaceDetail(p)
	if d.ID != "ChIJ1" {
		t.Errorf("ID = %q, want ChIJ1", d.ID)
	}
	if d.Location == nil || d.Location.Latitude != 41.5 {
		t.Error("Location not set correctly")
	}
	if len(d.Reviews) != 1 || d.Reviews[0].Author != "Bob" {
		t.Error("Reviews not set correctly")
	}
	if len(d.Photos) != 1 || d.Photos[0].WidthPx != 1000 {
		t.Error("Photos not set correctly")
	}
	if !d.Delivery || !d.DineIn || !d.Takeout {
		t.Error("Service flags not set correctly")
	}
}
