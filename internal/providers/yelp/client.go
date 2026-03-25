package yelp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/emdash-projects/agents/internal/auth"
)

const (
	yelpBaseURL = "https://www.yelp.com"
)

// Client is an HTTP client wrapper for Yelp's internal web API.
// Uses cookie-based session auth (bse + zss cookies).
type Client struct {
	http    *http.Client
	session *auth.YelpSession
	base    string

	mu   sync.Mutex
	csrf string // csrftok value, rotated from cookies/pages
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory that reads credentials from env vars.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		session, err := auth.NewYelpSession()
		if err != nil {
			return nil, fmt.Errorf("yelp auth: %w", err)
		}
		return &Client{
			http:    &http.Client{Timeout: 30 * time.Second},
			session: session,
			base:    yelpBaseURL,
			csrf:    session.CSRFToken,
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(httpClient *http.Client, base string) *Client {
	return &Client{
		http: httpClient,
		session: &auth.YelpSession{
			BSE:       "test-bse",
			ZSS:       "test-zss",
			CSRFToken: "test-csrf",
			UserAgent: "test-agent",
		},
		base: base,
		csrf: "test-csrf",
	}
}

// doYelp performs an HTTP request against Yelp's internal web API.
// path is relative to www.yelp.com (e.g. "/search/snippet").
func (c *Client) doYelp(ctx context.Context, method, path string, params url.Values) ([]byte, error) {
	rawURL := c.base + path
	if len(params) > 0 {
		rawURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http %s: %w", method, err)
	}
	defer resp.Body.Close()

	// Rotate CSRF token from Set-Cookie if present
	c.rotateCSRF(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(body))
	}

	return body, nil
}

// doYelpPost performs a POST request with form-encoded body.
func (c *Client) doYelpPost(ctx context.Context, path string, form url.Values) ([]byte, error) {
	rawURL := c.base + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http POST: %w", err)
	}
	defer resp.Body.Close()

	c.rotateCSRF(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(body))
	}

	return body, nil
}

// doYelpJSON performs a POST with JSON body (for GraphQL batch endpoint).
func (c *Client) doYelpJSON(ctx context.Context, path string, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	rawURL := c.base + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http POST: %w", err)
	}
	defer resp.Body.Close()

	c.rotateCSRF(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(body))
	}

	return body, nil
}

// setHeaders applies session cookies and common headers to a request.
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Cookie", c.session.CookieString())
	req.Header.Set("User-Agent", c.session.UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", yelpBaseURL)
	req.Header.Set("Referer", yelpBaseURL+"/")

	c.mu.Lock()
	csrf := c.csrf
	c.mu.Unlock()
	if csrf != "" {
		req.Header.Set("X-Csrf-Token", csrf)
	}
}

// rotateCSRF extracts the csrftok cookie from a response and updates the stored value.
func (c *Client) rotateCSRF(resp *http.Response) {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftok" && cookie.Value != "" {
			c.mu.Lock()
			c.csrf = cookie.Value
			c.mu.Unlock()
			return
		}
	}
}

// getCSRF returns the current CSRF token value.
func (c *Client) getCSRF() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.csrf
}

// RateLimitError is returned when Yelp responds with HTTP 429.
type RateLimitError struct{}

func (e *RateLimitError) Error() string {
	return "yelp rate limit exceeded; try again later"
}

// truncateBody returns the first 200 chars of a response body for error messages.
func truncateBody(body []byte) string {
	s := string(body)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
