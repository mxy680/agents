package yelp

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"hello world", 20, "hello world"},
		{"hello world", 8, "hello..."},
		{"hello", 5, "hello"},
		{"", 10, ""},
		{"abcdef", 6, "abcdef"},
		{"abcdefg", 6, "abc..."},
	}
	for _, tt := range tests {
		got := truncate(tt.s, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
		}
	}
}

func TestTruncateUnicode(t *testing.T) {
	// Multi-byte unicode characters should be truncated by rune count, not byte count.
	// truncate(s, max) returns s[:max-3]+"..." when len(s) > max.
	s := "日本語テスト" // 6 runes
	got := truncate(s, 6)
	// len is exactly max, so no truncation
	if got != "日本語テスト" {
		t.Errorf("truncate unicode exact = %q, want %q", got, "日本語テスト")
	}
	got = truncate(s, 5)
	// len=6 > max=5, so return s[:2]+"..."
	if got != "日本..." {
		t.Errorf("truncate unicode = %q, want %q", got, "日本...")
	}
}

func TestFormatRating(t *testing.T) {
	tests := []struct {
		r    float64
		want string
	}{
		{4.5, "4.5"},
		{3.0, "3.0"},
		{0.0, "0.0"},
		{5.0, "5.0"},
		{1.75, "1.8"},
	}
	for _, tt := range tests {
		got := formatRating(tt.r)
		if got != tt.want {
			t.Errorf("formatRating(%f) = %q, want %q", tt.r, got, tt.want)
		}
	}
}

func TestFormatAddress(t *testing.T) {
	t.Run("display_address preferred", func(t *testing.T) {
		loc := BusinessLocation{
			Address1:       "800 N Point St",
			City:           "San Francisco",
			State:          "CA",
			ZipCode:        "94109",
			DisplayAddress: []string{"800 N Point St", "San Francisco, CA 94109"},
		}
		got := formatAddress(loc)
		want := "800 N Point St, San Francisco, CA 94109"
		if got != want {
			t.Errorf("formatAddress with display_address = %q, want %q", got, want)
		}
	})

	t.Run("fallback to fields", func(t *testing.T) {
		loc := BusinessLocation{
			Address1: "800 N Point St",
			City:     "San Francisco",
			State:    "CA",
			ZipCode:  "94109",
		}
		got := formatAddress(loc)
		want := "800 N Point St, San Francisco, CA, 94109"
		if got != want {
			t.Errorf("formatAddress fallback = %q, want %q", got, want)
		}
	})

	t.Run("only city and state", func(t *testing.T) {
		loc := BusinessLocation{
			City:  "Denver",
			State: "CO",
		}
		got := formatAddress(loc)
		want := "Denver, CO"
		if got != want {
			t.Errorf("formatAddress city/state = %q, want %q", got, want)
		}
	})

	t.Run("empty location", func(t *testing.T) {
		got := formatAddress(BusinessLocation{})
		if got != "" {
			t.Errorf("formatAddress empty = %q, want empty string", got)
		}
	})
}

