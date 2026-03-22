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

// withDiscussionsMock registers mock handlers for discussion topic endpoints.
func withDiscussionsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/discussion_topics/201/entries", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 3001, "user_id": 1, "user_name": "Test User",
				"message": "This is my reply", "created_at": "2026-02-10T10:00:00Z",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 3001, "user_id": 1, "user_name": "Test User",
				"message": "First entry message", "created_at": "2026-02-10T09:00:00Z",
			},
			{
				"id": 3002, "user_id": 2, "user_name": "Another User",
				"message": "Second entry message", "created_at": "2026-02-10T10:00:00Z",
			},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/discussion_topics/201/read_all", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/v1/courses/101/discussion_topics/201", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		if r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 201, "title": "Updated Discussion Topic", "discussion_type": "side_comment",
				"published": true, "pinned": false, "locked": false,
				"posted_at": "2026-02-01T12:00:00Z",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 201, "title": "Discussion Topic One", "discussion_type": "side_comment",
			"published": true, "pinned": false, "locked": false,
			"posted_at": "2026-02-01T12:00:00Z", "user_name": "Instructor",
			"message": "This is the discussion prompt.",
		})
	})
	mux.HandleFunc("/api/v1/courses/101/discussion_topics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 202, "title": "New Discussion Topic", "discussion_type": "side_comment",
				"published": false, "pinned": false,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 201, "title": "Discussion Topic One", "discussion_type": "side_comment",
				"published": true, "pinned": false, "locked": false,
			},
			{
				"id": 202, "title": "Discussion Topic Two", "discussion_type": "threaded",
				"published": true, "pinned": true, "locked": false,
			},
		})
	})
}

// withAnnouncementsMock registers mock handlers for announcement endpoints.
func withAnnouncementsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/announcements", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 301, "title": "Welcome Announcement", "published": true,
				"posted_at": "2026-01-15T09:00:00Z", "user_name": "Instructor",
				"message": "Welcome to the course!",
			},
			{
				"id": 302, "title": "Midterm Reminder", "published": true,
				"posted_at": "2026-03-01T09:00:00Z", "user_name": "Instructor",
				"message": "Midterm is next week.",
			},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/discussion_topics/301", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		if r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 301, "title": "Updated Announcement", "published": true,
				"posted_at": "2026-01-15T09:00:00Z",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 301, "title": "Welcome Announcement", "published": true,
			"posted_at": "2026-01-15T09:00:00Z", "user_name": "Instructor",
			"message": "Welcome to the course!",
		})
	})
}

// withPagesMock registers mock handlers for wiki page endpoints.
func withPagesMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/pages/test-page/revisions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"revision_id": 2, "updated_at": "2026-02-15T10:00:00Z",
				"latest": true, "edited_by": "Instructor",
			},
			{
				"revision_id": 1, "updated_at": "2026-01-15T10:00:00Z",
				"latest": false, "edited_by": "Instructor",
			},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/pages/test-page", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"url": "test-page"})
			return
		}
		if r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"url": "test-page", "title": "Updated Test Page", "published": true,
				"front_page": false, "editing_roles": "teachers",
				"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-02-01T00:00:00Z",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"url": "test-page", "title": "Test Page", "published": true,
			"front_page": false, "editing_roles": "teachers",
			"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-02-01T00:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/courses/101/pages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"url": "new-page", "title": "New Page", "published": false, "front_page": false,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"url": "test-page", "title": "Test Page", "published": true, "front_page": true,
			},
			{
				"url": "syllabus", "title": "Syllabus", "published": true, "front_page": false,
			},
		})
	})
}

