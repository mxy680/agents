package imessage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- scheduled list ---

func TestScheduledListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/schedule") {
			t.Errorf("expected path containing message/schedule, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`[{"id":1,"chatGuid":"iMessage;-;+15551234567","message":"Scheduled hello","scheduledFor":1800000000000,"status":"pending"}]`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "scheduled", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Scheduled hello") {
		t.Errorf("expected message text in JSON output, got: %s", out)
	}
	if !containsStr(out, "iMessage;-;+15551234567") {
		t.Errorf("expected chat GUID in JSON output, got: %s", out)
	}
}

// --- scheduled get ---

func TestScheduledGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/schedule/7") {
			t.Errorf("expected path containing message/schedule/7, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"id":7,"chatGuid":"iMessage;-;+15559876543","message":"Get scheduled","scheduledFor":1800000000000,"status":"pending"}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "scheduled", "get", "--id", "7"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "7") {
		t.Errorf("expected ID 7 in output, got: %s", out)
	}
	if !containsStr(out, "Get scheduled") {
		t.Errorf("expected message text in output, got: %s", out)
	}
	if !containsStr(out, "iMessage;-;+15559876543") {
		t.Errorf("expected chat GUID in output, got: %s", out)
	}
}

// --- scheduled create ---

func TestScheduledCreateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/schedule") {
			t.Errorf("expected path containing message/schedule, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"id":10,"chatGuid":"iMessage;-;+15551234567","message":"Future message","scheduledFor":1893456000000,"status":"pending"}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"imessage", "scheduled", "create",
			"--chat-guid", "iMessage;-;+15551234567",
			"--text", "Future message",
			"--send-at", "2030-01-01T12:00:00Z",
			"--json",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Future message") {
		t.Errorf("expected message text in JSON output, got: %s", out)
	}
	if !containsStr(out, "10") {
		t.Errorf("expected ID 10 in JSON output, got: %s", out)
	}
}

func TestScheduledCreateDryRun(t *testing.T) {
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
		root.SetArgs([]string{
			"imessage", "scheduled", "create",
			"--chat-guid", "iMessage;-;+15551234567",
			"--text", "Future message",
			"--send-at", "2030-01-01T12:00:00Z",
			"--dry-run",
		})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !containsStr(out, "dry-run") && !containsStr(out, "dry_run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

// --- scheduled update ---

func TestScheduledUpdateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/schedule/3") {
			t.Errorf("expected path containing message/schedule/3, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"id":3,"chatGuid":"iMessage;-;+15551234567","message":"Updated message","scheduledFor":1893456000000,"status":"pending"}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"imessage", "scheduled", "update",
			"--id", "3",
			"--text", "Updated message",
			"--json",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Updated message") {
		t.Errorf("expected updated message text in JSON output, got: %s", out)
	}
	if !containsStr(out, "3") {
		t.Errorf("expected ID 3 in JSON output, got: %s", out)
	}
}

// --- scheduled delete ---

func TestScheduledDeleteRequiresConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"imessage", "scheduled", "delete", "--id", "5"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is not provided")
	}
	if !containsStr(err.Error(), "confirm") {
		t.Errorf("expected 'confirm' in error message, got: %s", err.Error())
	}
}

func TestScheduledDeleteJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/schedule/5") {
			t.Errorf("expected path containing message/schedule/5, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "scheduled", "delete", "--id", "5", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "5") {
		t.Errorf("expected ID 5 in JSON output, got: %s", out)
	}
	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in JSON output, got: %s", out)
	}
}
