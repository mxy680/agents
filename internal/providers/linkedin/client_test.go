package linkedin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGet_HeaderInjection(t *testing.T) {
	var capturedReq *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/voyager/api/test", url.Values{"foo": {"bar"}})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	assertHeader(t, capturedReq, "User-Agent", "TestAgent/1.0")
	assertHeader(t, capturedReq, "Csrf-Token", "ajax:test-jsessionid")
	assertHeader(t, capturedReq, "X-Li-Lang", "en_US")
	assertHeader(t, capturedReq, "X-Restli-Protocol-Version", restliProtocolVersion)
	assertHeader(t, capturedReq, "X-Requested-With", "XMLHttpRequest")

	cookie := capturedReq.Header.Get("Cookie")
	for _, want := range []string{"li_at=test-li-at", "JSESSIONID="} {
		if !containsStr(cookie, want) {
			t.Errorf("Cookie header missing %q, got %q", want, cookie)
		}
	}
}

func TestGet_QueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "test" {
			t.Errorf("query param q=%q, want %q", got, "test")
		}
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/test", url.Values{"q": {"test"}})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	resp.Body.Close()
}

func TestPost_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.PostJSON(context.Background(), "/test", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("PostJSON: %v", err)
	}
	resp.Body.Close()
}

func TestDecodeJSON_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"name":"test"}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, _ := client.Get(context.Background(), "/test", nil)

	var result struct{ Name string }
	if err := client.DecodeJSON(resp, &result); err != nil {
		t.Fatalf("DecodeJSON: %v", err)
	}
	if result.Name != "test" {
		t.Errorf("Name = %q, want %q", result.Name, "test")
	}
}

func TestDecodeJSON_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, _ := client.Get(context.Background(), "/test", nil)

	var result struct{}
	err := client.DecodeJSON(resp, &result)
	if err == nil {
		t.Fatal("expected rate limit error")
	}
	rl, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}
	if rl.RetryAfter.Seconds() != 60 {
		t.Errorf("RetryAfter = %v, want 60s", rl.RetryAfter)
	}
}

func TestDecodeJSON_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"status":403,"message":"Not authorized"}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, _ := client.Get(context.Background(), "/test", nil)

	var result struct{}
	err := client.DecodeJSON(resp, &result)
	if err == nil {
		t.Fatal("expected API error")
	}
	if !containsStr(err.Error(), "Not authorized") {
		t.Errorf("error should mention 'Not authorized', got %q", err.Error())
	}
}

func TestCSRFRotation(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Set-Cookie", `JSESSIONID="ajax:new-csrf-value"; Path=/; Secure`)
		}
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)

	// First request — server sends new JSESSIONID
	resp1, _ := client.Get(context.Background(), "/test1", nil)
	resp1.Body.Close()

	// Verify CSRF was rotated
	client.mu.Lock()
	csrf := client.csrf
	client.mu.Unlock()

	if csrf != "ajax:new-csrf-value" {
		t.Errorf("csrf after rotation = %q, want %q", csrf, "ajax:new-csrf-value")
	}
}

func TestDelete(t *testing.T) {
	var capturedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Delete(context.Background(), "/test")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	resp.Body.Close()

	if capturedMethod != "DELETE" {
		t.Errorf("method = %q, want DELETE", capturedMethod)
	}
}

// assertHeader checks that a request header matches the expected value.
func assertHeader(t *testing.T, req *http.Request, key, want string) {
	t.Helper()
	got := req.Header.Get(key)
	if got != want {
		t.Errorf("header %s = %q, want %q", key, got, want)
	}
}
