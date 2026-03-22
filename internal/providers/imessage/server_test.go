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

func newTestRootCmdServer(factory ClientFactory) *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)
	return root
}

func captureStdoutServer(root *cobra.Command, args []string) (string, error) {
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

func TestServerInfoJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "server/info") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"os_version":"macOS 14.0","server_version":"1.9.0","detected_icloud":"MacBookPro","private_api":true,"proxy_service":"Cloudflare"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "info", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "1.9.0") {
		t.Errorf("expected server version in output, got: %s", output)
	}
	if !strings.Contains(output, "macOS 14.0") {
		t.Errorf("expected OS version in output, got: %s", output)
	}
}

func TestServerLogsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "server/logs") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":"2024-01-01 12:00:00 [INFO] Server started\n2024-01-01 12:00:01 [INFO] Ready"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "logs", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Server started") {
		t.Errorf("expected log content in output, got: %s", output)
	}
}

func TestServerRestartDryRun(t *testing.T) {
	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(http.DefaultClient, "http://localhost", "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "restart", "--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run in output, got: %s", output)
	}
	if !strings.Contains(output, "soft") {
		t.Errorf("expected restart type in output, got: %s", output)
	}
}

func TestServerStatsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "server/statistics/totals") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"messages":1234,"chats":56,"attachments":789}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "stats", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "1234") {
		t.Errorf("expected messages count in output, got: %s", output)
	}
}

func TestServerAlertsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "server/alert") && !strings.Contains(r.URL.Path, "read") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":[{"type":"info","name":"iCloud Status","value":"Connected"},{"type":"warning","name":"Private API","value":"Disabled"}]}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "alerts", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "iCloud Status") {
		t.Errorf("expected alert name in output, got: %s", output)
	}
	if !strings.Contains(output, "Private API") {
		t.Errorf("expected second alert in output, got: %s", output)
	}
}

// --- server update-check ---

func TestServerUpdateCheckJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "server/update/check") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"available":false,"version":"1.9.0"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "update-check", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "1.9.0") {
		t.Errorf("expected version in output, got: %s", output)
	}
}

func TestServerUpdateCheckAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":200,"message":"OK","data":{"available":true,"version":"2.0.0"}}`)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "update-check"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "2.0.0") {
		t.Errorf("expected new version in output, got: %s", output)
	}
}

// --- server update-install ---

func TestServerUpdateInstallJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "server/update/install") && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"success":true}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "update-install", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "install") && !strings.Contains(output, "success") {
		t.Errorf("expected install-related content in output, got: %s", output)
	}
}

func TestServerUpdateInstallDryRun(t *testing.T) {
	apiCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":200,"message":"OK","data":{}}`)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "update-install", "--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run in output, got: %s", output)
	}
}

// --- server alerts-read ---

func TestServerAlertsReadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "server/alert/read") && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"success":true}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "alerts-read", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "read") && !strings.Contains(output, "success") {
		t.Errorf("expected read/success content in output, got: %s", output)
	}
}

func TestServerAlertsReadDryRun(t *testing.T) {
	apiCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":200,"message":"OK","data":{}}`)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "alerts-read", "--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiCalled {
		t.Error("expected no API call during dry-run")
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run in output, got: %s", output)
	}
}

// --- server restart (hard) ---

func TestServerRestartHard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "server/restart/hard") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"success":true}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdServer(factory)
	output, err := captureStdoutServer(root, []string{"imessage", "server", "restart", "--soft=false"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "hard") {
		t.Errorf("expected 'hard' in output, got: %s", output)
	}
}
