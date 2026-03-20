package linkedin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotificationsList_Deprecated(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"notifications", "list"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for deprecated notifications list endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestNotificationsList_AliasDeprecated(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNotificationsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"notif", "list"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for deprecated notifications list via alias")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message via alias, got: %s", err.Error())
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
