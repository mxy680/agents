package yelp

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
	return newClientWithBase(server.Client(), server.URL, "test-api-key")
}

// newTestClientFactory returns a ClientFactory pointing at the given test server.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return newTestClient(server), nil
	}
}

// newTestRootCmd creates a root command with --json flag wired up.
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

// newFullMockServer creates an httptest server with all Yelp API mock endpoints.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	withBusinessSearchMock(mux)
	withBusinessPhoneSearchMock(mux)
	withBusinessMatchMock(mux)
	withBusinessGetMock(mux)
	withBusinessReviewsMock(mux)
	withEventSearchMock(mux)
	withEventGetMock(mux)
	withEventFeaturedMock(mux)
	withCategoryListMock(mux)
	withCategoryGetMock(mux)
	withAutocompleteMock(mux)
	withTransactionSearchMock(mux)

	return httptest.NewServer(mux)
}

func sampleBusiness() map[string]any {
	return map[string]any{
		"id":           "gary-danko-san-francisco",
		"alias":        "gary-danko-san-francisco",
		"name":         "Gary Danko",
		"rating":       4.5,
		"review_count": 5682,
		"price":        "$$$$",
		"phone":        "+14152520800",
		"is_closed":    false,
		"distance":     1200.5,
		"url":          "https://www.yelp.com/biz/gary-danko-san-francisco",
		"location": map[string]any{
			"address1":        "800 N Point St",
			"city":            "San Francisco",
			"state":           "CA",
			"zip_code":        "94109",
			"country":         "US",
			"display_address": []string{"800 N Point St", "San Francisco, CA 94109"},
		},
		"categories": []map[string]any{
			{"alias": "newamerican", "title": "American (New)"},
		},
		"coordinates": map[string]any{
			"latitude":  37.8051,
			"longitude": -122.4212,
		},
	}
}

func sampleBusinessDetail() map[string]any {
	b := sampleBusiness()
	b["is_claimed"] = true
	b["photos"] = []string{
		"https://s3-media1.fl.yelpcdn.com/bphoto/photo1.jpg",
		"https://s3-media1.fl.yelpcdn.com/bphoto/photo2.jpg",
	}
	b["hours"] = []map[string]any{
		{
			"hours_type": "REGULAR",
			"is_open_now": true,
			"open": []map[string]any{
				{"day": 0, "start": "1730", "end": "2200", "is_overnight": false},
			},
		},
	}
	return b
}

