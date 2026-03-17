package github

import (
	"encoding/json"
	"testing"
)

// --- Orgs list tests ---

func TestOrgsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("orgs", []string{"org"},
		newOrgsListCmd(factory),
		newOrgsGetCmd(factory),
		newOrgsMembersCmd(factory),
		newOrgsReposCmd(factory),
	))

	out := runCmd(t, root, "orgs", "list")
	mustContain(t, out, "acme-corp")
}

func TestOrgsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("orgs", []string{"org"},
		newOrgsListCmd(factory),
		newOrgsGetCmd(factory),
		newOrgsMembersCmd(factory),
		newOrgsReposCmd(factory),
	))

	out := runCmd(t, root, "orgs", "list", "--json")

	var orgs []OrgSummary
	if err := json.Unmarshal([]byte(out), &orgs); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(orgs) == 0 {
		t.Fatal("expected at least one org in JSON output")
	}
	if orgs[0].Login != "acme-corp" {
		t.Errorf("orgs[0].Login = %q, want %q", orgs[0].Login, "acme-corp")
	}
}

// --- Orgs get tests ---

func TestOrgsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("orgs", []string{"org"},
		newOrgsListCmd(factory),
		newOrgsGetCmd(factory),
		newOrgsMembersCmd(factory),
		newOrgsReposCmd(factory),
	))

	out := runCmd(t, root, "orgs", "get", "--org=acme-corp")
	mustContain(t, out, "acme-corp")
	mustContain(t, out, "Acme Corporation")
	mustContain(t, out, "info@acme.com")
}

func TestOrgsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("orgs", []string{"org"},
		newOrgsListCmd(factory),
		newOrgsGetCmd(factory),
		newOrgsMembersCmd(factory),
		newOrgsReposCmd(factory),
	))

	out := runCmd(t, root, "orgs", "get", "--org=acme-corp", "--json")

	var org OrgDetail
	if err := json.Unmarshal([]byte(out), &org); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if org.Login != "acme-corp" {
		t.Errorf("org.Login = %q, want %q", org.Login, "acme-corp")
	}
	if org.Name != "Acme Corporation" {
		t.Errorf("org.Name = %q, want %q", org.Name, "Acme Corporation")
	}
	if org.PublicRepos != 25 {
		t.Errorf("org.PublicRepos = %d, want 25", org.PublicRepos)
	}
}

// --- Orgs members tests ---

func TestOrgsMembersText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("orgs", []string{"org"},
		newOrgsListCmd(factory),
		newOrgsGetCmd(factory),
		newOrgsMembersCmd(factory),
		newOrgsReposCmd(factory),
	))

	out := runCmd(t, root, "orgs", "members", "--org=acme-corp")
	mustContain(t, out, "alice")
	mustContain(t, out, "bob")
}

func TestOrgsMembersJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("orgs", []string{"org"},
		newOrgsListCmd(factory),
		newOrgsGetCmd(factory),
		newOrgsMembersCmd(factory),
		newOrgsReposCmd(factory),
	))

	out := runCmd(t, root, "orgs", "members", "--org=acme-corp", "--json")

	var members []MemberInfo
	if err := json.Unmarshal([]byte(out), &members); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(members) < 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
}

// --- Orgs repos tests ---

func TestOrgsReposText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("orgs", []string{"org"},
		newOrgsListCmd(factory),
		newOrgsGetCmd(factory),
		newOrgsMembersCmd(factory),
		newOrgsReposCmd(factory),
	))

	out := runCmd(t, root, "orgs", "repos", "--org=acme-corp")
	mustContain(t, out, "repo-alpha")
}

func TestOrgsReposJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("orgs", []string{"org"},
		newOrgsListCmd(factory),
		newOrgsGetCmd(factory),
		newOrgsMembersCmd(factory),
		newOrgsReposCmd(factory),
	))

	out := runCmd(t, root, "orgs", "repos", "--org=acme-corp", "--json")

	var repos []RepoSummary
	if err := json.Unmarshal([]byte(out), &repos); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(repos) == 0 {
		t.Fatal("expected at least one repo")
	}
	if repos[0].Name != "repo-alpha" {
		t.Errorf("repos[0].Name = %q, want %q", repos[0].Name, "repo-alpha")
	}
}
