package places

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
	api "google.golang.org/api/places/v1"
	"google.golang.org/api/option"
)

// newFullMockServer creates a test server that handles all Places API endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withSearchTextMock(mux)
	withSearchNearbyMock(mux)
	withGetMock(mux)
	withAutocompleteMock(mux)
	withPhotosMock(mux)
	return httptest.NewServer(mux)
}

func withSearchTextMock(mux *http.ServeMux) {
	mux.HandleFunc("/v1/places:searchText", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"places": []map[string]any{
				{
					"name":             "places/ChIJ1",
					"displayName":      map[string]string{"text": "Coffee Corner"},
					"formattedAddress": "123 Main St, Cleveland, OH 44101",
					"types":            []string{"cafe", "food", "point_of_interest"},
					"rating":           4.5,
					"userRatingCount":  120,
					"priceLevel":       "PRICE_LEVEL_MODERATE",
					"businessStatus":   "OPERATIONAL",
					"regularOpeningHours": map[string]any{
						"openNow": true,
					},
					"googleMapsUri":            "https://maps.google.com/?cid=123",
					"internationalPhoneNumber": "+1 216-555-0100",
					"websiteUri":               "https://coffeecorner.example.com",
				},
				{
					"name":             "places/ChIJ2",
					"displayName":      map[string]string{"text": "Bean & Leaf"},
					"formattedAddress": "456 Oak Ave, Cleveland, OH 44102",
					"types":            []string{"cafe", "food"},
					"rating":           4.2,
					"userRatingCount":  85,
					"priceLevel":       "PRICE_LEVEL_INEXPENSIVE",
					"businessStatus":   "OPERATIONAL",
				},
			},
			"nextPageToken": "token123",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withSearchNearbyMock(mux *http.ServeMux) {
	mux.HandleFunc("/v1/places:searchNearby", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"places": []map[string]any{
				{
					"name":             "places/ChIJ3",
					"displayName":      map[string]string{"text": "Nearby Diner"},
					"formattedAddress": "789 Elm St, Cleveland, OH 44103",
					"types":            []string{"restaurant", "food"},
					"rating":           3.8,
					"userRatingCount":  45,
					"businessStatus":   "OPERATIONAL",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withGetMock(mux *http.ServeMux) {
	mux.HandleFunc("/v1/places/ChIJ1", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"name":                     "places/ChIJ1",
			"id":                       "ChIJ1",
			"displayName":             map[string]string{"text": "Coffee Corner"},
			"formattedAddress":        "123 Main St, Cleveland, OH 44101",
			"shortFormattedAddress":   "123 Main St",
			"types":                   []string{"cafe", "food", "point_of_interest"},
			"primaryType":             "cafe",
			"rating":                  4.5,
			"userRatingCount":         120,
			"priceLevel":              "PRICE_LEVEL_MODERATE",
			"businessStatus":          "OPERATIONAL",
			"googleMapsUri":           "https://maps.google.com/?cid=123",
			"internationalPhoneNumber": "+1 216-555-0100",
			"websiteUri":              "https://coffeecorner.example.com",
			"editorialSummary":        map[string]string{"text": "A cozy coffee shop with excellent espresso."},
			"location": map[string]any{
				"latitude":  41.4993,
				"longitude": -81.6944,
			},
			"delivery":       true,
			"dineIn":         true,
			"takeout":        true,
			"curbsidePickup": false,
			"reservable":     false,
			"regularOpeningHours": map[string]any{
				"openNow": true,
				"weekdayDescriptions": []string{
					"Monday: 6:00 AM – 8:00 PM",
					"Tuesday: 6:00 AM – 8:00 PM",
					"Wednesday: 6:00 AM – 8:00 PM",
					"Thursday: 6:00 AM – 8:00 PM",
					"Friday: 6:00 AM – 9:00 PM",
					"Saturday: 7:00 AM – 9:00 PM",
					"Sunday: 7:00 AM – 6:00 PM",
				},
			},
			"reviews": []map[string]any{
				{
					"rating":                          5,
					"authorAttribution":               map[string]string{"displayName": "Alice"},
					"text":                            map[string]string{"text": "Best coffee in town!"},
					"relativePublishTimeDescription": "a week ago",
					"publishTime":                    "2026-03-09T10:00:00Z",
				},
			},
			"photos": []map[string]any{
				{
					"name":     "places/ChIJ1/photos/abc123",
					"widthPx":  4032,
					"heightPx": 3024,
				},
			},
			"addressComponents": []map[string]any{
				{
					"longText":  "123",
					"shortText": "123",
					"types":     []string{"street_number"},
				},
				{
					"longText":  "Main Street",
					"shortText": "Main St",
					"types":     []string{"route"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withAutocompleteMock(mux *http.ServeMux) {
	mux.HandleFunc("/v1/places:autocomplete", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"suggestions": []map[string]any{
				{
					"placePrediction": map[string]any{
						"place":          "places/ChIJ1",
						"text":           map[string]string{"text": "Coffee Corner, 123 Main St, Cleveland, OH"},
						"distanceMeters": 1500,
						"structuredFormat": map[string]any{
							"mainText": map[string]string{"text": "Coffee Corner"},
						},
					},
				},
				{
					"queryPrediction": map[string]any{
						"text": map[string]string{"text": "coffee shops near me"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func withPhotosMock(mux *http.ServeMux) {
	mux.HandleFunc("/v1/places/ChIJ1/photos/abc123/media", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"name":     "places/ChIJ1/photos/abc123/media",
			"photoUri": "https://lh3.googleusercontent.com/places/photo123",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// newTestServiceFactory returns a ServiceFactory that creates a Places service
// backed by the given httptest server, bypassing OAuth entirely.
func newTestServiceFactory(server *httptest.Server) ServiceFactory {
	return func(ctx context.Context) (*api.Service, error) {
		return api.NewService(ctx,
			option.WithoutAuthentication(),
			option.WithEndpoint(server.URL+"/"),
			option.WithHTTPClient(server.Client()),
		)
	}
}

// captureStdout runs f with os.Stdout redirected to a pipe and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	buf := make([]byte, 65536)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// newTestRootCmd creates a root command with global flags for testing.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}
