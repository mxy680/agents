package canvas

import (
	"strings"
	"testing"
)

func TestCoursesListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCoursesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"courses", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("output should contain course name, got: %s", output)
	}
	if !strings.Contains(output, "CS101") {
		t.Errorf("output should contain course code, got: %s", output)
	}
	if !strings.Contains(output, "Data Structures") {
		t.Errorf("output should contain second course name, got: %s", output)
	}
}

func TestCoursesListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCoursesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"courses", "list", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"name"`) {
		t.Errorf("JSON output should contain name field, got: %s", output)
	}
	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("JSON output should contain course name, got: %s", output)
	}
	if !strings.Contains(output, "CS101") {
		t.Errorf("JSON output should contain course code, got: %s", output)
	}
}

func TestCoursesGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCoursesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"courses", "get", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("output should contain course name, got: %s", output)
	}
	if !strings.Contains(output, "CS101") {
		t.Errorf("output should contain course code, got: %s", output)
	}
	if !strings.Contains(output, "available") {
		t.Errorf("output should contain workflow state, got: %s", output)
	}
	if !strings.Contains(output, "30") {
		t.Errorf("output should contain student count, got: %s", output)
	}
}

func TestCoursesGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCoursesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"courses", "get", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"name"`) {
		t.Errorf("JSON output should contain name field, got: %s", output)
	}
	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("JSON output should contain course name, got: %s", output)
	}
	if !strings.Contains(output, `"course_code"`) {
		t.Errorf("JSON output should contain course_code field, got: %s", output)
	}
}

func TestCoursesGetMissingID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCoursesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"courses", "get"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}
