package canvas

import (
	"strings"
	"testing"
)

func TestAssignmentGroupsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignment-groups", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework") {
		t.Errorf("expected 'Homework' assignment group, got: %s", output)
	}
	if !strings.Contains(output, "Exams") {
		t.Errorf("expected 'Exams' assignment group, got: %s", output)
	}
	if !strings.Contains(output, "40") {
		t.Errorf("expected group weight in output, got: %s", output)
	}
}

func TestAssignmentGroupsListMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"assignment-groups", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestAssignmentGroupsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignment-groups", "get", "--course-id", "101", "--group-id", "10001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework") {
		t.Errorf("expected group name, got: %s", output)
	}
	if !strings.Contains(output, "40") {
		t.Errorf("expected group weight in output, got: %s", output)
	}
}

func TestAssignmentGroupsDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"assignment-groups", "delete", "--course-id", "101", "--group-id", "10001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is missing")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestAssignmentGroupsCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignment-groups", "create", "--course-id", "101", "--name", "Quizzes"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "New Group") && !strings.Contains(output, "10001") {
		t.Errorf("expected creation output, got: %s", output)
	}
}

func TestAssignmentGroupsUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignment-groups", "update", "--course-id", "101", "--group-id", "10001", "--name", "Homework Updated"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "10001") || !strings.Contains(output, "updated") {
		t.Errorf("expected update output with ID 10001, got: %s", output)
	}
}

func TestAssignmentGroupsDeleteConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignment-groups", "delete", "--course-id", "101", "--group-id", "10001", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "10001") || !strings.Contains(output, "deleted") {
		t.Errorf("expected deletion output, got: %s", output)
	}
}

func TestAssignmentGroupsCreateWithWeight(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"assignment-groups", "create",
			"--course-id", "101",
			"--name", "Labs",
			"--weight", "20",
			"--position", "3",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "10001") {
		t.Errorf("expected assignment group creation output, got: %s", output)
	}
}

func TestAssignmentGroupsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignment-groups", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework") {
		t.Errorf("expected assignment group name in JSON output, got: %s", output)
	}
}

func TestAssignmentGroupsUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignment-groups", "update", "--course-id", "101", "--group-id", "10001", "--name", "Homework Updated", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "10001") {
		t.Errorf("expected group ID in JSON output, got: %s", output)
	}
}

func TestAssignmentGroupsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAssignmentGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"assignment-groups", "get", "--course-id", "101", "--group-id", "10001", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework") {
		t.Errorf("expected group name in JSON output, got: %s", output)
	}
}
