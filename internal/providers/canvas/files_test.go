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
