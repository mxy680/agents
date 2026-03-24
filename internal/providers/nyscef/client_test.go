package nyscef

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitError(t *testing.T) {
	e := &RateLimitError{}
	if e.Error() == "" {
		t.Error("RateLimitError.Error() returned empty string")
	}
	if got := e.Error(); got != "nyscef rate limit exceeded; try again later" {
		t.Errorf("RateLimitError.Error() = %q", got)
	}
}

func TestBlockedError(t *testing.T) {
	e := &BlockedError{StatusCode: 403, Body: "forbidden"}
	msg := e.Error()
	if msg == "" {
		t.Error("BlockedError.Error() returned empty string")
	}
	if got := e.Error(); got == "" {
		t.Error("BlockedError.Error() should not be empty")
	}
}

func TestGetRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.getWithStdHTTP(context.Background(), server.URL+"/test")
	if err == nil {
		t.Fatal("expected error for 429, got nil")
	}
	if _, ok := err.(*RateLimitError); !ok {
		t.Errorf("expected RateLimitError, got %T: %v", err, err)
	}
}

func TestGetBlocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("access denied"))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.getWithStdHTTP(context.Background(), server.URL+"/test")
	if err == nil {
		t.Fatal("expected error for 403, got nil")
	}
	if _, ok := err.(*BlockedError); !ok {
		t.Errorf("expected BlockedError, got %T: %v", err, err)
	}
}

func TestPostRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.postWithStdHTTP(context.Background(), server.URL+"/test", nil)
	if err == nil {
		t.Fatal("expected error for 429, got nil")
	}
	if _, ok := err.(*RateLimitError); !ok {
		t.Errorf("expected RateLimitError, got %T: %v", err, err)
	}
}

func TestPostBlocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("blocked"))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.postWithStdHTTP(context.Background(), server.URL+"/test", nil)
	if err == nil {
		t.Fatal("expected error for 403, got nil")
	}
	if _, ok := err.(*BlockedError); !ok {
		t.Errorf("expected BlockedError, got %T: %v", err, err)
	}
}

func TestGetHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.getWithStdHTTP(context.Background(), server.URL+"/test")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestPostHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.postWithStdHTTP(context.Background(), server.URL+"/test", nil)
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestGetSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html>ok</html>"))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	body, err := client.getWithStdHTTP(context.Background(), server.URL+"/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != "<html>ok</html>" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestPostSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html>results</html>"))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	body, err := client.postWithStdHTTP(context.Background(), server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != "<html>results</html>" {
		t.Errorf("unexpected body: %q", body)
	}
}
