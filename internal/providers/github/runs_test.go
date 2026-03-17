package github

import (
	"testing"

	"github.com/spf13/cobra"
)

// newRunsCmd is a helper to build a fresh runs command tree for each test.
func newRunsCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("runs", []string{"run"},
		newRunsListCmd(factory),
		newRunsGetCmd(factory),
		newRunsRerunCmd(factory),
		newRunsWorkflowsCmd(factory),
	)
}

func TestRunsList(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		output := runCmd(t, root, "runs", "list", "--owner=alice", "--repo=repo-alpha")
		mustContain(t, output, "1001")
		mustContain(t, output, "CI")
		mustContain(t, output, "completed")
		mustContain(t, output, "success")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		output := runCmd(t, root, "runs", "list", "--owner=alice", "--repo=repo-alpha", "--json")
		mustContain(t, output, `"id"`)
		mustContain(t, output, `"status"`)
		mustContain(t, output, "completed")
	})

	t.Run("alias run", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		output := runCmd(t, root, "run", "list", "--owner=alice", "--repo=repo-alpha")
		mustContain(t, output, "CI")
	})

	t.Run("missing owner flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		err := runCmdErr(t, root, "runs", "list", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error when --owner is missing")
		}
	})

	t.Run("missing repo flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		err := runCmdErr(t, root, "runs", "list", "--owner=alice")
		if err == nil {
			t.Fatal("expected error when --repo is missing")
		}
	})
}

func TestRunsGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		output := runCmd(t, root, "runs", "get", "--owner=alice", "--repo=repo-alpha", "--run-id=1001")
		mustContain(t, output, "1001")
		mustContain(t, output, "CI")
		mustContain(t, output, "completed")
		mustContain(t, output, "success")
		mustContain(t, output, "main")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		output := runCmd(t, root, "runs", "get", "--owner=alice", "--repo=repo-alpha", "--run-id=1001", "--json")
		mustContain(t, output, `"id"`)
		mustContain(t, output, `"workflowId"`)
		mustContain(t, output, `"runNumber"`)
		mustContain(t, output, "completed")
	})

	t.Run("missing run-id flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		err := runCmdErr(t, root, "runs", "get", "--owner=alice", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error when --run-id is missing")
		}
	})
}

func TestRunsRerun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		output := runCmd(t, root, "runs", "re-run", "--owner=alice", "--repo=repo-alpha", "--run-id=1001")
		mustContain(t, output, "Re-run triggered")
		mustContain(t, output, "1001")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		output := runCmd(t, root, "runs", "re-run", "--owner=alice", "--repo=repo-alpha", "--run-id=1001", "--json")
		mustContain(t, output, `"status"`)
		mustContain(t, output, "re-run triggered")
	})

	t.Run("missing run-id flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		err := runCmdErr(t, root, "runs", "re-run", "--owner=alice", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error when --run-id is missing")
		}
	})
}

func TestRunsWorkflows(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		output := runCmd(t, root, "runs", "workflows", "--owner=alice", "--repo=repo-alpha")
		mustContain(t, output, "100")
		mustContain(t, output, "CI")
		mustContain(t, output, ".github/workflows/ci.yml")
		mustContain(t, output, "active")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		output := runCmd(t, root, "runs", "workflows", "--owner=alice", "--repo=repo-alpha", "--json")
		mustContain(t, output, `"id"`)
		mustContain(t, output, `"path"`)
		mustContain(t, output, `"state"`)
		mustContain(t, output, "active")
	})

	t.Run("missing owner flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		err := runCmdErr(t, root, "runs", "workflows", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error when --owner is missing")
		}
	})

	t.Run("missing repo flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(newRunsCmd(factory))

		err := runCmdErr(t, root, "runs", "workflows", "--owner=alice")
		if err == nil {
			t.Fatal("expected error when --repo is missing")
		}
	})
}
