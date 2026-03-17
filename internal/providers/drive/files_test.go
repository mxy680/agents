package drive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "list")

	mustContain(t, out, "NAME")
	mustContain(t, out, "Project Plan.docx")
	mustContain(t, out, "Budget.xlsx")
}

func TestFilesListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "list", "--json")

	var files []FileSummary
	if err := json.Unmarshal([]byte(out), &files); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
	if files[0].Name != "Project Plan.docx" {
		t.Errorf("expected first file name=Project Plan.docx, got %s", files[0].Name)
	}
}

func TestFilesGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "get", "--file-id=file1")

	mustContain(t, out, "ID:           file1")
	mustContain(t, out, "Name:         Project Plan.docx")
	mustContain(t, out, "Description:  Q1 project plan")
	mustContain(t, out, "Owner:        alice@example.com")
}

func TestFilesGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "get", "--file-id=file1", "--json")

	var detail FileDetail
	if err := json.Unmarshal([]byte(out), &detail); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if detail.ID != "file1" {
		t.Errorf("expected ID=file1, got %s", detail.ID)
	}
	if detail.Description != "Q1 project plan" {
		t.Errorf("expected Description=Q1 project plan, got %s", detail.Description)
	}
}

func TestFilesDownloadToFile(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	outPath := filepath.Join(t.TempDir(), "downloaded.txt")

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))

	// Download writes to file, not stdout
	root.SetArgs([]string{"files", "download", "--file-id=file1", "--output=" + outPath})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if string(data) != "file-content-here" {
		t.Errorf("expected 'file-content-here', got %q", string(data))
	}
}

func TestFilesDownloadToStdout(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "download", "--file-id=file1")

	mustContain(t, out, "file-content-here")
}

func TestFilesUploadDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	// Create a temp file to "upload"
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(tmpFile, []byte("hello"), 0644)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "upload", "--path="+tmpFile, "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "Would upload")
}

func TestFilesUploadDryRunJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(tmpFile, []byte("hello"), 0644)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "upload", "--path="+tmpFile, "--dry-run", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["action"] != "upload" {
		t.Errorf("expected action=upload, got %v", result["action"])
	}
}

func TestFilesCopyDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "copy", "--file-id=file1", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "Would copy file file1")
}

func TestFilesCopyText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "copy", "--file-id=file1")

	mustContain(t, out, "Copied:")
	mustContain(t, out, "file1-copy")
}

func TestFilesCopyJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "copy", "--file-id=file1", "--json")

	var detail FileDetail
	if err := json.Unmarshal([]byte(out), &detail); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if detail.ID != "file1-copy" {
		t.Errorf("expected ID=file1-copy, got %s", detail.ID)
	}
}

func TestFilesMoveDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "move", "--file-id=file1", "--parent=folder2", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "Would move file file1")
}

func TestFilesMoveText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "move", "--file-id=file1", "--parent=folder2")

	mustContain(t, out, "Moved:")
	mustContain(t, out, "folder2")
}

func TestFilesTrashDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "trash", "--file-id=file1", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "Would trash file file1")
}

func TestFilesTrashText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "trash", "--file-id=file1")

	mustContain(t, out, "Trashed:")
}

func TestFilesTrashJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "trash", "--file-id=file1", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["status"] != "trashed" {
		t.Errorf("expected status=trashed, got %s", result["status"])
	}
}

func TestFilesUntrashDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "untrash", "--file-id=file1", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "Would untrash file file1")
}

func TestFilesUntrashText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "untrash", "--file-id=file1")

	mustContain(t, out, "Restored:")
}

func TestFilesDeleteDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "delete", "--file-id=file1", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "Would permanently delete file file1")
}

func TestFilesDeleteRequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	err := runCmdErr(t, root, "files", "delete", "--file-id=file1")

	if err == nil {
		t.Fatal("expected error without --confirm")
	}
	if err.Error() != "this action is irreversible; re-run with --confirm to proceed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFilesDeleteWithConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "delete", "--file-id=file1", "--confirm")

	mustContain(t, out, "Deleted: file1")
}

func TestFilesDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	out := runCmd(t, root, "files", "delete", "--file-id=file1", "--confirm", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
}

func TestFilesGetRequiresFileID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestFilesCmd(factory))
	err := runCmdErr(t, root, "files", "get")

	if err == nil {
		t.Fatal("expected error without --file-id")
	}
}
