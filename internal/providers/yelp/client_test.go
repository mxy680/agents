package yelp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestClientDoYelpBearerAuthHeader(t *testing.T) {
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"businesses":[],"total":0}`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL, "my-test-api-key")
	_, err := client.doYelp(context.Background(), "GET", "/businesses/search", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAuth != "Bearer my-test-api-key" {
		t.Errorf("Authorization header = %q, want %q", capturedAuth, "Bearer my-test-api-key")
	}
}

func TestClientDoYelpAcceptHeader(t *testing.T) {
	var capturedAccept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL, "key")
	_, err := client.doYelp(context.Background(), "GET", "/categories", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAccept != "application/json" {
		t.Errorf("Accept header = %q, want %q", capturedAccept, "application/json")
	}
}

func TestClientDoYelpRateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"code":"TOO_MANY_REQUESTS","description":"Rate limit exceeded"}}`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL, "key")
	_, err := client.doYelp(context.Background(), "GET", "/businesses/search", nil)
	if err == nil {
		t.Fatal("expected error for 429 response")
	}

	// Should be a RateLimitError
	if _, ok := err.(*RateLimitError); !ok {
		t.Errorf("expected *RateLimitError, got %T: %v", err, err)
	}

	if err.Error() != "yelp rate limit exceeded; try again later" {
		t.Errorf("RateLimitError message = %q", err.Error())
	}
}

func TestClientDoYelpAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"code":"TOKEN_MISSING","description":"An access token must be supplied."}}`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL, "bad-key")
	_, err := client.doYelp(context.Background(), "GET", "/businesses/search", nil)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}

	if !strings.Contains(err.Error(), "TOKEN_MISSING") {
		t.Errorf("error should contain error code, got: %v", err)
	}
	if !strings.Contains(err.Error(), "An access token must be supplied.") {
		t.Errorf("error should contain description, got: %v", err)
	}
}

func TestClientDoYelpHTTPError(t *testing.T) {
	// Server returns 404 with non-JSON body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`not found`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL, "key")
	_, err := client.doYelp(context.Background(), "GET", "/businesses/nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should contain status code, got: %v", err)
	}
}

func TestClientDoYelpQueryParams(t *testing.T) {
	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL, "key")
	params := url.Values{}
	params.Set("location", "San Francisco")
	params.Set("term", "pizza")

	_, err := client.doYelp(context.Background(), "GET", "/businesses/search", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedQuery, "location=San+Francisco") && !strings.Contains(capturedQuery, "location=San%20Francisco") {
		t.Errorf("expected location param in query, got: %s", capturedQuery)
	}
	if !strings.Contains(capturedQuery, "term=pizza") {
		t.Errorf("expected term param in query, got: %s", capturedQuery)
	}
}

func TestDefaultClientFactoryMissingKey(t *testing.T) {
	// Ensure YELP_API_KEY is not set
	t.Setenv("YELP_API_KEY", "")

	factory := DefaultClientFactory()
	_, err := factory(context.Background())
	if err == nil {
		t.Fatal("expected error when YELP_API_KEY is not set")
	}
	if !strings.Contains(err.Error(), "YELP_API_KEY") {
		t.Errorf("error should mention YELP_API_KEY, got: %v", err)
	}
}

func TestDefaultClientFactoryWithKey(t *testing.T) {
	t.Setenv("YELP_API_KEY", "test-key-12345")

	factory := DefaultClientFactory()
	client, err := factory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.apiKey != "test-key-12345" {
		t.Errorf("client.apiKey = %q, want %q", client.apiKey, "test-key-12345")
	}
	if client.base != yelpBaseURL {
		t.Errorf("client.base = %q, want %q", client.base, yelpBaseURL)
	}
}

func TestRateLimitErrorMessage(t *testing.T) {
	e := &RateLimitError{}
	want := "yelp rate limit exceeded; try again later"
	if e.Error() != want {
		t.Errorf("RateLimitError.Error() = %q, want %q", e.Error(), want)
	}
}

