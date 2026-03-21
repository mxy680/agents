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

func newTestRootCmdFindMy(factory ClientFactory) *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)
	return root
}

func captureStdoutFindMy(root *cobra.Command, args []string) (string, error) {
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

func TestFindMyDevicesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "icloud/findmy/devices") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":[{"id":"device-001","name":"Alice's iPhone","batteryLevel":0.75},{"id":"device-002","name":"Alice's MacBook","batteryLevel":0.50}]}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdFindMy(factory)
	output, err := captureStdoutFindMy(root, []string{"imessage", "findmy", "devices", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "device-001") {
		t.Errorf("expected device id in output, got: %s", output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("expected device name in output, got: %s", output)
	}
}

func TestFindMyDevicesRefreshJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "icloud/findmy/devices/refresh") && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"refreshed":true}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdFindMy(factory)
	output, err := captureStdoutFindMy(root, []string{"imessage", "findmy", "devices-refresh", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "refreshed") {
		t.Errorf("expected refreshed confirmation in output, got: %s", output)
	}
}

func TestFindMyFriendsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "icloud/findmy/friends") && !strings.Contains(r.URL.Path, "refresh") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":[{"id":"friend-001","handle":"bob@example.com","firstName":"Bob"},{"id":"friend-002","handle":"carol@example.com","firstName":"Carol"}]}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdFindMy(factory)
	output, err := captureStdoutFindMy(root, []string{"imessage", "findmy", "friends", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "friend-001") {
		t.Errorf("expected friend id in output, got: %s", output)
	}
	if !strings.Contains(output, "Bob") {
		t.Errorf("expected friend name in output, got: %s", output)
	}
}

func TestFindMyFriendsRefreshJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "icloud/findmy/friends/refresh") && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"refreshed":true}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdFindMy(factory)
	output, err := captureStdoutFindMy(root, []string{"imessage", "findmy", "friends-refresh", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "refreshed") {
		t.Errorf("expected refreshed confirmation in output, got: %s", output)
	}
}
