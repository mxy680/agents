package canvas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientGet(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := newTestClient(server)
	body, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !strings.Contains(string(body), "true") {
		t.Errorf("expected response body to contain 'true', got: %s", string(body))
	}
}

func TestClientPost(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"created": req["name"]})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := newTestClient(server)
	body, err := client.Post(context.Background(), "/test", map[string]any{"name": "test-resource"})
	if err != nil {
		t.Fatalf("Post returned error: %v", err)
	}
	if !strings.Contains(string(body), "test-resource") {
		t.Errorf("expected response to echo name, got: %s", string(body))
	}
}

func TestClientDelete(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/test/42", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"deleted": true})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := newTestClient(server)
	body, err := client.Delete(context.Background(), "/test/42")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if !strings.Contains(string(body), "deleted") {
		t.Errorf("expected response to confirm deletion, got: %s", string(body))
	}
}

func TestClientApplyHeaders(t *testing.T) {
	var capturedCookie, capturedAccept, capturedUserAgent, capturedCSRF string

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		capturedCookie = r.Header.Get("Cookie")
		capturedAccept = r.Header.Get("Accept")
		capturedUserAgent = r.Header.Get("User-Agent")
		capturedCSRF = r.Header.Get("X-CSRF-Token")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := newTestClient(server)

	// Test GET (should not send X-CSRF-Token).
	_, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if capturedAccept != "application/json" {
		t.Errorf("expected Accept: application/json, got: %q", capturedAccept)
	}
	if capturedUserAgent != "TestAgent/1.0" {
		t.Errorf("expected User-Agent TestAgent/1.0, got: %q", capturedUserAgent)
	}
	if capturedCSRF != "" {
		t.Errorf("GET should not send X-CSRF-Token header, got: %q", capturedCSRF)
	}
	if !strings.Contains(capturedCookie, "test-session-cookie") {
		t.Errorf("expected session cookie in Cookie header, got: %q", capturedCookie)
	}
}

func TestClientCaptureResponseHeaders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Set-Cookie", "_csrf_token=new-csrf-value; Path=/; HttpOnly")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	client.mu.Lock()
	csrf := client.csrf
	client.mu.Unlock()

	if csrf != "new-csrf-value" {
		t.Errorf("expected CSRF to be updated from Set-Cookie, got: %q", csrf)
	}
}

func TestParseLinkHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "next link present",
			header:   `<https://canvas.edu/api/v1/courses?page=2&per_page=10>; rel="next", <https://canvas.edu/api/v1/courses?page=1&per_page=10>; rel="first"`,
			expected: "https://canvas.edu/api/v1/courses?page=2&per_page=10",
		},
		{
			name:     "no next link",
			header:   `<https://canvas.edu/api/v1/courses?page=1&per_page=10>; rel="first", <https://canvas.edu/api/v1/courses?page=5&per_page=10>; rel="last"`,
			expected: "",
		},
		{
			name:     "empty header",
			header:   "",
			expected: "",
		},
		{
			name:     "only next link",
			header:   `<https://canvas.edu/api/v1/items?page=3>; rel="next"`,
			expected: "https://canvas.edu/api/v1/items?page=3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseLinkHeader(tc.header)
			if result != tc.expected {
				t.Errorf("ParseLinkHeader(%q) = %q, want %q", tc.header, result, tc.expected)
			}
		})
	}
}

func TestRateLimitError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Rate-Limit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"rate limit exceeded"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error for rate limit response")
	}

	rateLimitErr, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T: %v", err, err)
	}
	if !strings.Contains(rateLimitErr.Error(), "rate limit") {
		t.Errorf("error message should mention rate limit, got: %v", rateLimitErr)
	}
}

func TestClientHTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention HTTP 404, got: %v", err)
	}
}
