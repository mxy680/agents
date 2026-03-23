package zillow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultClientFactory(t *testing.T) {
	factory := DefaultClientFactory()
	if factory == nil {
		t.Fatal("DefaultClientFactory() returned nil")
	}

	client, err := factory(context.Background())
	if err != nil {
		t.Fatalf("factory() error: %v", err)
	}
	if client == nil {
		t.Fatal("factory() returned nil client")
	}
	if client.baseURL != zillowBaseURL {
		t.Errorf("expected baseURL %q, got %q", zillowBaseURL, client.baseURL)
	}
	if client.mortURL != mortgageAPIURL {
		t.Errorf("expected mortURL %q, got %q", mortgageAPIURL, client.mortURL)
	}
	if client.userAgent == "" {
		t.Error("expected non-empty userAgent")
	}
}

func TestClientGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("User-Agent") == "" {
				t.Error("expected non-empty User-Agent header")
			}
			if r.Header.Get("Accept") != "application/json" {
				t.Errorf("expected Accept: application/json, got %s", r.Header.Get("Accept"))
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok":true}`))
		}))
		defer server.Close()

		client := newClientWithBase(server.Client(), server.URL)
		body, err := client.Get(context.Background(), server.URL+"/test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(body), "ok") {
			t.Errorf("unexpected body: %s", body)
		}
	})

	t.Run("429_rate_limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := newClientWithBase(server.Client(), server.URL)
		_, err := client.Get(context.Background(), server.URL+"/test")
		if err == nil {
			t.Fatal("expected error for 429, got nil")
		}
		if _, ok := err.(*RateLimitError); !ok {
			t.Errorf("expected *RateLimitError, got %T: %v", err, err)
		}
		if !strings.Contains(err.Error(), "rate limit") {
			t.Errorf("expected rate limit message, got: %s", err.Error())
		}
	})

	t.Run("403_blocked", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "access denied", http.StatusForbidden)
		}))
		defer server.Close()

		client := newClientWithBase(server.Client(), server.URL)
		_, err := client.Get(context.Background(), server.URL+"/test")
		if err == nil {
			t.Fatal("expected error for 403, got nil")
		}
		blockedErr, ok := err.(*BlockedError)
		if !ok {
			t.Errorf("expected *BlockedError, got %T: %v", err, err)
		} else {
			if blockedErr.StatusCode != http.StatusForbidden {
				t.Errorf("expected status 403, got %d", blockedErr.StatusCode)
			}
			if !strings.Contains(blockedErr.Error(), "blocked") {
				t.Errorf("expected 'blocked' in error message, got: %s", blockedErr.Error())
			}
		}
	})

	t.Run("500_error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := newClientWithBase(server.Client(), server.URL)
		_, err := client.Get(context.Background(), server.URL+"/test")
		if err == nil {
			t.Fatal("expected error for 500, got nil")
		}
		if !strings.Contains(err.Error(), "http 500") {
			t.Errorf("expected 'http 500' in error, got: %s", err.Error())
		}
	})
}

func TestClientPutJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		client := newClientWithBase(server.Client(), server.URL)
		payload := map[string]any{"key": "value"}
		body, err := client.PutJSON(context.Background(), server.URL+"/test", payload)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(body), "status") {
			t.Errorf("unexpected body: %s", body)
		}
	})

	t.Run("429_rate_limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := newClientWithBase(server.Client(), server.URL)
		_, err := client.PutJSON(context.Background(), server.URL+"/test", map[string]any{})
		if err == nil {
			t.Fatal("expected error for 429, got nil")
		}
		if _, ok := err.(*RateLimitError); !ok {
			t.Errorf("expected *RateLimitError, got %T", err)
		}
	})

	t.Run("403_blocked", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "forbidden", http.StatusForbidden)
		}))
		defer server.Close()

		client := newClientWithBase(server.Client(), server.URL)
		_, err := client.PutJSON(context.Background(), server.URL+"/test", map[string]any{})
		if err == nil {
			t.Fatal("expected error for 403, got nil")
		}
		if _, ok := err.(*BlockedError); !ok {
			t.Errorf("expected *BlockedError, got %T", err)
		}
	})
}

func TestNewClientWithBase(t *testing.T) {
	httpClient := &http.Client{}
	client := newClientWithBase(httpClient, "http://localhost:9999")

	if client.baseURL != "http://localhost:9999" {
		t.Errorf("expected baseURL 'http://localhost:9999', got %q", client.baseURL)
	}
	if client.staticURL != "http://localhost:9999" {
		t.Errorf("expected staticURL 'http://localhost:9999', got %q", client.staticURL)
	}
	if client.mortURL != "http://localhost:9999" {
		t.Errorf("expected mortURL 'http://localhost:9999', got %q", client.mortURL)
	}
	if client.userAgent != "TestAgent/1.0" {
		t.Errorf("expected userAgent 'TestAgent/1.0', got %q", client.userAgent)
	}
	if client.http != httpClient {
		t.Error("expected http client to be the one passed in")
	}
}

func TestApplyHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all required headers are set
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent header")
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept: application/json, got %q", r.Header.Get("Accept"))
		}
		if r.Header.Get("Accept-Language") == "" {
			t.Error("missing Accept-Language header")
		}
		if r.Header.Get("Referer") == "" {
			t.Error("missing Referer header")
		}
		if r.Header.Get("Origin") == "" {
			t.Error("missing Origin header")
		}
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newClientWithBase(server.Client(), server.URL)
	_, err := client.Get(context.Background(), server.URL+"/headers-check")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTruncateBody(t *testing.T) {
	t.Run("short_body", func(t *testing.T) {
		body := []byte("short error message")
		result := truncateBody(body)
		if result != "short error message" {
			t.Errorf("expected unchanged body, got %q", result)
		}
	})

	t.Run("long_body_truncated", func(t *testing.T) {
		body := make([]byte, 300)
		for i := range body {
			body[i] = 'x'
		}
		result := truncateBody(body)
		if len(result) != 203 { // 200 chars + "..."
			t.Errorf("expected 203 chars, got %d: %q", len(result), result)
		}
		if !strings.HasSuffix(result, "...") {
			t.Errorf("expected '...' suffix, got %q", result)
		}
	})

	t.Run("exactly_200_chars", func(t *testing.T) {
		body := make([]byte, 200)
		for i := range body {
			body[i] = 'a'
		}
		result := truncateBody(body)
		if len(result) != 200 {
			t.Errorf("expected 200 chars, got %d", len(result))
		}
		if strings.HasSuffix(result, "...") {
			t.Errorf("should not truncate exactly 200 chars")
		}
	})
}
