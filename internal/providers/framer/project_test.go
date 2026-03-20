package framer

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectInfo(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	cmd := newProjectInfoCmd(factory)
	root.AddCommand(cmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"info"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Test Project") {
		t.Errorf("expected output to contain 'Test Project', got: %s", output)
	}
}

func TestProjectInfoJSON(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	cmd := newProjectInfoCmd(factory)
	root.AddCommand(cmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"info", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var info ProjectInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("expected valid JSON, got: %s", output)
	}
	if info.Name != "Test Project" {
		t.Errorf("expected name 'Test Project', got: %s", info.Name)
	}
}

func TestProjectUser(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	cmd := newProjectUserCmd(factory)
	root.AddCommand(cmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"user"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Test User") {
		t.Errorf("expected output to contain 'Test User', got: %s", output)
	}
}

func TestProjectUserJSON(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	cmd := newProjectUserCmd(factory)
	root.AddCommand(cmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"user", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var user User
	if err := json.Unmarshal([]byte(output), &user); err != nil {
		t.Fatalf("expected valid JSON, got: %s", output)
	}
	if user.Name != "Test User" {
		t.Errorf("expected name 'Test User', got: %s", user.Name)
	}
}

func TestProjectChangedPaths(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	cmd := newProjectChangedPathsCmd(factory)
	root.AddCommand(cmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"changed-paths"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "/new-page") {
		t.Errorf("expected output to contain '/new-page', got: %s", output)
	}
	if !strings.Contains(output, "/home") {
		t.Errorf("expected output to contain '/home', got: %s", output)
	}
}

func TestProjectContributors(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	cmd := newProjectContributorsCmd(factory)
	root.AddCommand(cmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"contributors"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "alice@example.com") {
		t.Errorf("expected output to contain 'alice@example.com', got: %s", output)
	}
}

func TestProjectContributorsWithVersions(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	cmd := newProjectContributorsCmd(factory)
	root.AddCommand(cmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"contributors", "--from-version", "v1", "--to-version", "v2"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Contributors") {
		t.Errorf("expected output to contain 'Contributors', got: %s", output)
	}
}
