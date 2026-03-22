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

func newTestRootCmdWebhooks(factory ClientFactory) *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)
	return root
}

func captureStdoutWebhooks(root *cobra.Command, args []string) (string, error) {
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

func TestWebhooksListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/webhook" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":[{"id":1,"url":"https://example.com/webhook1","events":["new-message"]},{"id":2,"url":"https://example.com/webhook2","events":["chat-read-status-changed"]}]}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdWebhooks(factory)
	output, err := captureStdoutWebhooks(root, []string{"imessage", "webhooks", "list", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "webhook1") {
		t.Errorf("expected webhook URL in output, got: %s", output)
	}
	if !strings.Contains(output, "new-message") {
		t.Errorf("expected event type in output, got: %s", output)
	}
}

func TestWebhooksCreateDryRun(t *testing.T) {
	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(http.DefaultClient, "http://localhost", "test-pass"), nil
	}

	root := newTestRootCmdWebhooks(factory)
	output, err := captureStdoutWebhooks(root, []string{
		"imessage", "webhooks", "create",
		"--url=https://example.com/hook",
		"--events=new-message,chat-read-status-changed",
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run in output, got: %s", output)
	}
	if !strings.Contains(output, "https://example.com/hook") {
		t.Errorf("expected webhook URL in output, got: %s", output)
	}
}

func TestWebhooksCreateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/webhook" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"id":3,"url":"https://example.com/hook","events":["new-message"]}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdWebhooks(factory)
	output, err := captureStdoutWebhooks(root, []string{
		"imessage", "webhooks", "create",
		"--url=https://example.com/hook",
		"--json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "https://example.com/hook") {
		t.Errorf("expected webhook URL in output, got: %s", output)
	}
	if !strings.Contains(output, "3") {
		t.Errorf("expected webhook ID in output, got: %s", output)
	}
}

func TestWebhooksDeleteRequiresConfirm(t *testing.T) {
	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(http.DefaultClient, "http://localhost", "test-pass"), nil
	}

	root := newTestRootCmdWebhooks(factory)
	// No --confirm flag; expect an error about requiring confirmation.
	root.SetArgs([]string{"imessage", "webhooks", "delete", "--id=1"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(err.Error(), "confirm") {
		t.Errorf("expected confirm error, got: %v", err)
	}
}

func TestWebhooksDeleteJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "webhook/5") && r.Method == http.MethodDelete {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"deleted":true}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdWebhooks(factory)
	output, err := captureStdoutWebhooks(root, []string{
		"imessage", "webhooks", "delete",
		"--id=5",
		"--confirm",
		"--json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected deleted confirmation in output, got: %s", output)
	}
}
