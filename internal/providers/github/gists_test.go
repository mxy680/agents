package github

import (
	"testing"
)

func TestGistsList(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output lists gists", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "list")
		mustContain(t, output, "gist1")
		mustContain(t, output, "My snippet")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "list", "--json")
		mustContain(t, output, "gist1")
		mustContain(t, output, "description")
	})

	t.Run("with limit flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "list", "--limit=5")
		mustContain(t, output, "gist1")
	})

	t.Run("with public flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "list", "--public")
		mustContain(t, output, "gist1")
	})
}

func TestGistsGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output shows gist detail", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "get", "--gist-id=gist1")
		mustContain(t, output, "gist1")
		mustContain(t, output, "My snippet")
		mustContain(t, output, "snippet.go")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "get", "--gist-id=gist1", "--json")
		mustContain(t, output, "gist1")
		mustContain(t, output, "description")
		mustContain(t, output, "files")
	})

	t.Run("missing gist-id returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "gists", "get")
		if err == nil {
			t.Fatal("expected error for missing --gist-id flag")
		}
	})
}

func TestGistsCreate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("dry-run text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "create",
			"--description=My new gist",
			`--files={"hello.txt":{"content":"hello"}}`,
			"--dry-run",
		)
		mustContain(t, output, "My new gist")
	})

	t.Run("dry-run json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "create",
			"--description=My new gist",
			"--dry-run", "--json",
		)
		mustContain(t, output, "create")
		mustContain(t, output, "description")
	})

	t.Run("create gist text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "create",
			"--description=My new gist",
			`--files={"hello.txt":{"content":"hello"}}`,
		)
		mustContain(t, output, "Created:")
		mustContain(t, output, "gist-new1")
	})

	t.Run("create gist json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "create",
			"--description=My new gist",
			`--files={"hello.txt":{"content":"hello"}}`,
			"--json",
		)
		mustContain(t, output, "id")
		mustContain(t, output, "gist-new1")
	})

	t.Run("create public gist", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "create",
			"--description=Public gist",
			`--files={"hello.txt":{"content":"hello"}}`,
			"--public",
		)
		mustContain(t, output, "Created:")
	})

	t.Run("mutually exclusive --files and --files-file returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "gists", "create",
			"--description=test",
			`--files={"a.txt":{"content":"a"}}`,
			"--files-file=/tmp/some.json",
		)
		if err == nil {
			t.Fatal("expected error when both --files and --files-file are provided")
		}
	})
}

func TestGistsUpdate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("dry-run text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "update",
			"--gist-id=gist1",
			"--description=Updated snippet",
			"--dry-run",
		)
		mustContain(t, output, "gist1")
	})

	t.Run("dry-run json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "update",
			"--gist-id=gist1",
			"--description=Updated snippet",
			"--dry-run", "--json",
		)
		mustContain(t, output, "update")
		mustContain(t, output, "gist1")
	})

	t.Run("update gist text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "update",
			"--gist-id=gist1",
			"--description=Updated snippet",
		)
		mustContain(t, output, "Updated:")
		mustContain(t, output, "gist1")
	})

	t.Run("update gist json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "update",
			"--gist-id=gist1",
			"--description=Updated snippet",
			"--json",
		)
		mustContain(t, output, "id")
		mustContain(t, output, "gist1")
	})

	t.Run("update gist with files", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "update",
			"--gist-id=gist1",
			`--files={"snippet.go":{"content":"package main\nfunc main(){}"}}`,
		)
		mustContain(t, output, "Updated:")
	})

	t.Run("missing gist-id returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "gists", "update", "--description=test")
		if err == nil {
			t.Fatal("expected error for missing --gist-id flag")
		}
	})
}

func TestGistsDelete(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("delete without --confirm returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "gists", "delete", "--gist-id=gist1")
		if err == nil {
			t.Fatal("expected error when --confirm is not provided")
		}
	})

	t.Run("delete with --confirm text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "delete", "--gist-id=gist1", "--confirm")
		mustContain(t, output, "Deleted:")
		mustContain(t, output, "gist1")
	})

	t.Run("delete with --confirm json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "delete", "--gist-id=gist1", "--confirm", "--json")
		mustContain(t, output, "deleted")
		mustContain(t, output, "gistId")
	})

	t.Run("dry-run delete", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		output := runCmd(t, root, "gists", "delete", "--gist-id=gist1", "--dry-run")
		mustContain(t, output, "gist1")
	})

	t.Run("missing gist-id returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("gists", []string{"gist"},
			newGistsListCmd(factory),
			newGistsGetCmd(factory),
			newGistsCreateCmd(factory),
			newGistsUpdateCmd(factory),
			newGistsDeleteCmd(factory),
		))
		err := runCmdErr(t, root, "gists", "delete", "--confirm")
		if err == nil {
			t.Fatal("expected error for missing --gist-id flag")
		}
	})
}
