package linkedin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotificationsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Alice Smith viewed your profile") {
		t.Errorf("expected notification text in output, got: %s", out)
	}
	if !containsStr(out, "Bob Jones liked your post") {
		t.Errorf("expected second notification text in output, got: %s", out)
	}
}

func TestNotificationsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"is_read"`) {
		t.Errorf("expected JSON field 'is_read' in output, got: %s", out)
	}
	if !containsStr(out, "urn:li:notification:111") {
		t.Errorf("expected notification URN in JSON output, got: %s", out)
	}
	if !containsStr(out, "urn:li:notification:222") {
		t.Errorf("expected second notification URN in JSON output, got: %s", out)
	}
}

func TestNotificationsList_Alias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notif", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Alice") {
		t.Errorf("expected output via 'notif' alias, got: %s", out)
	}
}

func TestNotificationsList_InvalidCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"notifications", "list", "--cursor", "notanumber"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid cursor")
	}
}

func TestNotificationsList_Empty(t *testing.T) {
	server := newEmptyNotificationsServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "No notifications found.") {
		t.Errorf("expected 'No notifications found.' in output, got: %s", out)
	}
}

func TestNotificationsMarkRead_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "mark-read", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestNotificationsMarkRead_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "mark-read"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "marked as read") {
		t.Errorf("expected 'marked as read' in output, got: %s", out)
	}
}

func TestNotificationsMarkRead_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"notifications", "mark-read", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON 'status' field in output, got: %s", out)
	}
	if !containsStr(out, "marked_read") {
		t.Errorf("expected 'marked_read' in JSON output, got: %s", out)
	}
}

// newEmptyNotificationsServer creates a test server that returns an empty notifications list.
func newEmptyNotificationsServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"elements":[],"paging":{"start":0,"count":20,"total":0}}`))
	}))
}
