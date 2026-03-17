package github

import (
	"testing"
)

func TestReleasesList(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output lists releases", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "list", "--owner=alice", "--repo=repo-alpha")
		mustContain(t, output, "v1.0.0")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "list", "--owner=alice", "--repo=repo-alpha", "--json")
		mustContain(t, output, "v1.0.0")
		mustContain(t, output, "tagName")
	})

	t.Run("with limit flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "list", "--owner=alice", "--repo=repo-alpha", "--limit=5")
		mustContain(t, output, "v1.0.0")
	})

	t.Run("missing owner returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "releases", "list", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error for missing --owner flag")
		}
	})
}

func TestReleasesGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("get by --latest", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "get", "--owner=alice", "--repo=repo-alpha", "--latest")
		mustContain(t, output, "v1.0.0")
		mustContain(t, output, "Release 1.0")
	})

	t.Run("get by --tag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "get", "--owner=alice", "--repo=repo-alpha", "--tag=v1.0.0")
		mustContain(t, output, "v1.0.0")
		mustContain(t, output, "Tag:")
	})

	t.Run("get by --release-id", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "get", "--owner=alice", "--repo=repo-alpha", "--release-id=500")
		mustContain(t, output, "v1.0.0")
	})

	t.Run("get by --latest json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "get", "--owner=alice", "--repo=repo-alpha", "--latest", "--json")
		mustContain(t, output, "tagName")
		mustContain(t, output, "v1.0.0")
	})

	t.Run("no selector returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "releases", "get", "--owner=alice", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error when no selector is provided")
		}
	})

	t.Run("multiple selectors returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "releases", "get", "--owner=alice", "--repo=repo-alpha", "--latest", "--tag=v1.0.0")
		if err == nil {
			t.Fatal("expected error when multiple selectors are provided")
		}
	})
}

func TestReleasesCreate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("dry-run text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "create",
			"--owner=alice", "--repo=repo-alpha", "--tag=v2.0.0",
			"--name=Release 2.0", "--body=Second release",
			"--dry-run",
		)
		mustContain(t, output, "v2.0.0")
	})

	t.Run("dry-run json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "create",
			"--owner=alice", "--repo=repo-alpha", "--tag=v2.0.0",
			"--dry-run", "--json",
		)
		mustContain(t, output, "create")
		mustContain(t, output, "v2.0.0")
	})

	t.Run("create release text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "create",
			"--owner=alice", "--repo=repo-alpha", "--tag=v2.0.0",
			"--name=Release 2.0",
		)
		mustContain(t, output, "Created:")
		mustContain(t, output, "v2.0.0")
	})

	t.Run("create release json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "create",
			"--owner=alice", "--repo=repo-alpha", "--tag=v2.0.0",
			"--json",
		)
		mustContain(t, output, "tagName")
		mustContain(t, output, "v2.0.0")
	})

	t.Run("create with draft and prerelease flags", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "create",
			"--owner=alice", "--repo=repo-alpha", "--tag=v2.0.0",
			"--draft", "--prerelease",
		)
		mustContain(t, output, "Created:")
	})

	t.Run("missing tag returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "releases", "create", "--owner=alice", "--repo=repo-alpha")
		if err == nil {
			t.Fatal("expected error for missing --tag flag")
		}
	})
}

func TestReleasesDelete(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("delete without --confirm returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "releases", "delete",
			"--owner=alice", "--repo=repo-alpha", "--release-id=500",
		)
		if err == nil {
			t.Fatal("expected error when --confirm is not provided")
		}
	})

	t.Run("delete with --confirm text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "delete",
			"--owner=alice", "--repo=repo-alpha", "--release-id=500", "--confirm",
		)
		mustContain(t, output, "Deleted:")
		mustContain(t, output, "500")
	})

	t.Run("delete with --confirm json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "delete",
			"--owner=alice", "--repo=repo-alpha", "--release-id=500", "--confirm", "--json",
		)
		mustContain(t, output, "deleted")
		mustContain(t, output, "releaseId")
	})

	t.Run("dry-run delete text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		output := runCmd(t, root, "releases", "delete",
			"--owner=alice", "--repo=repo-alpha", "--release-id=500", "--dry-run",
		)
		mustContain(t, output, "500")
	})

	t.Run("missing release-id returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("releases", []string{"release"},
			newReleasesListCmd(factory),
			newReleasesGetCmd(factory),
			newReleasesCreateCmd(factory),
			newReleasesDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "releases", "delete", "--owner=alice", "--repo=repo-alpha", "--confirm")
		if err == nil {
			t.Fatal("expected error for missing --release-id flag")
		}
	})
}
