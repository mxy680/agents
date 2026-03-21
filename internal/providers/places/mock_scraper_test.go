package places

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// mockScraper returns a ScraperFunc that returns canned entries.
func mockScraper(entries []Entry, err error) ScraperFunc {
	return func(ctx context.Context, opts ScraperOptions) (ScraperResult, error) {
		if err != nil {
			return ScraperResult{}, err
		}
		result := make([]Entry, len(entries))
		copy(result, entries)
		if opts.Limit > 0 && len(result) > opts.Limit {
			result = result[:opts.Limit]
		}
		return ScraperResult{Entries: result}, nil
	}
}

// mockScraperWithCapture returns a ScraperFunc that captures the options passed to it.
func mockScraperWithCapture(entries []Entry) (ScraperFunc, *ScraperOptions) {
	var captured ScraperOptions
	fn := func(ctx context.Context, opts ScraperOptions) (ScraperResult, error) {
		captured = opts
		result := make([]Entry, len(entries))
		copy(result, entries)
		if opts.Limit > 0 && len(result) > opts.Limit {
			result = result[:opts.Limit]
		}
		return ScraperResult{Entries: result}, nil
	}
	return fn, &captured
}

// testEntries returns a set of test place entries.
func testEntries() []Entry {
	return []Entry{
		{
			Title:       "Coffee Corner",
			Category:    "Coffee shop",
			Categories:  []string{"Coffee shop", "Cafe"},
			Address:     "123 Main St, Cleveland, OH 44101",
			Phone:       "+1 216-555-0100",
			Website:     "https://coffeecorner.example.com",
			Link:        "https://maps.google.com/?cid=123",
			PlaceID:     "ChIJ1",
			Latitude:    41.4993,
			Longitude:   -81.6944,
			Rating:      4.5,
			ReviewCount: 120,
			PriceRange:  "$$",
			Status:      "OPERATIONAL",
			Description: "A cozy coffee shop with excellent espresso.",
			Emails:      []string{"info@coffeecorner.example.com"},
			OpenHours: map[string][]string{
				"Monday":  {"6:00 AM – 8:00 PM"},
				"Tuesday": {"6:00 AM – 8:00 PM"},
			},
			UserReviews: []Review{
				{Name: "Alice", Rating: 5, Description: "Best coffee in town!", When: "a week ago"},
			},
			Images: []Image{
				{Title: "Storefront", Image: "https://example.com/photo1.jpg"},
			},
		},
		{
			Title:       "Bean & Leaf",
			Category:    "Coffee shop",
			Address:     "456 Oak Ave, Cleveland, OH 44102",
			Phone:       "+1 216-555-0200",
			Rating:      4.2,
			ReviewCount: 85,
			PriceRange:  "$",
			Status:      "OPERATIONAL",
		},
		{
			Title:       "The Daily Grind",
			Category:    "Coffee shop",
			Address:     "789 Elm St, Cleveland, OH 44103",
			Rating:      3.8,
			ReviewCount: 45,
			Status:      "OPERATIONAL",
		},
	}
}

// captureStdout runs f with os.Stdout redirected to a pipe and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	buf := make([]byte, 65536)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// newTestRootCmd creates a root command with global flags for testing.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}

// errScraper returns an error-producing ScraperFunc for testing error paths.
func errScraper(msg string) ScraperFunc {
	return mockScraper(nil, fmt.Errorf("%s", msg))
}
