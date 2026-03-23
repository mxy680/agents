package canvas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestModulesListText(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"modules", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Module One") {
		t.Errorf("expected first module name, got: %s", output)
	}
	if !strings.Contains(output, "Module Two") {
		t.Errorf("expected second module name, got: %s", output)
	}
}

func TestModulesListJSON(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"modules", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"name"`) {
		t.Errorf("JSON output should contain name field, got: %s", output)
	}
	if !strings.Contains(output, "Module One") {
		t.Errorf("JSON output should contain module name, got: %s", output)
	}
}

func TestModulesListMissingCourseID(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"modules", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestModulesGetText(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"modules", "get", "--course-id", "101", "--module-id", "401"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Module One") {
		t.Errorf("expected module name, got: %s", output)
	}
	if !strings.Contains(output, "401") {
		t.Errorf("expected module ID in output, got: %s", output)
	}
	if !strings.Contains(output, "completed") {
		t.Errorf("expected module state in output, got: %s", output)
	}
}

func TestModulesCreateDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"modules", "create",
			"--course-id", "101",
			"--name", "Week 3 Materials",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
	if !strings.Contains(output, "Week 3 Materials") {
		t.Errorf("expected module name in dry-run output, got: %s", output)
	}
}

func TestModulesDeleteNoConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"modules", "delete", "--course-id", "101", "--module-id", "401"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is absent")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestModulesDeleteConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"modules", "delete",
			"--course-id", "101",
			"--module-id", "401",
			"--confirm",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "401") {
		t.Errorf("expected module ID in output, got: %s", output)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}

func TestModulesItemsText(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"modules", "items",
			"--course-id", "101",
			"--module-id", "401",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Read Chapter 1") {
		t.Errorf("expected first item title, got: %s", output)
	}
	if !strings.Contains(output, "Homework 1") {
		t.Errorf("expected second item title, got: %s", output)
	}
	if !strings.Contains(output, "Page") {
		t.Errorf("expected item type in output, got: %s", output)
	}
}

func TestModulesAddItemDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"modules", "add-item",
			"--course-id", "101",
			"--module-id", "401",
			"--type", "Assignment",
			"--content-id", "501",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
	if !strings.Contains(output, "Assignment") {
		t.Errorf("expected item type in dry-run output, got: %s", output)
	}
}

func TestModulesRemoveItemNoConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{
			"modules", "remove-item",
			"--course-id", "101",
			"--module-id", "401",
			"--item-id", "4001",
		})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is absent")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestModulesRemoveItemConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withModulesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"modules", "remove-item",
			"--course-id", "101",
			"--module-id", "401",
			"--item-id", "4001",
			"--confirm",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "4001") {
		t.Errorf("expected item ID in output, got: %s", output)
	}
	if !strings.Contains(output, "removed") {
		t.Errorf("expected 'removed' in output, got: %s", output)
	}
}

func TestModulesCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"modules", "create", "--course-id", "101", "--name", "New Module"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "New Module") && !strings.Contains(output, "402") {
		t.Errorf("expected creation output, got: %s", output)
	}
}

func TestModulesUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"modules", "update", "--course-id", "101", "--module-id", "401", "--name", "Updated Module"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "401") || !strings.Contains(output, "updated") {
		t.Errorf("expected update output, got: %s", output)
	}
}

func TestModulesAddItemLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"modules", "add-item",
			"--course-id", "101",
			"--module-id", "401",
			"--type", "Assignment",
			"--content-id", "501",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "added") && !strings.Contains(output, "4002") && !strings.Contains(output, "Assignment") {
		t.Errorf("expected item added output, got: %s", output)
	}
}

func TestModulesItemsJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"modules", "items", "--course-id", "101", "--module-id", "401", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework 1") {
		t.Errorf("expected module item in JSON output, got: %s", output)
	}
}

func TestModulesUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newModulesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"modules", "update", "--course-id", "101", "--module-id", "401", "--name", "Updated Module", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "401") {
		t.Errorf("expected module ID in JSON output, got: %s", output)
	}
}
