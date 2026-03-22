package imessage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// --- messages send ---

func TestMessagesSendJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/text") {
			t.Errorf("expected path containing message/text, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"msg-guid-1","text":"Hello world","isFromMe":true,"dateCreated":1700000000000}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "send", "--to", "+15551234567", "--text", "Hello world", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "msg-guid-1") {
		t.Errorf("expected message GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "Hello world") {
		t.Errorf("expected message text in JSON output, got: %s", out)
	}
}

func TestMessagesSendDryRun(t *testing.T) {
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
		root.SetArgs([]string{"imessage", "messages", "send", "--to", "+15551234567", "--text", "Hello", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !containsStr(out, "dry-run") && !containsStr(out, "dry_run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

// --- messages send-group ---

func TestMessagesSendGroupJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/text") {
			t.Errorf("expected path containing message/text, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"group-msg-guid","text":"Hello group","isFromMe":true,"dateCreated":1700000000000}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "send-group", "--guid", "iMessage;+;chat-group-id", "--text", "Hello group", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "group-msg-guid") {
		t.Errorf("expected message GUID in JSON output, got: %s", out)
	}
}

// --- messages get ---

func TestMessagesGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/msg-guid-get") {
			t.Errorf("expected path containing message/msg-guid-get, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"msg-guid-get","text":"Retrieved message","isFromMe":false,"dateCreated":1700000000000,"handle":{"address":"+15559876543"}}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "get", "--guid", "msg-guid-get"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "msg-guid-get") {
		t.Errorf("expected message GUID in output, got: %s", out)
	}
	if !containsStr(out, "Retrieved message") {
		t.Errorf("expected message text in output, got: %s", out)
	}
}

// --- messages query ---

func TestMessagesQueryJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/query") {
			t.Errorf("expected path containing message/query, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`[{"guid":"q-msg-1","text":"Query result","isFromMe":true,"dateCreated":1700000000000}]`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "query", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "q-msg-1") {
		t.Errorf("expected message GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "Query result") {
		t.Errorf("expected message text in JSON output, got: %s", out)
	}
}

// --- messages edit ---

func TestMessagesEditJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/edit-guid/edit") {
			t.Errorf("expected path containing message/edit-guid/edit, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"edit-guid","text":"Edited text","isFromMe":true,"dateCreated":1700000000000}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "edit", "--guid", "edit-guid", "--text", "Edited text", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "edit-guid") {
		t.Errorf("expected message GUID in JSON output, got: %s", out)
	}
}

// --- messages unsend ---

func TestMessagesUnsendJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/unsend-guid/unsend") {
			t.Errorf("expected path containing message/unsend-guid/unsend, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "unsend", "--guid", "unsend-guid", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "unsend-guid") {
		t.Errorf("expected message GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "true") {
		t.Errorf("expected 'true' for unsent field in JSON output, got: %s", out)
	}
}

// --- messages react ---

func TestMessagesReactJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/react") {
			t.Errorf("expected path containing message/react, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "react",
			"--chat-guid", "iMessage;-;+15551234567",
			"--message-guid", "react-msg-guid",
			"--type", "love",
			"--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "react-msg-guid") {
		t.Errorf("expected message GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "love") {
		t.Errorf("expected reaction type in JSON output, got: %s", out)
	}
}

// --- messages delete ---

func TestMessagesDeleteRequiresConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"imessage", "messages", "delete",
		"--chat-guid", "iMessage;-;+15551234567",
		"--message-guid", "del-msg-guid"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is not provided")
	}
	if !containsStr(err.Error(), "confirm") {
		t.Errorf("expected 'confirm' in error message, got: %s", err.Error())
	}
}

// --- messages count ---

func TestMessagesCountJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/count") {
			t.Errorf("expected path containing message/count, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"total":1337}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "count", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "1337") {
		t.Errorf("expected count 1337 in JSON output, got: %s", out)
	}
	if !containsStr(out, "total") {
		t.Errorf("expected 'total' field in JSON output, got: %s", out)
	}
}

