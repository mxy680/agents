package instagram

import (
	"encoding/json"
	"testing"
)

func TestStoriesListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	out := runCmd(t, root, "stories", "list")

	mustContain(t, out, "ID")
	mustContain(t, out, "story_111")
}

func TestStoriesListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	out := runCmd(t, root, "stories", "list", "--json")

	var items []StorySummary
	if err := json.Unmarshal([]byte(out), &items); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one story")
	}
	if items[0].ID != "story_111" {
		t.Errorf("expected ID=story_111, got %s", items[0].ID)
	}
}

func TestStoriesListDefaultsToSelf(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	// No --user-id; should use DSUserID from session
	out := runCmd(t, root, "stories", "list")
	mustContain(t, out, "story_111")
}

func TestStoriesListWithUserID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	out := runCmd(t, root, "stories", "list", "--user-id=99999")
	mustContain(t, out, "story_111")
}

func TestStoriesGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	out := runCmd(t, root, "stories", "get", "--story-id=story_111")

	mustContain(t, out, "ID:")
	mustContain(t, out, "story_111")
	mustContain(t, out, "Taken At:")
}

func TestStoriesGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	out := runCmd(t, root, "stories", "get", "--story-id=story_111", "--json")

	var item StorySummary
	if err := json.Unmarshal([]byte(out), &item); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if item.ID != "story_111" {
		t.Errorf("expected ID=story_111, got %s", item.ID)
	}
}

func TestStoriesGetMissingFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	err := runCmdErr(t, root, "stories", "get")
	if err == nil {
		t.Error("expected error when --story-id not provided")
	}
}

func TestStoriesViewersText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	out := runCmd(t, root, "stories", "viewers", "--story-id=story_111")

	mustContain(t, out, "USERNAME")
	mustContain(t, out, "viewer_user")
}

func TestStoriesViewersJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	out := runCmd(t, root, "stories", "viewers", "--story-id=story_111", "--json")

	var users []UserSummary
	if err := json.Unmarshal([]byte(out), &users); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(users) == 0 {
		t.Fatal("expected at least one viewer")
	}
	if users[0].Username != "viewer_user" {
		t.Errorf("expected username=viewer_user, got %s", users[0].Username)
	}
	if !users[0].IsVerified {
		t.Error("expected is_verified=true for viewer_user")
	}
}

func TestStoriesFeedText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	out := runCmd(t, root, "stories", "feed")

	mustContain(t, out, "USERNAME")
	mustContain(t, out, "followed_user")
}

func TestStoriesFeedJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	out := runCmd(t, root, "stories", "feed", "--json")

	var entries []storiesTrayEntry
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one tray entry")
	}
	if entries[0].Username != "followed_user" {
		t.Errorf("expected username=followed_user, got %s", entries[0].Username)
	}
	if entries[0].Count != 2 {
		t.Errorf("expected story_count=2, got %d", entries[0].Count)
	}
}


func TestStoriesStoryAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestStoriesCmd(factory))
	// "story" and "st" are aliases for "stories"
	out := runCmd(t, root, "story", "list")
	mustContain(t, out, "story_111")
}
