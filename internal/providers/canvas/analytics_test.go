package canvas

import (
	"strings"
	"testing"
)

func TestAnalyticsCourseJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "course", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "page_views") {
		t.Errorf("expected 'page_views' in JSON output, got: %s", output)
	}
	if !strings.Contains(output, "500") {
		t.Errorf("expected page view count in output, got: %s", output)
	}
}

func TestAnalyticsCourseText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "course", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "101") {
		t.Errorf("expected course ID in output, got: %s", output)
	}
}

func TestAnalyticsCoursesMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "course"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestAnalyticsAssignmentsJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "assignments", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework 1") {
		t.Errorf("expected 'Homework 1' in analytics output, got: %s", output)
	}
	if !strings.Contains(output, "assignment_id") {
		t.Errorf("expected 'assignment_id' field in JSON, got: %s", output)
	}
}

func TestAnalyticsAssignmentsText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "assignments", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework 1") {
		t.Errorf("expected assignment title in output, got: %s", output)
	}
	if !strings.Contains(output, "501") {
		t.Errorf("expected assignment ID in output, got: %s", output)
	}
}

func TestAnalyticsStudentJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "student", "--course-id", "101", "--student-id", "42", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "page_views") {
		t.Errorf("expected student analytics JSON with page_views, got: %s", output)
	}
}

func TestAnalyticsStudentText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "student", "--course-id", "101", "--student-id", "42"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "42") && !strings.Contains(output, "101") {
		t.Errorf("expected student or course ID in output, got: %s", output)
	}
}

func TestAnalyticsStudentAssignmentsJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "student-assignments", "--course-id", "101", "--student-id", "42", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework 1") {
		t.Errorf("expected assignment title in JSON, got: %s", output)
	}
}

func TestAnalyticsStudentAssignmentsText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "student-assignments", "--course-id", "101", "--student-id", "42"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "501") && !strings.Contains(output, "Homework 1") {
		t.Errorf("expected assignment data in output, got: %s", output)
	}
}
