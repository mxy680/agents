package docs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	out, err := runCmd(t, root, "documents", "create", "--title=My Doc", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result DocumentSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.ID != "new-doc-id" {
		t.Errorf("expected id 'new-doc-id', got %q", result.ID)
	}
	if result.Title != "My Doc" {
		t.Errorf("expected title 'My Doc', got %q", result.Title)
	}
	if !strings.Contains(result.URL, "new-doc-id") {
		t.Errorf("expected URL to contain document ID, got %q", result.URL)
	}
}

func TestDocumentsCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	out, err := runCmd(t, root, "documents", "create", "--title=My Doc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "My Doc") {
		t.Errorf("expected title in output, got: %s", out)
	}
	if !strings.Contains(out, "new-doc-id") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestDocumentsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	out, err := runCmd(t, root, "documents", "create", "--title=Test", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestDocumentsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	out, err := runCmd(t, root, "documents", "get", "--document-id=doc1", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result DocumentDetail
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.ID != "doc1" {
		t.Errorf("expected id 'doc1', got %q", result.ID)
	}
	if result.Title != "Test Document" {
		t.Errorf("expected title 'Test Document', got %q", result.Title)
	}
	if !strings.Contains(result.Body, "Hello world") {
		t.Errorf("expected body to contain 'Hello world', got %q", result.Body)
	}
}

func TestDocumentsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	out, err := runCmd(t, root, "documents", "get", "--document-id=doc1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Test Document") {
		t.Errorf("expected title in output, got: %s", out)
	}
	if !strings.Contains(out, "Hello world") {
		t.Errorf("expected body text in output, got: %s", out)
	}
}

func TestDocumentsAppend_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	out, err := runCmd(t, root, "documents", "append", "--document-id=doc1", "--text=New text", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchUpdateResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.DocumentID != "doc1" {
		t.Errorf("expected documentId 'doc1', got %q", result.DocumentID)
	}
	if result.Replies != 1 {
		t.Errorf("expected 1 reply, got %d", result.Replies)
	}
}

func TestDocumentsAppend_TextFile(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "content.txt")
	if err := os.WriteFile(filePath, []byte("File content"), 0o644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	out, err := runCmd(t, root, "documents", "append", "--document-id=doc1", "--text-file="+filePath, "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchUpdateResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.DocumentID != "doc1" {
		t.Errorf("expected documentId 'doc1', got %q", result.DocumentID)
	}
}

func TestDocumentsAppend_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	out, err := runCmd(t, root, "documents", "append", "--document-id=doc1", "--text=Appended text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "doc1") {
		t.Errorf("expected doc ID in output, got: %s", out)
	}
}

func TestDocumentsAppend_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	out, err := runCmd(t, root, "documents", "append", "--document-id=doc1", "--text=hello", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestDocumentsAppend_MissingText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	_, err := runCmd(t, root, "documents", "append", "--document-id=doc1")
	if err == nil {
		t.Fatal("expected error when neither --text nor --text-file is provided")
	}
}

func TestDocumentsBatchUpdate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	requests := `[{"insertText":{"location":{"index":1},"text":"Hello"}}]`
	out, err := runCmd(t, root, "documents", "batch-update", "--document-id=doc1", "--requests="+requests, "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchUpdateResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.DocumentID != "doc1" {
		t.Errorf("expected documentId 'doc1', got %q", result.DocumentID)
	}
	if result.Replies != 1 {
		t.Errorf("expected 1 reply, got %d", result.Replies)
	}
}

func TestDocumentsBatchUpdate_File(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requests.json")
	requests := `[{"insertText":{"location":{"index":1},"text":"Hello"}},{"insertText":{"location":{"index":1},"text":"World"}}]`
	if err := os.WriteFile(filePath, []byte(requests), 0o644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	out, err := runCmd(t, root, "documents", "batch-update", "--document-id=doc1", "--requests-file="+filePath, "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchUpdateResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.Replies != 2 {
		t.Errorf("expected 2 replies, got %d", result.Replies)
	}
}

func TestDocumentsBatchUpdate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	requests := `[{"insertText":{"location":{"index":1},"text":"Hello"}}]`
	out, err := runCmd(t, root, "documents", "batch-update", "--document-id=doc1", "--requests="+requests)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "doc1") {
		t.Errorf("expected doc ID in output, got: %s", out)
	}
}

func TestDocumentsBatchUpdate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	requests := `[{"insertText":{"location":{"index":1},"text":"Hello"}}]`
	out, err := runCmd(t, root, "documents", "batch-update", "--document-id=doc1", "--requests="+requests, "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestDocumentsBatchUpdate_InvalidJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	_, err := runCmd(t, root, "documents", "batch-update", "--document-id=doc1", "--requests=not-json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDocumentsBatchUpdate_MissingRequests(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestDocumentsCmd(newTestDocsServiceFactory(server)))

	_, err := runCmd(t, root, "documents", "batch-update", "--document-id=doc1")
	if err == nil {
		t.Fatal("expected error when neither --requests nor --requests-file is provided")
	}
}
