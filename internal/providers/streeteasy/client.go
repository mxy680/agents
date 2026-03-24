package streeteasy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	streeteasyBaseURL = "https://streeteasy.com"
	defaultUserAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// Client is an HTTP client wrapper for StreetEasy's web pages.
type Client struct {
	http      *http.Client
	baseURL   string // https://streeteasy.com
	userAgent string
	cookies   string // raw cookie string from Playwright session capture
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory that reads config from env vars.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		transport := http.DefaultTransport.(*http.Transport).Clone()

		// Optional proxy support for production use
		if proxyURL := os.Getenv("STREETEASY_PROXY_URL"); proxyURL != "" {
			parsed, err := url.Parse(proxyURL)
			if err != nil {
				return nil, fmt.Errorf("parse STREETEASY_PROXY_URL: %w", err)
			}
			transport.Proxy = http.ProxyURL(parsed)
		}

		userAgent := os.Getenv("STREETEASY_USER_AGENT")
		if userAgent == "" {
			userAgent = defaultUserAgent
		}

		return &Client{
			http:      &http.Client{Timeout: 30 * time.Second, Transport: transport},
			baseURL:   streeteasyBaseURL,
			userAgent: userAgent,
			cookies:   os.Getenv("STREETEASY_COOKIES"),
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(httpClient *http.Client, base string) *Client {
	return &Client{
		http:      httpClient,
		baseURL:   base,
		userAgent: "TestAgent/1.0",
	}
}

// applyHeaders sets the required request headers for StreetEasy's web pages.
func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", c.baseURL+"/")
	if c.cookies != "" {
		req.Header.Set("Cookie", c.cookies)
	}
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

// RateLimitError is returned when StreetEasy responds with HTTP 429.
type RateLimitError struct{}

func (e *RateLimitError) Error() string {
	return "streeteasy rate limit exceeded; try again later or use a proxy"
}

// BlockedError is returned when StreetEasy blocks the request (HTTP 403).
type BlockedError struct {
	StatusCode int
	Body       string
}

func (e *BlockedError) Error() string {
	return fmt.Sprintf("streeteasy blocked request (HTTP %d); refresh cookies via Playwright session capture", e.StatusCode)
}

// truncateBody returns the first 200 chars of a response body for error messages.
func truncateBody(body []byte) string {
	s := string(body)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
