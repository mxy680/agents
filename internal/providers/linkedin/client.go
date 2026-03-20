package linkedin

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
	baseURL = "https://www.linkedin.com"

	// LinkedIn Rest-li protocol version.
	restliProtocolVersion = "2.0.0"
)

// RateLimitError is returned when LinkedIn responds with HTTP 429.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("linkedin rate limit exceeded; retry after %s", e.RetryAfter)
	}
	return "linkedin rate limit exceeded"
}

// ChallengeRequiredError is returned when LinkedIn requires additional verification.
type ChallengeRequiredError struct {
	Message string
}

func (e *ChallengeRequiredError) Error() string {
	return fmt.Sprintf("linkedin challenge required: %s", e.Message)
}

// Client is an HTTP client wrapper for the LinkedIn Voyager API.
type Client struct {
	http    *http.Client
	session *auth.LinkedInSession
	baseURL string

	mu   sync.Mutex
	csrf string // JSESSIONID value used as csrf-token header
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory that reads real credentials
// from environment variables.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		session, err := auth.NewLinkedInSession()
		if err != nil {
			return nil, fmt.Errorf("linkedin auth: %w", err)
		}
		return &Client{
			http:    &http.Client{Timeout: 30 * time.Second},
			session: session,
			baseURL: baseURL,
			csrf:    session.JSESSIONID,
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(session *auth.LinkedInSession, httpClient *http.Client, base string) *Client {
	return &Client{
		http:    httpClient,
		session: session,
		baseURL: base,
		csrf:    session.JSESSIONID,
	}
}

// applyHeaders sets all required LinkedIn Voyager API request headers.
func (c *Client) applyHeaders(req *http.Request) {
	c.mu.Lock()
	csrf := c.csrf
	c.mu.Unlock()

	req.Header.Set("User-Agent", c.session.UserAgent)
	req.Header.Set("Cookie", c.session.CookieString())
	req.Header.Set("Csrf-Token", csrf)
	req.Header.Set("X-Li-Lang", "en_US")
	req.Header.Set("X-RestLi-Protocol-Version", restliProtocolVersion)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/vnd.linkedin.normalized+json+2.1")
}

// captureResponseHeaders updates rotating state from response headers.
func (c *Client) captureResponseHeaders(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Rotate CSRF/JSESSIONID if LinkedIn sends a new one.
	for _, sc := range resp.Header["Set-Cookie"] {
		for _, part := range strings.Split(sc, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "JSESSIONID=") {
				newCSRF := strings.TrimPrefix(part, "JSESSIONID=")
				newCSRF = strings.Trim(newCSRF, "\"")
				if newCSRF != "" {
					c.csrf = newCSRF
				}
			}
		}
	}
}

// Get performs a GET request to a Voyager API path.
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

// Post performs a POST request with a form-encoded body.
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

// PostJSON performs a POST request with a JSON-encoded body.
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

// PutJSON performs a PUT request with a JSON-encoded body.
func (c *Client) PutJSON(ctx context.Context, path string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal PUT body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build PUT JSON request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PUT JSON %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// Patch performs a PATCH request with a JSON-encoded body.
func (c *Client) Patch(ctx context.Context, path string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal PATCH body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build PATCH request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PATCH %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build DELETE request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DELETE %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// DecodeJSON reads resp.Body into target, first checking for error status codes.
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
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("linkedin API error (http=%d): could not read response body: %w", resp.StatusCode, readErr)
	}

	var envelope struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}
	if jsonErr := json.Unmarshal(body, &envelope); jsonErr == nil {
		if strings.Contains(string(body), "CHALLENGE") {
			msg := envelope.Message
			if msg == "" {
				msg = "verification challenge required"
			}
			return &ChallengeRequiredError{Message: msg}
		}
		if envelope.Message != "" {
			return fmt.Errorf("linkedin API error (status=%d, http=%d): %s",
				envelope.Status, resp.StatusCode, envelope.Message)
		}
	}

	return fmt.Errorf("linkedin API error (http=%d): no details available", resp.StatusCode)
}

// parseRateLimitError builds a RateLimitError from a 429 response.
func parseRateLimitError(resp *http.Response) *RateLimitError {
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
