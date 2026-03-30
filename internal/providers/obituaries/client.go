package obituaries

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://www.legacy.com/api/obituaries"

// Client wraps net/http for the Legacy.com obituary search API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory using standard net/http.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return &Client{
			httpClient: &http.Client{Timeout: 30 * time.Second},
			baseURL:    defaultBaseURL,
			userAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(httpClient *http.Client, base string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    base,
		userAgent:  "TestAgent/1.0",
	}
}

// Get performs an HTTP GET with query parameters and returns the raw response body.
func (c *Client) Get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	reqURL := c.baseURL + "/" + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(body))
	}
	return body, nil
}

// truncateBody returns the first 200 chars of a response body for error messages.
func truncateBody(body []byte) string {
	s := string(body)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