func TestFormatCategories(t *testing.T) {
	t.Run("multiple categories", func(t *testing.T) {
		cats := []BusinessCategory{
			{Alias: "pizza", Title: "Pizza"},
			{Alias: "italian", Title: "Italian"},
		}
		got := formatCategories(cats)
		want := "Pizza, Italian"
		if got != want {
			t.Errorf("formatCategories = %q, want %q", got, want)
		}
	})

	t.Run("single category", func(t *testing.T) {
		cats := []BusinessCategory{
			{Alias: "coffee", Title: "Coffee & Tea"},
		}
		got := formatCategories(cats)
		if got != "Coffee & Tea" {
			t.Errorf("formatCategories single = %q, want %q", got, "Coffee & Tea")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := formatCategories([]BusinessCategory{})
		if got != "" {
			t.Errorf("formatCategories empty = %q, want empty string", got)
		}
	})
}

func TestFormatDistance(t *testing.T) {
	tests := []struct {
		meters float64
		want   string
	}{
		{0, "-"},
		{1609.344, "1.0 mi"},
		{804.672, "0.5 mi"},
		{3218.688, "2.0 mi"},
	}
	for _, tt := range tests {
		got := formatDistance(tt.meters)
		if got != tt.want {
			t.Errorf("formatDistance(%f) = %q, want %q", tt.meters, got, tt.want)
		}
	}
}

func TestOrDash(t *testing.T) {
	if got := orDash(""); got != "-" {
		t.Errorf("orDash(\"\") = %q, want \"-\"", got)
	}
	if got := orDash("hello"); got != "hello" {
		t.Errorf("orDash(\"hello\") = %q, want \"hello\"", got)
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		ss   []string
		sep  string
		want string
	}{
		{[]string{"a", "b", "c"}, ", ", "a, b, c"},
		{[]string{"only"}, ", ", "only"},
		{[]string{}, ", ", ""},
		{[]string{"x", "y"}, " | ", "x | y"},
	}
	for _, tt := range tests {
		got := joinStrings(tt.ss, tt.sep)
		if got != tt.want {
			t.Errorf("joinStrings(%v, %q) = %q, want %q", tt.ss, tt.sep, got, tt.want)
		}
	}
}

func TestTruncateBody(t *testing.T) {
	t.Run("short body", func(t *testing.T) {
		body := []byte("short error message")
		got := truncateBody(body)
		if got != "short error message" {
			t.Errorf("truncateBody short = %q", got)
		}
	})

	t.Run("long body truncated", func(t *testing.T) {
		body := make([]byte, 300)
		for i := range body {
			body[i] = 'x'
		}
		got := truncateBody(body)
		if len(got) != 203 { // 200 chars + "..."
			t.Errorf("truncateBody long: len=%d, want 203", len(got))
		}
		if got[200:] != "..." {
			t.Errorf("truncateBody should end with '...'")
		}
	})
}

func TestFormatBusinessDetail(t *testing.T) {
	detail := BusinessDetail{
		BusinessSummary: BusinessSummary{
			ID:          "test-biz-123",
			Name:        "Test Business",
			Rating:      4.5,
			ReviewCount: 100,
			Price:       "$$",
			Phone:       "+14155551234",
			URL:         "https://www.yelp.com/biz/test-biz-123",
			Location: BusinessLocation{
				Address1: "123 Main St",
				City:     "San Francisco",
				State:    "CA",
				ZipCode:  "94102",
			},
			Categories: []BusinessCategory{
				{Alias: "pizza", Title: "Pizza"},
			},
		},
		IsClaimed: true,
		Photos:    []string{"photo1.jpg", "photo2.jpg"},
		Hours: []BusinessHours{
			{HoursType: "REGULAR", IsOpenNow: true},
		},
	}

	lines := formatBusinessDetail(detail)

	if len(lines) == 0 {
		t.Fatal("formatBusinessDetail returned empty lines")
	}

	// Check key lines are present
	foundName := false
	foundRating := false
	foundPhotos := false
	foundHours := false
	for _, line := range lines {
		if line == "Name:        Test Business" {
			foundName = true
		}
		if line == "Rating:      4.5 (100 reviews)" {
			foundRating = true
		}
		if line == "Photos:      2 available" {
			foundPhotos = true
		}
		if line == "Hours (REGULAR): open now" {
			foundHours = true
		}
	}
	if !foundName {
		t.Errorf("missing Name line in %v", lines)
	}
	if !foundRating {
		t.Errorf("missing Rating line in %v", lines)
	}
	if !foundPhotos {
		t.Errorf("missing Photos line in %v", lines)
	}
	if !foundHours {
		t.Errorf("missing Hours line in %v", lines)
	}
}

func TestFormatEventDetail(t *testing.T) {
	event := EventSummary{
		ID:          "test-event-123",
		Name:        "Test Concert",
		Description: "A great concert.",
		TimeStart:   "2024-04-15 18:00:00",
		TimeEnd:     "2024-04-15 22:00:00",
		IsFree:      false,
		Cost:        50.0,
		Location: EventLocation{
			Address1: "100 Concert Ave",
			City:     "San Francisco",
			State:    "CA",
		},
		AttendingCount: 250,
		EventURL:       "https://example.com/concert",
	}

	lines := formatEventDetail(event)
	if len(lines) == 0 {
		t.Fatal("formatEventDetail returned empty lines")
	}

	foundName := false
	foundFree := false
	foundCost := false
	foundDesc := false
	for _, line := range lines {
		if line == "Name:        Test Concert" {
			foundName = true
		}
		if line == "Free:        no" {
			foundFree = true
		}
		if line == "Cost:        $50.00" {
			foundCost = true
		}
		if line == "Description: A great concert." {
			foundDesc = true
		}
	}
	if !foundName {
		t.Errorf("missing Name line in %v", lines)
	}
	if !foundFree {
		t.Errorf("missing Free line in %v", lines)
	}
	if !foundCost {
		t.Errorf("missing Cost line in %v", lines)
	}
	if !foundDesc {
		t.Errorf("missing Description line in %v", lines)
	}
}

func TestFormatEventDetailFree(t *testing.T) {
	event := EventSummary{
		ID:        "free-event",
		Name:      "Free Event",
		TimeStart: "2024-05-01 10:00:00",
		IsFree:    true,
		Cost:      0,
		Location: EventLocation{
			City:  "Chicago",
			State: "IL",
		},
	}

	lines := formatEventDetail(event)
	foundFree := false
	foundCost := false
	for _, line := range lines {
		if line == "Free:        yes" {
			foundFree = true
		}
		if line == "Cost:        -" {
			foundCost = true
		}
	}
	if !foundFree {
		t.Errorf("missing Free=yes line in %v", lines)
	}
	if !foundCost {
		t.Errorf("missing Cost=- line for free event in %v", lines)
	}
}

func TestFormatAutocompleteResult(t *testing.T) {
	t.Run("full result", func(t *testing.T) {
		result := AutocompleteResult{
			Terms: []AutocompleteTerm{
				{Text: "pizza"},
				{Text: "pizza delivery"},
			},
			Businesses: []AutocompleteBusinessResult{
				{ID: "pizzeria-abc", Name: "Pizzeria ABC"},
			},
			Categories: []AutocompleteCategory{
				{Alias: "pizza", Title: "Pizza"},
			},
		}
		lines := formatAutocompleteResult(result)
		if len(lines) == 0 {
			t.Fatal("formatAutocompleteResult returned empty lines")
		}

		foundTerms := false
		foundBiz := false
		foundCat := false
		for _, line := range lines {
			if line == "Terms:      pizza, pizza delivery" {
				foundTerms = true
			}
			if line == "Businesses:" {
				foundBiz = true
			}
			if line == "Categories:" {
				foundCat = true
			}
		}
		if !foundTerms {
			t.Errorf("missing Terms line in %v", lines)
		}
		if !foundBiz {
			t.Errorf("missing Businesses section in %v", lines)
		}
		if !foundCat {
			t.Errorf("missing Categories section in %v", lines)
		}
	})

	t.Run("empty result returns no suggestions message", func(t *testing.T) {
		result := AutocompleteResult{}
		lines := formatAutocompleteResult(result)
		if len(lines) != 1 || lines[0] != "No suggestions found." {
			t.Errorf("expected 'No suggestions found.' line, got %v", lines)
		}
	})
}
