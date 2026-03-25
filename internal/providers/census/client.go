package census

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const defaultBaseURL = "https://api.census.gov/data/2023/acs/acs5"

// Client wraps net/http for the US Census Bureau ACS public API.
// No API key required for fewer than 500 requests/day.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory using standard net/http.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return &Client{
			httpClient: &http.Client{Timeout: 60 * time.Second},
			baseURL:    defaultBaseURL,
			apiKey:     os.Getenv("CENSUS_API_KEY"), // optional
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

// Query fetches ACS data. vars is a comma-separated variable list.
// forClause is like "tract:*" or "tract:000100".
// inClauses is like ["state:36", "county:005"].
func (c *Client) Query(ctx context.Context, vars string, forClause string, inClauses []string) ([][]string, error) {
	params := url.Values{}
	params.Set("get", vars)
	params.Set("for", forClause)
	for _, in := range inClauses {
		params.Add("in", in)
	}
	if c.apiKey != "" {
		params.Set("key", c.apiKey)
	}

	reqURL := c.baseURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

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

	var result [][]string
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

// truncateBody returns the first 200 chars of a response body for error messages.
func truncateBody(body []byte) string {
	s := string(body)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
