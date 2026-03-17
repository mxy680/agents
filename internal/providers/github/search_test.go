package github

import (
	"testing"
)

func TestSearchRepos(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output lists matching repos", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "repos", "--query=repo-alpha")
		mustContain(t, output, "alice/repo-alpha")
		mustContain(t, output, "Found 1 results")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "repos", "--query=repo-alpha", "--json")
		mustContain(t, output, "totalCount")
		mustContain(t, output, "alice/repo-alpha")
	})

	t.Run("with sort and order flags", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "repos", "--query=repo-alpha", "--sort=stars", "--order=asc")
		mustContain(t, output, "alice/repo-alpha")
	})

	t.Run("with limit flag", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "repos", "--query=repo-alpha", "--limit=5")
		mustContain(t, output, "alice/repo-alpha")
	})

	t.Run("missing query returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		err := runCmdErr(t, root, "search", "repos")
		if err == nil {
			t.Fatal("expected error for missing --query flag")
		}
	})
}

func TestSearchCode(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output lists code results", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "code", "--query=main.go")
		mustContain(t, output, "main.go")
		mustContain(t, output, "alice/repo-alpha")
		mustContain(t, output, "Found 1 results")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "code", "--query=main.go", "--json")
		mustContain(t, output, "totalCount")
		mustContain(t, output, "main.go")
	})

	t.Run("with sort and order flags", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "code", "--query=main.go", "--sort=indexed", "--order=asc")
		mustContain(t, output, "main.go")
	})

	t.Run("missing query returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		err := runCmdErr(t, root, "search", "code")
		if err == nil {
			t.Fatal("expected error for missing --query flag")
		}
	})
}

func TestSearchIssues(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output lists issues", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "issues", "--query=bug")
		mustContain(t, output, "Bug report")
		mustContain(t, output, "Found 1 results")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "issues", "--query=bug", "--json")
		mustContain(t, output, "totalCount")
		mustContain(t, output, "Bug report")
	})

	t.Run("with sort and order flags", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "issues", "--query=bug", "--sort=created", "--order=asc")
		mustContain(t, output, "Bug report")
	})

	t.Run("missing query returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		err := runCmdErr(t, root, "search", "issues")
		if err == nil {
			t.Fatal("expected error for missing --query flag")
		}
	})
}

func TestSearchCommits(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output lists commits", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "commits", "--query=feat")
		mustContain(t, output, "feat: add feature")
		mustContain(t, output, "Found 1 results")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "commits", "--query=feat", "--json")
		mustContain(t, output, "totalCount")
		mustContain(t, output, "abc1234567890")
	})

	t.Run("with sort and order flags", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "commits", "--query=feat", "--sort=author-date", "--order=asc")
		mustContain(t, output, "feat: add feature")
	})

	t.Run("sha is truncated to 7 chars in text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "commits", "--query=feat")
		mustContain(t, output, "abc1234")
	})

	t.Run("missing query returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		err := runCmdErr(t, root, "search", "commits")
		if err == nil {
			t.Fatal("expected error for missing --query flag")
		}
	})
}

func TestSearchUsers(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output lists users", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "users", "--query=alice")
		mustContain(t, output, "alice")
		mustContain(t, output, "Found 1 results")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "users", "--query=alice", "--json")
		mustContain(t, output, "totalCount")
		mustContain(t, output, "alice")
	})

	t.Run("with sort and order flags", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "users", "--query=alice", "--sort=followers", "--order=asc")
		mustContain(t, output, "alice")
	})

	t.Run("user type is shown in text output", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		output := runCmd(t, root, "search", "users", "--query=alice")
		mustContain(t, output, "User")
	})

	t.Run("missing query returns error", func(t *testing.T) {
		root := newTestRootCmd()
		root.AddCommand(buildTestCmd("search", nil,
			newSearchReposCmd(factory),
			newSearchCodeCmd(factory),
			newSearchIssuesCmd(factory),
			newSearchCommitsCmd(factory),
			newSearchUsersCmd(factory),
		))
		err := runCmdErr(t, root, "search", "users")
		if err == nil {
			t.Fatal("expected error for missing --query flag")
		}
	})
}
