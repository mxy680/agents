package instagram

import (
	"encoding/json"
	"testing"
)

func TestRelationshipsFollowersText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestRelationshipsCmd(factory))
	out := runCmd(t, root, "relationships", "followers", "--user-id=99999")

	mustContain(t, out, "USERNAME")
	mustContain(t, out, "rel_user")
}

func TestRelationshipsFollowersDefaultsToSelf(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestRelationshipsCmd(factory))
	// No --user-id; should use DSUserID from session
	out := runCmd(t, root, "relationships", "followers")
	mustContain(t, out, "rel_user")
}

func TestRelationshipsFollowersJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestRelationshipsCmd(factory))
	out := runCmd(t, root, "relationships", "followers", "--json")

	var users []UserSummary
	if err := json.Unmarshal([]byte(out), &users); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(users) == 0 {
		t.Fatal("expected at least one follower")
	}
	if users[0].Username != "rel_user" {
		t.Errorf("expected username=rel_user, got %s", users[0].Username)
	}
}

func TestRelationshipsFollowingText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestRelationshipsCmd(factory))
	out := runCmd(t, root, "relationships", "following")

	mustContain(t, out, "rel_user")
}


func TestRelationshipsBlockedText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestRelationshipsCmd(factory))
	out := runCmd(t, root, "relationships", "blocked")

	mustContain(t, out, "USERNAME")
	mustContain(t, out, "blocked_user")
}

func TestRelationshipsBlockedJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestRelationshipsCmd(factory))
	out := runCmd(t, root, "relationships", "blocked", "--json")

	var users []UserSummary
	if err := json.Unmarshal([]byte(out), &users); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(users) == 0 {
		t.Fatal("expected at least one blocked user")
	}
	if users[0].Username != "blocked_user" {
		t.Errorf("expected username=blocked_user, got %s", users[0].Username)
	}
}


func TestRelationshipsStatusText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestRelationshipsCmd(factory))
	out := runCmd(t, root, "relationships", "status", "--user-id=user_999")

	mustContain(t, out, "User ID:")
	mustContain(t, out, "user_999")
	mustContain(t, out, "Following:")
	mustContain(t, out, "Blocking:")
}

func TestRelationshipsStatusJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestRelationshipsCmd(factory))
	out := runCmd(t, root, "relationships", "status", "--user-id=user_999", "--json")

	var status RelationshipStatusResult
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if status.UserID != "user_999" {
		t.Errorf("expected UserID=user_999, got %s", status.UserID)
	}
	if !status.Following {
		t.Error("expected Following=true")
	}
}

func TestRelationshipsAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	// Test "rel" alias
	root := newTestRootCmd()
	root.AddCommand(buildTestRelationshipsCmd(factory))
	out := runCmd(t, root, "rel", "followers")
	mustContain(t, out, "rel_user")

	// Test "friendship" alias
	root2 := newTestRootCmd()
	root2.AddCommand(buildTestRelationshipsCmd(factory))
	out2 := runCmd(t, root2, "friendship", "following")
	mustContain(t, out2, "rel_user")
}


func TestRelationshipsStatusMissingFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestRelationshipsCmd(factory))
	err := runCmdErr(t, root, "relationships", "status")
	if err == nil {
		t.Error("expected error when --user-id not provided")
	}
}
