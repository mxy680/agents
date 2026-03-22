package x

import (
	"testing"
)

func TestNotificationsAll_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "all", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestNotificationsAll_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "all"})
		root.Execute() //nolint:errcheck
	})

	// Either shows notifications or "No notifications found."
	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestNotificationsMentions_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "mentions", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestNotificationsVerified_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "verified", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestNotificationsAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notif", "all", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output via 'notif' alias, got empty string")
	}
}

func TestNotificationsAll_WithLimit(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "all", "--limit=5", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}
