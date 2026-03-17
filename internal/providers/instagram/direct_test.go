package instagram

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDirectThreadsText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "threads")

	mustContain(t, out, "THREAD ID")
	mustContain(t, out, "thread_111")
	mustContain(t, out, "Test Thread")
}

func TestDirectThreadsJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "threads", "--json")

	dec := json.NewDecoder(strings.NewReader(out))
	var threads []DirectThreadSummary
	if err := dec.Decode(&threads); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(threads) == 0 {
		t.Fatal("expected at least one thread")
	}
	if threads[0].ThreadID != "thread_111" {
		t.Errorf("expected ThreadID=thread_111, got %s", threads[0].ThreadID)
	}
}

func TestDirectGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "get", "--thread-id=thread_111")

	mustContain(t, out, "ITEM ID")
	mustContain(t, out, "item_aaa")
	mustContain(t, out, "Hello there")
}

func TestDirectGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "get", "--thread-id=thread_111", "--json")

	var messages []DirectMessageSummary
	if err := json.Unmarshal([]byte(out), &messages); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(messages) == 0 {
		t.Fatal("expected at least one message")
	}
	if messages[0].ItemID != "item_aaa" {
		t.Errorf("expected ItemID=item_aaa, got %s", messages[0].ItemID)
	}
}

func TestDirectGetMissingFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	err := runCmdErr(t, root, "direct", "get")
	if err == nil {
		t.Error("expected error when --thread-id not provided")
	}
}

func TestDirectSendDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "send", "--thread-id=thread_111", "--text=Hello", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "thread_111")
}

func TestDirectSend(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "send", "--thread-id=thread_111", "--text=Hello")

	mustContain(t, out, "Message sent to thread thread_111")
}

func TestDirectCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "create", "--user-ids=user_aaa,user_bbb", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestDirectCreate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "create", "--user-ids=user_aaa")

	mustContain(t, out, "Group thread created")
}

func TestDirectDeleteMessageDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "delete-message", "--thread-id=thread_111", "--item-id=item_aaa", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestDirectDeleteMessageNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	err := runCmdErr(t, root, "direct", "delete-message", "--thread-id=thread_111", "--item-id=item_aaa")
	if err == nil {
		t.Error("expected error when --confirm not provided")
	}
}

func TestDirectDeleteMessageConfirmed(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "delete-message", "--thread-id=thread_111", "--item-id=item_aaa", "--confirm")

	mustContain(t, out, "Deleted message item_aaa from thread thread_111")
}

func TestDirectMarkSeen(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "mark-seen", "--thread-id=thread_111", "--item-id=item_aaa")

	mustContain(t, out, "Marked message item_aaa as seen in thread thread_111")
}

func TestDirectMarkSeenJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "mark-seen", "--thread-id=thread_111", "--item-id=item_aaa", "--json")

	var result directActionResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result.Status != "ok" {
		t.Errorf("expected status=ok, got %s", result.Status)
	}
}

func TestDirectPendingText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "pending")

	mustContain(t, out, "pending_thread_222")
	mustContain(t, out, "Pending Thread")
}

func TestDirectApproveDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "approve", "--thread-id=thread_111", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestDirectApprove(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "approve", "--thread-id=thread_111")

	mustContain(t, out, "Approved DM request for thread thread_111")
}

func TestDirectDeclineDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "decline", "--thread-id=thread_111", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestDirectDecline(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "direct", "decline", "--thread-id=thread_111")

	mustContain(t, out, "Declined DM request for thread thread_111")
}

func TestDirectAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	// Test "dm" alias
	root := newTestRootCmd()
	root.AddCommand(buildTestDirectCmd(factory))
	out := runCmd(t, root, "dm", "threads")
	mustContain(t, out, "thread_111")

	// Test "msg" alias
	root2 := newTestRootCmd()
	root2.AddCommand(buildTestDirectCmd(factory))
	out2 := runCmd(t, root2, "msg", "threads")
	mustContain(t, out2, "thread_111")
}
