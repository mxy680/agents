package yelp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestClientDoYelpCookieHeader(t *testing.T) {
	var capturedCookie string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCookie = r.Header.Get("Cookie")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.doYelp(context.Background(), "GET", "/search/snippet", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedCookie, "bse=test-bse") {
		t.Errorf("Cookie header should contain bse, got: %q", capturedCookie)
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

	client := newClientWithBase(server.Client(), server.URL)
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
		w.Write([]byte(`rate limited`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.doYelp(context.Background(), "GET", "/search/snippet", nil)
	if err == nil {
		t.Fatal("expected error for 429 response")
	}

	if _, ok := err.(*RateLimitError); !ok {
		t.Errorf("expected *RateLimitError, got %T: %v", err, err)
	}
}

func TestClientDoYelpHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`not found`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.doYelp(context.Background(), "GET", "/biz/nonexistent", nil)
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

	client := newClientWithBase(server.Client(), server.URL)
	params := url.Values{}
	params.Set("find_loc", "San Francisco")
	params.Set("find_desc", "pizza")

	_, err := client.doYelp(context.Background(), "GET", "/search/snippet", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedQuery, "find_loc=San+Francisco") && !strings.Contains(capturedQuery, "find_loc=San%20Francisco") {
		t.Errorf("expected find_loc param in query, got: %s", capturedQuery)
	}
	if !strings.Contains(capturedQuery, "find_desc=pizza") {
		t.Errorf("expected find_desc param in query, got: %s", capturedQuery)
	}
}

func TestDefaultClientFactoryMissingBSE(t *testing.T) {
	t.Setenv("YELP_BSE", "")

	factory := DefaultClientFactory()
	_, err := factory(context.Background())
	if err == nil {
		t.Fatal("expected error when YELP_BSE is not set")
	}
	if !strings.Contains(err.Error(), "YELP_BSE") {
		t.Errorf("error should mention YELP_BSE, got: %v", err)
	}
}

func TestDefaultClientFactoryWithSession(t *testing.T) {
	t.Setenv("YELP_BSE", "test-bse-value")
	t.Setenv("YELP_ZSS", "test-zss-value")
	t.Setenv("YELP_CSRF_TOKEN", "test-csrf")

	factory := DefaultClientFactory()
	client, err := factory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.base != yelpBaseURL {
		t.Errorf("client.base = %q, want %q", client.base, yelpBaseURL)
	}
}

func TestClientCSRFRotation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "csrftok", Value: "new-csrf-value"})
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.doYelp(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.getCSRF() != "new-csrf-value" {
		t.Errorf("CSRF not rotated: got %q, want %q", client.getCSRF(), "new-csrf-value")
	}
}

func TestRateLimitErrorMessage(t *testing.T) {
	e := &RateLimitError{}
	want := "yelp rate limit exceeded; try again later"
	if e.Error() != want {
		t.Errorf("RateLimitError.Error() = %q, want %q", e.Error(), want)
	}
}

