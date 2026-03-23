package canvas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQuizzesListText(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Midterm Quiz") {
		t.Errorf("expected quiz title in output, got: %s", output)
	}
	if !strings.Contains(output, "Pop Quiz") {
		t.Errorf("expected second quiz title in output, got: %s", output)
	}
	if !strings.Contains(output, "assignment") {
		t.Errorf("expected quiz type in output, got: %s", output)
	}
	// Published quiz should show "yes".
	if !strings.Contains(output, "yes") {
		t.Errorf("expected published=yes in output, got: %s", output)
	}
}

func TestQuizzesListJSON(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"title"`) {
		t.Errorf("expected title field in JSON, got: %s", output)
	}
	if !strings.Contains(output, "Midterm Quiz") {
		t.Errorf("expected quiz title in JSON output, got: %s", output)
	}
	if !strings.Contains(output, `"quiz_type"`) {
		t.Errorf("expected quiz_type field in JSON, got: %s", output)
	}
}

func TestQuizzesListMissingCourseID(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestQuizzesGetText(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "get", "--course-id", "101", "--quiz-id", "1101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Midterm Quiz") {
		t.Errorf("expected quiz title in output, got: %s", output)
	}
	if !strings.Contains(output, "1101") {
		t.Errorf("expected quiz ID in output, got: %s", output)
	}
	if !strings.Contains(output, "assignment") {
		t.Errorf("expected quiz type in output, got: %s", output)
	}
	if !strings.Contains(output, "60") {
		t.Errorf("expected time limit in output, got: %s", output)
	}
	if !strings.Contains(output, "Covers chapters 1-5") {
		t.Errorf("expected description in output, got: %s", output)
	}
}

func TestQuizzesGetMissingID(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "get", "--course-id", "101"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --quiz-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--quiz-id") {
		t.Errorf("error should mention --quiz-id, got: %v", execErr)
	}
}

func TestQuizzesCreateDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"quizzes", "create",
			"--course-id", "101",
			"--title", "Final Exam",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
	if !strings.Contains(output, "Final Exam") {
		t.Errorf("expected quiz title in dry-run output, got: %s", output)
	}
}

func TestQuizzesCreateSuccess(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"quizzes", "create",
			"--course-id", "101",
			"--title", "Pop Quiz",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1102") {
		t.Errorf("expected quiz ID in output, got: %s", output)
	}
	if !strings.Contains(output, "Pop Quiz") {
		t.Errorf("expected quiz title in output, got: %s", output)
	}
}

func TestQuizzesDeleteNoConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "delete", "--course-id", "101", "--quiz-id", "1101"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is absent")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestQuizzesDeleteConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"quizzes", "delete",
			"--course-id", "101",
			"--quiz-id", "1101",
			"--confirm",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1101") {
		t.Errorf("expected quiz ID in deletion output, got: %s", output)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}

func TestQuizzesQuestionsText(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "questions", "--course-id", "101", "--quiz-id", "1101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "multiple_choice_question") {
		t.Errorf("expected question type in output, got: %s", output)
	}
	if !strings.Contains(output, "true_false_question") {
		t.Errorf("expected second question type in output, got: %s", output)
	}
	if !strings.Contains(output, "Question 1") {
		t.Errorf("expected question name in output, got: %s", output)
	}
}

func TestQuizzesSubmissionsText(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "submissions", "--course-id", "101", "--quiz-id", "1101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "complete") {
		t.Errorf("expected submission state in output, got: %s", output)
	}
	if !strings.Contains(output, "45") {
		t.Errorf("expected score in output, got: %s", output)
	}
	if !strings.Contains(output, "38") {
		t.Errorf("expected second submission score in output, got: %s", output)
	}
}

func TestQuizzesSubmissionsJSON(t *testing.T) {
	mux := http.NewServeMux()
	withQuizzesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "submissions", "--course-id", "101", "--quiz-id", "1101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"workflow_state"`) {
		t.Errorf("expected workflow_state field in JSON, got: %s", output)
	}
	if !strings.Contains(output, "complete") {
		t.Errorf("expected submission state in JSON output, got: %s", output)
	}
}

func TestQuizzesCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "create", "--course-id", "101", "--title", "Pop Quiz"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "Pop Quiz") && !strings.Contains(output, "1102") {
		t.Errorf("expected quiz creation output, got: %s", output)
	}
}

func TestQuizzesUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "update", "--course-id", "101", "--quiz-id", "1101", "--title", "Updated Quiz"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1101") {
		t.Errorf("expected quiz ID 1101 in update output, got: %s", output)
	}
}

func TestQuizzesCreateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"quizzes", "create",
			"--course-id", "101",
			"--title", "Final Quiz",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1102") {
		t.Errorf("expected quiz ID in JSON output, got: %s", output)
	}
}

func TestQuizzesUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"quizzes", "update",
			"--course-id", "101",
			"--quiz-id", "1101",
			"--title", "Updated Quiz",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1101") {
		t.Errorf("expected quiz ID in JSON output, got: %s", output)
	}
}

func TestQuizzesQuestionsJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"quizzes", "questions", "--course-id", "101", "--quiz-id", "1101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "multiple_choice") {
		t.Errorf("expected question type in JSON output, got: %s", output)
	}
}
func TestQuizzesCreateWithOptionalFlags(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"quizzes", "create",
			"--course-id", "101",
			"--title", "Pop Quiz",
			"--description", "A short quiz",
			"--quiz-type", "assignment",
			"--time-limit", "30",
			"--points", "50",
			"--published",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1102") {
		t.Errorf("expected quiz ID in output, got: %s", output)
	}
}

func TestQuizzesUpdateWithOptionalFlags(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newQuizzesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"quizzes", "update",
			"--course-id", "101",
			"--quiz-id", "1101",
			"--title", "Updated Quiz",
			"--description", "Updated description",
			"--time-limit", "90",
			"--published",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1101") {
		t.Errorf("expected quiz ID in output, got: %s", output)
	}
}
