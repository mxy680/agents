package nydos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	dailyFilingsURL = "https://data.ny.gov/resource/k4vb-judh.json"
	activeCorpsURL  = "https://data.ny.gov/resource/n9v6-gdp6.json"
	userAgent       = "Emdash-Agents/1.0"
)

// Client wraps net/http for the NY DOS Socrata public APIs.
// No authentication is required — these are public APIs.
type Client struct {
	httpClient      *http.Client
	dailyFilingsURL string
	activeCorpsURL  string
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory using standard net/http.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return &Client{
			httpClient:      &http.Client{Timeout: 60 * time.Second},
			dailyFilingsURL: dailyFilingsURL,
			activeCorpsURL:  activeCorpsURL,
		}, nil
	}
}

// newClientWithBase creates a Client targeting custom base URLs (used in tests).
func newClientWithBase(httpClient *http.Client, dailyBase, activeBase string) *Client {
	return &Client{
		httpClient:      httpClient,
		dailyFilingsURL: dailyBase,
		activeCorpsURL:  activeBase,
	}
}

// Query performs an HTTP GET to the given base URL with Socrata query parameters
// and returns the response body.
func (c *Client) Query(ctx context.Context, baseURL string, params url.Values) ([]byte, error) {
	reqURL := baseURL
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

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
