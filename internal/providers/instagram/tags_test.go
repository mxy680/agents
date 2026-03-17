package instagram

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTagsGetTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "tags", "get", "--name=golang")
	mustContain(t, out, "golang")
	mustContain(t, out, "Posts:")
}

func TestTagsGetJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "--json", "tags", "get", "--name=golang")
	var result TagSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if result.Name != "golang" {
		t.Errorf("expected tag name golang, got %s", result.Name)
	}
}

func TestTagsFeedTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "tags", "feed", "--name=golang")
	mustContain(t, out, "tag_media_111")
}

func TestTagsFeedJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "--json", "tags", "feed", "--name=golang")
	var result []MediaSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one media item")
	}
}

func TestTagsFeedWithTab(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "tags", "feed", "--name=golang", "--tab=recent")
	mustContain(t, out, "tag_media_111")
}

func TestTagsFollowDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "tags", "follow", "--name=golang", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestTagsFollowTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "tags", "follow", "--name=golang")
	mustContain(t, out, "Followed tag #golang")
}

func TestTagsFollowJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "--json", "tags", "follow", "--name=golang")
	var result tagActionResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}

func TestTagsUnfollowDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "tags", "unfollow", "--name=golang", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestTagsUnfollowTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "tags", "unfollow", "--name=golang")
	mustContain(t, out, "Unfollowed tag #golang")
}

func TestTagsFollowingTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "tags", "following")
	mustContain(t, out, "photography")
}

func TestTagsFollowingJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "--json", "tags", "following")
	var result []TagSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one tag")
	}
}

func TestTagsRelatedTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "tags", "related", "--name=golang")
	mustContain(t, out, "relatedgolang")
}

func TestTagsRelatedJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestTagsCmd(factory))

	out := runCmd(t, root, "--json", "tags", "related", "--name=golang")
	var result []TagSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
}

func TestTagsAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)

	for _, alias := range []string{"tag", "hashtag"} {
		root := newTestRootCmd()
		root.AddCommand(buildTestTagsCmd(factory))
		out := runCmd(t, root, alias, "get", "--name=golang")
		mustContain(t, out, "golang")
	}
}
