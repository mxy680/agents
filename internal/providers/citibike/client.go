package citibike

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://gbfs.citibikenyc.com/gbfs/2.3/en"

// Client wraps net/http for the Citi Bike GBFS public API.
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

// GetJSON performs an HTTP GET and JSON-decodes the response into out.
func (c *Client) GetJSON(ctx context.Context, path string, out any) error {
	url := c.baseURL + "/" + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// truncateBody returns the first 200 chars of a response body for error messages.
func truncateBody(body []byte) string {
	s := string(body)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
