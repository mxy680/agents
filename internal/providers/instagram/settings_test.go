package instagram

import (
	"encoding/json"
	"strings"
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

func TestSettingsPrivacyTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "settings", "privacy")
	mustContain(t, out, "Privacy settings")
}

func TestSettingsPrivacyJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "--json", "settings", "privacy")
	var result privacySettingsResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if result.Status != "ok" {
		t.Errorf("expected status ok, got %s", result.Status)
	}
}

func TestSettingsSetPrivateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "settings", "set-private", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestSettingsSetPrivateTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "settings", "set-private")
	mustContain(t, out, "private")
}

func TestSettingsSetPrivateJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "--json", "settings", "set-private")
	var result accountActionResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if result.Status != "ok" {
		t.Errorf("expected status ok, got %s", result.Status)
	}
}

func TestSettingsSetPublicDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "settings", "set-public", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestSettingsSetPublicTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "settings", "set-public")
	mustContain(t, out, "public")
}

func TestSettingsSetPublicJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "--json", "settings", "set-public")
	var result accountActionResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}

func TestSettingsTwoFactorStatusTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "settings", "two-factor-status")
	mustContain(t, out, "Two-factor authentication info")
}

func TestSettingsTwoFactorStatusJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	out := runCmd(t, root, "--json", "settings", "two-factor-status")
	var result twoFactorInfoResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if result.Status != "ok" {
		t.Errorf("expected status ok, got %s", result.Status)
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
