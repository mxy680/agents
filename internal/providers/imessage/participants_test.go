package imessage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- participants add ---

func TestParticipantsAddJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/group-guid/participant") {
			t.Errorf("expected path containing chat/group-guid/participant, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"group-guid","displayName":"My Group","chatIdentifier":"group-id","participants":[{"handle":{"address":"+15551234567"}},{"handle":{"address":"+15559999999"}}]}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "participants", "add", "--guid", "group-guid", "--address", "+15559999999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "group-guid") {
		t.Errorf("expected chat GUID in JSON output, got: %s", out)
	}
}

func TestParticipantsAddDryRun(t *testing.T) {
	apiCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "participants", "add", "--guid", "group-guid", "--address", "+15559999999", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !containsStr(out, "dry-run") && !containsStr(out, "dry_run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
	if !containsStr(out, "+15559999999") {
		t.Errorf("expected address in dry-run output, got: %s", out)
	}
}

// --- participants remove ---

func TestParticipantsRemoveRequiresConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"imessage", "participants", "remove", "--guid", "group-guid", "--address", "+15551234567"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is not provided")
	}
	if !containsStr(err.Error(), "confirm") {
		t.Errorf("expected 'confirm' in error message, got: %s", err.Error())
	}
}

func TestParticipantsRemoveJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/group-guid/participant/remove") {
			t.Errorf("expected path containing chat/group-guid/participant/remove, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"group-guid","displayName":"My Group","chatIdentifier":"group-id","participants":[{"handle":{"address":"+15551234567"}}]}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "participants", "remove", "--guid", "group-guid", "--address", "+15559999999", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "group-guid") {
		t.Errorf("expected chat GUID in JSON output, got: %s", out)
	}
}
