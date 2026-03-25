package yelp

import (
	"github.com/spf13/cobra"
)

// Provider implements the Yelp Fusion API v3 integration.
type Provider struct {
	clientFactory ClientFactory
}

// New creates a new Yelp provider using the real Yelp Fusion API.
func New() *Provider {
	return &Provider{
		clientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "yelp"
}

// RegisterCommands adds all Yelp subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	yelpCmd := &cobra.Command{
		Use:   "yelp",
		Short: "Search Yelp for businesses, events, and more",
		Long:  "Access Yelp Fusion API v3 to search businesses, read reviews, find events, and more. Requires YELP_API_KEY.",
	}

	yelpCmd.AddCommand(newBusinessesCmd(p.clientFactory))
	yelpCmd.AddCommand(newReviewsCmd(p.clientFactory))
	yelpCmd.AddCommand(newEventsCmd(p.clientFactory))
	yelpCmd.AddCommand(newCategoriesCmd(p.clientFactory))
	yelpCmd.AddCommand(newAutocompleteCmd(p.clientFactory))
	yelpCmd.AddCommand(newTransactionsCmd(p.clientFactory))

	parent.AddCommand(yelpCmd)
}
