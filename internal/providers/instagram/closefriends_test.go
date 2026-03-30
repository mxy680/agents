package instagram

import (
	"encoding/json"
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
