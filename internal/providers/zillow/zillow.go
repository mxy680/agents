package zillow

import (
	"github.com/spf13/cobra"
)

// Provider implements the Zillow real estate integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new Zillow provider using the real Zillow APIs.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "zillow"
}

// RegisterCommands adds all Zillow subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	zillowCmd := &cobra.Command{
		Use:     "zillow",
		Short:   "Search Zillow for properties, agents, and real estate data",
		Long:    "Search Zillow's real estate listings, get property details, Zestimates, agent info, mortgage rates, and more — no API key needed.",
		Aliases: []string{"zw"},
	}

	zillowCmd.AddCommand(newPropertiesCmd(p.ClientFactory))
	zillowCmd.AddCommand(newZestimatesCmd(p.ClientFactory))
	zillowCmd.AddCommand(newAgentsCmd(p.ClientFactory))
	zillowCmd.AddCommand(newMortgageCmd(p.ClientFactory))
	zillowCmd.AddCommand(newSearchCmd(p.ClientFactory))
	zillowCmd.AddCommand(newWalkScoreCmd(p.ClientFactory))
	zillowCmd.AddCommand(newSchoolsCmd(p.ClientFactory))
	zillowCmd.AddCommand(newNeighborhoodsCmd(p.ClientFactory))
	zillowCmd.AddCommand(newBuildersCmd(p.ClientFactory))
	zillowCmd.AddCommand(newRentalsCmd(p.ClientFactory))

	parent.AddCommand(zillowCmd)
}
