package nyscef

import (
	"context"
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

// buildCaseSearchPage builds a mock HTML page simulating NYSCEF case search results.
func buildCaseSearchPage(cases []map[string]string) []byte {
	html := `<!DOCTYPE html><html><body><table><thead><tr><th>Index Number</th><th>Case Type</th><th>Caption</th><th>Filed</th><th>Court</th><th>Status</th></tr></thead><tbody>`
	for _, c := range cases {
		docketID := c["docketId"]
		indexNum := c["indexNumber"]
		caseType := c["caseType"]
		caption := c["caption"]
		filed := c["filingDate"]
		court := c["court"]
		status := c["status"]
		html += `<tr>`
		html += `<td><a href="/nyscef/CaseDetails?docketId=` + docketID + `">` + indexNum + `</a></td>`
		html += `<td>` + caseType + `</td>`
		html += `<td>` + caption + `</td>`
		html += `<td>` + filed + `</td>`
		html += `<td>` + court + `</td>`
		html += `<td>` + status + `</td>`
		html += `</tr>`
	}
	html += `</tbody></table></body></html>`
	return []byte(html)
}

// buildCaseDetailPage builds a mock HTML page simulating NYSCEF case details.
func buildCaseDetailPage(c map[string]string) []byte {
	html := `<!DOCTYPE html><html><body>`
	if caption, ok := c["caption"]; ok {
		html += `<h1>` + caption + `</h1>`
	}
	html += `<table>`
	if v, ok := c["indexNumber"]; ok {
		html += `<tr><td>Index Number</td><td>` + v + `</td></tr>`
	}
	if v, ok := c["caseType"]; ok {
		html += `<tr><td>Case Type</td><td>` + v + `</td></tr>`
	}
	if v, ok := c["filingDate"]; ok {
		html += `<tr><td>Date Filed</td><td>` + v + `</td></tr>`
	}
	if v, ok := c["court"]; ok {
		html += `<tr><td>Court</td><td>` + v + `</td></tr>`
	}
	if v, ok := c["status"]; ok {
		html += `<tr><td>Status</td><td>` + v + `</td></tr>`
	}
	html += `</table></body></html>`
	return []byte(html)
}

// newFullMockServer creates an httptest server with all NYSCEF mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withCaseSearchMock(mux)
	withCaseDetailMock(mux)
	return httptest.NewServer(mux)
}

// withCaseSearchMock adds the CaseSearch endpoint mock.
func withCaseSearchMock(mux *http.ServeMux) {
	cases := []map[string]string{
		{
			"docketId":    "123456",
			"indexNumber": "600001/2024",
			"caseType":    "Probate",
			"caption":     "In re Estate of John Doe",
			"filingDate":  "01/15/2024",
			"court":       "Bronx Surrogate",
			"status":      "Active",
		},
		{
			"docketId":    "789012",
			"indexNumber": "600002/2024",
			"caseType":    "Partition",
			"caption":     "Smith v. Smith",
			"filingDate":  "02/20/2024",
			"court":       "Bronx Supreme",
			"status":      "Active",
		},
	}

	mux.HandleFunc("/nyscef/CaseSearch", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(buildCaseSearchPage(cases))
	})
}

// withCaseDetailMock adds the CaseDetails endpoint mock.
func withCaseDetailMock(mux *http.ServeMux) {
	mux.HandleFunc("/nyscef/CaseDetails", func(w http.ResponseWriter, r *http.Request) {
		docketID := r.URL.Query().Get("docketId")
		var c map[string]string
		switch docketID {
		case "789012":
			c = map[string]string{
				"caption":     "Smith v. Smith",
				"indexNumber": "600002/2024",
				"caseType":    "Partition",
				"filingDate":  "02/20/2024",
				"court":       "Bronx Supreme Court",
				"status":      "Active",
			}
		default:
			c = map[string]string{
				"caption":     "In re Estate of John Doe",
				"indexNumber": "600001/2024",
				"caseType":    "Probate",
				"filingDate":  "01/15/2024",
				"court":       "Bronx Surrogate Court",
				"status":      "Active",
			}
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(buildCaseDetailPage(c))
	})
}