// --- messages count-sent ---

func TestMessagesCountSentJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/count/me") {
			t.Errorf("expected path containing message/count/me, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"total":500}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "count-sent"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "500") {
		t.Errorf("expected count 500 in output, got: %s", out)
	}
}

// --- messages send-attachment ---

func TestMessagesSendAttachmentDryRun(t *testing.T) {
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
		root.SetArgs([]string{"imessage", "messages", "send-attachment", "--to", "+15551234567", "--path", "/tmp/file.jpg", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !containsStr(out, "dry-run") && !containsStr(out, "dry_run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestMessagesSendAttachmentWithText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"att-msg-guid","text":"accompanying text","isFromMe":true,"dateCreated":1700000000000}`))
	}))
	defer server.Close()

	// Create a real temp file so stat check passes
	tmpFile, err := os.CreateTemp(t.TempDir(), "att-*.jpg")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmpFile.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "send-attachment",
			"--to", "+15551234567",
			"--path", tmpFile.Name(),
			"--text", "accompanying text",
			"--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "att-msg-guid") {
		t.Errorf("expected message GUID in JSON output, got: %s", out)
	}
}

// --- messages send-multipart ---

func TestMessagesSendMultipartJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/multipart") {
			t.Errorf("expected path containing message/multipart, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"guid":"multi-msg-guid","text":"part1","isFromMe":true,"dateCreated":1700000000000}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "send-multipart",
			"--to", "+15551234567",
			"--parts", `[{"type":"text","content":"part1"}]`,
			"--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "multi-msg-guid") {
		t.Errorf("expected message GUID in JSON output, got: %s", out)
	}
}

func TestMessagesSendMultipartDryRun(t *testing.T) {
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
		root.SetArgs([]string{"imessage", "messages", "send-multipart",
			"--to", "+15551234567",
			"--parts", `[{"type":"text","content":"hello"}]`,
			"--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !containsStr(out, "dry-run") && !containsStr(out, "dry_run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestMessagesSendMultipartRequiresParts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"imessage", "messages", "send-multipart", "--to", "+15551234567"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when neither --parts nor --parts-file is provided")
	}
}

// --- messages count-updated ---

func TestMessagesCountUpdatedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/count/updated") {
			t.Errorf("expected path containing message/count/updated, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{"total":17}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "count-updated", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "17") {
		t.Errorf("expected count 17 in JSON output, got: %s", out)
	}
	if !containsStr(out, "total") {
		t.Errorf("expected 'total' field in JSON output, got: %s", out)
	}
}

// --- messages embedded-media ---

func TestMessagesEmbeddedMediaJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/embed-guid/embedded-media") {
			t.Errorf("expected path containing message/embed-guid/embedded-media, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`[{"guid":"att-embed-1","transferName":"video.mp4","mimeType":"video/mp4","totalBytes":1024000}]`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "embedded-media", "--guid", "embed-guid"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "video.mp4") {
		t.Errorf("expected filename in output, got: %s", out)
	}
}

func TestMessagesEmbeddedMediaEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`[]`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "embedded-media", "--guid", "no-media-guid"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "No embedded media") {
		t.Errorf("expected 'No embedded media' in output, got: %s", out)
	}
}

// --- messages notify ---

func TestMessagesNotifyJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "message/notify-guid/notify") {
			t.Errorf("expected path containing message/notify-guid/notify, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "notify", "--guid", "notify-guid", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "notify-guid") {
		t.Errorf("expected GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "true") {
		t.Errorf("expected 'true' for notified field, got: %s", out)
	}
}

// --- messages delete (with --confirm) ---

func TestMessagesDeleteConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, bbJSONResponse(`{}`))
	}))
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"imessage", "messages", "delete",
			"--chat-guid", "iMessage;-;+15551234567",
			"--message-guid", "del-msg-guid-2",
			"--confirm",
			"--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "del-msg-guid-2") {
		t.Errorf("expected message GUID in JSON output, got: %s", out)
	}
	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in JSON output, got: %s", out)
	}
}