// withBusinessSearchMock adds the business search endpoint mock.
// The client base URL is server.URL (no /v3 prefix), and paths start with /businesses/...
func withBusinessSearchMock(mux *http.ServeMux) {
	mux.HandleFunc("/businesses/search", func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"total": 1,
			"businesses": []map[string]any{
				sampleBusiness(),
			},
			"region": map[string]any{
				"center": map[string]any{
					"latitude":  37.8051,
					"longitude": -122.4212,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// withBusinessPhoneSearchMock adds the phone search endpoint mock.
func withBusinessPhoneSearchMock(mux *http.ServeMux) {
	mux.HandleFunc("/businesses/search/phone", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"total": 1,
			"businesses": []map[string]any{
				sampleBusiness(),
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// withBusinessMatchMock adds the business match endpoint mock.
func withBusinessMatchMock(mux *http.ServeMux) {
	mux.HandleFunc("/businesses/matches", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"businesses": []map[string]any{
				sampleBusiness(),
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// withBusinessGetMock adds the business detail endpoint mock.
func withBusinessGetMock(mux *http.ServeMux) {
	mux.HandleFunc("/businesses/gary-danko-san-francisco", func(w http.ResponseWriter, r *http.Request) {
		// Return reviews if reviews path
		if strings.HasSuffix(r.URL.Path, "/reviews") {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleBusinessDetail())
	})
}

// withBusinessReviewsMock adds the business reviews endpoint mock.
func withBusinessReviewsMock(mux *http.ServeMux) {
	mux.HandleFunc("/businesses/gary-danko-san-francisco/reviews", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"total": 3,
			"reviews": []map[string]any{
				{
					"id":           "xAG4O7l-t1ubbwVAlPnDKg",
					"rating":       5,
					"text":         "Absolutely amazing. The food was outstanding and the service was impeccable.",
					"time_created": "2024-03-15 18:22:03",
					"url":          "https://www.yelp.com/biz/gary-danko-san-francisco?hrid=xAG4",
					"user": map[string]any{
						"name":      "John D.",
						"image_url": "https://s3-media1.fl.yelpcdn.com/photo/john.jpg",
					},
				},
				{
					"id":           "bBBQ8EzInCVBf4o7l2XYZA",
					"rating":       4,
					"text":         "Great experience overall. A bit pricey but worth it.",
					"time_created": "2024-02-28 20:15:00",
					"url":          "https://www.yelp.com/biz/gary-danko-san-francisco?hrid=bBBQ",
					"user": map[string]any{
						"name": "Sarah M.",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// withEventSearchMock adds the events search endpoint mock.
func withEventSearchMock(mux *http.ServeMux) {
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		// Exact path only — featured and specific IDs are handled by their own handlers
		if r.URL.Path != "/events" {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"total": 2,
			"events": []map[string]any{
				{
					"id":              "san-francisco-tech-meetup",
					"name":            "SF Tech Meetup",
					"description":     "Monthly tech meetup in San Francisco.",
					"time_start":      "2024-04-15 18:00:00",
					"time_end":        "2024-04-15 21:00:00",
					"is_free":         true,
					"cost":            0.0,
					"attending_count": 142,
					"event_site_url":  "https://www.meetup.com/sf-tech",
					"location": map[string]any{
						"address1": "123 Market St",
						"city":     "San Francisco",
						"state":    "CA",
						"zip_code": "94105",
						"country":  "US",
					},
				},
				{
					"id":              "san-francisco-food-festival",
					"name":            "SF Food Festival",
					"description":     "Annual food festival with local vendors.",
					"time_start":      "2024-04-20 10:00:00",
					"time_end":        "2024-04-20 22:00:00",
					"is_free":         false,
					"cost":            25.0,
					"attending_count": 500,
					"event_site_url":  "https://sffoodfestival.com",
					"location": map[string]any{
						"address1": "Golden Gate Park",
						"city":     "San Francisco",
						"state":    "CA",
						"zip_code": "94117",
						"country":  "US",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// withEventGetMock adds the event detail endpoint mock.
func withEventGetMock(mux *http.ServeMux) {
	mux.HandleFunc("/events/san-francisco-tech-meetup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		event := map[string]any{
			"id":              "san-francisco-tech-meetup",
			"name":            "SF Tech Meetup",
			"description":     "Monthly tech meetup in San Francisco.",
			"time_start":      "2024-04-15 18:00:00",
			"time_end":        "2024-04-15 21:00:00",
			"is_free":         true,
			"cost":            0.0,
			"attending_count": 142,
			"event_site_url":  "https://www.meetup.com/sf-tech",
			"location": map[string]any{
				"address1": "123 Market St",
				"city":     "San Francisco",
				"state":    "CA",
				"zip_code": "94105",
				"country":  "US",
			},
		}
		json.NewEncoder(w).Encode(event)
	})
}

// withEventFeaturedMock adds the featured event endpoint mock.
func withEventFeaturedMock(mux *http.ServeMux) {
	mux.HandleFunc("/events/featured", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		event := map[string]any{
			"id":              "sf-featured-concert",
			"name":            "SF Symphony Concert",
			"description":     "An evening of classical music.",
			"time_start":      "2024-04-25 19:30:00",
			"time_end":        "2024-04-25 22:00:00",
			"is_free":         false,
			"cost":            75.0,
			"attending_count": 300,
			"event_site_url":  "https://sfsymphony.org",
			"location": map[string]any{
				"address1": "201 Van Ness Ave",
				"city":     "San Francisco",
				"state":    "CA",
				"zip_code": "94102",
				"country":  "US",
			},
		}
		json.NewEncoder(w).Encode(event)
	})
}

// withCategoryListMock adds the categories list endpoint mock.
func withCategoryListMock(mux *http.ServeMux) {
	mux.HandleFunc("/categories", func(w http.ResponseWriter, r *http.Request) {
		// Exact path only — specific aliases handled by withCategoryGetMock
		if r.URL.Path != "/categories" {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"categories": []map[string]any{
				{
					"alias":          "restaurants",
					"title":          "Restaurants",
					"parent_aliases": []string{},
				},
				{
					"alias":          "pizza",
					"title":          "Pizza",
					"parent_aliases": []string{"restaurants"},
				},
				{
					"alias":          "coffee",
					"title":          "Coffee & Tea",
					"parent_aliases": []string{"food"},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// withCategoryGetMock adds the category detail endpoint mock.
func withCategoryGetMock(mux *http.ServeMux) {
	mux.HandleFunc("/categories/pizza", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"category": map[string]any{
				"alias":          "pizza",
				"title":          "Pizza",
				"parent_aliases": []string{"restaurants"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// withAutocompleteMock adds the autocomplete endpoint mock.
func withAutocompleteMock(mux *http.ServeMux) {
	mux.HandleFunc("/autocomplete", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"terms": []map[string]any{
				{"text": "pizza"},
				{"text": "pizza delivery"},
			},
			"businesses": []map[string]any{
				{"id": "gary-danko-san-francisco", "name": "Gary Danko"},
				{"id": "nopa-san-francisco", "name": "Nopa"},
			},
			"categories": []map[string]any{
				{"alias": "pizza", "title": "Pizza"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

// withTransactionSearchMock adds the transactions search endpoint mock.
func withTransactionSearchMock(mux *http.ServeMux) {
	mux.HandleFunc("/transactions/delivery/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"total": 1,
			"businesses": []map[string]any{
				sampleBusiness(),
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}
