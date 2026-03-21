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

func newTestRootCmdAttach(factory ClientFactory) *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)
	return root
}

func captureStdoutAttach(root *cobra.Command, args []string) (string, error) {
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

func TestAttachmentsGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/attachment/abc-guid-123") && !strings.Contains(r.URL.Path, "blurhash") && !strings.Contains(r.URL.Path, "download") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"guid":"abc-guid-123","transferName":"photo.jpg","mimeType":"image/jpeg","totalBytes":204800,"isOutgoing":false,"createdDate":1700000000000}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdAttach(factory)
	output, err := captureStdoutAttach(root, []string{"imessage", "attachments", "get", "--guid=abc-guid-123", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "abc-guid-123") {
		t.Errorf("expected GUID in output, got: %s", output)
	}
	if !strings.Contains(output, "photo.jpg") {
		t.Errorf("expected filename in output, got: %s", output)
	}
}

func TestAttachmentsCountJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/attachment/count" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"total":15}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdAttach(factory)
	output, err := captureStdoutAttach(root, []string{"imessage", "attachments", "count", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "15") {
		t.Errorf("expected count 15 in output, got: %s", output)
	}
}

func TestAttachmentsBlurhashJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/attachment/abc-guid-123/blurhash") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":200,"message":"OK","data":{"blurhash":"LKO2blurhashvalue"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	root := newTestRootCmdAttach(factory)
	output, err := captureStdoutAttach(root, []string{"imessage", "attachments", "blurhash", "--guid=abc-guid-123", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "LKO2blurhashvalue") {
		t.Errorf("expected blurhash value in output, got: %s", output)
	}
}

func TestAttachmentsDownload(t *testing.T) {
	fileContent := []byte("fake-attachment-data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/attachment/dl-guid/download") {
			w.WriteHeader(http.StatusOK)
			w.Write(fileContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), "attachment-*.bin")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmpFile.Close()
	outPath := tmpFile.Name()

	root := newTestRootCmdAttach(factory)
	root.SetArgs([]string{"imessage", "attachments", "download", "--guid=dl-guid", fmt.Sprintf("--output=%s", outPath)})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(got) != string(fileContent) {
		t.Errorf("expected %q in file, got %q", fileContent, got)
	}
}

func TestAttachmentsUploadDryRun(t *testing.T) {
	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(http.DefaultClient, "http://localhost", "test-pass"), nil
	}

	root := newTestRootCmdAttach(factory)
	output, err := captureStdoutAttach(root, []string{"imessage", "attachments", "upload", "--path=/tmp/test.jpg", "--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run in output, got: %s", output)
	}
	if !strings.Contains(output, "/tmp/test.jpg") {
		t.Errorf("expected file path in output, got: %s", output)
	}
}

// --- attachments download-force ---

func TestAttachmentsDownloadForce(t *testing.T) {
	fileContent := []byte("force-downloaded-data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/attachment/force-guid/download/force") {
			t.Errorf("expected path containing /attachment/force-guid/download/force, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(fileContent)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), "force-*.bin")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmpFile.Close()
	outPath := tmpFile.Name()

	root := newTestRootCmdAttach(factory)
	_, execErr := captureStdoutAttach(root, []string{"imessage", "attachments", "download-force", "--guid=force-guid", fmt.Sprintf("--output=%s", outPath)})
	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(got) != string(fileContent) {
		t.Errorf("expected %q in file, got %q", fileContent, got)
	}
}

// --- attachments live ---

func TestAttachmentsLive(t *testing.T) {
	liveContent := []byte("live-photo-data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/attachment/live-guid/live") {
			t.Errorf("expected path containing /attachment/live-guid/live, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(liveContent)
	}))
	defer server.Close()

	factory := func(ctx context.Context) (*Client, error) {
		return newClientWithBase(server.Client(), server.URL, "test-pass"), nil
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), "live-*.mov")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmpFile.Close()
	outPath := tmpFile.Name()

	root := newTestRootCmdAttach(factory)
	_, execErr := captureStdoutAttach(root, []string{"imessage", "attachments", "live", "--guid=live-guid", fmt.Sprintf("--output=%s", outPath)})
	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(got) != string(liveContent) {
		t.Errorf("expected %q in file, got %q", liveContent, got)
	}
}
