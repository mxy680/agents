package canvas

import (
	"strings"
	"testing"
)

func TestContentExportsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentExportsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"content-exports", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "common_cartridge") {
		t.Errorf("expected export type 'common_cartridge' in output, got: %s", output)
	}
	if !strings.Contains(output, "exported") {
		t.Errorf("expected workflow state 'exported' in output, got: %s", output)
	}
}

func TestContentExportsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentExportsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"content-exports", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "export_type") {
		t.Errorf("JSON output should contain export_type field, got: %s", output)
	}
	if !strings.Contains(output, "common_cartridge") {
		t.Errorf("JSON output should contain export type value, got: %s", output)
	}
}

func TestContentExportsListMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentExportsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"content-exports", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestContentExportsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentExportsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"content-exports", "get", "--course-id", "101", "--export-id", "15001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "common_cartridge") {
		t.Errorf("expected export type, got: %s", output)
	}
	if !strings.Contains(output, "exported") {
		t.Errorf("expected workflow state, got: %s", output)
	}
	if !strings.Contains(output, "export.imscc") {
		t.Errorf("expected download URL in output, got: %s", output)
	}
}

func TestContentExportsCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newContentExportsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"content-exports", "create", "--course-id", "101", "--type", "common_cartridge"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "15001") && !strings.Contains(output, "created") && !strings.Contains(output, "common_cartridge") {
		t.Errorf("expected creation output, got: %s", output)
	}
}
