package linkedin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConnectionsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"connections", "list"})
		root.Execute() //nolint:errcheck
	})

	// The normalized connections response only includes URNs — profile names
	// require a separate API call which connections list does not make.
	if !containsStr(out, "urn:li:fsd_profile:ACoAAConn1") {
		t.Errorf("expected connection URN in output, got: %s", out)
	}
	if !containsStr(out, "urn:li:fsd_profile:ACoAAConn2") {
		t.Errorf("expected second connection URN in output, got: %s", out)
	}
}

func TestConnectionsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"connections", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"urn"`) {
		t.Errorf("expected JSON field 'urn' in output, got: %s", out)
	}
	if !containsStr(out, "ACoAAConn1") {
		t.Errorf("expected connection URN in JSON output, got: %s", out)
	}
}

func TestConnectionsList_WithAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"conn", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "ACoAAConn1") {
		t.Errorf("expected connection URN in output via alias, got: %s", out)
	}
}

func TestConnectionsList_EmptyResult(t *testing.T) {
	emptyServer := newEmptyConnectionsServer(t)
	defer emptyServer.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(emptyServer)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"connections", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "No connections found.") {
		t.Errorf("expected 'No connections found.' in output, got: %s", out)
	}
}

func TestConnectionsGet_MissingURN(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"connections", "get"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --urn is missing")
	}
}

func TestConnectionsGet_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"connections", "get", "--urn", "testuser"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Test") {
		t.Errorf("expected 'Test' in profile output, got: %s", out)
	}
}

func TestConnectionsRemove_MissingURN(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"connections", "remove"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --urn is missing")
	}
}

func TestConnectionsRemove_RequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"connections", "remove", "--urn", "urn:li:fs_miniProfile:ACoAAConn1"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is not provided")
	}
	if !containsStr(err.Error(), "irreversible") {
		t.Errorf("expected 'irreversible' in error message, got: %s", err.Error())
	}
}

func TestConnectionsRemove_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"connections", "remove", "--urn", "urn:li:fs_miniProfile:ACoAAConn1", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestConnectionsRemove_WithConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"connections", "remove", "--urn", "urn:li:fs_miniProfile:ACoAAConn1", "--confirm"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "removed") {
		t.Errorf("expected 'removed' in output, got: %s", out)
	}
}

func TestConnectionsRemove_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newConnectionsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"connections", "remove", "--urn", "urn:li:fs_miniProfile:ACoAAConn1", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON 'status' field in output, got: %s", out)
	}
	if !containsStr(out, "removed") {
		t.Errorf("expected 'removed' in JSON output, got: %s", out)
	}
}

// newEmptyConnectionsServer creates a test server that returns an empty normalized connections response.
func newEmptyConnectionsServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"paging":{"count":0,"start":0},"$type":"com.linkedin.restli.common.CollectionResponse"},"included":[]}`))
	}))
}
