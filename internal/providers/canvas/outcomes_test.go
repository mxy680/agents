package canvas

import (
	"strings"
	"testing"
)

func TestOutcomesGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "get", "--outcome-id", "11001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Critical Thinking") {
		t.Errorf("expected outcome title, got: %s", output)
	}
	if !strings.Contains(output, "3") {
		t.Errorf("expected mastery points in output, got: %s", output)
	}
}

func TestOutcomesGetMissingID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "get"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --outcome-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--outcome-id") {
		t.Errorf("error should mention --outcome-id, got: %v", execErr)
	}
}

func TestOutcomesDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "delete", "--outcome-id", "11001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is missing")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestOutcomesDeleteConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "delete", "--outcome-id", "11001", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "11001") {
		t.Errorf("expected outcome ID 11001 in deletion output, got: %s", output)
	}
}

func TestOutcomesListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "list", "--context-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Critical Thinking") {
		t.Errorf("expected outcome title, got: %s", output)
	}
}

func TestOutcomesCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "create", "--title", "New Outcome", "--context-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "New Outcome") && !strings.Contains(output, "11002") {
		t.Errorf("expected creation output, got: %s", output)
	}
}

func TestOutcomesUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "update", "--outcome-id", "11001", "--title", "Updated Outcome"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "11001") || !strings.Contains(output, "updated") {
		t.Errorf("expected update output, got: %s", output)
	}
}

func TestOutcomesGroupsText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "groups", "--context-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Root Outcome Group") && !strings.Contains(output, "20001") && !strings.Contains(output, "group") {
		t.Errorf("expected outcome groups output, got: %s", output)
	}
}

func TestOutcomesResultsText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "results", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "101") {
		t.Errorf("expected course ID in results output, got: %s", output)
	}
}

func TestOutcomesListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "list", "--context-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Critical Thinking") {
		t.Errorf("expected outcome title in JSON output, got: %s", output)
	}
}

func TestOutcomesCreateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "create", "--title", "New Outcome", "--context-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "outcome") {
		t.Errorf("expected outcome in JSON output, got: %s", output)
	}
}

func TestOutcomesUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "update", "--outcome-id", "11001", "--title", "Updated Outcome", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "11001") {
		t.Errorf("expected outcome ID in JSON output, got: %s", output)
	}
}

func TestOutcomesGroupsJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "groups", "--context-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Root Outcome Group") {
		t.Errorf("expected outcome groups in JSON output, got: %s", output)
	}
}

func TestOutcomesListMissingContextID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --context-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--context-id") {
		t.Errorf("error should mention --context-id, got: %v", execErr)
	}
}

func TestOutcomesCreateWithOptionalFlags(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"outcomes", "create",
			"--title", "New Outcome",
			"--context-id", "101",
			"--description", "An important outcome",
			"--mastery-points", "4",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "11002") && !strings.Contains(output, "New Outcome") && !strings.Contains(output, "created") {
		t.Errorf("expected outcome creation output, got: %s", output)
	}
}

func TestOutcomesInvalidContextType(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newOutcomesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"outcomes", "list", "--context-id", "101", "--context-type", "invalid"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error for invalid context-type")
	}
	if !strings.Contains(execErr.Error(), "context-type") {
		t.Errorf("error should mention context-type, got: %v", execErr)
	}
}
