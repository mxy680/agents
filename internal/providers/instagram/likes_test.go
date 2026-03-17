package instagram

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLikesLikeDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	out := runCmd(t, root, "likes", "like", "--media-id=111222333", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "111222333")
}

func TestLikesLike(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	out := runCmd(t, root, "likes", "like", "--media-id=111222333")

	mustContain(t, out, "Liked media 111222333")
}

func TestLikesLikeJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	out := runCmd(t, root, "likes", "like", "--media-id=111222333", "--json")

	var result likeActionResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result.Status != "ok" {
		t.Errorf("expected status=ok, got %s", result.Status)
	}
}

func TestLikesUnlikeDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	out := runCmd(t, root, "likes", "unlike", "--media-id=111222333", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestLikesUnlike(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	out := runCmd(t, root, "likes", "unlike", "--media-id=111222333")

	mustContain(t, out, "Unliked media 111222333")
}

func TestLikesListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	out := runCmd(t, root, "likes", "list", "--media-id=111222333")

	mustContain(t, out, "USERNAME")
	mustContain(t, out, "liker_user")
}

func TestLikesListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	out := runCmd(t, root, "likes", "list", "--media-id=111222333", "--json")

	var users []UserSummary
	if err := json.Unmarshal([]byte(out), &users); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(users) == 0 {
		t.Fatal("expected at least one liker")
	}
	if users[0].Username != "liker_user" {
		t.Errorf("expected username=liker_user, got %s", users[0].Username)
	}
}

func TestLikesListMissingFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	err := runCmdErr(t, root, "likes", "list")
	if err == nil {
		t.Error("expected error when --media-id not provided")
	}
}

func TestLikesLikedText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	out := runCmd(t, root, "likes", "liked")

	mustContain(t, out, "ID")
	mustContain(t, out, "liked_post_111")
	mustContain(t, out, "Liked post caption")
}

func TestLikesLikedJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	out := runCmd(t, root, "likes", "liked", "--json")

	dec := json.NewDecoder(strings.NewReader(out))
	var items []MediaSummary
	if err := dec.Decode(&items); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one liked post")
	}
	if items[0].ID != "liked_post_111" {
		t.Errorf("expected ID=liked_post_111, got %s", items[0].ID)
	}
	if items[0].LikeCount != 300 {
		t.Errorf("expected like_count=300, got %d", items[0].LikeCount)
	}
}

func TestLikesAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestLikesCmd(factory))
	// "like" is an alias for "likes"
	out := runCmd(t, root, "like", "liked")
	mustContain(t, out, "liked_post_111")
}
