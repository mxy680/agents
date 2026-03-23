package zillow

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

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		price int64
		want  string
	}{
		{0, "-"},
		{500, "$500"},
		{1000, "$1,000"},
		{450000, "$450,000"},
		{1250000, "$1,250,000"},
		{10000000, "$10,000,000"},
		{999, "$999"},
		{1001, "$1,001"},
	}
	for _, tt := range tests {
		got := formatPrice(tt.price)
		if got != tt.want {
			t.Errorf("formatPrice(%d) = %q, want %q", tt.price, got, tt.want)
		}
	}
}

func TestJsonStr(t *testing.T) {
	m := map[string]any{
		"name":  "Denver",
		"zpid":  float64(12345678),
		"other": true,
	}

	got := jsonStr(m, "name")
	if got != "Denver" {
		t.Errorf("jsonStr string value = %q, want %q", got, "Denver")
	}

	got = jsonStr(m, "zpid")
	if got != "12345678" {
		t.Errorf("jsonStr float64 value = %q, want %q", got, "12345678")
	}

	got = jsonStr(m, "missing")
	if got != "" {
		t.Errorf("jsonStr missing key = %q, want empty string", got)
	}

	got = jsonStr(m, "other")
	if got != "" {
		t.Errorf("jsonStr bool value = %q, want empty string (unsupported type)", got)
	}
}

func TestExtractZPIDFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			"https://www.zillow.com/homedetails/123-Main-St-Denver-CO-80202/12345678_zpid/",
			"12345678",
		},
		{
			"https://www.zillow.com/homedetails/456-Oak-Ave/87654321_zpid/",
			"87654321",
		},
		{
			"https://www.zillow.com/homedetails/some-address/",
			"",
		},
		{
			"",
			"",
		},
		{
			"https://www.zillow.com/homes/for_sale/",
			"",
		},
	}
	for _, tt := range tests {
		got := extractZPIDFromURL(tt.url)
		if got != tt.want {
			t.Errorf("extractZPIDFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
