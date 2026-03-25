package citibike

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

// testStations is the set of stations served by the mock server.
// Stations are near Times Square (40.7580, -73.9855).
var testStations = []gbfsStationInfo{
	{
		StationID: "station-001",
		Name:      "W 42 St & 8 Ave",
		Lat:       40.7580,
		Lon:       -73.9855,
		Capacity:  35,
	},
	{
		StationID: "station-002",
		Name:      "W 38 St & 8 Ave",
		Lat:       40.7544,
		Lon:       -73.9940,
		Capacity:  27,
	},
	{
		StationID: "station-003",
		Name:      "1 Ave & E 30 St",
		Lat:       40.7418,
		Lon:       -73.9759,
		Capacity:  42,
	},
}

// buildStationInformationResponse serialises stations into the GBFS wire format.
func buildStationInformationResponse(stations []gbfsStationInfo) []byte {
	resp := map[string]any{
		"data": map[string]any{
			"stations": stations,
		},
	}
	b, _ := json.Marshal(resp)
	return b
}

// newFullMockServer creates an httptest server with all Citi Bike mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withStationInformationMock(mux)
	return httptest.NewServer(mux)
}

// withStationInformationMock adds the station_information.json endpoint.
func withStationInformationMock(mux *http.ServeMux) {
	mux.HandleFunc("/station_information.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(buildStationInformationResponse(testStations))
	})
}
