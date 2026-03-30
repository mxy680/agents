package dof

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://data.cityofnewyork.us/resource/w7rz-68fs.json"

// Client wraps net/http for the NYC DOF Socrata public API.
// No authentication is required — this is a public API.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory using standard net/http.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return &Client{
			httpClient: &http.Client{Timeout: 30 * time.Second},
			baseURL:    defaultBaseURL,
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(httpClient *http.Client, base string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    base,
	}
}

// Get performs an HTTP GET to the base URL with SoQL query parameters and returns the response body.
func (c *Client) Get(ctx context.Context, params url.Values) ([]byte, error) {
	reqURL := c.baseURL
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Emdash-Agents/1.0")
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
