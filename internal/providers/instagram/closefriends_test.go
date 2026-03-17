package instagram

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCloseFriendsListTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCloseFriendsCmd(factory))

	out := runCmd(t, root, "closefriends", "list")
	mustContain(t, out, "bestie_user")
}

func TestCloseFriendsListJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCloseFriendsCmd(factory))

	out := runCmd(t, root, "--json", "closefriends", "list")
	var result []UserSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one user")
	}
	if result[0].Username != "bestie_user" {
		t.Errorf("expected bestie_user, got %s", result[0].Username)
	}
}

func TestCloseFriendsAddDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCloseFriendsCmd(factory))

	out := runCmd(t, root, "closefriends", "add", "--user-id=user_123", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestCloseFriendsAddTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCloseFriendsCmd(factory))

	out := runCmd(t, root, "closefriends", "add", "--user-id=user_123")
	mustContain(t, out, "Added user user_123 to close friends")
}

func TestCloseFriendsAddJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCloseFriendsCmd(factory))

	out := runCmd(t, root, "--json", "closefriends", "add", "--user-id=user_123")
	var result setBestiesResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if result.Status != "ok" {
		t.Errorf("expected status ok, got %s", result.Status)
	}
}

func TestCloseFriendsRemoveDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCloseFriendsCmd(factory))

	out := runCmd(t, root, "closefriends", "remove", "--user-id=user_123", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestCloseFriendsRemoveTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCloseFriendsCmd(factory))

	out := runCmd(t, root, "closefriends", "remove", "--user-id=user_123")
	mustContain(t, out, "Removed user user_123 from close friends")
}

func TestCloseFriendsRemoveJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCloseFriendsCmd(factory))

	out := runCmd(t, root, "--json", "closefriends", "remove", "--user-id=user_123")
	var result setBestiesResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}

func TestCloseFriendsAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)

	for _, alias := range []string{"cf", "besties"} {
		root := newTestRootCmd()
		root.AddCommand(buildTestCloseFriendsCmd(factory))
		out := runCmd(t, root, alias, "list")
		mustContain(t, out, "bestie_user")
	}
}
