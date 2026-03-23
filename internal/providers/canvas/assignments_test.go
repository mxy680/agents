package canvas

import (
	"strings"
	"testing"
)

func TestAssignmentsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignments", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework 1") {
		t.Errorf("output should contain assignment name, got: %s", output)
	}
	if !strings.Contains(output, "Homework 2") {
		t.Errorf("output should contain second assignment name, got: %s", output)
	}
	if !strings.Contains(output, "100") {
		t.Errorf("output should contain points, got: %s", output)
	}
}

func TestAssignmentsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignments", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"name"`) {
		t.Errorf("JSON output should contain name field, got: %s", output)
	}
	if !strings.Contains(output, "Homework 1") {
		t.Errorf("JSON output should contain assignment name, got: %s", output)
	}
	if !strings.Contains(output, `"points_possible"`) {
		t.Errorf("JSON output should contain points_possible field, got: %s", output)
	}
}

func TestAssignmentsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignments", "get", "--course-id", "101", "--assignment-id", "501"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework 1") {
		t.Errorf("output should contain assignment name, got: %s", output)
	}
	if !strings.Contains(output, "100") {
		t.Errorf("output should contain points, got: %s", output)
	}
	if !strings.Contains(output, "points") {
		t.Errorf("output should contain grading type, got: %s", output)
	}
	if !strings.Contains(output, "Complete exercises 1-10") {
		t.Errorf("output should contain description, got: %s", output)
	}
}

func TestAssignmentsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignments", "get", "--course-id", "101", "--assignment-id", "501", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"name"`) {
		t.Errorf("JSON output should contain name field, got: %s", output)
	}
	if !strings.Contains(output, "Homework 1") {
		t.Errorf("JSON output should contain assignment name, got: %s", output)
	}
	if !strings.Contains(output, `"points_possible"`) {
		t.Errorf("JSON output should contain points_possible field, got: %s", output)
	}
}

func TestAssignmentsGetMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"assignments", "get", "--assignment-id", "501"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestAssignmentsDeleteDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignments", "delete", "--course-id", "101", "--assignment-id", "501", "--confirm", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("output should indicate dry run, got: %s", output)
	}
}

func TestAssignmentsDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"assignments", "delete", "--course-id", "101", "--assignment-id", "501"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is missing")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestAssignmentsDeleteConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignments", "delete", "--course-id", "101", "--assignment-id", "501", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "501") {
		t.Errorf("output should contain assignment ID, got: %s", output)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("output should confirm deletion, got: %s", output)
	}
}
