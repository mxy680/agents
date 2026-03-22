package imessage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// --- chats list ---

func TestChatsListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/query") {
			t.Errorf("expected path containing chat/query, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`[{"guid":"chat-guid-1","displayName":"Test Chat","chatIdentifier":"+15551234567"}]`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "chat-guid-1") {
		t.Errorf("expected chat GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "Test Chat") {
		t.Errorf("expected display name in JSON output, got: %s", out)
	}
}

func TestChatsListText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`[{"guid":"chat-guid-1","displayName":"My Group","chatIdentifier":"group-id"}]`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "My Group") {
		t.Errorf("expected display name in text output, got: %s", out)
	}
	if !containsStr(out, "chat-guid-1") {
		t.Errorf("expected GUID in text output, got: %s", out)
	}
}

// --- chats get ---

func TestChatsGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/test-chat-guid") {
			t.Errorf("expected path containing chat/test-chat-guid, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"test-chat-guid","displayName":"Direct Chat","chatIdentifier":"+15559876543"}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "get", "--guid", "test-chat-guid", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "test-chat-guid") {
		t.Errorf("expected chat GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "Direct Chat") {
		t.Errorf("expected display name in JSON output, got: %s", out)
	}
}

// --- chats create ---

func TestChatsCreateDryRun(t *testing.T) {
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
		root.SetArgs([]string{"imessage", "chats", "create", "--participants", "+15551234567", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !containsStr(out, "dry-run") && !containsStr(out, "dry_run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestChatsCreateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/new") {
			t.Errorf("expected path containing chat/new, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"new-chat-guid","displayName":"","chatIdentifier":"+15551234567"}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "create", "--participants", "+15551234567", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "new-chat-guid") {
		t.Errorf("expected new chat GUID in JSON output, got: %s", out)
	}
}

// --- chats delete ---

func TestChatsDeleteRequiresConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"imessage", "chats", "delete", "--guid", "some-guid"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is not provided")
	}
	if !containsStr(err.Error(), "confirm") {
		t.Errorf("expected 'confirm' in error message, got: %s", err.Error())
	}
}

func TestChatsDeleteJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/del-guid") {
			t.Errorf("expected path containing chat/del-guid, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "delete", "--guid", "del-guid", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "del-guid") {
		t.Errorf("expected GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in JSON output, got: %s", out)
	}
}

// --- chats read ---

func TestChatsReadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/read-guid/read") {
			t.Errorf("expected path containing chat/read-guid/read, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "read", "--guid", "read-guid", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "read-guid") {
		t.Errorf("expected GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "true") {
		t.Errorf("expected 'true' for read field in JSON output, got: %s", out)
	}
}

// --- chats count ---

func TestChatsCountJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/count") {
			t.Errorf("expected path containing chat/count, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"total":42}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "count", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "42") {
		t.Errorf("expected count 42 in JSON output, got: %s", out)
	}
	if !containsStr(out, "total") {
		t.Errorf("expected 'total' field in JSON output, got: %s", out)
	}
}

// --- chats typing (start) ---

func TestChatsTypingJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/typing-guid/typing") {
			t.Errorf("expected path containing chat/typing-guid/typing, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "typing", "--guid", "typing-guid", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "typing-guid") {
		t.Errorf("expected GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "typing") {
		t.Errorf("expected 'typing' field in JSON output, got: %s", out)
	}
}

// --- chats typing (stop) ---

func TestChatsTypingStopJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE for stop typing, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/stop-guid/typing") {
			t.Errorf("expected path containing chat/stop-guid/typing, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "typing", "--guid", "stop-guid", "--stop", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "stop-guid") {
		t.Errorf("expected GUID in JSON output, got: %s", out)
	}
}

// --- chats update ---

func TestChatsUpdateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/update-guid") {
			t.Errorf("expected path containing chat/update-guid, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"update-guid","displayName":"New Name","chatIdentifier":"update-guid"}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "update", "--guid", "update-guid", "--name", "New Name", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "update-guid") {
		t.Errorf("expected GUID in JSON output, got: %s", out)
	}
}

