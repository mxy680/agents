package places

import (
	"github.com/spf13/cobra"
)

// Provider implements the Google Places integration using google-maps-scraper.
type Provider struct {
	// ScraperFunc runs the scraper. Defaults to shelling out to the
	// google-maps-scraper binary. Override in tests with a mock.
	ScraperFunc ScraperFunc
}

// New creates a new Places provider using the real scraper binary.
func New() *Provider {
	return &Provider{
		ScraperFunc: defaultScraperFunc(defaultScraperBinary()),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "places"
}

// RegisterCommands adds all Places subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	placesCmd := &cobra.Command{
		Use:     "places",
		Short:   "Search Google Maps for businesses and places",
		Long:    "Search Google Maps by scraping — no API key needed. Returns rich data including address, phone, hours, reviews, ratings, and optionally emails.",
		Aliases: []string{"place"},
	}

	placesCmd.AddCommand(newSearchCmd(p.ScraperFunc))
	placesCmd.AddCommand(newLookupCmd(p.ScraperFunc))

	parent.AddCommand(placesCmd)
}
