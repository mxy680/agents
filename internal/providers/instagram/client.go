package instagram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emdash-projects/agents/internal/auth"
)

const (
	baseURL       = "https://www.instagram.com"
	mobileBaseURL = "https://i.instagram.com"

	// Instagram web app headers.
	igAppID      = "936619743392459"
	igAjaxToken  = "1035277278"
	igASBDID     = "359341"

	// Mobile user agent for i.instagram.com endpoints.
	mobileUserAgent = "Instagram 275.0.0.27.98 Android (33/13; 420dpi; 1080x2400; samsung; SM-S918B; dm3q; qcom; en_US; 458229258)"
)

// RateLimitError is returned when Instagram responds with HTTP 429.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("instagram rate limit exceeded; retry after %s", e.RetryAfter)
	}
	return "instagram rate limit exceeded"
}

// ChallengeRequiredError is returned when Instagram requires a challenge (e.g. 2FA checkpoint).
type ChallengeRequiredError struct {
	Message string
}

func (e *ChallengeRequiredError) Error() string {
	return fmt.Sprintf("instagram challenge required: %s", e.Message)
}

// Client is an HTTP client wrapper for the Instagram web API. It handles
// header injection, cookie attachment, CSRF rotation, and www-claim tracking.
// It supports both web (www.instagram.com) and mobile (i.instagram.com) endpoints.
type Client struct {
	http           *http.Client
	session        *auth.InstagramSession
	baseURL        string
	mobileBase     string // i.instagram.com for mobile-only endpoints

	mu       sync.Mutex
	csrf     string // rotated from Set-Cookie responses
	wwwClaim string // rotated from X-IG-Set-WWW-Claim responses
}

// ClientFactory is the function signature for creating a Client.
// Mirrors the ServiceFactory pattern used by the Google providers.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory that reads real credentials
// from environment variables.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		session, err := auth.NewInstagramSession()
		if err != nil {
			return nil, fmt.Errorf("instagram auth: %w", err)
		}
		return &Client{
			http:       &http.Client{Timeout: 30 * time.Second},
			session:    session,
			baseURL:    baseURL,
			mobileBase: mobileBaseURL,
			csrf:       session.CSRFToken,
			wwwClaim:   "0",
		}, nil
	}
}

// newClientWithBase creates a Client that targets a custom base URL (used in tests).
// Both web and mobile base URLs point to the same test server.
func newClientWithBase(session *auth.InstagramSession, httpClient *http.Client, base string) *Client {
	return &Client{
		http:       httpClient,
		session:    session,
		baseURL:    base,
		mobileBase: base, // tests use same server for both
		csrf:       session.CSRFToken,
		wwwClaim:   "0",
	}
}

// applyHeaders sets all required Instagram web request headers on req.
func (c *Client) applyHeaders(req *http.Request) {
	c.mu.Lock()
	csrf := c.csrf
	wwwClaim := c.wwwClaim
	c.mu.Unlock()

	req.Header.Set("User-Agent", c.session.UserAgent)
	req.Header.Set("Cookie", c.session.CookieString())
	req.Header.Set("X-IG-App-ID", igAppID)
	req.Header.Set("X-CSRFToken", csrf)
	req.Header.Set("X-IG-WWW-Claim", wwwClaim)
	req.Header.Set("X-Instagram-AJAX", igAjaxToken)
	req.Header.Set("X-ASBD-ID", igASBDID)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
}

// applyMobileHeaders sets headers for i.instagram.com mobile API requests.
// Uses a mobile user agent to avoid "useragent mismatch" errors.
func (c *Client) applyMobileHeaders(req *http.Request) {
	c.mu.Lock()
	csrf := c.csrf
	wwwClaim := c.wwwClaim
	c.mu.Unlock()

	req.Header.Set("User-Agent", mobileUserAgent)
	req.Header.Set("Cookie", c.session.CookieString())
	req.Header.Set("X-IG-App-ID", igAppID)
	req.Header.Set("X-CSRFToken", csrf)
	req.Header.Set("X-IG-WWW-Claim", wwwClaim)
}

