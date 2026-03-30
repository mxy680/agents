package instagram

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReelsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	out := runCmd(t, root, "reels", "list")

	mustContain(t, out, "ID")
	mustContain(t, out, "reel_111")
}

func TestReelsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	out := runCmd(t, root, "reels", "list", "--json")

	// Output may include a "Next cursor:" line after the JSON; decode only the JSON portion.
	dec := json.NewDecoder(strings.NewReader(out))
	var items []ReelSummary
	if err := dec.Decode(&items); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one reel")
	}
	if items[0].ID != "reel_111" {
		t.Errorf("expected ID=reel_111, got %s", items[0].ID)
	}
	if items[0].PlayCount != 1000 {
		t.Errorf("expected play_count=1000, got %d", items[0].PlayCount)
	}
}

func TestReelsListDefaultsToSelf(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	// No --user-id; should use DSUserID from session
	out := runCmd(t, root, "reels", "list")
	mustContain(t, out, "reel_111")
}

func TestReelsListWithUserID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	out := runCmd(t, root, "reels", "list", "--user-id=99999")
	mustContain(t, out, "reel_111")
}

func TestReelsListShowsCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	out := runCmd(t, root, "reels", "list")
	// Mock returns more_available=true with a cursor
	mustContain(t, out, "reel_cursor_1")
}

func TestReelsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	out := runCmd(t, root, "reels", "get", "--reel-id=reel_111")

	mustContain(t, out, "ID:")
	mustContain(t, out, "reel_111")
	mustContain(t, out, "Plays:")
}

func TestReelsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	out := runCmd(t, root, "reels", "get", "--reel-id=reel_111", "--json")

	var item ReelSummary
	if err := json.Unmarshal([]byte(out), &item); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if item.ID != "reel_111" {
		t.Errorf("expected ID=reel_111, got %s", item.ID)
	}
	if item.PlayCount != 500 {
		// The /media/info/ mock returns play_count=500
		t.Errorf("expected play_count=500, got %d", item.PlayCount)
	}
}

func TestReelsGetMissingFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	err := runCmdErr(t, root, "reels", "get")
	if err == nil {
		t.Error("expected error when --reel-id not provided")
	}
}

func TestReelsFeedText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	out := runCmd(t, root, "reels", "feed")

	mustContain(t, out, "ID")
	mustContain(t, out, "feed_reel_222")
}

func TestReelsFeedJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	out := runCmd(t, root, "reels", "feed", "--json")

	var items []ReelSummary
	if err := json.Unmarshal([]byte(out), &items); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one reel in feed")
	}
	if items[0].ID != "feed_reel_222" {
		t.Errorf("expected ID=feed_reel_222, got %s", items[0].ID)
	}
	if items[0].PlayCount != 5000 {
		t.Errorf("expected play_count=5000, got %d", items[0].PlayCount)
	}
}


func TestReelsReelAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestReelsCmd(factory))
	// "reel" is an alias for "reels"
	out := runCmd(t, root, "reel", "list")
	mustContain(t, out, "reel_111")
}
