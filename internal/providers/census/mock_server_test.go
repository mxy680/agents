package census

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

// testTractRows are representative Bronx tracts for mock responses.
var testTractRows = []map[string]string{
	{
		"NAME":         "Census Tract 1, Bronx County, New York",
		varPopulation:  "4521",
		varMedianIncome: "52300",
		varMedianRent:  "1450",
		varMedianHomeVal: "385000",
		varTotalHousing: "2100",
		varVacantUnits: "180",
		varOwnerOccupied: "480",
		varRenterOccupied: "1380",
		"state":  "36",
		"county": "005",
		"tract":  "000100",
	},
	{
		"NAME":         "Census Tract 2, Bronx County, New York",
		varPopulation:  "3200",
		varMedianIncome: "38750",
		varMedianRent:  "1200",
		varMedianHomeVal: "-",
		varTotalHousing: "1600",
		varVacantUnits: "200",
		varOwnerOccupied: "200",
		varRenterOccupied: "1200",
		"state":  "36",
		"county": "005",
		"tract":  "000200",
	},
	{
		"NAME":         "Census Tract 3, Bronx County, New York",
		varPopulation:  "5800",
		varMedianIncome: "61200",
		varMedianRent:  "1800",
		varMedianHomeVal: "520000",
		varTotalHousing: "2500",
		varVacantUnits: "125",
		varOwnerOccupied: "900",
		varRenterOccupied: "1450",
		"state":  "36",
		"county": "005",
		"tract":  "000300",
	},
	{
		"NAME":         "Census Tract 4, Bronx County, New York",
		varPopulation:  "2900",
		varMedianIncome: "29500",
		varMedianRent:  "1050",
		varMedianHomeVal: "-",
		varTotalHousing: "1400",
		varVacantUnits: "210",
		varOwnerOccupied: "140",
		varRenterOccupied: "1050",
		"state":  "36",
		"county": "005",
		"tract":  "000400",
	},
}

// buildCensusResponse serialises tract rows into the Census API [][]string wire format.
func buildCensusResponse(rows []map[string]string) []byte {
	headers := []string{
		"NAME",
		varPopulation, varMedianIncome, varMedianRent, varMedianHomeVal,
		varTotalHousing, varVacantUnits, varOwnerOccupied, varRenterOccupied,
		"state", "county", "tract",
	}
	var result [][]string
	result = append(result, headers)
	for _, row := range rows {
		r := make([]string, len(headers))
		for i, h := range headers {
			r[i] = row[h]
		}
		result = append(result, r)
	}
	data, _ := json.Marshal(result)
	return data
}

// newFullMockServer creates an httptest server with all Census mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withCensusMock(mux)
	return httptest.NewServer(mux)
}

// withCensusMock adds the Census ACS data endpoint mock.
// The Census API is a single endpoint with query parameters; we handle "/" for all years.
func withCensusMock(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// If a specific tract is requested (for the profile command), return only that tract.
		forParam := r.URL.Query().Get("for")
		tractRows := testTractRows
		if len(forParam) > 6 && forParam[:6] == "tract:" {
			tractCode := forParam[6:]
			if tractCode != "*" {
				var filtered []map[string]string
				for _, row := range testTractRows {
					if row["tract"] == tractCode {
						filtered = append(filtered, row)
					}
				}
				tractRows = filtered
			}
		}

		w.Write(buildCensusResponse(tractRows))
	})
}
