package github

import (
	"encoding/json"
	"testing"
)

// --- Teams list tests ---

func TestTeamsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "list", "--org=acme-corp")
	mustContain(t, out, "backend")
	mustContain(t, out, "Backend")
}

func TestTeamsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "list", "--org=acme-corp", "--json")

	var teams []TeamSummary
	if err := json.Unmarshal([]byte(out), &teams); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(teams) == 0 {
		t.Fatal("expected at least one team")
	}
	if teams[0].Slug != "backend" {
		t.Errorf("teams[0].Slug = %q, want %q", teams[0].Slug, "backend")
	}
}

// --- Teams get tests ---

func TestTeamsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "get", "--org=acme-corp", "--team-slug=backend")
	mustContain(t, out, "Backend")
	mustContain(t, out, "backend")
	mustContain(t, out, "push")
}

func TestTeamsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "get", "--org=acme-corp", "--team-slug=backend", "--json")

	var team TeamSummary
	if err := json.Unmarshal([]byte(out), &team); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if team.Name != "Backend" {
		t.Errorf("team.Name = %q, want %q", team.Name, "Backend")
	}
	if team.Permission != "push" {
		t.Errorf("team.Permission = %q, want %q", team.Permission, "push")
	}
}

// --- Teams members tests ---

func TestTeamsMembersText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "members", "--org=acme-corp", "--team-slug=backend")
	mustContain(t, out, "alice")
}

func TestTeamsMembersJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "members", "--org=acme-corp", "--team-slug=backend", "--json")

	var members []MemberInfo
	if err := json.Unmarshal([]byte(out), &members); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(members) == 0 {
		t.Fatal("expected at least one member")
	}
}

// --- Teams repos tests ---

func TestTeamsReposText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "repos", "--org=acme-corp", "--team-slug=backend")
	mustContain(t, out, "repo-alpha")
}

func TestTeamsReposJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "repos", "--org=acme-corp", "--team-slug=backend", "--json")

	var repos []RepoSummary
	if err := json.Unmarshal([]byte(out), &repos); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(repos) == 0 {
		t.Fatal("expected at least one repo")
	}
}

// --- Teams add-repo tests ---

func TestTeamsAddRepoDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "add-repo",
		"--org=acme-corp", "--team-slug=backend",
		"--owner=alice", "--repo=repo-alpha", "--dry-run")
	mustContain(t, out, "Would add repo")
	mustContain(t, out, "alice/repo-alpha")
}

func TestTeamsAddRepoText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "add-repo",
		"--org=acme-corp", "--team-slug=backend",
		"--owner=alice", "--repo=repo-alpha", "--permission=push")
	mustContain(t, out, "Added")
	mustContain(t, out, "alice/repo-alpha")
}

func TestTeamsAddRepoJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "add-repo",
		"--org=acme-corp", "--team-slug=backend",
		"--owner=alice", "--repo=repo-alpha", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["status"] != "added" {
		t.Errorf("status = %q, want %q", result["status"], "added")
	}
}

// --- Teams remove-repo tests ---

func TestTeamsRemoveRepoConfirmRequired(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	err := runCmdErr(t, root, "teams", "remove-repo",
		"--org=acme-corp", "--team-slug=backend",
		"--owner=alice", "--repo=repo-alpha")
	if err == nil {
		t.Error("expected error when --confirm not provided")
	}
}

func TestTeamsRemoveRepoWithConfirmText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "remove-repo",
		"--org=acme-corp", "--team-slug=backend",
		"--owner=alice", "--repo=repo-alpha", "--confirm")
	mustContain(t, out, "Removed")
	mustContain(t, out, "alice/repo-alpha")
}

func TestTeamsRemoveRepoWithConfirmJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("teams", []string{"team"},
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
		newTeamsReposCmd(factory),
		newTeamsAddRepoCmd(factory),
		newTeamsRemoveRepoCmd(factory),
	))

	out := runCmd(t, root, "teams", "remove-repo",
		"--org=acme-corp", "--team-slug=backend",
		"--owner=alice", "--repo=repo-alpha", "--confirm", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["status"] != "removed" {
		t.Errorf("status = %q, want %q", result["status"], "removed")
	}
}
