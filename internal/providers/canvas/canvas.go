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
	canvasCmd.AddCommand(newDiscussionsCmd(p.ClientFactory))
	canvasCmd.AddCommand(newAnnouncementsCmd(p.ClientFactory))
	canvasCmd.AddCommand(newPagesCmd(p.ClientFactory))
	canvasCmd.AddCommand(newModulesCmd(p.ClientFactory))
	canvasCmd.AddCommand(newFilesCmd(p.ClientFactory))
	canvasCmd.AddCommand(newEnrollmentsCmd(p.ClientFactory))
	canvasCmd.AddCommand(newCalendarCmd(p.ClientFactory))
	canvasCmd.AddCommand(newConversationsCmd(p.ClientFactory))
	canvasCmd.AddCommand(newQuizzesCmd(p.ClientFactory))
	canvasCmd.AddCommand(newGroupsCmd(p.ClientFactory))
	canvasCmd.AddCommand(newRubricsCmd(p.ClientFactory))
	canvasCmd.AddCommand(newGradesCmd(p.ClientFactory))
	canvasCmd.AddCommand(newSectionsCmd(p.ClientFactory))

	parent.AddCommand(canvasCmd)
}
