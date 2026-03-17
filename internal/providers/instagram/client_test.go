package instagram

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/emdash-projects/agents/internal/auth"
)

// newTestSession returns a minimal InstagramSession for use in tests.
func newTestSession() *auth.InstagramSession {
	return &auth.InstagramSession{
		SessionID: "test-session-id",
		CSRFToken: "test-csrf-token",
		DSUserID:  "123456789",
		Mid:       "test-mid",
		IgDid:     "test-ig-did",
		UserAgent: "TestAgent/1.0",
	}
}

// newTestClient creates a Client pointing at the given test server.
func newTestClient(server *httptest.Server) *Client {
	return newClientWithBase(newTestSession(), server.Client(), server.URL)
}

// TestGet_HeaderInjection verifies that all required headers are sent with GET requests.
func TestGet_HeaderInjection(t *testing.T) {
	var capturedReq *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/api/v1/test", url.Values{"foo": {"bar"}})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	assertHeader(t, capturedReq, "User-Agent", "TestAgent/1.0")
	assertHeader(t, capturedReq, "X-Ig-App-Id", igAppID)
	assertHeader(t, capturedReq, "X-Csrftoken", "test-csrf-token")
	assertHeader(t, capturedReq, "X-Instagram-Ajax", igAjaxToken)
	assertHeader(t, capturedReq, "X-Asbd-Id", igASBDID)
	assertHeader(t, capturedReq, "X-Requested-With", "XMLHttpRequest")

	// Cookie header must contain all session cookie values.
	cookie := capturedReq.Header.Get("Cookie")
	for _, want := range []string{"sessionid=test-session-id", "csrftoken=test-csrf-token", "ds_user_id=123456789", "mid=test-mid", "ig_did=test-ig-did"} {
		if !containsStr(cookie, want) {
			t.Errorf("Cookie header %q missing %q", cookie, want)
		}
	}

	// Query param should be forwarded.
	if q := capturedReq.URL.Query().Get("foo"); q != "bar" {
		t.Errorf("query param foo = %q, want bar", q)
	}
}

// TestPost_HeaderInjection verifies that POST requests have correct headers and body encoding.
func TestPost_HeaderInjection(t *testing.T) {
	var capturedReq *http.Request
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		capturedBody, _ = readBody(r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	params := url.Values{"key": {"value"}}
	resp, err := client.Post(context.Background(), "/api/v1/action", params)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	defer resp.Body.Close()

	assertHeader(t, capturedReq, "Content-Type", "application/x-www-form-urlencoded")
	assertHeader(t, capturedReq, "X-Csrftoken", "test-csrf-token")

	if !containsStr(string(capturedBody), "key=value") {
		t.Errorf("POST body %q missing form param key=value", capturedBody)
	}
}

// TestCSRFRotation verifies that a new csrftoken in Set-Cookie is captured and used
// in the next request.
func TestCSRFRotation(t *testing.T) {
	callCount := 0
	var secondCSRF string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First request: rotate CSRF via Set-Cookie.
			w.Header().Set("Set-Cookie", "csrftoken=new-csrf-xyz; Path=/")
			w.Write([]byte(`{}`))
			return
		}
		// Second request: capture the CSRF token used.
		secondCSRF = r.Header.Get("X-Csrftoken")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)

	resp, err := client.Get(context.Background(), "/first", nil)
	if err != nil {
		t.Fatalf("first Get: %v", err)
	}
	resp.Body.Close()

	resp, err = client.Get(context.Background(), "/second", nil)
	if err != nil {
		t.Fatalf("second Get: %v", err)
	}
	resp.Body.Close()

	if secondCSRF != "new-csrf-xyz" {
		t.Errorf("rotated CSRF = %q, want %q", secondCSRF, "new-csrf-xyz")
	}
}

// TestWWWClaimTracking verifies that the X-IG-Set-WWW-Claim response header is
// captured and forwarded in subsequent requests.
func TestWWWClaimTracking(t *testing.T) {
	callCount := 0
	var secondClaim string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("X-IG-Set-WWW-Claim", "claim-token-abc")
			w.Write([]byte(`{}`))
			return
		}
		secondClaim = r.Header.Get("X-Ig-Www-Claim")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)

	resp, _ := client.Get(context.Background(), "/first", nil)
	resp.Body.Close()

	resp, _ = client.Get(context.Background(), "/second", nil)
	resp.Body.Close()

	if secondClaim != "claim-token-abc" {
		t.Errorf("www-claim = %q, want %q", secondClaim, "claim-token-abc")
	}
}