// captureResponseHeaders updates rotating state (CSRF, www-claim) from response headers.
func (c *Client) captureResponseHeaders(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Rotate CSRF token if Instagram sends a new one via Set-Cookie.
	for _, sc := range resp.Header["Set-Cookie"] {
		for _, part := range strings.Split(sc, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "csrftoken=") {
				newCSRF := strings.TrimPrefix(part, "csrftoken=")
				if newCSRF != "" {
					c.csrf = newCSRF
				}
			}
		}
	}

	// Capture www-claim from response.
	if claim := resp.Header.Get("X-IG-Set-WWW-Claim"); claim != "" {
		c.wwwClaim = claim
	}
}

// Get performs a GET request to path with optional query parameters.
func (c *Client) Get(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build GET request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// Post performs a POST request to path with a form-encoded body.
func (c *Client) Post(ctx context.Context, path string, body url.Values) (*http.Response, error) {
	encoded := ""
	if body != nil {
		encoded = body.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("build POST request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// MobileGet performs a GET request via i.instagram.com with a mobile user agent.
// Use this for endpoints that return "useragent mismatch" on www.instagram.com.
func (c *Client) MobileGet(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	u := c.mobileBase + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build mobile GET request: %w", err)
	}
	c.applyMobileHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mobile GET %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// MobilePost performs a POST request via i.instagram.com with a mobile user agent.
func (c *Client) MobilePost(ctx context.Context, path string, body url.Values) (*http.Response, error) {
	encoded := ""
	if body != nil {
		encoded = body.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.mobileBase+path, strings.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("build mobile POST request: %w", err)
	}
	c.applyMobileHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mobile POST %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// PostJSON performs a POST request to path with a JSON-encoded body.
func (c *Client) PostJSON(ctx context.Context, path string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal POST body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build POST JSON request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST JSON %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// graphQLRequest is the payload sent to /graphql/query.
type graphQLRequest struct {
	FriendlyName string         `json:"friendly_name"`
	Variables    map[string]any `json:"variables"`
	DocID        string         `json:"doc_id"`
}

// PostGraphQL executes a GraphQL query via POST /graphql/query and returns the
// raw "data" field from the response envelope.
func (c *Client) PostGraphQL(ctx context.Context, friendlyName string, variables map[string]any, docID string) (json.RawMessage, error) {
	payload := graphQLRequest{
		FriendlyName: friendlyName,
		Variables:    variables,
		DocID:        docID,
	}

	resp, err := c.PostJSON(ctx, "/graphql/query", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, parseRateLimitError(resp)
	}
	if resp.StatusCode >= 400 {
		return nil, c.handleError(resp)
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Status string          `json:"status"`
		// GraphQL errors surface here.
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode graphql response: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", envelope.Errors[0].Message)
	}
	return envelope.Data, nil
}

// DecodeJSON reads resp.Body into target, first checking for error status codes.
// The caller should not close resp.Body before calling DecodeJSON; this method
// closes it.
func (c *Client) DecodeJSON(resp *http.Response, target any) error {
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return parseRateLimitError(resp)
	}
	if resp.StatusCode >= 400 {
		return c.handleError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// handleError reads an error response body and returns a typed error.
func (c *Client) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var envelope struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		// challenge_required responses include this key.
		CheckpointURL string `json:"checkpoint_url"`
	}
	if jsonErr := json.Unmarshal(body, &envelope); jsonErr == nil {
		if strings.Contains(envelope.Message, "challenge_required") ||
			envelope.CheckpointURL != "" ||
			strings.Contains(string(body), "challenge_required") {
			msg := envelope.Message
			if msg == "" {
				msg = "checkpoint required"
			}
			return &ChallengeRequiredError{Message: msg}
		}
		if envelope.Message != "" {
			return fmt.Errorf("instagram API error (status=%s, http=%d): %s",
				envelope.Status, resp.StatusCode, envelope.Message)
		}
	}

	if len(body) > 256 {
		body = body[:256]
	}
	return fmt.Errorf("instagram API error (http=%d): %s", resp.StatusCode, string(body))
}

// parseRateLimitError builds a RateLimitError from a 429 response, parsing
// the Retry-After header when present.
func parseRateLimitError(resp *http.Response) *RateLimitError {
	// Drain body so the connection can be reused.
	io.Copy(io.Discard, resp.Body) //nolint:errcheck
	resp.Body.Close()

	e := &RateLimitError{}
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil {
			e.RetryAfter = time.Duration(secs) * time.Second
		}
	}
	return e
}