func TestChatsUpdateDryRun(t *testing.T) {
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
		root.SetArgs([]string{"imessage", "chats", "update", "--guid", "some-guid", "--name", "New Name", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !containsStr(out, "dry-run") && !containsStr(out, "dry_run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

// --- chats messages ---

func TestChatsMessagesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/msg-chat-guid/message") {
			t.Errorf("expected path containing chat/msg-chat-guid/message, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`[{"guid":"msg-in-chat","text":"Hello","isFromMe":false,"dateCreated":1700000000000}]`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "messages", "--guid", "msg-chat-guid", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "msg-in-chat") {
		t.Errorf("expected message GUID in JSON output, got: %s", out)
	}
}

func TestChatsMessagesWithParams(t *testing.T) {
	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`[{"guid":"msg-param","text":"Param test","isFromMe":true,"dateCreated":1700000000000}]`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "messages", "--guid", "param-chat", "--limit", "10", "--offset", "5"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(capturedURL, "limit=10") {
		t.Errorf("expected limit=10 in query string, got: %s", capturedURL)
	}
	if !containsStr(capturedURL, "offset=5") {
		t.Errorf("expected offset=5 in query string, got: %s", capturedURL)
	}
}

// --- chats unread ---

func TestChatsUnreadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/unread-guid/unread") {
			t.Errorf("expected path containing chat/unread-guid/unread, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "unread", "--guid", "unread-guid", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "unread-guid") {
		t.Errorf("expected GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "true") {
		t.Errorf("expected 'true' for unread field in JSON output, got: %s", out)
	}
}

func TestChatsUnreadDryRun(t *testing.T) {
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

	captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "unread", "--guid", "some-guid", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
}

// --- chats leave ---

func TestChatsLeaveJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/leave-guid/leave") {
			t.Errorf("expected path containing chat/leave-guid/leave, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "leave", "--guid", "leave-guid", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "leave-guid") {
		t.Errorf("expected GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "true") {
		t.Errorf("expected 'true' for left field in JSON output, got: %s", out)
	}
}

func TestChatsLeaveDryRun(t *testing.T) {
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

	captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "leave", "--guid", "some-guid", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
}

// --- chats icon get ---

func TestChatsIconGetNoOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "icon", "get", "--guid", "icon-chat-guid", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "icon-chat-guid") {
		t.Errorf("expected GUID in JSON output, got: %s", out)
	}
}

func TestChatsIconGetWithOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "chat/icon-dl-guid/icon") {
			t.Errorf("expected path containing chat/icon-dl-guid/icon, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte("fake-icon-data"))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	tmpFile, err := os.CreateTemp(t.TempDir(), "icon-*.png")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmpFile.Close()
	outPath := tmpFile.Name()

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "icon", "get", "--guid", "icon-dl-guid", "--output", outPath})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, outPath) {
		t.Errorf("expected output path in output, got: %s", out)
	}
}

// --- chats icon set ---

func TestChatsIconSetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/icon-set-guid/icon") {
			t.Errorf("expected path containing chat/icon-set-guid/icon, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"updated": true}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "icon", "set", "--guid", "icon-set-guid", "--path", "/tmp/icon.png", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "true") {
		t.Errorf("expected 'true' in JSON output, got: %s", out)
	}
}

func TestChatsIconSetDryRun(t *testing.T) {
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
		root.SetArgs([]string{"imessage", "chats", "icon", "set", "--guid", "some-guid", "--path", "/tmp/icon.png", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !containsStr(out, "dry-run") && !containsStr(out, "dry_run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

// --- chats icon remove ---

func TestChatsIconRemoveJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "chat/icon-rm-guid/icon") {
			t.Errorf("expected path containing chat/icon-rm-guid/icon, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "chats", "icon", "remove", "--guid", "icon-rm-guid", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "icon-rm-guid") {
		t.Errorf("expected GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "true") {
		t.Errorf("expected 'true' for icon_removed field, got: %s", out)
	}
}

func TestChatsIconRemoveRequiresConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"imessage", "chats", "icon", "remove", "--guid", "some-guid"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is not provided")
	}
	if !containsStr(err.Error(), "confirm") {
		t.Errorf("expected 'confirm' in error message, got: %s", err.Error())
	}
}

func TestChatsIconRemoveDryRun(t *testing.T) {
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
		root.SetArgs([]string{"imessage", "chats", "icon", "remove", "--guid", "some-guid", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !containsStr(out, "dry-run") && !containsStr(out, "dry_run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}
