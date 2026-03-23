package canvas

import (
	"strings"
	"testing"
)

func TestGroupsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Study Group") {
		t.Errorf("expected group name 'Study Group', got: %s", output)
	}
	if !strings.Contains(output, "Lab Partners") {
		t.Errorf("expected group name 'Lab Partners', got: %s", output)
	}
}

func TestGroupsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "get", "--group-id", "2001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Study Group") {
		t.Errorf("expected group name, got: %s", output)
	}
	if !strings.Contains(output, "invitation_only") {
		t.Errorf("expected join level, got: %s", output)
	}
}

func TestGroupsGetMissingID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"groups", "get"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --group-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--group-id") {
		t.Errorf("error should mention --group-id, got: %v", execErr)
	}
}

func TestGroupsDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"groups", "delete", "--group-id", "2001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is missing")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestGroupsMembersText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "members", "--group-id", "2001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "42") {
		t.Errorf("expected member user_id 42 in output, got: %s", output)
	}
	if !strings.Contains(output, "accepted") {
		t.Errorf("expected membership state 'accepted', got: %s", output)
	}
}

func TestGroupsCategoriesText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "categories", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Project Groups") {
		t.Errorf("expected category 'Project Groups', got: %s", output)
	}
	if !strings.Contains(output, "Lab Groups") {
		t.Errorf("expected category 'Lab Groups', got: %s", output)
	}
}

func TestGroupsMembersJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "members", "--group-id", "2001", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"user_id"`) {
		t.Errorf("JSON output should contain user_id field, got: %s", output)
	}
}

func TestGroupsCategoriesJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "categories", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Project Groups") {
		t.Errorf("JSON output should contain category names, got: %s", output)
	}
}

func TestGroupsCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "create", "--name", "New Study Group", "--group-category-id", "3001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "New Group") && !strings.Contains(output, "2003") {
		t.Errorf("expected group creation output, got: %s", output)
	}
}

func TestGroupsUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "update", "--group-id", "2001", "--name", "Updated Study Group"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "2001") || !strings.Contains(output, "updated") {
		t.Errorf("expected update output, got: %s", output)
	}
}

func TestGroupsDeleteConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "delete", "--group-id", "2001", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "2001") || !strings.Contains(output, "deleted") {
		t.Errorf("expected deletion output, got: %s", output)
	}
}

func TestGroupsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "list", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Study Group") {
		t.Errorf("expected group name in JSON output, got: %s", output)
	}
}
