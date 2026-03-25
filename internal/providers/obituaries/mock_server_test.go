package obituaries

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// newTestClient creates a Client pointing at the given test server.
func newTestClient(server *httptest.Server) *Client {
	return newClientWithBase(server.Client(), server.URL)
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

	buf := make([]byte, 256*1024)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}

// testObituaries is the set of obituaries served by the mock server.
var testObituaries = []rawObituary{
	{
		ID:          100001,
		Name:        ObituaryName{First: "Maria", Last: "Gonzalez", Full: "Maria Gonzalez"},
		City:        "Bronx",
		State:       "New York",
		PublishDate: "2026-03-20",
		Age:         82,
		URL:         "https://www.legacy.com/obituaries/maria-gonzalez-100001",
		Publication: "New York Daily News",
	},
	{
		ID:          100002,
		Name:        ObituaryName{First: "James", Last: "Washington", Full: "James Washington"},
		City:        "Bronx",
		State:       "New York",
		PublishDate: "2026-03-18",
		Age:         75,
		URL:         "https://www.legacy.com/obituaries/james-washington-100002",
		Publication: "New York Post",
	},
	{
		ID:          100003,
		Name:        ObituaryName{First: "Rosa", Last: "Chen", Full: "Rosa Chen"},
		City:        "Bronx",
		State:       "New York",
		PublishDate: "2026-03-15",
		Age:         91,
		URL:         "https://www.legacy.com/obituaries/rosa-chen-100003",
		Publication: "New York Times",
	},
	{
		ID:          100004,
		Name:        ObituaryName{First: "Anthony", Last: "DeMarco", Full: "Anthony DeMarco"},
		City:        "Bronx",
		State:       "New York",
		PublishDate: "2026-03-10",
		URL:         "https://www.legacy.com/obituaries/anthony-demarco-100004",
		Publication: "New York Daily News",
	},
}

// buildSearchResponse serialises obituaries into the Legacy.com search response wire format.
func buildSearchResponse(obits []rawObituary) []byte {
	resp := map[string]any{
		"obituaries": obits,
		"totalCount": len(obits),
		"page":       1,
		"pageSize":   25,
	}
	b, _ := json.Marshal(resp)
	return b
}

// newFullMockServer creates an httptest server with all obituaries mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withSearchMock(mux)
	return httptest.NewServer(mux)
}

// withSearchMock adds the /search endpoint mock.
func withSearchMock(mux *http.ServeMux) {
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(buildSearchResponse(testObituaries))
	})
}
