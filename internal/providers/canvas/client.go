package canvas

import (
	"bytes"
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

// RateLimitError is returned when Canvas responds with HTTP 403 with rate limit headers.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("canvas rate limit exceeded; retry after %s", e.RetryAfter)
	}
	return "canvas rate limit exceeded"
}

// Client is an HTTP client wrapper for the Canvas LMS REST API.
type Client struct {
	http    *http.Client
	session *auth.CanvasSession
	baseURL string

	mu   sync.Mutex
	csrf string // _csrf_token cookie value, also sent as X-CSRF-Token header
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory that reads real credentials
// from environment variables.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		session, err := auth.NewCanvasSession()
		if err != nil {
			return nil, fmt.Errorf("canvas auth: %w", err)
		}
		return &Client{
			http:    &http.Client{Timeout: 30 * time.Second},
			session: session,
			baseURL: session.BaseURL,
			csrf:    extractCookieValue(session.Cookies, "_csrf_token"),
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(session *auth.CanvasSession, httpClient *http.Client, base string) *Client {
	return &Client{
		http:    httpClient,
		session: session,
		baseURL: strings.TrimRight(base, "/"),
		csrf:    extractCookieValue(session.Cookies, "_csrf_token"),
	}
}

// extractCookieValue finds a cookie value by name from a raw cookie string.
func extractCookieValue(cookieStr, name string) string {
	for _, part := range strings.Split(cookieStr, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, name+"=") {
			return strings.TrimPrefix(part, name+"=")
		}
	}
	return ""
}

// applyHeaders sets all required Canvas API request headers.
func (c *Client) applyHeaders(req *http.Request) {
	c.mu.Lock()
	csrf := c.csrf
	c.mu.Unlock()

	req.Header.Set("Cookie", c.session.CookieString())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.session.UserAgent)
	// CSRF token needed for mutating requests
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		req.Header.Set("X-CSRF-Token", csrf)
	}
}

// captureResponseHeaders rotates the CSRF token from Set-Cookie headers.
func (c *Client) captureResponseHeaders(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, sc := range resp.Header["Set-Cookie"] {
		for _, part := range strings.Split(sc, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "_csrf_token=") {
				newCSRF := strings.TrimPrefix(part, "_csrf_token=")
				if newCSRF != "" {
					c.csrf = newCSRF
				}
			}
		}
	}
}

// Get performs a GET request to a Canvas API endpoint.
func (c *Client) Get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	u := c.baseURL + "/api/v1" + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build GET request: %w", err)
	}
	c.applyHeaders(req)
	return c.do(req)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body any) ([]byte, error) {
	return c.mutate(ctx, http.MethodPost, path, body)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body any) ([]byte, error) {
	return c.mutate(ctx, http.MethodPut, path, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	u := c.baseURL + "/api/v1" + path
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build DELETE request: %w", err)
	}
	c.applyHeaders(req)
	return c.do(req)
}

// Download performs a GET and writes the response body to a writer.
func (c *Client) Download(ctx context.Context, path string, w io.Writer) error {
	u := c.baseURL + "/api/v1" + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("build download request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()
	c.captureResponseHeaders(resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

// mutate performs a POST/PUT/PATCH with a JSON body.
func (c *Client) mutate(ctx context.Context, method, path string, body any) ([]byte, error) {
	u := c.baseURL + "/api/v1" + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build %s request: %w", method, err)
	}
	c.applyHeaders(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

// do executes an HTTP request and returns the response body.
func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()
	c.captureResponseHeaders(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
		if resp.Header.Get("X-Rate-Limit-Remaining") == "0" {
			return nil, &RateLimitError{}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ParseLinkHeader extracts the "next" URL from a Link header for pagination.
// Canvas returns: <https://...?page=2&per_page=10>; rel="next"
func ParseLinkHeader(header string) string {
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, `rel="next"`) {
			// Extract URL between < and >
			start := strings.Index(part, "<")
			end := strings.Index(part, ">")
			if start >= 0 && end > start {
				return part[start+1 : end]
			}
		}
	}
	return ""
}
