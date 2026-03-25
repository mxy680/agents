package yelp

import (
	"github.com/spf13/cobra"
)

// Provider implements the Yelp integration using session-based auth.
type Provider struct {
	clientFactory ClientFactory
}

// New creates a new Yelp provider using the Yelp internal web API.
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
		Use:     "yelp",
		Short:   "Interact with Yelp",
		Long:    "Search businesses, read reviews, manage collections, and more via Yelp's internal API.",
		Aliases: []string{"y"},
	}

	yelpCmd.AddCommand(newBusinessesCmd(p.clientFactory))
	yelpCmd.AddCommand(newReviewsCmd(p.clientFactory))
	yelpCmd.AddCommand(newEventsCmd(p.clientFactory))
	yelpCmd.AddCommand(newCategoriesCmd(p.clientFactory))
	yelpCmd.AddCommand(newAutocompleteCmd(p.clientFactory))
	yelpCmd.AddCommand(newTransactionsCmd(p.clientFactory))

	parent.AddCommand(yelpCmd)
}
