package instagram

import (
	"encoding/json"
	"testing"
)

func TestActivityFeedTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestActivityCmd(factory))

	out := runCmd(t, root, "activity", "feed")
	mustContain(t, out, "liked your photo")
}

func TestActivityFeedJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestActivityCmd(factory))

	out := runCmd(t, root, "--json", "activity", "feed")
	var result []ActivityItem
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one activity item")
	}
	if result[0].Text != "liked your photo" {
		t.Errorf("expected text 'liked your photo', got %s", result[0].Text)
	}
}

func TestActivityFeedWithLimit(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestActivityCmd(factory))

	out := runCmd(t, root, "activity", "feed", "--limit=5")
	mustContain(t, out, "liked your photo")
}


func TestActivityAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)

	for _, alias := range []string{"notifications", "notif"} {
		root := newTestRootCmd()
		root.AddCommand(buildTestActivityCmd(factory))
		out := runCmd(t, root, alias, "feed")
		mustContain(t, out, "liked your photo")
	}
}
