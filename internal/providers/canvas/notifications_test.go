package canvas

import (
	"strings"
	"testing"
)

func TestNotificationsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Canvas Maintenance") {
		t.Errorf("expected notification subject 'Canvas Maintenance', got: %s", output)
	}
}

func TestNotificationsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "list", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"subject"`) {
		t.Errorf("JSON output should contain subject field, got: %s", output)
	}
	if !strings.Contains(output, "Canvas Maintenance") {
		t.Errorf("JSON output should contain notification subject, got: %s", output)
	}
}

func TestNotificationsPreferencesText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "preferences"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Assignment Graded") {
		t.Errorf("expected preference 'Assignment Graded' in output, got: %s", output)
	}
	if !strings.Contains(output, "immediately") {
		t.Errorf("expected frequency 'immediately' in output, got: %s", output)
	}
}

func TestNotificationsUpdatePreferenceDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "update-preference", "--category", "Assignment Graded", "--frequency", "daily", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
}

func TestNotificationsUpdatePreferenceLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "update-preference", "--category", "Assignment Graded", "--frequency", "daily"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Assignment Graded") || !strings.Contains(output, "daily") {
		t.Errorf("expected preference update confirmation with category and frequency, got: %s", output)
	}
}

func TestNotificationsPreferencesJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "preferences", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "notification_preferences") {
		t.Errorf("expected notification_preferences in JSON output, got: %s", output)
	}
}

func TestNotificationsUpdatePreferenceJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "update-preference", "--category", "Assignment Graded", "--frequency", "daily", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "notification_preferences") {
		t.Errorf("expected notification_preferences in JSON output, got: %s", output)
	}
}

func TestNotificationsUpdatePreferenceMissingCategory(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "update-preference", "--frequency", "daily"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --category is missing")
	}
	if !strings.Contains(execErr.Error(), "--category") {
		t.Errorf("error should mention --category, got: %v", execErr)
	}
}

func TestNotificationsUpdatePreferenceInvalidFrequency(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "update-preference", "--category", "Assignment Graded", "--frequency", "invalid"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error for invalid frequency")
	}
	if !strings.Contains(execErr.Error(), "frequency") {
		t.Errorf("error should mention frequency, got: %v", execErr)
	}
}
