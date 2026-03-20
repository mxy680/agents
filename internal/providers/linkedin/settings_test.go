package linkedin

import (
	"testing"
)

func TestSettingsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "PUBLIC") {
		t.Errorf("expected 'PUBLIC' visibility in output, got: %s", out)
	}
	if !containsStr(out, "CONNECTIONS") {
		t.Errorf("expected 'CONNECTIONS' messaging preference in output, got: %s", out)
	}
	if !containsStr(out, "true") {
		t.Errorf("expected 'true' active status in output, got: %s", out)
	}
}

func TestSettingsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"profile_visibility"`) {
		t.Errorf("expected 'profile_visibility' field in JSON output, got: %s", out)
	}
	if !containsStr(out, "PUBLIC") {
		t.Errorf("expected 'PUBLIC' in JSON output, got: %s", out)
	}
	if !containsStr(out, `"messaging_preference"`) {
		t.Errorf("expected 'messaging_preference' field in JSON output, got: %s", out)
	}
	if !containsStr(out, `"active_status"`) {
		t.Errorf("expected 'active_status' field in JSON output, got: %s", out)
	}
}

func TestSettingsPrivacy_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"settings", "privacy"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "PUBLIC") {
		t.Errorf("expected 'PUBLIC' visibility in output, got: %s", out)
	}
	if !containsStr(out, "ALL") {
		t.Errorf("expected 'ALL' connections visibility in output, got: %s", out)
	}
}

func TestSettingsPrivacy_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"settings", "privacy", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"profile_visibility"`) {
		t.Errorf("expected 'profile_visibility' field in JSON output, got: %s", out)
	}
	if !containsStr(out, `"connections_visibility"`) {
		t.Errorf("expected 'connections_visibility' field in JSON output, got: %s", out)
	}
	if !containsStr(out, `"last_name_visibility"`) {
		t.Errorf("expected 'last_name_visibility' field in JSON output, got: %s", out)
	}
	if !containsStr(out, `"profile_photo_visibility"`) {
		t.Errorf("expected 'profile_photo_visibility' field in JSON output, got: %s", out)
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

func TestSettingsAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSettingsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"setting", "get"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "PUBLIC") {
		t.Errorf("expected settings via alias 'setting' to work, got: %s", out)
	}
}
