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

func newTestRootCmdMac(factory ClientFactory) *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)
	return root
}

func captureStdoutMac(root *cobra.Command, args []string) (string, error) {
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

func TestMacLockDryRun(t *testing.T) {
	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(http.DefaultClient, "http://localhost", "test-pass"), nil
	}

	root := newTestRootCmdMac(factory)
	output, err := captureStdoutMac(root, []string{"imessage", "mac", "lock", "--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run in output, got: %s", output)
	}
	if !strings.Contains(output, "lock") {
		t.Errorf("expected lock action in output, got: %s", output)
	}
}

func TestMacLockJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "mac/lock") && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"locked":true}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdMac(factory)
	output, err := captureStdoutMac(root, []string{"imessage", "mac", "lock", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "locked") {
		t.Errorf("expected lock confirmation in output, got: %s", output)
	}
}

func TestMacRestartMessagesDryRun(t *testing.T) {
	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(http.DefaultClient, "http://localhost", "test-pass"), nil
	}

	root := newTestRootCmdMac(factory)
	output, err := captureStdoutMac(root, []string{"imessage", "mac", "restart-messages", "--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run in output, got: %s", output)
	}
	if !strings.Contains(output, "Messages") {
		t.Errorf("expected Messages app reference in output, got: %s", output)
	}
}

func TestMacRestartMessagesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "mac/imessage/restart") && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"restarted":true}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdMac(factory)
	output, err := captureStdoutMac(root, []string{"imessage", "mac", "restart-messages", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "restarted") {
		t.Errorf("expected restart confirmation in output, got: %s", output)
	}
}
