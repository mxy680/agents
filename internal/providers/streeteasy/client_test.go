package streeteasy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRateLimitError(t *testing.T) {
	e := &RateLimitError{}
	if !strings.Contains(e.Error(), "rate limit") {
		t.Errorf("RateLimitError.Error() = %q, expected 'rate limit'", e.Error())
	}
}

func TestBlockedError(t *testing.T) {
	e := &BlockedError{StatusCode: 403, Body: "forbidden"}
	msg := e.Error()
	if !strings.Contains(msg, "403") {
		t.Errorf("BlockedError.Error() = %q, expected '403'", msg)
	}
	if !strings.Contains(msg, "Playwright") {
		t.Errorf("BlockedError.Error() = %q, expected mention of Playwright", msg)
	}
}

func TestClientGet_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.Get(context.Background(), server.URL+"/test")
	if err == nil {
		t.Fatal("expected rate limit error")
	}
	if _, ok := err.(*RateLimitError); !ok {
		t.Errorf("expected *RateLimitError, got %T: %v", err, err)
	}
}

func TestClientGet_Blocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.Get(context.Background(), server.URL+"/test")
	if err == nil {
		t.Fatal("expected blocked error")
	}
	if _, ok := err.(*BlockedError); !ok {
		t.Errorf("expected *BlockedError, got %T: %v", err, err)
	}
}

func TestClientGet_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.Get(context.Background(), server.URL+"/test")
	if err == nil {
		t.Fatal("expected error for 5xx status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected '500' in error, got: %v", err)
	}
}

func TestClientGet_WithCookies(t *testing.T) {
	var capturedCookie string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCookie = r.Header.Get("Cookie")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	client.cookies = "_px3=abc123; SE_VISITOR_ID=xyz"

	_, err := client.Get(context.Background(), server.URL+"/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedCookie, "_px3=abc123") {
		t.Errorf("expected cookie header to include _px3, got: %s", capturedCookie)
	}
}

func TestClientGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html>OK</html>`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	body, err := client.Get(context.Background(), server.URL+"/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `<html>OK</html>` {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestClientGet_InvalidURL(t *testing.T) {
	client := newClientWithBase(http.DefaultClient, "http://localhost")
	_, err := client.Get(context.Background(), "://invalid-url")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestDefaultClientFactory(t *testing.T) {
	factory := DefaultClientFactory()
	if factory == nil {
		t.Fatal("DefaultClientFactory() returned nil")
	}
	client, err := factory(context.Background())
	if err != nil {
		t.Fatalf("factory(ctx) error: %v", err)
	}
	if client == nil {
		t.Fatal("factory returned nil client")
	}
	if client.baseURL != streeteasyBaseURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, streeteasyBaseURL)
	}
	if client.userAgent == "" {
		t.Error("userAgent should not be empty")
	}
}
