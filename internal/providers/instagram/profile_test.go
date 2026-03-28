package instagram

import (
	"encoding/json"
	"testing"
)

func TestProfileGetByUsernameText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestProfileCmd(factory))
	out := runCmd(t, root, "profile", "get", "--username=omniclaw680")

	mustContain(t, out, "Username:")
	mustContain(t, out, "omniclaw680")
	mustContain(t, out, "Biography:")
}

func TestProfileGetByUsernameJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestProfileCmd(factory))
	out := runCmd(t, root, "profile", "get", "--username=omniclaw680", "--json")

	var detail UserDetail
	if err := json.Unmarshal([]byte(out), &detail); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if detail.Username != "omniclaw680" {
		t.Errorf("expected username=omniclaw680, got %s", detail.Username)
	}
	if detail.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestProfileGetByUserIDText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestProfileCmd(factory))
	out := runCmd(t, root, "profile", "get", "--user-id=42544748138")

	mustContain(t, out, "ID:")
	mustContain(t, out, "42544748138")
	mustContain(t, out, "Username:")
}

func TestProfileGetByUserIDJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestProfileCmd(factory))
	out := runCmd(t, root, "profile", "get", "--user-id=42544748138", "--json")

	var detail UserDetail
	if err := json.Unmarshal([]byte(out), &detail); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if detail.ID != "42544748138" {
		t.Errorf("expected ID=42544748138, got %s", detail.ID)
	}
	if detail.MediaCount != 15 {
		t.Errorf("expected MediaCount=15, got %d", detail.MediaCount)
	}
}

func TestProfileGetNoFlagsError(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestProfileCmd(factory))
	err := runCmdErr(t, root, "profile", "get")
	if err == nil {
		t.Error("expected error when no --username or --user-id provided")
	}
}

