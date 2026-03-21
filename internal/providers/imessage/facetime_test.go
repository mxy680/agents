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

func newTestRootCmdFaceTime(factory ClientFactory) *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)
	return root
}

func captureStdoutFaceTime(root *cobra.Command, args []string) (string, error) {
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

func TestFaceTimeCallDryRun(t *testing.T) {
	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(http.DefaultClient, "http://localhost", "test-pass"), nil
	}

	root := newTestRootCmdFaceTime(factory)
	output, err := captureStdoutFaceTime(root, []string{
		"imessage", "facetime", "call",
		"--addresses=alice@example.com,bob@example.com",
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run in output, got: %s", output)
	}
	if !strings.Contains(output, "alice@example.com") {
		t.Errorf("expected address in output, got: %s", output)
	}
}

func TestFaceTimeCallJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "facetime/session") && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"sessionId":"session-xyz","status":"initiated"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdFaceTime(factory)
	output, err := captureStdoutFaceTime(root, []string{
		"imessage", "facetime", "call",
		"--addresses=alice@example.com",
		"--json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// JSON output contains the raw API response data; text line includes the address.
	if !strings.Contains(output, "session-xyz") {
		t.Errorf("expected sessionId in JSON output, got: %s", output)
	}
}

func TestFaceTimeAnswerJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "facetime/answer/call-uuid-001") && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"callUuid":"call-uuid-001","status":"answered"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdFaceTime(factory)
	output, err := captureStdoutFaceTime(root, []string{
		"imessage", "facetime", "answer",
		"--call-uuid=call-uuid-001",
		"--json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "call-uuid-001") {
		t.Errorf("expected call UUID in output, got: %s", output)
	}
}

func TestFaceTimeLeaveJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "facetime/leave/call-uuid-002") && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"callUuid":"call-uuid-002","status":"left"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdFaceTime(factory)
	output, err := captureStdoutFaceTime(root, []string{
		"imessage", "facetime", "leave",
		"--call-uuid=call-uuid-002",
		"--json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "call-uuid-002") {
		t.Errorf("expected call UUID in output, got: %s", output)
	}
}
