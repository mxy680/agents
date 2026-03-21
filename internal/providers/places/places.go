package places

import (
	"context"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
	api "google.golang.org/api/places/v1"
)

// ServiceFactory is a function that creates a Places API service.
type ServiceFactory func(ctx context.Context) (*api.Service, error)

// Provider implements the Google Places integration.
type Provider struct {
	// ServiceFactory creates the Places API service. Defaults to auth.NewPlacesService.
	// Override in tests to inject a mock service pointing at a test server.
	ServiceFactory ServiceFactory
}

// New creates a new Places provider using the real Places API.
func New() *Provider {
	return &Provider{
		ServiceFactory: auth.NewPlacesService,
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
		Short:   "Interact with Google Places",
		Long:    "Search for places, get details, autocomplete, and download photos via the Google Places API (New).",
		Aliases: []string{"place"},
	}

	searchCmd := &cobra.Command{
		Use:     "search",
		Short:   "Search for places",
		Aliases: []string{"find"},
	}
	searchCmd.AddCommand(newSearchTextCmd(p.ServiceFactory))
	searchCmd.AddCommand(newSearchNearbyCmd(p.ServiceFactory))
	placesCmd.AddCommand(searchCmd)

	placesCmd.AddCommand(newGetCmd(p.ServiceFactory))
	placesCmd.AddCommand(newAutocompleteCmd(p.ServiceFactory))

	photosCmd := &cobra.Command{
		Use:     "photos",
		Short:   "Manage place photos",
		Aliases: []string{"photo"},
	}
	photosCmd.AddCommand(newPhotosListCmd(p.ServiceFactory))
	photosCmd.AddCommand(newPhotosGetCmd(p.ServiceFactory))
	placesCmd.AddCommand(photosCmd)

	parent.AddCommand(placesCmd)
}
