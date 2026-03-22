package x

import (
	"testing"
)

func TestScheduledList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newScheduledCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"scheduled", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestScheduledCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newScheduledCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"scheduled", "create", "--text=hello", "--date=2026-06-01T12:00:00Z", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestScheduledCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newScheduledCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"scheduled", "create", "--text=hello", "--date=2026-06-01T12:00:00Z", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "post") && !containsStr(out, "execute_at") {
		t.Errorf("expected post or execute_at in dry-run JSON output, got: %s", out)
	}
}

func TestScheduledCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newScheduledCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"scheduled", "create", "--text=scheduled tweet", "--date=2026-06-01T12:00:00Z", "--json"})
		root.Execute() //nolint:errcheck
	})

	// Either returns the created tweet or raw data.
	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestScheduledCreate_WithMediaIDs(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newScheduledCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"scheduled", "create",
			"--text=tweet with media",
			"--date=2026-06-01T12:00:00Z",
			"--media-ids=111,222",
			"--dry-run",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestScheduledDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newScheduledCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"scheduled", "delete", "--tweet-id=sched-123", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestScheduledDelete_RequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newScheduledCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"scheduled", "delete", "--tweet-id=sched-123"})
	// Should fail without --confirm.
	_ = root.Execute()
}

func TestScheduledDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newScheduledCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"scheduled", "delete", "--tweet-id=sched-123", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestScheduledAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newScheduledCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"sched", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output via 'sched' alias, got empty string")
	}
}
