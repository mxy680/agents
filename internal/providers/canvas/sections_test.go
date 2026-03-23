package canvas

import (
	"strings"
	"testing"
)

func TestSectionsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Section A") {
		t.Errorf("expected 'Section A' in output, got: %s", output)
	}
	if !strings.Contains(output, "Section B") {
		t.Errorf("expected 'Section B' in output, got: %s", output)
	}
	if !strings.Contains(output, "15") {
		t.Errorf("expected student count in output, got: %s", output)
	}
}

func TestSectionsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "get", "--section-id", "6001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Section A") {
		t.Errorf("expected section name, got: %s", output)
	}
	if !strings.Contains(output, "101") {
		t.Errorf("expected course ID in output, got: %s", output)
	}
}

func TestSectionsCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "create", "--course-id", "101", "--name", "Section C", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
}

func TestSectionsDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"sections", "delete", "--section-id", "6001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is missing")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestSectionsDeleteConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "delete", "--section-id", "6001", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "6001") {
		t.Errorf("expected section ID 6001 in deletion output, got: %s", output)
	}
}

func TestSectionsCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "create", "--course-id", "101", "--name", "Section C"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "New Section") && !strings.Contains(output, "6001") {
		t.Errorf("expected creation output, got: %s", output)
	}
}

func TestSectionsUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "update", "--section-id", "6001", "--name", "Updated Section"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "6001") || !strings.Contains(output, "updated") {
		t.Errorf("expected update output, got: %s", output)
	}
}

func TestSectionsCrosslistLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "crosslist", "--section-id", "6001", "--new-course-id", "102"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "6001") && !strings.Contains(output, "cross") && !strings.Contains(output, "102") {
		t.Errorf("expected crosslist output, got: %s", output)
	}
}

func TestSectionsUncrosslistLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "uncrosslist", "--section-id", "6001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "6001") && !strings.Contains(output, "cross") {
		t.Errorf("expected uncrosslist output, got: %s", output)
	}
}

func TestSectionsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Section A") {
		t.Errorf("expected section name in JSON output, got: %s", output)
	}
}

func TestSectionsUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "update", "--section-id", "6001", "--name", "Updated Section", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "6001") {
		t.Errorf("expected section ID in JSON output, got: %s", output)
	}
}

func TestSectionsCrosslistJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "crosslist", "--section-id", "6001", "--new-course-id", "102", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "6001") {
		t.Errorf("expected section ID in JSON output, got: %s", output)
	}
}

func TestSectionsUncrosslistJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSectionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"sections", "uncrosslist", "--section-id", "6001", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "6001") {
		t.Errorf("expected section ID in JSON output, got: %s", output)
	}
}
