package nydos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// newTestClient creates a Client pointing at the given test server for both endpoints.
func newTestClient(server *httptest.Server) *Client {
	return newClientWithBase(server.Client(), server.URL+"/daily", server.URL+"/corps")
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

// defaultDailyFilings returns representative test daily filing records.
func defaultDailyFilings() []DailyFilingRecord {
	return []DailyFilingRecord{
		{
			DOSID:      "6789012",
			CorpName:   "SEMINOLE HOLDINGS LLC",
			FilingDate: "2026-03-10T00:00:00.000",
			EntityType: "DOMESTIC LIMITED LIABILITY COMPANY",
			SOPName:    "JOHN DOE",
			SOPAddr1:   "1776 SEMINOLE AVE",
			SOPCity:    "BRONX",
			SOPState:   "NY",
			SOPZip5:    "10462",
			FilerName:  "JANE SMITH",
			FilerAddr1: "100 MAIN ST",
			FilerCity:  "NEW YORK",
			FilerState: "NY",
		},
		{
			DOSID:      "6789013",
			CorpName:   "ACME CORP INC",
			FilingDate: "2026-03-05T00:00:00.000",
			EntityType: "DOMESTIC BUSINESS CORPORATION",
			SOPName:    "BOB JONES",
			SOPAddr1:   "200 PARK AVE",
			SOPCity:    "NEW YORK",
			SOPState:   "NY",
			SOPZip5:    "10001",
			FilerName:  "BOB JONES",
			FilerAddr1: "200 PARK AVE",
			FilerCity:  "NEW YORK",
			FilerState: "NY",
		},
	}
}

// defaultActiveCorps returns representative test active corp records.
func defaultActiveCorps() []ActiveCorpRecord {
	return []ActiveCorpRecord{
		{
			DOSID:                "1234567",
			CurrentEntityName:    "SEMINOLE REALTY LLC",
			InitialDOSFilingDate: "2020-06-15T00:00:00.000",
			County:               "BRONX",
			Jurisdiction:         "NEW YORK",
			EntityTypeCode:       "DOMESTIC LIMITED LIABILITY COMPANY",
			DOSProcessName:       "REGISTERED AGENT INC",
			DOSProcessAddr1:      "99 WASHINGTON ST",
			DOSProcessCity:       "NEW YORK",
			DOSProcessState:      "NY",
			DOSProcessZip:        "10006",
		},
		{
			DOSID:                "2345678",
			CurrentEntityName:    "SEMINOLE PARTNERS LP",
			InitialDOSFilingDate: "2018-01-20T00:00:00.000",
			County:               "NEW YORK",
			Jurisdiction:         "NEW YORK",
			EntityTypeCode:       "DOMESTIC LIMITED PARTNERSHIP",
			DOSProcessName:       "PARTNER A",
			DOSProcessAddr1:      "55 WATER ST",
			DOSProcessCity:       "NEW YORK",
			DOSProcessState:      "NY",
			DOSProcessZip:        "10041",
		},
	}
}

// newFullMockServer creates an httptest server with all NY DOS mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	withDailyFilingsMock(mux)
	withActiveCorpsMock(mux)

	return httptest.NewServer(mux)
}

// withDailyFilingsMock adds the daily filings endpoint mock.
func withDailyFilingsMock(mux *http.ServeMux) {
	mux.HandleFunc("/daily", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(defaultDailyFilings())
		w.Write(b)
	})
}

// withActiveCorpsMock adds the active corporations endpoint mock.
func withActiveCorpsMock(mux *http.ServeMux) {
	mux.HandleFunc("/corps", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(defaultActiveCorps())
		w.Write(b)
	})
}
