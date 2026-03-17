package github

import (
	"testing"

	"github.com/spf13/cobra"
)

// newIssuesTestCmd builds the issues command tree for tests.
func newIssuesTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("issues", []string{"issue"},
		newIssuesListCmd(factory),
		newIssuesGetCmd(factory),
		newIssuesCreateCmd(factory),
		newIssuesUpdateCmd(factory),
		newIssuesCloseCmd(factory),
		newIssuesCommentCmd(factory),
	)
}

func TestIssuesList(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "list", "--owner", "alice", "--repo", "repo-alpha")
		mustContain(t, output, "Bug report")
		mustContain(t, output, "Feature request")
		mustContain(t, output, "open")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "list", "--owner", "alice", "--repo", "repo-alpha", "--json")
		mustContain(t, output, `"number"`)
		mustContain(t, output, `"Bug report"`)
		mustContain(t, output, `"Feature request"`)
	})

	t.Run("with state filter", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "list", "--owner", "alice", "--repo", "repo-alpha", "--state", "closed")
		// Mock returns the same data regardless of filter; just verify command runs cleanly.
		mustContain(t, output, "Bug report")
	})

	t.Run("missing required owner flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		err := runCmdErr(t, root, "issues", "list", "--repo", "repo-alpha")
		if err == nil {
			t.Error("expected error when --owner is missing")
		}
	})

	t.Run("missing required repo flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		err := runCmdErr(t, root, "issues", "list", "--owner", "alice")
		if err == nil {
			t.Error("expected error when --repo is missing")
		}
	})
}

func TestIssuesGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "get", "--owner", "alice", "--repo", "repo-alpha", "--number", "1")
		mustContain(t, output, "Bug report")
		mustContain(t, output, "open")
		mustContain(t, output, "alice")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "get", "--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--json")
		mustContain(t, output, `"title"`)
		mustContain(t, output, `"Bug report"`)
		mustContain(t, output, `"body"`)
		mustContain(t, output, `"url"`)
	})

	t.Run("missing required number flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		err := runCmdErr(t, root, "issues", "get", "--owner", "alice", "--repo", "repo-alpha")
		if err == nil {
			t.Error("expected error when --number is missing")
		}
	})
}

func TestIssuesCreate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "create",
			"--owner", "alice", "--repo", "repo-alpha", "--title", "New issue")
		mustContain(t, output, "Created issue")
		mustContain(t, output, "#3")
		mustContain(t, output, "New issue")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "create",
			"--owner", "alice", "--repo", "repo-alpha", "--title", "New issue", "--json")
		mustContain(t, output, `"number"`)
		mustContain(t, output, `"New issue"`)
		mustContain(t, output, `"state"`)
	})

	t.Run("with body and labels", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "create",
			"--owner", "alice", "--repo", "repo-alpha",
			"--title", "Bug with details",
			"--body", "Detailed description",
			"--labels", "bug,enhancement")
		mustContain(t, output, "Created issue")
	})

	t.Run("missing required title flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		err := runCmdErr(t, root, "issues", "create", "--owner", "alice", "--repo", "repo-alpha")
		if err == nil {
			t.Error("expected error when --title is missing")
		}
	})
}

func TestIssuesUpdate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output update title", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "update",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--title", "Updated bug report")
		mustContain(t, output, "Updated issue")
		mustContain(t, output, "#1")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "update",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--title", "Updated", "--json")
		mustContain(t, output, `"number"`)
		mustContain(t, output, `"title"`)
	})

	t.Run("update state to closed", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "update",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--state", "closed")
		mustContain(t, output, "Updated issue")
	})

	t.Run("missing required number flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		err := runCmdErr(t, root, "issues", "update", "--owner", "alice", "--repo", "repo-alpha", "--title", "New")
		if err == nil {
			t.Error("expected error when --number is missing")
		}
	})
}

func TestIssuesClose(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "close",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1")
		mustContain(t, output, "Closed issue")
		mustContain(t, output, "#1")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "close",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--json")
		mustContain(t, output, `"number"`)
		mustContain(t, output, `"state"`)
		mustContain(t, output, `"closed"`)
	})

	t.Run("dry-run text", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "close",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--dry-run")
		mustContain(t, output, "[DRY RUN]")
		mustContain(t, output, "#1")
		mustContain(t, output, "alice/repo-alpha")
	})

	t.Run("dry-run json", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "close",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--dry-run", "--json")
		mustContain(t, output, `"action"`)
		mustContain(t, output, `"close"`)
		mustContain(t, output, `"state"`)
	})

	t.Run("missing required number flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		err := runCmdErr(t, root, "issues", "close", "--owner", "alice", "--repo", "repo-alpha")
		if err == nil {
			t.Error("expected error when --number is missing")
		}
	})
}

func TestIssuesComment(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "comment",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--body", "Great catch!")
		mustContain(t, output, "Added comment")
		mustContain(t, output, "101")
		mustContain(t, output, "#1")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "comment",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--body", "Great catch!", "--json")
		mustContain(t, output, `"id"`)
		mustContain(t, output, `"body"`)
		mustContain(t, output, `"user"`)
		mustContain(t, output, `"alice"`)
	})

	t.Run("dry-run text", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "comment",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--body", "Great catch!", "--dry-run")
		mustContain(t, output, "[DRY RUN]")
		mustContain(t, output, "#1")
		mustContain(t, output, "alice/repo-alpha")
	})

	t.Run("dry-run json", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		output := runCmd(t, root, "issues", "comment",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1", "--body", "Great catch!", "--dry-run", "--json")
		mustContain(t, output, `"action"`)
		mustContain(t, output, `"comment"`)
		mustContain(t, output, `"body"`)
	})

	t.Run("missing required body flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newIssuesTestCmd(factory))
		err := runCmdErr(t, root, "issues", "comment",
			"--owner", "alice", "--repo", "repo-alpha", "--number", "1")
		if err == nil {
			t.Error("expected error when --body is missing")
		}
	})
}
