package zillow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	zillowBaseURL     = "https://www.zillow.com"
	zillowStaticURL   = "https://www.zillowstatic.com"
	mortgageAPIURL    = "https://mortgageapi.zillow.com"
	defaultUserAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// Client is an HTTP client wrapper for Zillow's internal APIs.
type Client struct {
	http      *http.Client
	baseURL   string // https://www.zillow.com
	staticURL string // https://www.zillowstatic.com
	mortURL   string // https://mortgageapi.zillow.com
	userAgent string
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory that reads config from env vars.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		transport := http.DefaultTransport.(*http.Transport).Clone()

		// Optional proxy support for production use
		if proxyURL := os.Getenv("ZILLOW_PROXY_URL"); proxyURL != "" {
			parsed, err := url.Parse(proxyURL)
			if err != nil {
				return nil, fmt.Errorf("parse ZILLOW_PROXY_URL: %w", err)
			}
			transport.Proxy = http.ProxyURL(parsed)
		}

		userAgent := os.Getenv("ZILLOW_USER_AGENT")
		if userAgent == "" {
			userAgent = defaultUserAgent
		}

		return &Client{
			http:      &http.Client{Timeout: 30 * time.Second, Transport: transport},
			baseURL:   zillowBaseURL,
			staticURL: zillowStaticURL,
			mortURL:   mortgageAPIURL,
			userAgent: userAgent,
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(httpClient *http.Client, base string) *Client {
	return &Client{
		http:      httpClient,
		baseURL:   base,
		staticURL: base,
		mortURL:   base,
		userAgent: "TestAgent/1.0",
	}
}

// applyHeaders sets the required request headers for Zillow's web API.
func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("Origin", c.baseURL)
}

// Get performs an HTTP GET and returns the response body.
func (c *Client) Get(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{}
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, &BlockedError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(body))
	}

	return body, nil
}

// PutJSON performs an HTTP PUT with a JSON body and returns the response body.
func (c *Client) PutJSON(ctx context.Context, rawURL string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, rawURL, strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http put: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{}
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, &BlockedError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(body))
	}

	return body, nil
}

// PostJSON performs an HTTP POST with a JSON body and returns the response body.
func (c *Client) PostJSON(ctx context.Context, rawURL string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{}
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, &BlockedError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(body))
	}

	return body, nil
}

// RateLimitError is returned when Zillow responds with HTTP 429.
type RateLimitError struct{}

func (e *RateLimitError) Error() string {
	return "zillow rate limit exceeded; try again later or use a proxy"
}

// BlockedError is returned when Zillow blocks the request (HTTP 403).
type BlockedError struct {
	StatusCode int
	Body       string
}

func (e *BlockedError) Error() string {
	return fmt.Sprintf("zillow blocked request (HTTP %d); try using a proxy via ZILLOW_PROXY_URL", e.StatusCode)
}

// truncateBody returns the first 200 chars of a response body for error messages.
func truncateBody(body []byte) string {
	s := string(body)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