// withModulesMock registers mock handlers for module endpoints.
func withModulesMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/modules/401/items/4001", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"deleted": true})
	})
	mux.HandleFunc("/api/v1/courses/101/modules/401/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 4002, "title": "New Assignment Item", "type": "Assignment",
				"position": 3, "module_id": 401,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 4001, "title": "Read Chapter 1", "type": "Page",
				"position": 1, "module_id": 401,
			},
			{
				"id": 4002, "title": "Homework 1", "type": "Assignment",
				"position": 2, "module_id": 401,
			},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/modules/401", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"deleted": true})
			return
		}
		if r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 401, "name": "Updated Module One", "position": 1,
				"items_count": 2, "published": true,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 401, "name": "Module One", "position": 1,
			"items_count": 2, "published": true, "state": "completed",
		})
	})
	mux.HandleFunc("/api/v1/courses/101/modules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 402, "name": "New Module", "position": 2,
				"items_count": 0, "published": false,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 401, "name": "Module One", "position": 1,
				"items_count": 2, "published": true,
			},
			{
				"id": 402, "name": "Module Two", "position": 2,
				"items_count": 5, "published": false,
			},
		})
	})
}

// withFilesMock registers mock handlers for file and folder endpoints.
func withFilesMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 601, "display_name": "homework.pdf", "filename": "homework.pdf",
				"content-type": "application/pdf", "size": 204800,
				"locked": false, "hidden": false,
				"created_at": "2026-01-15T10:00:00Z", "updated_at": "2026-01-15T10:00:00Z",
			},
			{
				"id": 602, "display_name": "syllabus.docx", "filename": "syllabus.docx",
				"content-type": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				"size": 51200, "locked": true, "hidden": false,
			},
		})
	})
	mux.HandleFunc("/api/v1/files/601", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 601, "display_name": "homework.pdf", "filename": "homework.pdf",
				"content-type": "application/pdf", "size": 204800,
				"locked": false, "hidden": false,
				"created_at": "2026-01-15T10:00:00Z", "updated_at": "2026-01-15T10:00:00Z",
				"url": "https://canvas.test.edu/files/601/download",
			})
		case http.MethodPut:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 601, "display_name": "homework-renamed.pdf", "filename": "homework-renamed.pdf",
				"content-type": "application/pdf", "size": 204800,
				"locked": false, "hidden": false,
			})
		case http.MethodDelete:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"id": 601, "display_name": "homework.pdf"})
		}
	})
	mux.HandleFunc("/api/v1/courses/101/folders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 702, "name": "New Folder", "full_name": "course files/New Folder",
				"files_count": 0, "folders_count": 0,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 701, "name": "course files", "full_name": "course files",
				"files_count": 5, "folders_count": 2,
			},
			{
				"id": 702, "name": "assignments", "full_name": "course files/assignments",
				"files_count": 3, "folders_count": 0,
			},
		})
	})
	mux.HandleFunc("/api/v1/folders/701/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 601, "display_name": "homework.pdf", "filename": "homework.pdf",
				"content-type": "application/pdf", "size": 204800, "locked": false,
			},
		})
	})
}

// withEnrollmentsMock registers mock handlers for enrollment endpoints.
func withEnrollmentsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/enrollments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 801, "course_id": 101, "user_id": 42,
				"type": "StudentEnrollment", "enrollment_state": "active",
				"user_name": "Jane Doe",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 801, "course_id": 101, "user_id": 42,
				"type": "StudentEnrollment", "enrollment_state": "active",
				"user_name": "Jane Doe",
			},
			{
				"id": 802, "course_id": 101, "user_id": 43,
				"type": "TeacherEnrollment", "enrollment_state": "active",
				"user_name": "Prof Smith",
			},
		})
	})
	mux.HandleFunc("/api/v1/accounts/self/enrollments/801", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 801, "course_id": 101, "user_id": 42,
			"type": "StudentEnrollment", "enrollment_state": "active",
			"user_name": "Jane Doe", "current_grade": "A", "current_score": 95.5,
		})
	})
	mux.HandleFunc("/api/v1/courses/101/enrollments/801", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		task := r.URL.Query().Get("task")
		switch task {
		case "delete":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 801, "course_id": 101, "user_id": 42,
				"type": "StudentEnrollment", "enrollment_state": "deleted",
			})
		case "deactivate":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 801, "course_id": 101, "user_id": 42,
				"type": "StudentEnrollment", "enrollment_state": "inactive",
			})
		case "conclude":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 801, "course_id": 101, "user_id": 42,
				"type": "StudentEnrollment", "enrollment_state": "completed",
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"id": 801, "course_id": 101, "user_id": 42,
				"type": "StudentEnrollment", "enrollment_state": "deleted",
			})
		}
	})
	mux.HandleFunc("/api/v1/courses/101/enrollments/801/reactivate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 801, "course_id": 101, "user_id": 42,
			"type": "StudentEnrollment", "enrollment_state": "active",
			"user_name": "Jane Doe",
		})
	})
}

