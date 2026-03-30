package hmda

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

// buildAggregationsResponse builds a mock HMDA aggregations API JSON response.
func buildAggregationsResponse(items []AggregationItem) []byte {
	resp := AggregationResponse{Aggregations: items}
	if items == nil {
		resp.Aggregations = []AggregationItem{}
	}
	b, _ := json.Marshal(resp)
	return b
}

// defaultAggregationItems returns a representative set of test aggregation items
// for the Bronx county (FIPS 36005).
func defaultAggregationItems() []AggregationItem {
	return []AggregationItem{
		{CensusTract: "36005000100", Count: 15, Sum: 7500000},
		{CensusTract: "36005000200", Count: 8, Sum: 3200000},
		{CensusTract: "36005000300", Count: 22, Sum: 11000000},
	}
}

// newFullMockServer creates an httptest server with all HMDA mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	withAggregationsMock(mux)

	return httptest.NewServer(mux)
}

// withAggregationsMock adds the aggregations endpoint mock.
func withAggregationsMock(mux *http.ServeMux) {
	mux.HandleFunc("/aggregations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(buildAggregationsResponse(defaultAggregationItems()))
	})
}
