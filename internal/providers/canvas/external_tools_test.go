package canvas

import (
	"strings"
	"testing"
)

func TestExternalToolsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newExternalToolsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"external-tools", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Khan Academy") {
		t.Errorf("expected 'Khan Academy' in output, got: %s", output)
	}
	if !strings.Contains(output, "Turnitin") {
		t.Errorf("expected 'Turnitin' in output, got: %s", output)
	}
	if !strings.Contains(output, "public") {
		t.Errorf("expected privacy level in output, got: %s", output)
	}
}

func TestExternalToolsListMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newExternalToolsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"external-tools", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestExternalToolsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newExternalToolsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"external-tools", "get", "--course-id", "101", "--tool-id", "13001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Khan Academy") {
		t.Errorf("expected tool name, got: %s", output)
	}
	if !strings.Contains(output, "khan.example.com") {
		t.Errorf("expected tool URL in output, got: %s", output)
	}
}

func TestExternalToolsDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newExternalToolsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"external-tools", "delete", "--course-id", "101", "--tool-id", "13001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is missing")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestExternalToolsCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newExternalToolsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"external-tools", "create",
			"--course-id", "101",
			"--name", "New LTI Tool",
			"--url", "https://lti.example.com/launch",
			"--consumer-key", "key123",
			"--shared-secret", "secret456",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "New LTI Tool") && !strings.Contains(output, "13001") {
		t.Errorf("expected tool creation output, got: %s", output)
	}
}

func TestExternalToolsUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newExternalToolsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"external-tools", "update",
			"--course-id", "101",
			"--tool-id", "13001",
			"--name", "Khan Academy Updated",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "13001") || !strings.Contains(output, "updated") {
		t.Errorf("expected tool update output, got: %s", output)
	}
}

func TestExternalToolsDeleteConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newExternalToolsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"external-tools", "delete", "--course-id", "101", "--tool-id", "13001", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "13001") || !strings.Contains(output, "deleted") {
		t.Errorf("expected tool deletion output, got: %s", output)
	}
}

func TestExternalToolsSessionlessLaunch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newExternalToolsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"external-tools", "sessionless-launch", "--course-id", "101", "--tool-id", "13001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "khan.example.com") && !strings.Contains(output, "Launch URL") && !strings.Contains(output, "sessionless") {
		t.Errorf("expected sessionless launch URL, got: %s", output)
	}
}
