package canvas

import (
	"strings"
	"testing"
)

func TestUsersMeText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"users", "me"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Test User") {
		t.Errorf("output should contain user name, got: %s", output)
	}
	if !strings.Contains(output, "test@example.com") {
		t.Errorf("output should contain email, got: %s", output)
	}
	if !strings.Contains(output, "testuser") {
		t.Errorf("output should contain login ID, got: %s", output)
	}
}

func TestUsersMeJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"users", "me", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"name"`) {
		t.Errorf("JSON output should contain name field, got: %s", output)
	}
	if !strings.Contains(output, "Test User") {
		t.Errorf("JSON output should contain user name, got: %s", output)
	}
	if !strings.Contains(output, `"email"`) {
		t.Errorf("JSON output should contain email field, got: %s", output)
	}
}

func TestUsersTodoText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"users", "todo"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "grading") {
		t.Errorf("output should contain todo type, got: %s", output)
	}
	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("output should contain course name, got: %s", output)
	}
}

func TestUsersTodoJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"users", "todo", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"type"`) {
		t.Errorf("JSON output should contain type field, got: %s", output)
	}
	if !strings.Contains(output, "grading") {
		t.Errorf("JSON output should contain todo type value, got: %s", output)
	}
	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("JSON output should contain context name, got: %s", output)
	}
}

func TestUsersUpcomingText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"users", "upcoming"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Midterm Exam") {
		t.Errorf("output should contain event title, got: %s", output)
	}
	if !strings.Contains(output, "2026-03-15") {
		t.Errorf("output should contain event start date, got: %s", output)
	}
}

func TestUsersUpcomingJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"users", "upcoming", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"title"`) {
		t.Errorf("JSON output should contain title field, got: %s", output)
	}
	if !strings.Contains(output, "Midterm Exam") {
		t.Errorf("JSON output should contain event title, got: %s", output)
	}
	if !strings.Contains(output, `"start_at"`) {
		t.Errorf("JSON output should contain start_at field, got: %s", output)
	}
}

func TestUsersMissingText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"users", "missing"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Lab Report 1") {
		t.Errorf("output should contain missing assignment name, got: %s", output)
	}
	if !strings.Contains(output, "2026-01-20") {
		t.Errorf("output should contain due date, got: %s", output)
	}
	if !strings.Contains(output, "101") {
		t.Errorf("output should contain course ID, got: %s", output)
	}
}

func TestUsersMissingJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"users", "missing", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Lab Report 1") {
		t.Errorf("JSON output should contain missing assignment name, got: %s", output)
	}
}
