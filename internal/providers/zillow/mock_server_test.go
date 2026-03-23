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
	withPropertyDetailMock(mux)
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

// withPropertyDetailMock adds the property detail GraphQL endpoint mock.
func withPropertyDetailMock(mux *http.ServeMux) {
	mux.HandleFunc("/graphql/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"property": map[string]any{
					"zpid":          12345678,
					"address": map[string]any{
						"streetAddress": "123 Main St",
						"city":          "Denver",
						"state":         "CO",
						"zipcode":       "80202",
					},
					"price":         450000,
					"bedrooms":      3,
					"bathrooms":     2.0,
					"livingArea":    1800,
					"lotSize":       5000,
					"yearBuilt":     1995,
					"homeType":      "SINGLE_FAMILY",
					"homeStatus":    "FOR_SALE",
					"description":   "Beautiful single family home with updated kitchen.",
					"zestimate":     460000,
					"rentZestimate": 2200,
					"latitude":      39.7392,
					"longitude":     -104.9903,
					"daysOnZillow":  14,
					"monthlyHoaFee": 0,
					"listingAgent": map[string]any{
						"name": "Jane Smith",
					},
					"brokerageName": "Denver Realty",
					"responsivePhotos": []map[string]any{
						{"url": "https://photos.zillowstatic.com/photo1.jpg"},
						{"url": "https://photos.zillowstatic.com/photo2.jpg"},
					},
					"priceHistory": []map[string]any{
						{"date": "2024-01-15", "event": "Listed", "price": 450000, "source": "MLS"},
						{"date": "2020-06-01", "event": "Sold", "price": 380000, "source": "MLS"},
					},
					"taxHistory": []map[string]any{
						{"year": 2023, "taxPaid": 4500, "taxAssessment": 420000},
						{"year": 2022, "taxPaid": 4200, "taxAssessment": 400000},
					},
					"schools": []map[string]any{
						{"name": "Lincoln Elementary", "rating": 8, "level": "Elementary", "type": "Public", "grades": "K-5", "distance": 0.3, "link": "/school/lincoln-elementary"},
					},
					"walkScore": map[string]any{
						"walkscore":    82,
						"description":  "Very Walkable",
						"ws_link":      "/walk-score/123-main-st",
					},
					"transitScore": map[string]any{
						"transit_score": 65,
						"description":   "Excellent Transit",
					},
					"bikeScore": map[string]any{
						"bike_score":  70,
						"description": "Very Bikeable",
					},
					"nearbyHomes": []map[string]any{
						{
							"zpid":    "11111111",
							"price":   475000,
							"address": map[string]any{"streetAddress": "125 Main St", "city": "Denver", "state": "CO", "zipcode": "80202"},
							"bedrooms":   3,
							"bathrooms":  2.5,
							"livingArea": 2000,
						},
					},
					"comps": []map[string]any{
						{
							"zpid":    "22222222",
							"price":   440000,
							"address": map[string]any{"streetAddress": "789 Elm St", "city": "Denver", "state": "CO", "zipcode": "80202"},
							"bedrooms":   3,
							"bathrooms":  2.0,
							"livingArea": 1750,
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Property detail via URL scraping (fallback)
	mux.HandleFunc("/homedetails/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		// Return HTML with embedded __NEXT_DATA__ JSON
		propData := map[string]any{
			"zpid":          12345678,
			"streetAddress": "123 Main St",
			"city":          "Denver",
			"state":         "CO",
			"zipcode":       "80202",
			"price":         450000,
			"bedrooms":      3,
			"bathrooms":     2.0,
			"livingArea":    1800,
			"homeType":      "SINGLE_FAMILY",
			"homeStatus":    "FOR_SALE",
			"zestimate":     460000,
			"rentZestimate": 2200,
		}
		nextData := map[string]any{
			"props": map[string]any{
				"pageProps": map[string]any{
					"componentProps": map[string]any{
						"gdpClientCache": map[string]any{
							"property": propData,
						},
					},
				},
			},
		}
		encoded, _ := json.Marshal(nextData)
		w.Write([]byte(`<html><body><script id="__NEXT_DATA__" type="application/json">` + string(encoded) + `</script></body></html>`))
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
