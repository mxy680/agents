package canvas

import (
	"github.com/spf13/cobra"
)

// Provider implements the Canvas LMS integration.
type Provider struct {
	ClientFactory ClientFactory
}

// New creates a new Canvas provider using the Canvas LMS REST API.
func New() *Provider {
	return &Provider{
		ClientFactory: DefaultClientFactory(),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "canvas"
}

// RegisterCommands adds all Canvas subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	canvasCmd := &cobra.Command{
		Use:     "canvas",
		Short:   "Interact with Canvas LMS",
		Long:    "View and manage courses, assignments, submissions, grades, discussions, and more via the Canvas LMS REST API.",
		Aliases: []string{"cvs"},
	}

	canvasCmd.AddCommand(newCoursesCmd(p.ClientFactory))
	canvasCmd.AddCommand(newAssignmentsCmd(p.ClientFactory))
	canvasCmd.AddCommand(newSubmissionsCmd(p.ClientFactory))
	canvasCmd.AddCommand(newUsersCmd(p.ClientFactory))

	parent.AddCommand(canvasCmd)
}
