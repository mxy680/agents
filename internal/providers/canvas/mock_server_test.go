package canvas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// newTestSession returns a minimal CanvasSession for use in tests.
func newTestSession() *auth.CanvasSession {
	return &auth.CanvasSession{
		BaseURL:       "https://canvas.test.edu",
		SessionCookie: "test-session-cookie",
		CSRFToken:     "test-csrf-token",
		UserAgent:     "TestAgent/1.0",
	}
}

// newTestClient creates a Client pointing at the given test server.
func newTestClient(server *httptest.Server) *Client {
	return newClientWithBase(newTestSession(), server.Client(), server.URL)
}

// newTestClientFactory returns a ClientFactory pointing at the given test server.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return newTestClient(server), nil
	}
}

// newTestRootCmd creates a root command with --json and --dry-run flags wired up.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "Output as JSON")
	root.PersistentFlags().Bool("dry-run", false, "Preview actions without executing them")
	return root
}

// captureStdout captures stdout during f() and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}

// mustJSON marshals v to json.RawMessage, panicking on error.
func mustJSON(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// withCoursesMock registers mock handlers for course endpoints.
func withCoursesMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 101, "name": "Intro to CS", "course_code": "CS101",
				"workflow_state": "available", "total_students": 30,
			},
			{
				"id": 102, "name": "Data Structures", "course_code": "CS201",
				"workflow_state": "available", "total_students": 25,
			},
		})
	})
	mux.HandleFunc("/api/v1/courses/101", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 101, "name": "Intro to CS", "course_code": "CS101",
			"workflow_state": "available", "total_students": 30,
			"start_at": "2026-01-15T00:00:00Z", "end_at": "2026-05-15T00:00:00Z",
		})
	})
}

// withAssignmentsMock registers mock handlers for assignment endpoints.
func withAssignmentsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/assignments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 501, "name": "New Assignment", "course_id": 101,
				"points_possible": 100, "published": true,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 501, "name": "Homework 1", "course_id": 101,
				"due_at": "2026-02-01T23:59:00Z", "points_possible": 100,
				"grading_type": "points", "published": true,
			},
			{
				"id": 502, "name": "Homework 2", "course_id": 101,
				"due_at": "2026-02-15T23:59:00Z", "points_possible": 50,
				"grading_type": "points", "published": true,
			},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/assignments/501", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 501, "name": "Homework 1", "course_id": 101,
			"due_at": "2026-02-01T23:59:00Z", "points_possible": 100,
			"grading_type": "points", "published": true,
			"description": "Complete exercises 1-10",
			"submission_types": []string{"online_text_entry", "online_upload"},
		})
	})
}

// withSubmissionsMock registers mock handlers for submission endpoints.
func withSubmissionsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/assignments/501/submissions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1001, "assignment_id": 501, "user_id": 1,
				"workflow_state": "submitted", "submitted_at": "2026-02-01T20:00:00Z",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 1001, "assignment_id": 501, "user_id": 1,
				"workflow_state": "graded", "grade": "A", "score": 95,
				"submitted_at": "2026-02-01T20:00:00Z", "graded_at": "2026-02-03T10:00:00Z",
			},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/assignments/501/submissions/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 1001, "assignment_id": 501, "user_id": 1,
			"workflow_state": "graded", "grade": "A", "score": 95,
			"submitted_at": "2026-02-01T20:00:00Z", "graded_at": "2026-02-03T10:00:00Z",
		})
	})
}

// withUsersMock registers mock handlers for user endpoints.
func withUsersMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/users/self", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 1, "name": "Test User", "short_name": "Test",
			"email": "test@example.com", "login_id": "testuser",
			"avatar_url": "https://canvas.test.edu/avatar.png",
		})
	})
	mux.HandleFunc("/api/v1/users/self/todo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"type": "grading", "assignment": map[string]any{
					"id": 501, "name": "Homework 1",
				},
				"context_name": "Intro to CS",
			},
		})
	})
	mux.HandleFunc("/api/v1/users/self/upcoming_events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 1, "title": "Midterm Exam",
				"start_at": "2026-03-15T14:00:00Z",
				"context_code": "course_101",
			},
		})
	})
	mux.HandleFunc("/api/v1/users/self/missing_submissions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 503, "name": "Lab Report 1", "course_id": 101,
				"due_at": "2026-01-20T23:59:00Z", "points_possible": 25,
			},
		})
	})
}

// newFullMockServer creates a test server with all Canvas mock endpoints registered.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withCoursesMock(mux)
	withAssignmentsMock(mux)
	withSubmissionsMock(mux)
	withUsersMock(mux)
	return httptest.NewServer(mux)
}
