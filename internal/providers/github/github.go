package github

import (
	"context"
	"net/http"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// ClientFactory is a function that creates an authenticated HTTP client for the GitHub API.
type ClientFactory func(ctx context.Context) (*http.Client, error)

// Provider implements the GitHub integration.
type Provider struct {
	// ClientFactory creates the GitHub HTTP client. Defaults to auth.NewGitHubClient.
	// Override in tests to inject a mock client pointing at a test server.
	ClientFactory ClientFactory
}

// New creates a new GitHub provider using the real GitHub API.
func New() *Provider {
	return &Provider{
		ClientFactory: auth.NewGitHubClient,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "github"
}

// RegisterCommands adds all GitHub subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	githubCmd := &cobra.Command{
		Use:   "github",
		Short: "Interact with GitHub",
		Long:  "Manage repositories, issues, pull requests, workflows, releases, gists, and more via the GitHub API.",
	}

	reposCmd := &cobra.Command{
		Use:     "repos",
		Short:   "Manage repositories",
		Aliases: []string{"repo"},
	}
	reposCmd.AddCommand(newReposListCmd(p.ClientFactory))
	reposCmd.AddCommand(newReposGetCmd(p.ClientFactory))
	reposCmd.AddCommand(newReposCreateCmd(p.ClientFactory))
	reposCmd.AddCommand(newReposForkCmd(p.ClientFactory))
	reposCmd.AddCommand(newReposDeleteCmd(p.ClientFactory))
	githubCmd.AddCommand(reposCmd)

	issuesCmd := &cobra.Command{
		Use:     "issues",
		Short:   "Manage issues",
		Aliases: []string{"issue"},
	}
	issuesCmd.AddCommand(newIssuesListCmd(p.ClientFactory))
	issuesCmd.AddCommand(newIssuesGetCmd(p.ClientFactory))
	issuesCmd.AddCommand(newIssuesCreateCmd(p.ClientFactory))
	issuesCmd.AddCommand(newIssuesUpdateCmd(p.ClientFactory))
	issuesCmd.AddCommand(newIssuesCloseCmd(p.ClientFactory))
	issuesCmd.AddCommand(newIssuesCommentCmd(p.ClientFactory))
	githubCmd.AddCommand(issuesCmd)

	pullsCmd := &cobra.Command{
		Use:     "pulls",
		Short:   "Manage pull requests",
		Aliases: []string{"pull", "pr"},
	}
	pullsCmd.AddCommand(newPullsListCmd(p.ClientFactory))
	pullsCmd.AddCommand(newPullsGetCmd(p.ClientFactory))
	pullsCmd.AddCommand(newPullsCreateCmd(p.ClientFactory))
	pullsCmd.AddCommand(newPullsUpdateCmd(p.ClientFactory))
	pullsCmd.AddCommand(newPullsMergeCmd(p.ClientFactory))
	pullsCmd.AddCommand(newPullsReviewCmd(p.ClientFactory))
	githubCmd.AddCommand(pullsCmd)

	runsCmd := &cobra.Command{
		Use:     "runs",
		Short:   "Manage workflow runs",
		Aliases: []string{"run"},
	}
	runsCmd.AddCommand(newRunsListCmd(p.ClientFactory))
	runsCmd.AddCommand(newRunsGetCmd(p.ClientFactory))
	runsCmd.AddCommand(newRunsRerunCmd(p.ClientFactory))
	runsCmd.AddCommand(newRunsWorkflowsCmd(p.ClientFactory))
	githubCmd.AddCommand(runsCmd)

	releasesCmd := &cobra.Command{
		Use:     "releases",
		Short:   "Manage releases",
		Aliases: []string{"release"},
	}
	releasesCmd.AddCommand(newReleasesListCmd(p.ClientFactory))
	releasesCmd.AddCommand(newReleasesGetCmd(p.ClientFactory))
	releasesCmd.AddCommand(newReleasesCreateCmd(p.ClientFactory))
	releasesCmd.AddCommand(newReleasesDeleteCmd(p.ClientFactory))
	githubCmd.AddCommand(releasesCmd)

	gistsCmd := &cobra.Command{
		Use:     "gists",
		Short:   "Manage gists",
		Aliases: []string{"gist"},
	}
	gistsCmd.AddCommand(newGistsListCmd(p.ClientFactory))
	gistsCmd.AddCommand(newGistsGetCmd(p.ClientFactory))
	gistsCmd.AddCommand(newGistsCreateCmd(p.ClientFactory))
	gistsCmd.AddCommand(newGistsUpdateCmd(p.ClientFactory))
	gistsCmd.AddCommand(newGistsDeleteCmd(p.ClientFactory))
	githubCmd.AddCommand(gistsCmd)

	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search GitHub resources",
	}
	searchCmd.AddCommand(newSearchReposCmd(p.ClientFactory))
	searchCmd.AddCommand(newSearchCodeCmd(p.ClientFactory))
	searchCmd.AddCommand(newSearchIssuesCmd(p.ClientFactory))
	searchCmd.AddCommand(newSearchCommitsCmd(p.ClientFactory))
	searchCmd.AddCommand(newSearchUsersCmd(p.ClientFactory))
	githubCmd.AddCommand(searchCmd)

	gitCmd := &cobra.Command{
		Use:   "git",
		Short: "Low-level Git data operations",
	}

	refsCmd := &cobra.Command{
		Use:     "refs",
		Short:   "Manage Git references",
		Aliases: []string{"ref"},
	}
	refsCmd.AddCommand(newRefsListCmd(p.ClientFactory))
	refsCmd.AddCommand(newRefsGetCmd(p.ClientFactory))
	refsCmd.AddCommand(newRefsCreateCmd(p.ClientFactory))
	refsCmd.AddCommand(newRefsUpdateCmd(p.ClientFactory))
	refsCmd.AddCommand(newRefsDeleteCmd(p.ClientFactory))
	gitCmd.AddCommand(refsCmd)

	commitsCmd := &cobra.Command{
		Use:     "commits",
		Short:   "Manage Git commits",
		Aliases: []string{"commit"},
	}
	commitsCmd.AddCommand(newGitCommitsGetCmd(p.ClientFactory))
	commitsCmd.AddCommand(newGitCommitsCreateCmd(p.ClientFactory))
	gitCmd.AddCommand(commitsCmd)

	treesCmd := &cobra.Command{
		Use:     "trees",
		Short:   "Manage Git trees",
		Aliases: []string{"tree"},
	}
	treesCmd.AddCommand(newGitTreesGetCmd(p.ClientFactory))
	treesCmd.AddCommand(newGitTreesCreateCmd(p.ClientFactory))
	gitCmd.AddCommand(treesCmd)

	blobsCmd := &cobra.Command{
		Use:     "blobs",
		Short:   "Manage Git blobs",
		Aliases: []string{"blob"},
	}
	blobsCmd.AddCommand(newGitBlobsGetCmd(p.ClientFactory))
	blobsCmd.AddCommand(newGitBlobsCreateCmd(p.ClientFactory))
	gitCmd.AddCommand(blobsCmd)

	tagsCmd := &cobra.Command{
		Use:     "tags",
		Short:   "Manage Git tags",
		Aliases: []string{"tag"},
	}
	tagsCmd.AddCommand(newGitTagsGetCmd(p.ClientFactory))
	tagsCmd.AddCommand(newGitTagsCreateCmd(p.ClientFactory))
	gitCmd.AddCommand(tagsCmd)

	githubCmd.AddCommand(gitCmd)

	orgsCmd := &cobra.Command{
		Use:     "orgs",
		Short:   "Manage organizations",
		Aliases: []string{"org"},
	}
	orgsCmd.AddCommand(newOrgsListCmd(p.ClientFactory))
	orgsCmd.AddCommand(newOrgsGetCmd(p.ClientFactory))
	orgsCmd.AddCommand(newOrgsMembersCmd(p.ClientFactory))
	orgsCmd.AddCommand(newOrgsReposCmd(p.ClientFactory))
	githubCmd.AddCommand(orgsCmd)

	teamsCmd := &cobra.Command{
		Use:     "teams",
		Short:   "Manage teams",
		Aliases: []string{"team"},
	}
	teamsCmd.AddCommand(newTeamsListCmd(p.ClientFactory))
	teamsCmd.AddCommand(newTeamsGetCmd(p.ClientFactory))
	teamsCmd.AddCommand(newTeamsMembersCmd(p.ClientFactory))
	teamsCmd.AddCommand(newTeamsReposCmd(p.ClientFactory))
	teamsCmd.AddCommand(newTeamsAddRepoCmd(p.ClientFactory))
	teamsCmd.AddCommand(newTeamsRemoveRepoCmd(p.ClientFactory))
	githubCmd.AddCommand(teamsCmd)

	labelsCmd := &cobra.Command{
		Use:     "labels",
		Short:   "Manage labels",
		Aliases: []string{"label"},
	}
	labelsCmd.AddCommand(newLabelsListCmd(p.ClientFactory))
	labelsCmd.AddCommand(newLabelsGetCmd(p.ClientFactory))
	labelsCmd.AddCommand(newLabelsCreateCmd(p.ClientFactory))
	labelsCmd.AddCommand(newLabelsUpdateCmd(p.ClientFactory))
	labelsCmd.AddCommand(newLabelsDeleteCmd(p.ClientFactory))
	githubCmd.AddCommand(labelsCmd)

	branchesCmd := &cobra.Command{
		Use:     "branches",
		Short:   "Manage branches",
		Aliases: []string{"branch"},
	}
	branchesCmd.AddCommand(newBranchesListCmd(p.ClientFactory))
	branchesCmd.AddCommand(newBranchesGetCmd(p.ClientFactory))

	protectionCmd := &cobra.Command{
		Use:   "protection",
		Short: "Manage branch protection rules",
	}
	protectionCmd.AddCommand(newProtectionGetCmd(p.ClientFactory))
	protectionCmd.AddCommand(newProtectionUpdateCmd(p.ClientFactory))
	protectionCmd.AddCommand(newProtectionDeleteCmd(p.ClientFactory))
	branchesCmd.AddCommand(protectionCmd)

	githubCmd.AddCommand(branchesCmd)

	parent.AddCommand(githubCmd)
}
