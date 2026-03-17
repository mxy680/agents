package github

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newPullsCmd is a helper to build a fresh pulls command tree for each test.
func newPullsCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("pulls", []string{"pull", "pr"},
		newPullsListCmd(factory),
		newPullsGetCmd(factory),
		newPullsCreateCmd(factory),
		newPullsUpdateCmd(factory),
		newPullsMergeCmd(factory),
		newPullsReviewCmd(factory),
	)
}

func TestPullsList(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "list", "--owner=alice", "--repo=repo-alpha")
		mustContain(t, output, "feature-x")
		mustContain(t, output, "Add feature X")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "list", "--owner=alice", "--repo=repo-alpha", "--json")
		mustContain(t, output, `"number"`)
		mustContain(t, output, "feature-x")
	})

	t.Run("missing owner flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		err := runCmdErr(t, root, "pulls", "list", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error when --owner is missing")
		}
	})

	t.Run("missing repo flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		err := runCmdErr(t, root, "pulls", "list", "--owner=alice")
		if err == nil {
			t.Fatal("expected error when --repo is missing")
		}
	})

	t.Run("alias pr", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pr", "list", "--owner=alice", "--repo=repo-alpha")
		mustContain(t, output, "Add feature X")
	})
}

func TestPullsGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "get", "--owner=alice", "--repo=repo-alpha", "--number=5")
		mustContain(t, output, "Add feature X")
		mustContain(t, output, "feature-x")
		mustContain(t, output, "main")
		mustContain(t, output, "#5")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "get", "--owner=alice", "--repo=repo-alpha", "--number=5", "--json")
		mustContain(t, output, `"number"`)
		mustContain(t, output, `"title"`)
		mustContain(t, output, "Add feature X")
	})

	t.Run("missing number flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		err := runCmdErr(t, root, "pulls", "get", "--owner=alice", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error when --number is missing")
		}
	})
}

func TestPullsCreate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "create",
			"--owner=alice", "--repo=repo-alpha",
			"--title=My new PR", "--head=feature-y", "--base=main",
		)
		mustContain(t, output, "Created PR #10")
		mustContain(t, output, "My new PR")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "create",
			"--owner=alice", "--repo=repo-alpha",
			"--title=My new PR", "--head=feature-y", "--base=main",
			"--json",
		)
		mustContain(t, output, `"number"`)
		mustContain(t, output, "My new PR")
	})

	t.Run("dry-run text", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "create",
			"--owner=alice", "--repo=repo-alpha",
			"--title=My new PR", "--head=feature-y", "--base=main",
			"--dry-run",
		)
		mustContain(t, output, "[DRY RUN]")
		mustContain(t, output, "My new PR")
	})

	t.Run("dry-run json", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "create",
			"--owner=alice", "--repo=repo-alpha",
			"--title=My new PR", "--head=feature-y", "--base=main",
			"--dry-run", "--json",
		)
		mustContain(t, output, `"action"`)
		mustContain(t, output, "create_pull")
	})

	t.Run("missing required flags", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		err := runCmdErr(t, root, "pulls", "create", "--owner=alice", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error when --title, --head, --base are missing")
		}
	})
}

func TestPullsUpdate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "update",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--title=Updated PR title",
		)
		mustContain(t, output, "Updated PR #5")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "update",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--title=Updated PR title", "--json",
		)
		mustContain(t, output, `"number"`)
	})

	t.Run("dry-run text", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "update",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--title=Updated PR title", "--dry-run",
		)
		mustContain(t, output, "[DRY RUN]")
		mustContain(t, output, "#5")
	})

	t.Run("dry-run json", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "update",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--dry-run", "--json",
		)
		mustContain(t, output, "update_pull")
	})

	t.Run("missing number flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		err := runCmdErr(t, root, "pulls", "update", "--owner=alice", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error when --number is missing")
		}
	})
}

func TestPullsMerge(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "merge",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
		)
		mustContain(t, output, "Merged: true")
		mustContain(t, output, "abc123def456")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "merge",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--json",
		)
		mustContain(t, output, `"merged"`)
		mustContain(t, output, "abc123def456")
	})

	t.Run("dry-run text", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "merge",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--dry-run",
		)
		mustContain(t, output, "[DRY RUN]")
		mustContain(t, output, "#5")
	})

	t.Run("dry-run json", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "merge",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--dry-run", "--json",
		)
		mustContain(t, output, "merge_pull")
	})

	t.Run("invalid merge method", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		err := runCmdErr(t, root, "pulls", "merge",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--method=invalid",
		)
		if err == nil {
			t.Fatal("expected error for invalid merge method")
		}
		if !strings.Contains(err.Error(), "invalid merge method") {
			t.Errorf("expected 'invalid merge method' in error, got: %v", err)
		}
	})

	t.Run("squash method", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "merge",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--method=squash",
		)
		mustContain(t, output, "Merged: true")
	})

	t.Run("missing number flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		err := runCmdErr(t, root, "pulls", "merge", "--owner=alice", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error when --number is missing")
		}
	})
}

func TestPullsReview(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output approve", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "review",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--event=APPROVE",
		)
		mustContain(t, output, "Review submitted")
		mustContain(t, output, "APPROVE")
	})

	t.Run("text output request changes", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "review",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--event=REQUEST_CHANGES", "--body=Please fix the tests",
		)
		mustContain(t, output, "Review submitted")
		mustContain(t, output, "REQUEST_CHANGES")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "review",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--event=COMMENT", "--body=Looks good",
			"--json",
		)
		mustContain(t, output, `"id"`)
		mustContain(t, output, "COMMENT")
	})

	t.Run("dry-run text", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "review",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--event=APPROVE", "--dry-run",
		)
		mustContain(t, output, "[DRY RUN]")
		mustContain(t, output, "APPROVE")
	})

	t.Run("dry-run json", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		output := runCmd(t, root, "pulls", "review",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--event=APPROVE", "--dry-run", "--json",
		)
		mustContain(t, output, "review_pull")
	})

	t.Run("invalid review event", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		err := runCmdErr(t, root, "pulls", "review",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
			"--event=INVALID",
		)
		if err == nil {
			t.Fatal("expected error for invalid review event")
		}
		if !strings.Contains(err.Error(), "invalid review event") {
			t.Errorf("expected 'invalid review event' in error, got: %v", err)
		}
	})

	t.Run("missing event flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newPullsCmd(factory))

		err := runCmdErr(t, root, "pulls", "review",
			"--owner=alice", "--repo=repo-alpha", "--number=5",
		)
		if err == nil {
			t.Fatal("expected error when --event is missing")
		}
	})
}
