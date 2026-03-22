package canvas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnrollmentsListText(t *testing.T) {
	mux := http.NewServeMux()
	withEnrollmentsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newEnrollmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"enrollments", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Jane Doe") {
		t.Errorf("expected student name in output, got: %s", output)
	}
	if !strings.Contains(output, "StudentEnrollment") {
		t.Errorf("expected enrollment type in output, got: %s", output)
	}
	if !strings.Contains(output, "Prof Smith") {
		t.Errorf("expected teacher name in output, got: %s", output)
	}
	if !strings.Contains(output, "TeacherEnrollment") {
		t.Errorf("expected teacher enrollment type in output, got: %s", output)
	}
}

func TestEnrollmentsListJSON(t *testing.T) {
	mux := http.NewServeMux()
	withEnrollmentsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newEnrollmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"enrollments", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"type"`) {
		t.Errorf("expected type field in JSON, got: %s", output)
	}
	if !strings.Contains(output, "StudentEnrollment") {
		t.Errorf("expected enrollment type in JSON output, got: %s", output)
	}
	if !strings.Contains(output, "Jane Doe") {
		t.Errorf("expected user name in JSON output, got: %s", output)
	}
}

func TestEnrollmentsListMissingCourseID(t *testing.T) {
	mux := http.NewServeMux()
	withEnrollmentsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newEnrollmentsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"enrollments", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestEnrollmentsGetText(t *testing.T) {
	mux := http.NewServeMux()
	withEnrollmentsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newEnrollmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"enrollments", "get", "--enrollment-id", "801"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "801") {
		t.Errorf("expected enrollment ID in output, got: %s", output)
	}
	if !strings.Contains(output, "StudentEnrollment") {
		t.Errorf("expected enrollment type in output, got: %s", output)
	}
	if !strings.Contains(output, "active") {
		t.Errorf("expected enrollment state in output, got: %s", output)
	}
	if !strings.Contains(output, "Jane Doe") {
		t.Errorf("expected user name in output, got: %s", output)
	}
	if !strings.Contains(output, "A") {
		t.Errorf("expected grade in output, got: %s", output)
	}
}

func TestEnrollmentsCreateDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withEnrollmentsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newEnrollmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"enrollments", "create",
			"--course-id", "101",
			"--user-id", "42",
			"--type", "StudentEnrollment",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
	if !strings.Contains(output, "StudentEnrollment") {
		t.Errorf("expected enrollment type in dry-run output, got: %s", output)
	}
}

func TestEnrollmentsCreateSuccess(t *testing.T) {
	mux := http.NewServeMux()
	withEnrollmentsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newEnrollmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"enrollments", "create",
			"--course-id", "101",
			"--user-id", "42",
			"--type", "StudentEnrollment",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "801") {
		t.Errorf("expected enrollment ID in output, got: %s", output)
	}
	if !strings.Contains(output, "StudentEnrollment") {
		t.Errorf("expected enrollment type in output, got: %s", output)
	}
}

func TestEnrollmentsDeleteNoConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withEnrollmentsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newEnrollmentsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"enrollments", "delete", "--course-id", "101", "--enrollment-id", "801"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is absent")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestEnrollmentsDeleteConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withEnrollmentsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newEnrollmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"enrollments", "delete",
			"--course-id", "101",
			"--enrollment-id", "801",
			"--confirm",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "801") {
		t.Errorf("expected enrollment ID in deletion output, got: %s", output)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}

func TestEnrollmentsDeactivateDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withEnrollmentsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newEnrollmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"enrollments", "deactivate",
			"--course-id", "101",
			"--enrollment-id", "801",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
}

func TestEnrollmentsReactivateDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withEnrollmentsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newEnrollmentsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"enrollments", "reactivate",
			"--course-id", "101",
			"--enrollment-id", "801",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
}