// TestDecodeJSON_Success verifies successful JSON decoding.
func TestDecodeJSON_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "value": "hello"})
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/data", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	var result map[string]string
	if err := client.DecodeJSON(resp, &result); err != nil {
		t.Fatalf("DecodeJSON: %v", err)
	}
	if result["value"] != "hello" {
		t.Errorf("decoded value = %q, want %q", result["value"], "hello")
	}
}

// TestDecodeJSON_RateLimit verifies that a 429 response returns RateLimitError.
func TestDecodeJSON_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/rate-limited", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	var result map[string]any
	err = client.DecodeJSON(resp, &result)
	if err == nil {
		t.Fatal("expected RateLimitError, got nil")
	}

	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Errorf("expected *RateLimitError, got %T: %v", err, err)
	}
	if rlErr.RetryAfter != 60*time.Second {
		t.Errorf("RetryAfter = %v, want 60s", rlErr.RetryAfter)
	}
}

// TestDecodeJSON_RateLimit_NoRetryAfter verifies RateLimitError without Retry-After header.
func TestDecodeJSON_RateLimit_NoRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/rate-limited", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	err = client.DecodeJSON(resp, nil)

	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Errorf("expected *RateLimitError, got %T: %v", err, err)
	}
	if rlErr.RetryAfter != 0 {
		t.Errorf("RetryAfter = %v, want 0", rlErr.RetryAfter)
	}
}

// TestHandleError_ChallengeRequired verifies that challenge_required errors are typed.
func TestHandleError_ChallengeRequired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "fail",
			"message": "challenge_required",
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/protected", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	var result map[string]any
	err = client.DecodeJSON(resp, &result)
	if err == nil {
		t.Fatal("expected ChallengeRequiredError, got nil")
	}

	var challengeErr *ChallengeRequiredError
	if !errors.As(err, &challengeErr) {
		t.Errorf("expected *ChallengeRequiredError, got %T: %v", err, err)
	}
}

// TestHandleError_ChallengeRequired_CheckpointURL verifies detection via checkpoint_url.
func TestHandleError_ChallengeRequired_CheckpointURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"status":         "fail",
			"message":        "blocked",
			"checkpoint_url": "/challenge/",
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/protected", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	var result map[string]any
	err = client.DecodeJSON(resp, &result)

	var challengeErr *ChallengeRequiredError
	if !errors.As(err, &challengeErr) {
		t.Errorf("expected *ChallengeRequiredError, got %T: %v", err, err)
	}
}

// TestHandleError_GenericAPIError verifies that non-challenge API errors are wrapped.
func TestHandleError_GenericAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "fail",
			"message": "media not found",
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	resp, err := client.Get(context.Background(), "/missing", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	var result map[string]any
	err = client.DecodeJSON(resp, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if containsStr(err.Error(), "media not found") == false {
		t.Errorf("error %q should mention 'media not found'", err.Error())
	}
}

// TestPostGraphQL_Success verifies GraphQL query execution and data extraction.
func TestPostGraphQL_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql/query" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		var body graphQLRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.DocID != "doc123" {
			t.Errorf("doc_id = %q, want doc123", body.DocID)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"data":   map[string]string{"result": "success"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	data, err := client.PostGraphQL(context.Background(), "TestQuery", map[string]any{"id": "123"}, "doc123")
	if err != nil {
		t.Fatalf("PostGraphQL: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if result["result"] != "success" {
		t.Errorf("result = %q, want success", result["result"])
	}
}

// TestPostGraphQL_GraphQLError verifies that GraphQL-level errors are returned.
func TestPostGraphQL_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": nil,
			"errors": []map[string]string{
				{"message": "not authorized"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.PostGraphQL(context.Background(), "TestQuery", nil, "doc456")
	if err == nil {
		t.Fatal("expected error for graphql errors, got nil")
	}
	if !containsStr(err.Error(), "not authorized") {
		t.Errorf("error %q should mention 'not authorized'", err.Error())
	}
}

// TestDefaultClientFactory_MissingEnv verifies that DefaultClientFactory returns an error
// when required environment variables are absent.
func TestDefaultClientFactory_MissingEnv(t *testing.T) {
	t.Setenv("INSTAGRAM_SESSION_ID", "")
	t.Setenv("INSTAGRAM_CSRF_TOKEN", "")
	t.Setenv("INSTAGRAM_DS_USER_ID", "")

	factory := DefaultClientFactory()
	_, err := factory(context.Background())
	if err == nil {
		t.Fatal("expected error for missing env vars, got nil")
	}
}

// --- helpers ---

func assertHeader(t *testing.T, r *http.Request, key, want string) {
	t.Helper()
	got := r.Header.Get(key)
	if got != want {
		t.Errorf("header %q = %q, want %q", key, got, want)
	}
}

func readBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	var buf []byte
	tmp := make([]byte, 512)
	for {
		n, err := r.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return buf, nil
}

func containsStr(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
