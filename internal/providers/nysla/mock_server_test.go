package nysla

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

// testLicenses is a representative set of Bronx liquor license records.
var testLicenses = []rawLicense{
	{
		SerialNumber:    "1234567",
		LicenseTypeName: "RESTAURANT/BAR",
		LicenseTypeCode: "OP",
		PremisesName:    "THE BRONX GRILL",
		PremisesAddress: "123 GRAND CONCOURSE",
		City:            "BRONX",
		CountyName:      "BRONX",
		ZIP:             "10451",
		EffectiveDate:   "2025-03-01T00:00:00.000",
		ExpirationDate:  "2027-03-01T00:00:00.000",
		LicenseStatus:   "ACTIVE",
	},
	{
		SerialNumber:    "2345678",
		LicenseTypeName: "TAVERN",
		LicenseTypeCode: "TW",
		PremisesName:    "SOUTH BRONX TAVERN",
		PremisesAddress: "456 MELROSE AVE",
		City:            "BRONX",
		CountyName:      "BRONX",
		ZIP:             "10451",
		EffectiveDate:   "2024-11-15T00:00:00.000",
		ExpirationDate:  "2026-11-15T00:00:00.000",
		LicenseStatus:   "ACTIVE",
	},
	{
		SerialNumber:    "3456789",
		LicenseTypeName: "LIQUOR STORE",
		LicenseTypeCode: "L",
		PremisesName:    "FORDHAM LIQUORS",
		PremisesAddress: "789 FORDHAM RD",
		City:            "BRONX",
		CountyName:      "BRONX",
		ZIP:             "10458",
		EffectiveDate:   "2023-06-01T00:00:00.000",
		ExpirationDate:  "2025-06-01T00:00:00.000",
		LicenseStatus:   "ACTIVE",
	},
	{
		SerialNumber:    "4567890",
		LicenseTypeName: "RESTAURANT/BAR",
		LicenseTypeCode: "OP",
		PremisesName:    "HUNTS POINT KITCHEN",
		PremisesAddress: "321 HUNTS POINT AVE",
		City:            "BRONX",
		CountyName:      "BRONX",
		ZIP:             "10474",
		EffectiveDate:   "2025-01-10T00:00:00.000",
		ExpirationDate:  "2027-01-10T00:00:00.000",
		LicenseStatus:   "ACTIVE",
	},
	{
		SerialNumber:    "5678901",
		LicenseTypeName: "WINE BAR",
		LicenseTypeCode: "WB",
		PremisesName:    "MOTT HAVEN WINE",
		PremisesAddress: "100 THIRD AVE",
		City:            "BRONX",
		CountyName:      "BRONX",
		ZIP:             "10454",
		EffectiveDate:   "2025-06-20T00:00:00.000",
		ExpirationDate:  "2027-06-20T00:00:00.000",
		LicenseStatus:   "ACTIVE",
	},
}

// buildLicensesResponse serialises a license slice to JSON bytes.
func buildLicensesResponse(licenses []rawLicense) []byte {
	if licenses == nil {
		licenses = []rawLicense{}
	}
	b, _ := json.Marshal(licenses)
	return b
}

// newFullMockServer creates an httptest server that serves all NYSLA mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withLicensesMock(mux)
	return httptest.NewServer(mux)
}

// withLicensesMock adds the root Socrata endpoint that returns testLicenses.
// The mock ignores query parameters and always returns the full test data set,
// which is sufficient for verifying command wiring and output formatting.
func withLicensesMock(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(buildLicensesResponse(testLicenses))
	})
}