// withCalendarMock registers mock handlers for calendar event endpoints.
func withCalendarMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/calendar_events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 902, "title": "Office Hours", "context_code": "course_101",
				"start_at": "2026-02-10T14:00:00Z", "end_at": "2026-02-10T15:00:00Z",
				"workflow_state": "active",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 901, "title": "Midterm Exam", "context_code": "course_101",
				"start_at": "2026-03-15T14:00:00Z", "end_at": "2026-03-15T16:00:00Z",
				"workflow_state": "active",
			},
			{
				"id": 902, "title": "Office Hours", "context_code": "course_101",
				"start_at": "2026-02-10T14:00:00Z", "end_at": "2026-02-10T15:00:00Z",
				"workflow_state": "active",
			},
		})
	})
	mux.HandleFunc("/api/v1/calendar_events/901", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 901, "title": "Midterm Exam", "context_code": "course_101",
				"start_at": "2026-03-15T14:00:00Z", "end_at": "2026-03-15T16:00:00Z",
				"workflow_state": "active", "location_name": "Room 204",
				"description": "Chapters 1-5",
			})
		case http.MethodPut:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 901, "title": "Midterm Exam Updated", "context_code": "course_101",
				"start_at": "2026-03-15T14:00:00Z", "end_at": "2026-03-15T16:00:00Z",
				"workflow_state": "active",
			})
		case http.MethodDelete:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"id": 901, "title": "Midterm Exam"})
		}
	})
}

// withConversationsMock registers mock handlers for conversation endpoints.
func withConversationsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/conversations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode([]map[string]any{
				{
					"id": 1002, "subject": "Question about homework",
					"workflow_state": "read", "message_count": 1,
					"last_message_at": "2026-02-01T10:00:00Z",
				},
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 1001, "subject": "Welcome to class",
				"workflow_state": "read", "message_count": 3,
				"last_message":    "See you Monday!",
				"last_message_at": "2026-01-20T09:00:00Z",
				"starred":         false,
			},
			{
				"id": 1002, "subject": "Question about homework",
				"workflow_state": "unread", "message_count": 1,
				"last_message":    "What is due Friday?",
				"last_message_at": "2026-02-01T10:00:00Z",
				"starred":         true,
			},
		})
	})
	mux.HandleFunc("/api/v1/conversations/1001", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1001, "subject": "Welcome to class",
				"workflow_state": "read", "message_count": 3,
				"last_message":    "See you Monday!",
				"last_message_at": "2026-01-20T09:00:00Z",
				"starred":         false,
				"participants":    []string{"Prof Smith", "Jane Doe"},
			})
		case http.MethodPut:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1001, "subject": "Welcome to class",
				"workflow_state": "read", "message_count": 3,
				"starred": true,
			})
		case http.MethodDelete:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"id": 1001})
		}
	})
	mux.HandleFunc("/api/v1/conversations/1001/add_message", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 1001, "subject": "Welcome to class",
			"workflow_state": "read", "message_count": 4,
			"last_message":    "Thanks for the update!",
			"last_message_at": "2026-01-21T08:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/conversations/mark_all_as_read", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{})
	})
	mux.HandleFunc("/api/v1/conversations/unread_count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"unread_count": 5})
	})
}

