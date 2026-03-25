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

// buildListingsPage builds a mock JSON-LD page response containing listings.
// Each listing map should have keys: streetAddress, addressLocality, addressRegion,
// price (string like "$1,500,000"), url (optional), and type (optional, defaults to "ApartmentComplex").
func buildListingsPage(listings []map[string]any) []byte {
	var graphItems []map[string]any
	for _, l := range listings {
		itemType := "ApartmentComplex"
		if t, ok := l["type"].(string); ok && t != "" {
			itemType = t
		}
		item := map[string]any{
			"@type": itemType,
		}
		// Build address
		addr := map[string]any{
			"@type": "PostalAddress",
		}
		if v, ok := l["streetAddress"]; ok {
			addr["streetAddress"] = v
		}
		if v, ok := l["addressLocality"]; ok {
			addr["addressLocality"] = v
		}
		if v, ok := l["addressRegion"]; ok {
			addr["addressRegion"] = v
		}
		if v, ok := l["postalCode"]; ok {
			addr["postalCode"] = v
		}
		// Allow a flat "address" string for backward-compat in tests.
		if flatAddr, ok := l["address"].(string); ok && flatAddr != "" {
			// Store it in the item directly so extractListingSummary can find it,
			// but for JSON-LD we must parse it into components. For test simplicity
			// we set streetAddress to the full flat string.
			addr["streetAddress"] = flatAddr
		}
		item["address"] = addr

		// Build additionalProperty for price.
		if price, ok := l["price"]; ok {
			var priceStr string
			switch p := price.(type) {
			case string:
				priceStr = p
			case float64:
				// Format as "$X,XXX,XXX" — use json encoding for simplicity.
				priceStr = formatPriceString(int64(p))
			case int:
				priceStr = formatPriceString(int64(p))
			case int64:
				priceStr = formatPriceString(p)
			}
			if priceStr != "" {
				item["additionalProperty"] = map[string]any{
					"@type": "PropertyValue",
					"value": priceStr,
				}
			}
		}

		// URL
		if u, ok := l["url"].(string); ok && u != "" {
			item["url"] = u
		}

		graphItems = append(graphItems, item)
	}

	graph := map[string]any{
		"@context": "http://schema.org",
		"@graph":   graphItems,
	}
	if graphItems == nil {
		graph["@graph"] = []map[string]any{}
	}
	graphJSON, _ := json.Marshal(graph)
	page := `<!DOCTYPE html><html><head></head><body>` +
		`<script type="application/ld+json">` +
		string(graphJSON) +
		`</script></body></html>`
	return []byte(page)
}

// formatPriceString formats an int64 price as "$X,XXX,XXX".
func formatPriceString(price int64) string {
	if price == 0 {
		return ""
	}
	s := formatPrice(price)
	return s
}

// buildListingDetailPage builds a mock listing detail page.
// Price history is not available in JSON-LD, so this returns a minimal page.
// The history command will return an empty array (parsePriceHistory always returns empty).
func buildListingDetailPage(priceHistory []map[string]any) []byte {
	// Build a minimal JSON-LD page — price history is not in JSON-LD.
	graph := map[string]any{
		"@context": "http://schema.org",
		"@graph": []map[string]any{
			{
				"@type": "ApartmentComplex",
				"address": map[string]any{
					"@type":         "PostalAddress",
					"streetAddress": "100 Riverside Blvd",
					"addressLocality": "New York",
					"addressRegion": "NY",
				},
				"additionalProperty": map[string]any{
					"@type": "PropertyValue",
					"value": "$2,500,000",
				},
				"url": "/nyc/real_estate/12345678",
			},
		},
	}
	graphJSON, _ := json.Marshal(graph)
	page := `<!DOCTYPE html><html><head></head><body>` +
		`<script type="application/ld+json">` +
		string(graphJSON) +
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
			"streetAddress":   "100 Riverside Blvd",
			"addressLocality": "New York",
			"addressRegion":   "NY",
			"price":           float64(2500000),
			"url":             "/nyc/real_estate/12345678",
		},
		{
			"streetAddress":   "200 Central Park West",
			"addressLocality": "New York",
			"addressRegion":   "NY",
			"price":           float64(4750000),
			"url":             "/nyc/real_estate/87654321",
		},
	}

	// Match any /for-sale/* or /for-rent/* path that isn't a specific listing.
	mux.HandleFunc("/for-sale/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(buildListingsPage(listings))
	})
	mux.HandleFunc("/for-rent/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		rentListings := []map[string]any{
			{
				"streetAddress":   "300 West 23rd St",
				"addressLocality": "New York",
				"addressRegion":   "NY",
				"price":           float64(4500),
				"url":             "/nyc/real_estate/11111111",
			},
		}
		w.Write(buildListingsPage(rentListings))
	})
}

// withListingDetailMock adds a listing detail endpoint mock.
func withListingDetailMock(mux *http.ServeMux) {
	mux.HandleFunc("/nyc/real_estate/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(buildListingDetailPage(nil))
	})
	mux.HandleFunc("/building/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(buildListingDetailPage(nil))
	})
}
