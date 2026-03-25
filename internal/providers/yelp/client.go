package yelp

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

const (
	yelpBaseURL = "https://api.yelp.com/v3"
)

// Client is an HTTP client wrapper for the Yelp Fusion API v3.
type Client struct {
	http   *http.Client
	apiKey string
	base   string
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory that reads config from env vars.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		apiKey := os.Getenv("YELP_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("YELP_API_KEY environment variable is not set")
		}
		return &Client{
			http:   &http.Client{Timeout: 30 * time.Second},
			apiKey: apiKey,
			base:   yelpBaseURL,
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(httpClient *http.Client, base, apiKey string) *Client {
	return &Client{
		http:   httpClient,
		apiKey: apiKey,
		base:   base,
	}
}

// doYelp performs an HTTP request against the Yelp Fusion API.
// path is the API path (e.g. "/businesses/search").
func (c *Client) doYelp(ctx context.Context, method, path string, params url.Values) ([]byte, error) {
	rawURL := c.base + path
	if len(params) > 0 {
		rawURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http %s: %w", method, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{}
	}
	if resp.StatusCode >= 400 {
		var errResp struct {
			Error struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"error"`
		}
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Error.Description != "" {
			return nil, fmt.Errorf("yelp api error (HTTP %d) %s: %s", resp.StatusCode, errResp.Error.Code, errResp.Error.Description)
		}
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(body))
	}

	return body, nil
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
