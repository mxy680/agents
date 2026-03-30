package dof

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

// defaultOwnerRecords returns a representative set of test owner records.
func defaultOwnerRecords() []rawOwnerRecord {
	return []rawOwnerRecord{
		{
			BBLE:          "2029640028",
			OwnerName:     "ACME REALTY LLC",
			TaxClass:      "4",
			AssessedValue: "1200000",
			Borough:       "2",
			Block:         "02964",
			Lot:           "0028",
			Address:       "123 MAIN ST",
		},
		{
			BBLE:          "1000010001",
			OwnerName:     "SMITH PROPERTIES LLC",
			TaxClass:      "2",
			AssessedValue: "850000",
			Borough:       "1",
			Block:         "00001",
			Lot:           "0001",
			Address:       "456 BROADWAY",
		},
		{
			BBLE:          "3012340056",
			OwnerName:     "BROOKLYN LLC HOLDINGS",
			TaxClass:      "4",
			AssessedValue: "2100000",
			Borough:       "3",
			Block:         "01234",
			Lot:           "0056",
			Address:       "789 FLATBUSH AVE",
		},
	}
}

// buildOwnerResponse builds a JSON response for owner records.
func buildOwnerResponse(records []rawOwnerRecord) []byte {
	if records == nil {
		records = []rawOwnerRecord{}
	}
	b, _ := json.Marshal(records)
	return b
}

// newFullMockServer creates an httptest server with all DOF mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	withOwnersMock(mux)
	return httptest.NewServer(mux)
}

// withOwnersMock adds the owners endpoint mock routing based on query parameters.
func withOwnersMock(mux *http.ServeMux) {
	// The DOF dataset is a single Socrata endpoint; route based on $where param.
	mux.HandleFunc("/resource/w7rz-68fs.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		where := r.URL.Query().Get("$where")

		// Route to specific record for BBL lookup.
		if strings.Contains(where, "bble=") {
			if strings.Contains(where, "2029640028") {
				recs := []rawOwnerRecord{defaultOwnerRecords()[0]}
				w.Write(buildOwnerResponse(recs))
				return
			}
			// Unknown BBL — return empty.
			w.Write(buildOwnerResponse(nil))
			return
		}

		// Search / by-entity — return matching records filtered by owner name.
		all := defaultOwnerRecords()
		upper := strings.ToUpper(where)
		var filtered []rawOwnerRecord
		for _, rec := range all {
			if strings.Contains(upper, "LLC") && strings.Contains(rec.OwnerName, "LLC") {
				filtered = append(filtered, rec)
				continue
			}
			if strings.Contains(upper, "ACME") && strings.Contains(rec.OwnerName, "ACME") {
				filtered = append(filtered, rec)
				continue
			}
			if strings.Contains(upper, "SMITH") && strings.Contains(rec.OwnerName, "SMITH") {
				filtered = append(filtered, rec)
				continue
			}
		}
		w.Write(buildOwnerResponse(filtered))
	})

	// Also handle the root URL pattern used by newClientWithBase in tests.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		where := r.URL.Query().Get("$where")

		if strings.Contains(where, "bble=") {
			if strings.Contains(where, "2029640028") {
				recs := []rawOwnerRecord{defaultOwnerRecords()[0]}
				w.Write(buildOwnerResponse(recs))
				return
			}
			w.Write(buildOwnerResponse(nil))
			return
		}

		all := defaultOwnerRecords()
		upper := strings.ToUpper(where)
		var filtered []rawOwnerRecord
		for _, rec := range all {
			if strings.Contains(upper, "LLC") && strings.Contains(rec.OwnerName, "LLC") {
				filtered = append(filtered, rec)
				continue
			}
			if strings.Contains(upper, "ACME") && strings.Contains(rec.OwnerName, "ACME") {
				filtered = append(filtered, rec)
				continue
			}
			if strings.Contains(upper, "SMITH") && strings.Contains(rec.OwnerName, "SMITH") {
				filtered = append(filtered, rec)
				continue
			}
		}
		w.Write(buildOwnerResponse(filtered))
	})
}
