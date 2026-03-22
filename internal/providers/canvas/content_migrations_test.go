package canvas

import (
	"strings"
	"testing"
)

func TestContentMigrationsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentMigrationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"content-migrations", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "course_copy_importer") {
		t.Errorf("expected migration type in output, got: %s", output)
	}
	if !strings.Contains(output, "completed") {
		t.Errorf("expected workflow state in output, got: %s", output)
	}
}

func TestContentMigrationsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentMigrationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"content-migrations", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "migration_type") {
		t.Errorf("JSON output should contain migration_type field, got: %s", output)
	}
	if !strings.Contains(output, "course_copy_importer") {
		t.Errorf("JSON output should contain migration type value, got: %s", output)
	}
}

func TestContentMigrationsListMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentMigrationsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"content-migrations", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestContentMigrationsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentMigrationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"content-migrations", "get", "--course-id", "101", "--migration-id", "14001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "course_copy_importer") {
		t.Errorf("expected migration type, got: %s", output)
	}
	if !strings.Contains(output, "completed") {
		t.Errorf("expected workflow state, got: %s", output)
	}
	if !strings.Contains(output, "progress") {
		t.Errorf("expected progress URL in output, got: %s", output)
	}
}

func TestContentMigrationsCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentMigrationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"content-migrations", "create", "--course-id", "101", "--type", "course_copy_importer"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "14001") && !strings.Contains(output, "created") && !strings.Contains(output, "running") {
		t.Errorf("expected creation output, got: %s", output)
	}
}

func TestContentMigrationsProgressText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentMigrationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"content-migrations", "progress", "--course-id", "101", "--migration-id", "14001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "14001") || !strings.Contains(output, "completed") {
		t.Errorf("expected migration ID and state in output, got: %s", output)
	}
}

func TestContentMigrationsContentListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentMigrationsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"content-migrations", "content-list", "--course-id", "101", "--migration-id", "14001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "assignment") && !strings.Contains(output, "Homework 1") {
		t.Errorf("expected content list items in output, got: %s", output)
	}
}
