package instagram

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLiveListTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "live", "list")
	mustContain(t, out, "broadcast_111")
	mustContain(t, out, "active")
}

func TestLiveListJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "--json", "live", "list")
	var result []LiveBroadcast
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one broadcast")
	}
	if result[0].ID != "broadcast_111" {
		t.Errorf("expected broadcast_111, got %s", result[0].ID)
	}
}

func TestLiveGetTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "live", "get", "--broadcast-id=broadcast_111")
	mustContain(t, out, "active")
	mustContain(t, out, "Viewers:")
}

func TestLiveGetJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "--json", "live", "get", "--broadcast-id=broadcast_111")
	var result LiveBroadcast
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if result.BroadcastStatus != "active" {
		t.Errorf("expected status active, got %s", result.BroadcastStatus)
	}
}

func TestLiveCommentsTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "live", "comments", "--broadcast-id=broadcast_111")
	mustContain(t, out, "Hello!")
}

func TestLiveCommentsJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "--json", "live", "comments", "--broadcast-id=broadcast_111")
	var result []rawLiveComment
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one comment")
	}
}

func TestLiveHeartbeatTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "live", "heartbeat", "--broadcast-id=broadcast_111")
	mustContain(t, out, "Viewer count:")
}

func TestLiveHeartbeatJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "--json", "live", "heartbeat", "--broadcast-id=broadcast_111")
	var result liveHeartbeatResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if result.ViewerCount != 505 {
		t.Errorf("expected 505 viewers, got %d", result.ViewerCount)
	}
}

func TestLiveLikeDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "live", "like", "--broadcast-id=broadcast_111", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestLiveLikeTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "live", "like", "--broadcast-id=broadcast_111")
	mustContain(t, out, "Liked broadcast")
}

func TestLiveLikeJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "--json", "live", "like", "--broadcast-id=broadcast_111")
	var result liveLikeResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}

func TestLivePostCommentDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "live", "post-comment", "--broadcast-id=broadcast_111", "--text=Hello", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestLivePostCommentTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "live", "post-comment", "--broadcast-id=broadcast_111", "--text=Hello")
	mustContain(t, out, "Posted comment")
}

func TestLivePostCommentJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "--json", "live", "post-comment", "--broadcast-id=broadcast_111", "--text=Hello")
	var result liveCommentResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}

func TestLiveAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLiveCmd(factory))

	out := runCmd(t, root, "broadcast", "list")
	mustContain(t, out, "broadcast_111")
}