// withQuizzesMock registers mock handlers for quiz endpoints.
func withQuizzesMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/quizzes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1102, "title": "Pop Quiz", "quiz_type": "practice_quiz",
				"published": false, "question_count": 0, "points_possible": 0,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 1101, "title": "Midterm Quiz", "quiz_type": "assignment",
				"published": true, "question_count": 10, "points_possible": 50,
				"due_at": "2026-03-10T23:59:00Z",
			},
			{
				"id": 1102, "title": "Pop Quiz", "quiz_type": "practice_quiz",
				"published": false, "question_count": 5, "points_possible": 25,
			},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/quizzes/1101", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1101, "title": "Midterm Quiz", "quiz_type": "assignment",
				"published": true, "question_count": 10, "points_possible": 50,
				"time_limit": 60, "due_at": "2026-03-10T23:59:00Z",
				"description": "Covers chapters 1-5",
			})
		case http.MethodPut:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1101, "title": "Midterm Quiz Updated", "quiz_type": "assignment",
				"published": true, "question_count": 10, "points_possible": 50,
			})
		case http.MethodDelete:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"id": 1101})
		}
	})
	mux.HandleFunc("/api/v1/courses/101/quizzes/1101/questions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 201, "position": 1, "question_type": "multiple_choice_question",
				"question_name": "Question 1", "points_possible": 5,
			},
			{
				"id": 202, "position": 2, "question_type": "true_false_question",
				"question_name": "Question 2", "points_possible": 5,
			},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/quizzes/1101/submissions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"quiz_submissions": []map[string]any{
				{
					"id": 301, "quiz_id": 1101, "user_id": 42,
					"workflow_state": "complete", "score": 45, "attempt": 1,
					"finished_at": "2026-03-10T22:30:00Z",
				},
				{
					"id": 302, "quiz_id": 1101, "user_id": 43,
					"workflow_state": "complete", "score": 38, "attempt": 1,
					"finished_at": "2026-03-10T23:00:00Z",
				},
			},
		})
	})
}

// withGroupsMock registers mock handlers for group endpoints.
func withGroupsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/users/self/groups", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 2001, "name": "Study Group", "join_level": "invitation_only", "members_count": 5},
			{"id": 2002, "name": "Lab Partners", "join_level": "parent_context_auto_join", "members_count": 3},
		})
	})
	mux.HandleFunc("/api/v1/groups/2001", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 2001, "name": "Study Group", "join_level": "invitation_only",
			"members_count": 5, "context_type": "Course", "description": "Weekly study sessions",
		})
	})
	mux.HandleFunc("/api/v1/groups/2001/memberships", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 9001, "group_id": 2001, "user_id": 42, "workflow_state": "accepted"},
			{"id": 9002, "group_id": 2001, "user_id": 43, "workflow_state": "accepted"},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/group_categories", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 3001, "name": "Project Groups", "groups_count": 4, "self_signup": "enabled"},
			{"id": 3002, "name": "Lab Groups", "groups_count": 6, "self_signup": "restricted"},
		})
	})
}

// withRubricsMock registers mock handlers for rubric endpoints.
func withRubricsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/rubrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 4001, "title": "New Rubric", "points_possible": 50.0,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 4001, "title": "Essay Rubric", "points_possible": 100.0, "context_type": "Course"},
			{"id": 4002, "title": "Lab Rubric", "points_possible": 50.0, "context_type": "Course"},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/rubrics/4001", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 4001, "title": "Essay Rubric", "points_possible": 100.0, "context_type": "Course",
		})
	})
}

// withGradesMock is a no-op because grades tests reuse /api/v1/courses/101/enrollments
// already registered by withEnrollmentsMock. The enrollment mock returns grade fields.
func withGradesMock(_ *http.ServeMux) {}

// withSectionsMock registers mock handlers for section endpoints.
func withSectionsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/sections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 6001, "name": "New Section", "course_id": 101,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 6001, "name": "Section A", "course_id": 101, "total_students": 15},
			{"id": 6002, "name": "Section B", "course_id": 101, "total_students": 12},
		})
	})
	mux.HandleFunc("/api/v1/sections/6001", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 6001, "name": "Section A", "course_id": 101,
			"total_students": 15, "start_at": "2026-01-15T00:00:00Z",
		})
	})
}

