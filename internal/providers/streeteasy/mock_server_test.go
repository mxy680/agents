package streeteasy

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

// buildListingsPage builds a mock __NEXT_DATA__ page response containing listings.
func buildListingsPage(listings []map[string]any) []byte {
	nextData := map[string]any{
		"props": map[string]any{
			"pageProps": map[string]any{
				"listings": listings,
			},
		},
	}
	nextDataJSON, _ := json.Marshal(nextData)
	page := `<!DOCTYPE html><html><head></head><body>` +
		`<script id="__NEXT_DATA__" type="application/json">` +
		string(nextDataJSON) +
		`</script></body></html>`
	return []byte(page)
}

// buildListingDetailPage builds a mock listing detail page with price history.
func buildListingDetailPage(priceHistory []map[string]any) []byte {
	nextData := map[string]any{
		"props": map[string]any{
			"pageProps": map[string]any{
				"priceHistory": priceHistory,
			},
		},
	}
	nextDataJSON, _ := json.Marshal(nextData)
	page := `<!DOCTYPE html><html><head></head><body>` +
		`<script id="__NEXT_DATA__" type="application/json">` +
		string(nextDataJSON) +
		`</script></body></html>`
	return []byte(page)
}

// newFullMockServer creates an httptest server with all StreetEasy mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	withSearchMock(mux)
	withListingDetailMock(mux)

	return httptest.NewServer(mux)
}

// withSearchMock adds the search page endpoint mock.
func withSearchMock(mux *http.ServeMux) {
	listings := []map[string]any{
		{
			"id":           "12345678",
			"address":      "100 Riverside Blvd, New York, NY 10069",
			"price":        2500000,
			"bedrooms":     3,
			"bathrooms":    2.5,
			"sqft":         1800,
			"daysOnMarket": 14,
			"status":       "for_sale",
			"url":          "/nyc/real_estate/12345678",
		},
		{
			"id":           "87654321",
			"address":      "200 Central Park West, New York, NY 10024",
			"price":        4750000,
			"bedrooms":     4,
			"bathrooms":    3.0,
			"sqft":         2400,
			"daysOnMarket": 7,
			"status":       "for_sale",
			"url":          "/nyc/real_estate/87654321",
		},
	}

	// Match any /for-sale/* or /for-rent/* path that isn't a specific listing
	mux.HandleFunc("/for-sale/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(buildListingsPage(listings))
	})
	mux.HandleFunc("/for-rent/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		// Return empty listings for rent (different status)
		rentListings := []map[string]any{
			{
				"id":      "11111111",
				"address": "300 West 23rd St, New York, NY 10011",
				"price":   4500,
				"status":  "for_rent",
				"url":     "/nyc/real_estate/11111111",
			},
		}
		w.Write(buildListingsPage(rentListings))
	})
}

// withListingDetailMock adds a listing detail endpoint mock.
func withListingDetailMock(mux *http.ServeMux) {
	priceHistory := []map[string]any{
		{
			"date":  "2024-01-15",
			"event": "Listed",
			"price": 2750000,
		},
		{
			"date":  "2024-03-01",
			"event": "Price Change",
			"price": 2600000,
		},
		{
			"date":  "2024-06-10",
			"event": "Price Change",
			"price": 2500000,
		},
	}

	mux.HandleFunc("/nyc/real_estate/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(buildListingDetailPage(priceHistory))
	})
	mux.HandleFunc("/building/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(buildListingDetailPage(priceHistory))
	})
}
