package canvas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFilesListText(t *testing.T) {
	mux := http.NewServeMux()
	withFilesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "homework.pdf") {
		t.Errorf("expected file name in output, got: %s", output)
	}
	if !strings.Contains(output, "syllabus.docx") {
		t.Errorf("expected second file name in output, got: %s", output)
	}
	// Locked file should have "L" marker.
	if !strings.Contains(output, "L") {
		t.Errorf("expected locked marker 'L' for syllabus.docx, got: %s", output)
	}
}

func TestFilesListJSON(t *testing.T) {
	mux := http.NewServeMux()
	withFilesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"display_name"`) {
		t.Errorf("expected display_name field in JSON, got: %s", output)
	}
	if !strings.Contains(output, "homework.pdf") {
		t.Errorf("expected file name in JSON output, got: %s", output)
	}
}

func TestFilesListMissingCourseID(t *testing.T) {
	mux := http.NewServeMux()
	withFilesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"files", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestFilesGetText(t *testing.T) {
	mux := http.NewServeMux()
	withFilesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "get", "--file-id", "601"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "homework.pdf") {
		t.Errorf("expected file name in output, got: %s", output)
	}
	if !strings.Contains(output, "601") {
		t.Errorf("expected file ID in output, got: %s", output)
	}
	if !strings.Contains(output, "application/pdf") {
		t.Errorf("expected content type in output, got: %s", output)
	}
}

func TestFilesDeleteNoConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withFilesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"files", "delete", "--file-id", "601"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is absent")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestFilesDeleteConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withFilesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "delete", "--file-id", "601", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "601") {
		t.Errorf("expected file ID in deletion output, got: %s", output)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}

func TestFilesFoldersText(t *testing.T) {
	mux := http.NewServeMux()
	withFilesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "folders", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "course files") {
		t.Errorf("expected folder name in output, got: %s", output)
	}
	if !strings.Contains(output, "assignments") {
		t.Errorf("expected second folder name in output, got: %s", output)
	}
}

func TestFilesFolderContentsText(t *testing.T) {
	mux := http.NewServeMux()
	withFilesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "folder-contents", "--folder-id", "701"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "homework.pdf") {
		t.Errorf("expected file name in folder contents output, got: %s", output)
	}
}

func TestFilesCreateFolderDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withFilesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"files", "create-folder",
			"--course-id", "101",
			"--name", "My Folder",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
	if !strings.Contains(output, "My Folder") {
		t.Errorf("expected folder name in dry-run output, got: %s", output)
	}
}

func TestFilesCreateFolderSuccess(t *testing.T) {
	mux := http.NewServeMux()
	withFilesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"files", "create-folder",
			"--course-id", "101",
			"--name", "New Folder",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "New Folder") {
		t.Errorf("expected folder name in output, got: %s", output)
	}
}
func TestFilesUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "update", "--file-id", "601", "--name", "homework-renamed.pdf"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "601") || !strings.Contains(output, "updated") {
		t.Errorf("expected file update output, got: %s", output)
	}
}

func TestFilesCreateFolderLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "create-folder", "--course-id", "101", "--name", "New Folder"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "New Folder") && !strings.Contains(output, "702") {
		t.Errorf("expected folder creation output, got: %s", output)
	}
}

func TestFilesDownloadMissingFileID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"files", "download", "--output", "/tmp/out.pdf"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --file-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--file-id") {
		t.Errorf("error should mention --file-id, got: %v", execErr)
	}
}

func TestFilesDownloadMissingOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"files", "download", "--file-id", "601"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --output is missing")
	}
	if !strings.Contains(execErr.Error(), "--output") {
		t.Errorf("error should mention --output, got: %v", execErr)
	}
}

func TestFilesFolderContentsJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "folder-contents", "--folder-id", "701", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "homework.pdf") {
		t.Errorf("expected file name in JSON output, got: %s", output)
	}
}

func TestFilesFoldersJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "folders", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "assignments") {
		t.Errorf("expected folder name in JSON output, got: %s", output)
	}
}

func TestFilesUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"files", "update", "--file-id", "601", "--name", "Updated File", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "601") {
		t.Errorf("expected file ID in JSON output, got: %s", output)
	}
}

func TestFilesUpdateWithAllFlags(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"files", "update",
			"--file-id", "601",
			"--name", "Updated File",
			"--parent-folder-id", "701",
			"--locked",
			"--hidden",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "601") {
		t.Errorf("expected file ID in output, got: %s", output)
	}
}

func TestFilesCreateFolderJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"files", "create-folder",
			"--course-id", "101",
			"--name", "New Folder",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "702") {
		t.Errorf("expected folder ID in JSON output, got: %s", output)
	}
}

func TestFilesCreateFolderWithOptionalFlags(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFilesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"files", "create-folder",
			"--course-id", "101",
			"--name", "New Folder",
			"--parent-folder", "course files",
			"--locked",
			"--hidden",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "702") {
		t.Errorf("expected folder ID in output, got: %s", output)
	}
}
