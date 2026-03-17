package github

import (
	"testing"

	"github.com/spf13/cobra"
)

func buildTestBranchesCmd(factory ClientFactory) *cobra.Command {
	branchesCmd := buildTestCmd("branches", []string{"branch"},
		newBranchesListCmd(factory),
		newBranchesGetCmd(factory),
	)
	protectionCmd := buildTestCmd("protection", nil,
		newProtectionGetCmd(factory),
		newProtectionUpdateCmd(factory),
		newProtectionDeleteCmd(factory),
	)
	branchesCmd.AddCommand(protectionCmd)
	return branchesCmd
}

func TestBranchesList(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		output := runCmd(t, root, "branches", "list", "--owner=alice", "--repo=repo-alpha")
		mustContain(t, output, "main")
		mustContain(t, output, "dev")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		output := runCmd(t, root, "branches", "list", "--owner=alice", "--repo=repo-alpha", "--json")
		mustContain(t, output, `"name"`)
	})
}

func TestBranchesGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		output := runCmd(t, root, "branches", "get", "--owner=alice", "--repo=repo-alpha", "--branch=main")
		mustContain(t, output, "main")
		mustContain(t, output, "abc123")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		output := runCmd(t, root, "branches", "get", "--owner=alice", "--repo=repo-alpha", "--branch=main", "--json")
		mustContain(t, output, `"sha"`)
	})
}

func TestProtectionGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		output := runCmd(t, root, "branches", "protection", "get", "--owner=alice", "--repo=repo-alpha", "--branch=main")
		mustContain(t, output, "Enforce Admins")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		output := runCmd(t, root, "branches", "protection", "get", "--owner=alice", "--repo=repo-alpha", "--branch=main", "--json")
		mustContain(t, output, `"enforceAdmins"`)
	})
}

func TestProtectionDelete(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("requires confirm", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		err := runCmdErr(t, root, "branches", "protection", "delete", "--owner=alice", "--repo=repo-alpha", "--branch=main")
		if err == nil {
			t.Fatal("expected error without --confirm")
		}
	})

	t.Run("with confirm", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		output := runCmd(t, root, "branches", "protection", "delete", "--owner=alice", "--repo=repo-alpha", "--branch=main", "--confirm")
		mustContain(t, output, "Deleted")
	})

	t.Run("dry run", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		output := runCmd(t, root, "branches", "protection", "delete", "--owner=alice", "--repo=repo-alpha", "--branch=main", "--dry-run")
		mustContain(t, output, "DRY RUN")
	})
}

func TestProtectionUpdate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("dry run", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		output := runCmd(t, root, "branches", "protection", "update", "--owner=alice", "--repo=repo-alpha", "--branch=main",
			"--settings", `{"enforce_admins":true}`, "--dry-run")
		mustContain(t, output, "DRY RUN")
	})

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestBranchesCmd(factory))
		output := runCmd(t, root, "branches", "protection", "update", "--owner=alice", "--repo=repo-alpha", "--branch=main",
			"--settings", `{"enforce_admins":true}`)
		mustContain(t, output, "Branch protection updated")
	})
}
