package canvas

import (
	"strings"
	"testing"
)

func TestFavoritesCoursesText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "courses"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("expected 'Intro to CS' in favorites, got: %s", output)
	}
	if !strings.Contains(output, "CS101") {
		t.Errorf("expected course code in output, got: %s", output)
	}
}

func TestFavoritesGroupsText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "groups"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Study Group") {
		t.Errorf("expected 'Study Group' in favorites, got: %s", output)
	}
}

func TestFavoritesAddCourseDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "add-course", "--course-id", "101", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
}

func TestFavoritesAddCourseMissingID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "add-course"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestFavoritesAddCourseLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "add-course", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "101") {
		t.Errorf("expected course ID in output, got: %s", output)
	}
}

func TestFavoritesRemoveCourseLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "remove-course", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "101") {
		t.Errorf("expected course ID in output, got: %s", output)
	}
}

func TestFavoritesAddGroupLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "add-group", "--group-id", "2001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "2001") {
		t.Errorf("expected group ID in output, got: %s", output)
	}
}

func TestFavoritesRemoveGroupLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "remove-group", "--group-id", "2001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "2001") {
		t.Errorf("expected group ID in output, got: %s", output)
	}
}

func TestFavoritesCoursesJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "courses", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("expected course name in JSON output, got: %s", output)
	}
}

func TestFavoritesGroupsJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "groups", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Study Group") {
		t.Errorf("expected group name in JSON output, got: %s", output)
	}
}

func TestFavoritesAddCourseJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "add-course", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "101") {
		t.Errorf("expected course ID in JSON output, got: %s", output)
	}
}

func TestFavoritesAddGroupJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "add-group", "--group-id", "2001", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "2001") {
		t.Errorf("expected group ID in JSON output, got: %s", output)
	}
}

func TestFavoritesRemoveCourseJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "remove-course", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "101") {
		t.Errorf("expected course ID in JSON output, got: %s", output)
	}
}

func TestFavoritesRemoveGroupJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newFavoritesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"favorites", "remove-group", "--group-id", "2001", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "2001") {
		t.Errorf("expected group ID in JSON output, got: %s", output)
	}
}
