package imessage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestRootCmdHandles(factory ClientFactory) *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)
	return root
}

func captureStdoutHandles(root *cobra.Command, args []string) (string, error) {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	root.SetArgs(args)
	err := root.Execute()
	w.Close()
	os.Stdout = old
	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	return string(buf[:n]), err
}

func TestHandlesListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/handle/query" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":[{"address":"+15555550100","service":"iMessage","country":"US"},{"address":"+15555550200","service":"SMS","country":"US"}]}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdHandles(factory)
	output, err := captureStdoutHandles(root, []string{"imessage", "handles", "list", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "+15555550100") {
		t.Errorf("expected address in output, got: %s", output)
	}
	if !strings.Contains(output, "iMessage") {
		t.Errorf("expected service in output, got: %s", output)
	}
}

func TestHandlesGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/handle/iMessage;-;+15555550100") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"address":"+15555550100","service":"iMessage","country":"US","uncanonicalizedId":"+1 (555) 555-0100"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdHandles(factory)
	output, err := captureStdoutHandles(root, []string{"imessage", "handles", "get", "--guid=iMessage;-;+15555550100", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "+15555550100") {
		t.Errorf("expected address in output, got: %s", output)
	}
}

func TestHandlesCountJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/handle/count" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"total":42}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdHandles(factory)
	output, err := captureStdoutHandles(root, []string{"imessage", "handles", "count", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "42") {
		t.Errorf("expected count 42 in output, got: %s", output)
	}
}

func TestHandlesFocusJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/handle/iMessage;-;+15555550100/focus") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"focusStatus":"Do Not Disturb"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdHandles(factory)
	output, err := captureStdoutHandles(root, []string{"imessage", "handles", "focus", "--guid=iMessage;-;+15555550100", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Do Not Disturb") {
		t.Errorf("expected focusStatus in output, got: %s", output)
	}
}

func TestHandlesAvailabilityJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/handle/availability/imessage") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"available":true,"address":"+15555550100"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdHandles(factory)
	output, err := captureStdoutHandles(root, []string{"imessage", "handles", "availability", "--address=+15555550100", "--service=imessage", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "true") {
		t.Errorf("expected available:true in output, got: %s", output)
	}
}
