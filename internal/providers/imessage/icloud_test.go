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

func newTestRootCmdICloud(factory ClientFactory) *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)
	return root
}

func captureStdoutICloud(root *cobra.Command, args []string) (string, error) {
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

func TestICloudAccountJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "icloud/account") && !strings.Contains(r.URL.Path, "alias") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"appleId":"alice@icloud.com","fullName":"Alice Smith"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdICloud(factory)
	output, err := captureStdoutICloud(root, []string{"imessage", "icloud", "account", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "alice@icloud.com") {
		t.Errorf("expected Apple ID in output, got: %s", output)
	}
	if !strings.Contains(output, "Alice Smith") {
		t.Errorf("expected full name in output, got: %s", output)
	}
}

func TestICloudChangeAliasDryRun(t *testing.T) {
	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(http.DefaultClient, "http://localhost", "test-pass"), nil
	}

	root := newTestRootCmdICloud(factory)
	output, err := captureStdoutICloud(root, []string{
		"imessage", "icloud", "change-alias",
		"--alias=newalias@icloud.com",
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run in output, got: %s", output)
	}
	if !strings.Contains(output, "newalias@icloud.com") {
		t.Errorf("expected alias in output, got: %s", output)
	}
}

func TestICloudChangeAliasJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "icloud/account/alias") && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"alias":"newalias@icloud.com","changed":true}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdICloud(factory)
	output, err := captureStdoutICloud(root, []string{
		"imessage", "icloud", "change-alias",
		"--alias=newalias@icloud.com",
		"--json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "newalias@icloud.com") {
		t.Errorf("expected alias in output, got: %s", output)
	}
}

func TestICloudContactCardJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "icloud/contact") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"firstName":"Alice","lastName":"Smith","email":"alice@icloud.com","phoneNumber":"+15555550100"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdICloud(factory)
	output, err := captureStdoutICloud(root, []string{"imessage", "icloud", "contact-card", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "alice@icloud.com") {
		t.Errorf("expected email in output, got: %s", output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("expected first name in output, got: %s", output)
	}
}