// withPlannerMock registers mock handlers for planner endpoints.
func withPlannerMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/planner/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"plannable_type": "assignment",
				"plannable_date": "2026-02-01T23:59:00Z",
				"plannable":      map[string]any{"title": "Homework 1"},
			},
			{
				"plannable_type": "discussion_topic",
				"plannable_date": "2026-02-05T23:59:00Z",
				"plannable":      map[string]any{"title": "Class Discussion"},
			},
		})
	})
	mux.HandleFunc("/api/v1/planner/notes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 7001, "title": "Study for Exam", "todo_date": "2026-03-01T00:00:00Z",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 7001, "title": "Study for Exam", "todo_date": "2026-03-01T00:00:00Z"},
			{"id": 7002, "title": "Review Notes", "todo_date": "2026-03-05T00:00:00Z"},
		})
	})
	mux.HandleFunc("/api/v1/planner/notes/7001", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 7001, "title": "Study for Exam", "todo_date": "2026-03-01T00:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/planner/overrides", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{})
	})
}

// withBookmarksMock registers mock handlers for bookmark endpoints.
func withBookmarksMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/users/self/bookmarks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 8001, "name": "New Bookmark", "url": "https://canvas.edu/courses/101", "position": 1,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 8001, "name": "Intro to CS", "url": "https://canvas.edu/courses/101", "position": 1},
			{"id": 8002, "name": "Data Structures", "url": "https://canvas.edu/courses/102", "position": 2},
		})
	})
	mux.HandleFunc("/api/v1/users/self/bookmarks/8001", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 8001, "name": "Intro to CS", "url": "https://canvas.edu/courses/101", "position": 1,
		})
	})
}

// withFavoritesMock registers mock handlers for favorites endpoints.
func withFavoritesMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/users/self/favorites/courses", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 101, "name": "Intro to CS", "course_code": "CS101", "workflow_state": "available"},
		})
	})
	mux.HandleFunc("/api/v1/users/self/favorites/groups", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 2001, "name": "Study Group", "join_level": "invitation_only"},
		})
	})
	mux.HandleFunc("/api/v1/users/self/favorites/courses/101", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 101, "name": "Intro to CS", "course_code": "CS101",
		})
	})
}

// withSearchMock registers mock handlers for search endpoints.
func withSearchMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/search/recipients", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "42", "name": "Alice Student", "type": "user"},
			{"id": "course_101", "name": "Intro to CS", "type": "context"},
		})
	})
	mux.HandleFunc("/api/v1/search/all_courses", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 101, "name": "Intro to CS", "course_code": "CS101"},
			{"id": 102, "name": "Data Structures", "course_code": "CS201"},
		})
	})
	mux.HandleFunc("/api/v1/search/all", func(w http.ResponseWriter, r *http.Request) {
		// Return 404 to trigger fallback to /search/recipients in search all command.
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
}

// withAssignmentGroupsMock registers mock handlers for assignment group endpoints.
func withAssignmentGroupsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/assignment_groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 10001, "name": "New Group", "position": 1, "group_weight": 20.0,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 10001, "name": "Homework", "position": 1, "group_weight": 40.0},
			{"id": 10002, "name": "Exams", "position": 2, "group_weight": 60.0},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/assignment_groups/10001", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 10001, "name": "Homework", "position": 1, "group_weight": 40.0,
		})
	})
}

// withOutcomesMock registers mock handlers for outcomes endpoints.
func withOutcomesMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/outcomes/11001", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 11001, "title": "Critical Thinking",
			"display_name": "CT", "mastery_points": 3.0,
			"context_type": "Course", "context_id": 101,
			"description": "Students will demonstrate critical thinking skills.",
		})
	})
	mux.HandleFunc("/api/v1/courses/101/outcome_groups/root/outcomes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"outcome_id": 11001, "outcome": map[string]any{"title": "Critical Thinking"}},
		})
	})
}

// withAnalyticsMock registers mock handlers for analytics endpoints.
func withAnalyticsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/analytics/activity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"page_views":     map[string]any{"total": 500},
			"participations": map[string]any{"total": 120},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/analytics/assignments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"assignment_id": 501, "title": "Homework 1", "max_score": 100, "min_score": 55},
			{"assignment_id": 502, "title": "Homework 2", "max_score": 50, "min_score": 20},
		})
	})
}

