package github

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

// buildTestGitCmd assembles the git command tree used in tests.
func buildTestGitCmd(factory ClientFactory) *cobra.Command {
	gitCmd := &cobra.Command{Use: "git"}

	refsCmd := buildTestCmd("refs", []string{"ref"},
		newRefsListCmd(factory),
		newRefsGetCmd(factory),
		newRefsCreateCmd(factory),
		newRefsUpdateCmd(factory),
		newRefsDeleteCmd(factory),
	)
	gitCmd.AddCommand(refsCmd)

	commitsCmd := buildTestCmd("commits", []string{"commit"},
		newGitCommitsGetCmd(factory),
		newGitCommitsCreateCmd(factory),
	)
	gitCmd.AddCommand(commitsCmd)

	treesCmd := buildTestCmd("trees", []string{"tree"},
		newGitTreesGetCmd(factory),
		newGitTreesCreateCmd(factory),
	)
	gitCmd.AddCommand(treesCmd)

	blobsCmd := buildTestCmd("blobs", []string{"blob"},
		newGitBlobsGetCmd(factory),
		newGitBlobsCreateCmd(factory),
	)
	gitCmd.AddCommand(blobsCmd)

	tagsCmd := buildTestCmd("tags", []string{"tag"},
		newGitTagsGetCmd(factory),
		newGitTagsCreateCmd(factory),
	)
	gitCmd.AddCommand(tagsCmd)

	return gitCmd
}

// --- Refs tests ---

func TestRefsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "refs", "list",
		"--owner=alice", "--repo=repo-alpha", "--namespace=heads")
	mustContain(t, out, "refs/heads/main")
	mustContain(t, out, "commit")
	mustContain(t, out, "abc123")
}

func TestRefsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "refs", "list",
		"--owner=alice", "--repo=repo-alpha", "--namespace=heads", "--json")

	var refs []RefInfo
	if err := json.Unmarshal([]byte(out), &refs); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(refs) == 0 {
		t.Fatal("expected at least one ref in JSON output")
	}
	if refs[0].Ref != "refs/heads/main" {
		t.Errorf("refs[0].Ref = %q, want %q", refs[0].Ref, "refs/heads/main")
	}
}

func TestRefsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "refs", "get",
		"--owner=alice", "--repo=repo-alpha", "--ref=heads/main")
	mustContain(t, out, "refs/heads/main")
	mustContain(t, out, "abc123")
}

func TestRefsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "refs", "get",
		"--owner=alice", "--repo=repo-alpha", "--ref=heads/main", "--json")

	var ref RefInfo
	if err := json.Unmarshal([]byte(out), &ref); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if ref.Ref != "refs/heads/main" {
		t.Errorf("ref.Ref = %q, want %q", ref.Ref, "refs/heads/main")
	}
	if ref.Object.SHA != "abc123" {
		t.Errorf("ref.Object.SHA = %q, want %q", ref.Object.SHA, "abc123")
	}
}

func TestRefsCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "refs", "create",
		"--owner=alice", "--repo=repo-alpha",
		"--ref=refs/heads/new-branch", "--sha=abc123", "--dry-run")
	mustContain(t, out, "Would create ref")
	mustContain(t, out, "refs/heads/new-branch")
}

func TestRefsCreateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "refs", "create",
		"--owner=alice", "--repo=repo-alpha",
		"--ref=refs/heads/new-branch", "--sha=abc123")
	mustContain(t, out, "Created ref")
}

func TestRefsDeleteConfirmRequired(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	err := runCmdErr(t, root, "git", "refs", "delete",
		"--owner=alice", "--repo=repo-alpha", "--ref=heads/old-branch")
	if err == nil {
		t.Error("expected error when --confirm not provided")
	}
}

// --- Commits tests ---

func TestGitCommitsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "commits", "get",
		"--owner=alice", "--repo=repo-alpha", "--sha=abc123")
	mustContain(t, out, "abc123")
	mustContain(t, out, "Initial commit")
	mustContain(t, out, "Alice")
}

func TestGitCommitsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "commits", "get",
		"--owner=alice", "--repo=repo-alpha", "--sha=abc123", "--json")

	var commit GitCommitInfo
	if err := json.Unmarshal([]byte(out), &commit); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if commit.SHA != "abc123" {
		t.Errorf("commit.SHA = %q, want %q", commit.SHA, "abc123")
	}
	if commit.Message != "Initial commit" {
		t.Errorf("commit.Message = %q, want %q", commit.Message, "Initial commit")
	}
}

func TestGitCommitsCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "commits", "create",
		"--owner=alice", "--repo=repo-alpha",
		"--message=My commit", "--tree=tree123", "--dry-run")
	mustContain(t, out, "Would create commit")
	mustContain(t, out, "My commit")
}

// --- Trees tests ---

func TestGitTreesGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "trees", "get",
		"--owner=alice", "--repo=repo-alpha", "--sha=tree123")
	mustContain(t, out, "tree123")
	mustContain(t, out, "main.go")
}

func TestGitTreesGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "trees", "get",
		"--owner=alice", "--repo=repo-alpha", "--sha=tree123", "--json")

	var tree GitTreeInfo
	if err := json.Unmarshal([]byte(out), &tree); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if tree.SHA != "tree123" {
		t.Errorf("tree.SHA = %q, want %q", tree.SHA, "tree123")
	}
	if len(tree.Tree) == 0 {
		t.Error("expected at least one tree entry")
	}
}

// --- Blobs tests ---

func TestGitBlobsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "blobs", "get",
		"--owner=alice", "--repo=repo-alpha", "--sha=blob1")
	mustContain(t, out, "blob1")
	mustContain(t, out, "base64")
}

func TestGitBlobsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "blobs", "get",
		"--owner=alice", "--repo=repo-alpha", "--sha=blob1", "--json")

	var blob GitBlobInfo
	if err := json.Unmarshal([]byte(out), &blob); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if blob.SHA != "blob1" {
		t.Errorf("blob.SHA = %q, want %q", blob.SHA, "blob1")
	}
	if blob.Encoding != "base64" {
		t.Errorf("blob.Encoding = %q, want %q", blob.Encoding, "base64")
	}
}

// --- Tags tests ---

func TestGitTagsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "tags", "get",
		"--owner=alice", "--repo=repo-alpha", "--sha=tag123")
	mustContain(t, out, "v1.0.0")
	mustContain(t, out, "Release v1.0.0")
	mustContain(t, out, "Alice")
}

func TestGitTagsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestGitCmd(factory))

	out := runCmd(t, root, "git", "tags", "get",
		"--owner=alice", "--repo=repo-alpha", "--sha=tag123", "--json")

	var tag GitTagInfo
	if err := json.Unmarshal([]byte(out), &tag); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if tag.Tag != "v1.0.0" {
		t.Errorf("tag.Tag = %q, want %q", tag.Tag, "v1.0.0")
	}
	if tag.Object.SHA != "abc123" {
		t.Errorf("tag.Object.SHA = %q, want %q", tag.Object.SHA, "abc123")
	}
}
