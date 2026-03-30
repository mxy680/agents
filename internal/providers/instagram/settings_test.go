package instagram

import (
	"encoding/json"
	"testing"
)

func TestSettingsGetTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "settings", "get")
	mustContain(t, out, "testuser")
	mustContain(t, out, "Username:")
}

func TestSettingsGetJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "--json", "settings", "get")
	var result rawUserSettings
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if result.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", result.Username)
	}
}


func TestSettingsLoginActivityTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "settings", "login-activity")
	mustContain(t, out, "Login activity:")
}

func TestSettingsLoginActivityJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "--json", "settings", "login-activity")
	var result loginActivityResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if len(result.LoginActivity) == 0 {
		t.Fatal("expected at least one login activity entry")
	}
}

func TestSettingsAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)

	for _, alias := range []string{"setting", "account"} {
		root := newTestRootCmd()
		root.AddCommand(buildTestSettingsCmd(factory))
		out := runCmd(t, root, alias, "get")
		mustContain(t, out, "testuser")
	}
}