// withNotificationsMock registers mock handlers for notifications endpoints.
func withNotificationsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/users/self/account_notifications", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id": 12001, "subject": "Canvas Maintenance",
				"message": "Canvas will be down Saturday.", "icon": "warning",
				"start_at": "2026-03-01T00:00:00Z", "end_at": "2026-03-02T00:00:00Z",
			},
		})
	})
	mux.HandleFunc("/api/v1/users/self/communication_channels/email/self/notification_preferences", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"notification_preferences": []map[string]any{
				{"notification": "Assignment Graded", "frequency": "immediately"},
				{"notification": "Assignment Due Date", "frequency": "daily"},
			},
		})
	})
}

// withExternalToolsMock registers mock handlers for external tool endpoints.
func withExternalToolsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/external_tools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 13001, "name": "New LTI Tool",
				"url": "https://lti.example.com/launch", "privacy_level": "public",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 13001, "name": "Khan Academy", "url": "https://khan.example.com/lti", "privacy_level": "public"},
			{"id": 13002, "name": "Turnitin", "url": "https://turnitin.example.com/lti", "privacy_level": "name_only"},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/external_tools/13001", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"delete": true})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 13001, "name": "Khan Academy",
			"url": "https://khan.example.com/lti", "privacy_level": "public",
			"description": "Khan Academy LTI integration",
		})
	})
}

// withPeerReviewsMock registers mock handlers for peer review endpoints.
func withPeerReviewsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/assignments/501/peer_reviews", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"assessor_id": 43, "asset_id": 1001, "workflow_state": "assigned"},
			{"assessor_id": 44, "asset_id": 1001, "workflow_state": "completed"},
		})
	})
}

// withContentMigrationsMock registers mock handlers for content migration endpoints.
func withContentMigrationsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/content_migrations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 14001, "migration_type": "course_copy_importer", "workflow_state": "running",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 14001, "migration_type": "course_copy_importer", "workflow_state": "completed"},
			{"id": 14002, "migration_type": "canvas_cartridge_importer", "workflow_state": "failed"},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/content_migrations/14001", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 14001, "migration_type": "course_copy_importer", "workflow_state": "completed",
			"progress_url": "https://canvas.test.edu/api/v1/progress/9999",
		})
	})
}

// withContentExportsMock registers mock handlers for content export endpoints.
func withContentExportsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/courses/101/content_exports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id": 15001, "export_type": "common_cartridge", "workflow_state": "created",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 15001, "export_type": "common_cartridge", "workflow_state": "exported"},
			{"id": 15002, "export_type": "qti", "workflow_state": "exported"},
		})
	})
	mux.HandleFunc("/api/v1/courses/101/content_exports/15001", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 15001, "export_type": "common_cartridge", "workflow_state": "exported",
			"attachment": map[string]any{
				"url": "https://canvas.test.edu/files/export.imscc",
			},
		})
	})
}

// withAuditMock registers mock handlers for audit endpoints used by grades history.
func withAuditMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/audit/grade_change/courses/101", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"events": []map[string]any{
				{"id": "evt1", "created_at": "2026-02-03T10:00:00Z", "event_type": "grade_change"},
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
	withDiscussionsMock(mux)
	withAnnouncementsMock(mux)
	withPagesMock(mux)
	withModulesMock(mux)
	withFilesMock(mux)
	withEnrollmentsMock(mux)
	withCalendarMock(mux)
	withConversationsMock(mux)
	withQuizzesMock(mux)
	withGroupsMock(mux)
	withRubricsMock(mux)
	withGradesMock(mux)
	withSectionsMock(mux)
	withPlannerMock(mux)
	withBookmarksMock(mux)
	withFavoritesMock(mux)
	withSearchMock(mux)
	withAssignmentGroupsMock(mux)
	withOutcomesMock(mux)
	withAnalyticsMock(mux)
	withNotificationsMock(mux)
	withExternalToolsMock(mux)
	withPeerReviewsMock(mux)
	withContentMigrationsMock(mux)
	withContentExportsMock(mux)
	withAuditMock(mux)
	return httptest.NewServer(mux)
}
