package linear

import (
	"github.com/spf13/cobra"
)

// Provider implements the Linear integration.
type Provider struct {
	// ClientFactory creates the Linear GraphQL client. Defaults to NewClient.
	// Override in tests to inject a mock client pointing at a test server.
	ClientFactory ClientFactory
}

// New creates a new Linear provider using the real Linear API.
func New() *Provider {
	return &Provider{
		ClientFactory: NewClient,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "linear"
}

// RegisterCommands adds all Linear subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	linearCmd := &cobra.Command{
		Use:     "linear",
		Short:   "Interact with Linear",
		Long:    "Manage issues, projects, cycles, teams, and more via the Linear GraphQL API.",
		Aliases: []string{"ln"},
	}

	issuesCmd := &cobra.Command{
		Use:     "issues",
		Short:   "Manage Linear issues",
		Aliases: []string{"issue", "i"},
	}
	issuesCmd.AddCommand(newIssuesListCmd(p.ClientFactory))
	issuesCmd.AddCommand(newIssuesGetCmd(p.ClientFactory))
	issuesCmd.AddCommand(newIssuesCreateCmd(p.ClientFactory))
	issuesCmd.AddCommand(newIssuesUpdateCmd(p.ClientFactory))
	issuesCmd.AddCommand(newIssuesDeleteCmd(p.ClientFactory))
	linearCmd.AddCommand(issuesCmd)

	projectsCmd := &cobra.Command{
		Use:     "projects",
		Short:   "Manage Linear projects",
		Aliases: []string{"proj", "p"},
	}
	projectsCmd.AddCommand(newProjectsListCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsGetCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsCreateCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsUpdateCmd(p.ClientFactory))
	projectsCmd.AddCommand(newProjectsDeleteCmd(p.ClientFactory))
	linearCmd.AddCommand(projectsCmd)

	cyclesCmd := &cobra.Command{
		Use:     "cycles",
		Short:   "View Linear cycles",
		Aliases: []string{"cycle", "c"},
	}
	cyclesCmd.AddCommand(newCyclesListCmd(p.ClientFactory))
	cyclesCmd.AddCommand(newCyclesGetCmd(p.ClientFactory))
	cyclesCmd.AddCommand(newCyclesCurrentCmd(p.ClientFactory))
	linearCmd.AddCommand(cyclesCmd)

	teamsCmd := &cobra.Command{
		Use:     "teams",
		Short:   "Manage Linear teams",
		Aliases: []string{"team", "t"},
	}
	teamsCmd.AddCommand(newTeamsListCmd(p.ClientFactory))
	teamsCmd.AddCommand(newTeamsGetCmd(p.ClientFactory))
	teamsCmd.AddCommand(newTeamsCreateCmd(p.ClientFactory))
	teamsCmd.AddCommand(newTeamsMembersCmd(p.ClientFactory))
	linearCmd.AddCommand(teamsCmd)

	commentsCmd := &cobra.Command{
		Use:     "comments",
		Short:   "Manage issue comments",
		Aliases: []string{"comment"},
	}
	commentsCmd.AddCommand(newCommentsListCmd(p.ClientFactory))
	commentsCmd.AddCommand(newCommentsCreateCmd(p.ClientFactory))
	commentsCmd.AddCommand(newCommentsDeleteCmd(p.ClientFactory))
	linearCmd.AddCommand(commentsCmd)

	labelsCmd := &cobra.Command{
		Use:     "labels",
		Short:   "Manage issue labels",
		Aliases: []string{"label", "l"},
	}
	labelsCmd.AddCommand(newLabelsListCmd(p.ClientFactory))
	labelsCmd.AddCommand(newLabelsCreateCmd(p.ClientFactory))
	labelsCmd.AddCommand(newLabelsDeleteCmd(p.ClientFactory))
	linearCmd.AddCommand(labelsCmd)

	usersCmd := &cobra.Command{
		Use:     "users",
		Short:   "View Linear users",
		Aliases: []string{"user", "u"},
	}
	usersCmd.AddCommand(newUsersListCmd(p.ClientFactory))
	usersCmd.AddCommand(newUsersGetCmd(p.ClientFactory))
	usersCmd.AddCommand(newUsersMeCmd(p.ClientFactory))
	linearCmd.AddCommand(usersCmd)

	workflowsCmd := &cobra.Command{
		Use:     "workflows",
		Short:   "View workflow states",
		Aliases: []string{"wf"},
	}
	workflowsCmd.AddCommand(newWorkflowsListCmd(p.ClientFactory))
	linearCmd.AddCommand(workflowsCmd)

	webhooksCmd := &cobra.Command{
		Use:     "webhooks",
		Short:   "Manage webhooks",
		Aliases: []string{"wh"},
	}
	webhooksCmd.AddCommand(newWebhooksListCmd(p.ClientFactory))
	webhooksCmd.AddCommand(newWebhooksCreateCmd(p.ClientFactory))
	webhooksCmd.AddCommand(newWebhooksDeleteCmd(p.ClientFactory))
	linearCmd.AddCommand(webhooksCmd)

	parent.AddCommand(linearCmd)
}
