package instagram

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMediaListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "list")

	mustContain(t, out, "ID")
	mustContain(t, out, "111222333")
	mustContain(t, out, "Test caption")
}

func TestMediaListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "list", "--json")

	// Output may include a "Next cursor:" line after the JSON; decode only the JSON portion.
	dec := json.NewDecoder(strings.NewReader(out))
	var items []MediaSummary
	if err := dec.Decode(&items); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one media item")
	}
	if items[0].ID != "111222333" {
		t.Errorf("expected ID=111222333, got %s", items[0].ID)
	}
	if items[0].LikeCount != 42 {
		t.Errorf("expected like_count=42, got %d", items[0].LikeCount)
	}
}

func TestMediaListWithUserID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "list", "--user-id=99999")

	mustContain(t, out, "111222333")
}

func TestMediaListDefaultsToSelf(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	// No --user-id; should use DSUserID from session
	out := runCmd(t, root, "media", "list")
	mustContain(t, out, "111222333")
}

func TestMediaGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "get", "--media-id=111222333")

	mustContain(t, out, "ID:")
	mustContain(t, out, "111222333")
	mustContain(t, out, "Likes:")
}

func TestMediaGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "get", "--media-id=111222333", "--json")

	var item MediaSummary
	if err := json.Unmarshal([]byte(out), &item); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if item.ID != "111222333" {
		t.Errorf("expected ID=111222333, got %s", item.ID)
	}
	if item.LikeCount != 100 {
		t.Errorf("expected like_count=100, got %d", item.LikeCount)
	}
}

func TestMediaGetMissingFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	err := runCmdErr(t, root, "media", "get")
	if err == nil {
		t.Error("expected error when --media-id not provided")
	}
}

func TestMediaDeleteDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "delete", "--media-id=111222333", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "delete media 111222333")
}

func TestMediaDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	err := runCmdErr(t, root, "media", "delete", "--media-id=111222333")
	if err == nil {
		t.Error("expected error when --confirm not provided")
	}
}

func TestMediaDeleteConfirmed(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "delete", "--media-id=111222333", "--confirm")

	mustContain(t, out, "Deleted media 111222333")
}

func TestMediaDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "delete", "--media-id=111222333", "--confirm", "--json")

	var result mediaDeleteResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if !result.DidDelete {
		t.Error("expected did_delete=true")
	}
}

func TestMediaArchiveDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "archive", "--media-id=111222333", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestMediaArchive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "archive", "--media-id=111222333")

	mustContain(t, out, "Archived media 111222333")
}

func TestMediaUnarchive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "unarchive", "--media-id=111222333")

	mustContain(t, out, "Unarchived media 111222333")
}

func TestMediaLikersText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "likers", "--media-id=111222333")

	mustContain(t, out, "USERNAME")
	mustContain(t, out, "liker_user")
}

func TestMediaLikersJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "likers", "--media-id=111222333", "--json")

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

func TestMediaSaveDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "save", "--media-id=111222333", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestMediaSave(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "save", "--media-id=111222333")

	mustContain(t, out, "Saved media 111222333")
}

func TestMediaUnsave(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	out := runCmd(t, root, "media", "unsave", "--media-id=111222333")

	mustContain(t, out, "Unsaved media 111222333")
}

func TestMediaPostAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestMediaCmd(factory))
	// "post" and "posts" are aliases for "media"
	out := runCmd(t, root, "post", "list")
	mustContain(t, out, "111222333")
}
