package x

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestClientApplyHeaders(t *testing.T) {
	var capturedReq *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GraphQL(context.Background(), "hash123", "TestOp", map[string]any{}, DefaultFeatures)
	if err != nil {
		t.Fatalf("GraphQL error: %v", err)
	}

	if capturedReq == nil {
		t.Fatal("no request captured")
	}

	checks := map[string]string{
		"Authorization":           "Bearer " + xBearerToken,
		"X-CSRF-Token":            "test-csrf-token",
		"X-Twitter-Auth-Type":     "OAuth2Session",
		"X-Twitter-Active-User":   "yes",
		"X-Twitter-Client-Language": "en",
		"User-Agent":              "TestAgent/1.0",
	}
	for header, want := range checks {
		got := capturedReq.Header.Get(header)
		if got != want {
			t.Errorf("header %s: got %q, want %q", header, got, want)
		}
	}

	cookie := capturedReq.Header.Get("Cookie")
	if !strings.Contains(cookie, "auth_token=test-auth-token") {
		t.Errorf("expected auth_token in Cookie header, got: %s", cookie)
	}
	if !strings.Contains(cookie, "ct0=test-csrf-token") {
		t.Errorf("expected ct0 in Cookie header, got: %s", cookie)
	}
}

func TestClientCaptureResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Set-Cookie", "ct0=new-csrf-token; Path=/; Secure")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GraphQL(context.Background(), "hash123", "TestOp", map[string]any{}, DefaultFeatures)
	if err != nil {
		t.Fatalf("GraphQL error: %v", err)
	}

	client.mu.Lock()
	csrf := client.csrf
	client.mu.Unlock()

	if csrf != "new-csrf-token" {
		t.Errorf("expected CSRF to rotate to 'new-csrf-token', got %q", csrf)
	}
}

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"key": "value"}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	params := url.Values{"foo": {"bar"}}
	resp, err := client.Get(context.Background(), "/test/path", params)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "field=value") {
			t.Errorf("expected form body, got: %s", body)
		}
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	body := url.Values{"field": {"value"}}
	resp, err := client.Post(context.Background(), "/test/path", body)
	if err != nil {
		t.Fatalf("Post error: %v", err)
	}
	defer resp.Body.Close()
}

func TestClientPostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["key"] != "val" {
			t.Errorf("expected key=val in body, got: %v", body)
		}
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.PostJSON(context.Background(), "/test/path", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("PostJSON error: %v", err)
	}
	defer resp.Body.Close()
}

func TestClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Delete(context.Background(), "/test/path")
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestClientDecodeJSON_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"value": 42}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}

	var result struct {
		Value int `json:"value"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		t.Fatalf("DecodeJSON error: %v", err)
	}
	if result.Value != 42 {
		t.Errorf("expected value=42, got %d", result.Value)
	}
}

func TestClientDecodeJSON_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}

	var target map[string]any
	err = client.DecodeJSON(resp, &target)
	if err == nil {
		t.Fatal("expected rate limit error, got nil")
	}

	rlErr, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T: %v", err, err)
	}
	if rlErr.RetryAfter != 30*time.Second {
		t.Errorf("expected RetryAfter=30s, got %v", rlErr.RetryAfter)
	}
}

func TestClientDecodeJSON_Error4xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"errors": [{"message": "Forbidden", "code": 403}]}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}

	var target map[string]any
	err = client.DecodeJSON(resp, &target)
	if err == nil {
		t.Fatal("expected error for 403, got nil")
	}
	if !strings.Contains(err.Error(), "Forbidden") {
		t.Errorf("expected error to contain 'Forbidden', got: %v", err)
	}
}

func TestClientDecodeJSON_AccountLocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"errors": [{"message": "Account locked", "code": 326}]}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}

	var target map[string]any
	err = client.DecodeJSON(resp, &target)
	if err == nil {
		t.Fatal("expected account locked error, got nil")
	}

	_, ok := err.(*AccountLockedError)
	if !ok {
		t.Fatalf("expected *AccountLockedError, got %T: %v", err, err)
	}
}

func TestClientGraphQL_ErrorInEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errors": [{"message": "Something went wrong"}]}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GraphQL(context.Background(), "hash123", "TestOp", map[string]any{}, DefaultFeatures)
	if err == nil {
		t.Fatal("expected error from GraphQL errors envelope, got nil")
	}
	if !strings.Contains(err.Error(), "Something went wrong") {
		t.Errorf("expected error message in error, got: %v", err)
	}
}

func TestClientGraphQLPost_ErrorInEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errors": [{"message": "POST failed"}]}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GraphQLPost(context.Background(), "hash123", "TestOp", map[string]any{}, DefaultFeatures)
	if err == nil {
		t.Fatal("expected error from GraphQLPost errors envelope, got nil")
	}
	if !strings.Contains(err.Error(), "POST failed") {
		t.Errorf("expected 'POST failed' in error, got: %v", err)
	}
}

func TestClientGraphQL_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GraphQL(context.Background(), "hash", "Op", map[string]any{}, DefaultFeatures)
	if err == nil {
		t.Fatal("expected rate limit error, got nil")
	}
	rlErr, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}
	if rlErr.RetryAfter != 60*time.Second {
		t.Errorf("expected 60s retry, got %v", rlErr.RetryAfter)
	}
}

func TestRateLimitError_WithRetryAfter(t *testing.T) {
	e := &RateLimitError{RetryAfter: 30 * time.Second}
	if !strings.Contains(e.Error(), "30s") {
		t.Errorf("expected retry after duration in error, got: %s", e.Error())
	}
}

func TestAccountLockedError_NoMessage(t *testing.T) {
	e := &AccountLockedError{}
	if e.Error() != "x account locked" {
		t.Errorf("unexpected error string: %s", e.Error())
	}
}
