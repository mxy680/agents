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

func newTestRootCmdContacts(factory ClientFactory) *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)
	return root
}

func captureStdoutContacts(root *cobra.Command, args []string) (string, error) {
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

func TestContactsListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/contact" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":[{"id":"1","firstName":"Alice","lastName":"Smith","displayName":"Alice Smith","phoneNumbers":[{"address":"+15555550100"}],"emails":[{"address":"alice@example.com"}]},{"id":"2","firstName":"Bob","lastName":"Jones","displayName":"Bob Jones","phoneNumbers":[{"address":"+15555550200"}]}]}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdContacts(factory)
	output, err := captureStdoutContacts(root, []string{"imessage", "contacts", "list", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Alice Smith") {
		t.Errorf("expected contact name in output, got: %s", output)
	}
	if !strings.Contains(output, "Bob Jones") {
		t.Errorf("expected second contact name in output, got: %s", output)
	}
}

func TestContactsGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/contact/query" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":[{"id":"1","firstName":"Alice","lastName":"Smith","displayName":"Alice Smith","phoneNumbers":[{"address":"+15555550100"}],"emails":[{"address":"alice@example.com"}]}]}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdContacts(factory)
	output, err := captureStdoutContacts(root, []string{"imessage", "contacts", "get", "--query=Alice", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Alice Smith") {
		t.Errorf("expected contact name in output, got: %s", output)
	}
	if !strings.Contains(output, "alice@example.com") {
		t.Errorf("expected email in output, got: %s", output)
	}
}

func TestContactsCreateDryRun(t *testing.T) {
	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(http.DefaultClient, "http://localhost", "test-pass"), nil
	}

	root := newTestRootCmdContacts(factory)
	output, err := captureStdoutContacts(root, []string{
		"imessage", "contacts", "create",
		`--data={"firstName":"Test","lastName":"User"}`,
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run in output, got: %s", output)
	}
}

func TestContactsCreateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/contact" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"id":"99","firstName":"New","lastName":"Contact","displayName":"New Contact"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdContacts(factory)
	output, err := captureStdoutContacts(root, []string{
		"imessage", "contacts", "create",
		`--data={"firstName":"New","lastName":"Contact"}`,
		"--json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "New Contact") {
		t.Errorf("expected contact name in output, got: %s", output)
	}
}
