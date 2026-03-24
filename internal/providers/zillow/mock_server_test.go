package zillow

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

// newFullMockServer creates an httptest server with all Zillow API mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	withSearchMock(mux)
	withAutocompleteMock(mux)
	withMortgageMock(mux)
	withLenderReviewsMock(mux)

	return httptest.NewServer(mux)
}

// withSearchMock adds the property search endpoint mock.
func withSearchMock(mux *http.ServeMux) {
	mux.HandleFunc("/async-create-search-page-state", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"cat1": map[string]any{
				"searchResults": map[string]any{
					"listResults": []map[string]any{
						{
							"zpid":       "12345678",
							"statusText": "FOR SALE",
							"unformattedPrice": 450000,
							"address":    "123 Main St, Denver, CO 80202",
							"addressStreet": "123 Main St",
							"addressCity":   "Denver",
							"addressState":  "CO",
							"addressZipcode": "80202",
							"beds":       3,
							"baths":      2.0,
							"area":       1800,
							"latLong": map[string]any{
								"latitude":  39.7392,
								"longitude": -104.9903,
							},
							"detailUrl": "/homedetails/123-Main-St-Denver-CO-80202/12345678_zpid/",
							"hdpData": map[string]any{
								"homeInfo": map[string]any{
									"zpid":     12345678,
									"homeType": "SINGLE_FAMILY",
									"daysOnZillow": 14,
								},
							},
						},
						{
							"zpid":       "87654321",
							"statusText": "FOR SALE",
							"unformattedPrice": 325000,
							"address":    "456 Oak Ave, Denver, CO 80203",
							"beds":       2,
							"baths":      1.5,
							"area":       1200,
							"latLong": map[string]any{
								"latitude":  39.7294,
								"longitude": -104.9815,
							},
							"detailUrl": "/homedetails/456-Oak-Ave-Denver-CO-80203/87654321_zpid/",
							"hdpData": map[string]any{
								"homeInfo": map[string]any{
									"zpid":     87654321,
									"homeType": "CONDO",
									"daysOnZillow": 7,
								},
							},
						},
					},
					"totalResultCount": 2,
				},
			},
			"cat2": map[string]any{
				"searchList": map[string]any{
					"totalPages":       1,
					"totalResultCount": 2,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// withAutocompleteMock adds the autocomplete suggestions endpoint mock.
func withAutocompleteMock(mux *http.ServeMux) {
	mux.HandleFunc("/autocomplete/v3/suggestions", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		results := []map[string]any{
			{
				"display":    "Denver, CO",
				"metaData": map[string]any{
					"zpid": "",
					"lat":  39.7392,
					"lng":  -104.9903,
				},
				"resultType": "city",
			},
		}
		if strings.Contains(q, "123") {
			results = []map[string]any{
				{
					"display":    "123 Main St, Denver, CO 80202",
					"metaData": map[string]any{
						"zpid": "12345678",
						"lat":  39.7392,
						"lng":  -104.9903,
					},
					"resultType": "address",
				},
			}
		}
		json.NewEncoder(w).Encode(map[string]any{
			"results": results,
		})
	})
}

// withMortgageMock adds the mortgage rates endpoint mock.
func withMortgageMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/getRates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "OK",
			"rates": map[string]any{
				"Fixed30Year": map[string]any{
					"samples": []map[string]any{
						{"rate": 6.875, "apr": 7.012, "date": "2024-12-15"},
					},
				},
				"Fixed15Year": map[string]any{
					"samples": []map[string]any{
						{"rate": 6.125, "apr": 6.287, "date": "2024-12-15"},
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/getCurrentRates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "OK",
			"rates": map[string]any{
				"Fixed30Year": map[string]any{
					"rate": 6.875,
					"apr":  7.012,
				},
				"Fixed15Year": map[string]any{
					"rate": 6.125,
					"apr":  6.287,
				},
			},
		})
	})
}

// withLenderReviewsMock adds the lender reviews endpoint mock.
func withLenderReviewsMock(mux *http.ServeMux) {
	mux.HandleFunc("/api/zillowLenderReviews", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"profileURL":   "https://www.zillow.com/lender-profile/12345/",
			"reviewURL":    "https://www.zillow.com/lender-reviews/12345/",
			"totalReviews": 42,
			"rating":       4.8,
			"reviews": []map[string]any{
				{
					"rating":                   5.0,
					"title":                    "Great experience",
					"content":                  "Very helpful and responsive.",
					"loanType":                 "Conventional",
					"loanProgram":              "Fixed30Year",
					"closingCostsSatisfaction": 5.0,
					"interestRateSatisfaction": 4.5,
					"verifiedReviewer":         true,
				},
			},
		})
	})
}
