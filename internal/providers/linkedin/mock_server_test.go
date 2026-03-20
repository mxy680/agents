package linkedin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// newTestSession returns a minimal LinkedInSession for use in tests.
func newTestSession() *auth.LinkedInSession {
	return &auth.LinkedInSession{
		LiAt:       "test-li-at",
		JSESSIONID: "ajax:test-jsessionid",
		UserAgent:  "TestAgent/1.0",
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

// newTestRootCmd creates a root command with --json flag wired up.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "Output as JSON")
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

// containsStr is a test helper that checks if s contains sub.
func containsStr(s, sub string) bool {
	return strings.Contains(s, sub)
}

// withProfileMock registers profile-related mock handlers on mux.
func withProfileMock(mux *http.ServeMux) {
	// GET /voyager/api/identity/profiles/{publicId}
	mux.HandleFunc("/voyager/api/identity/profiles/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/voyager/api/identity/profiles/")
		publicID := strings.TrimSuffix(path, "/")

		if publicID == "" {
			http.Error(w, `{"message":"missing publicId"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"profile": {
				"entityUrn": "urn:li:fs_profile:ACoAABtest123",
				"firstName": "Test",
				"lastName": "User",
				"headline": "Software Engineer at TestCorp",
				"summary": "A test user profile.",
				"locationName": "San Francisco, CA",
				"industryName": "Computer Software",
				"profilePicture": {
					"displayImageReference": {
						"vectorImage": {
							"rootUrl": "https://example.com/pic/",
							"artifacts": [{"fileIdentifyingUrlPathSegment": "200x200.jpg"}]
						}
					}
				}
			},
			"connectionCount": 500,
			"followerCount": 1200
		}`))
	})

	// GET /voyager/api/me (current user profile)
	mux.HandleFunc("/voyager/api/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"miniProfile": {
				"entityUrn": "urn:li:fs_miniProfile:ACoAABtest123",
				"firstName": "Mark",
				"lastName": "Test",
				"occupation": "Software Engineer",
				"publicIdentifier": "marktest"
			}
		}`))
	})
}

// newFullMockServer creates a test server with all LinkedIn mock handlers.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withProfileMock(mux)
	return httptest.NewServer(mux)
}
