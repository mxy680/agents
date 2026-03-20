package linkedin

import (
	"testing"
)

func TestSettingsGet_Deprecated(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"settings", "get"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for deprecated settings get endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestSettingsGet_AliasDeprecated(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"setting", "get"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for deprecated settings get via alias")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message via alias, got: %s", err.Error())
	}
}

func TestSettingsPrivacy_Deprecated(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"settings", "privacy"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for deprecated settings privacy endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestSettingsVisibility_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"settings", "visibility", "--field", "profileVisibility", "--value", "PRIVATE", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected 'DRY RUN' in output, got: %s", out)
	}
	if !containsStr(out, "profileVisibility") {
		t.Errorf("expected field name in dry-run output, got: %s", out)
	}
}

func TestSettingsVisibility_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"settings", "visibility", "--field", "profileVisibility", "--value", "PRIVATE", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "profileVisibility") {
		t.Errorf("expected field name in dry-run JSON output, got: %s", out)
	}
	if !containsStr(out, "PRIVATE") {
		t.Errorf("expected value 'PRIVATE' in dry-run JSON output, got: %s", out)
	}
}

func TestSettingsVisibility_Update(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"settings", "visibility", "--field", "profileVisibility", "--value", "PRIVATE"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "profileVisibility") {
		t.Errorf("expected field name in update output, got: %s", out)
	}
	if !containsStr(out, "PRIVATE") {
		t.Errorf("expected value 'PRIVATE' in update output, got: %s", out)
	}
}

func TestSettingsVisibility_MissingField(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"settings", "visibility", "--value", "PRIVATE"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --field is missing")
	}
}

func TestSettingsVisibility_MissingValue(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"settings", "visibility", "--field", "profileVisibility"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --value is missing")
	}
}
