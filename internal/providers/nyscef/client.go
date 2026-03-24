package nyscef

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	azuretls "github.com/Noooste/azuretls-client"
)

const (
	nyscefBaseURL    = "https://iapps.courts.state.ny.us"
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

	// Chrome 131 JA3 fingerprint
	chromeJA3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,45-13-43-0-16-65281-51-18-11-27-35-23-10-5-17613-21,29-23-24,0"

	// Chrome HTTP/2 SETTINGS fingerprint
	chromeHTTP2 = "1:65536;2:0;3:1000;4:6291456;6:262144|15663105|0|m,s,a,p"
)

// Client wraps azuretls to impersonate Chrome's TLS + HTTP/2 fingerprint.
// This bypasses Cloudflare protection on NYSCEF without requiring cookies.
type Client struct {
	session   *azuretls.Session
	baseURL   string
	userAgent string
	testHTTP  *http.Client // only for tests (bypasses azuretls)
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory using Chrome TLS impersonation.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		session := azuretls.NewSession()
		session.SetTimeout(30 * time.Second)

		// Impersonate Chrome's TLS fingerprint
		if err := session.ApplyJa3(chromeJA3, azuretls.Chrome); err != nil {
			return nil, fmt.Errorf("apply chrome JA3: %w", err)
		}

		// Impersonate Chrome's HTTP/2 fingerprint
		if err := session.ApplyHTTP2(chromeHTTP2); err != nil {
			return nil, fmt.Errorf("apply chrome HTTP/2: %w", err)
		}

		// Set default headers to match Chrome
		session.OrderedHeaders = azuretls.OrderedHeaders{
			{"User-Agent", defaultUserAgent},
			{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
			{"Accept-Language", "en-US,en;q=0.9"},
			{"Accept-Encoding", "gzip, deflate, br"},
			{"Upgrade-Insecure-Requests", "1"},
			{"sec-ch-ua", `"Chromium";v="131", "Not A(Brand";v="24"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-ch-ua-platform", `"macOS"`},
			{"Sec-Fetch-Dest", "document"},
			{"Sec-Fetch-Mode", "navigate"},
			{"Sec-Fetch-Site", "none"},
			{"Sec-Fetch-User", "?1"},
		}

		return &Client{
			session:   session,
			baseURL:   nyscefBaseURL,
			userAgent: defaultUserAgent,
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(httpClient *http.Client, base string) *Client {
	return &Client{
		baseURL:   base,
		userAgent: "TestAgent/1.0",
		testHTTP:  httpClient,
	}
}

// Get performs an HTTP GET and returns the response body.
func (c *Client) Get(ctx context.Context, rawURL string) ([]byte, error) {
	if c.testHTTP != nil {
		return c.getWithStdHTTP(ctx, rawURL)
	}

	headers := azuretls.OrderedHeaders{
		{"Referer", c.baseURL + "/"},
	}

	resp, err := c.session.Get(rawURL, headers)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{}
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, &BlockedError{StatusCode: resp.StatusCode, Body: string(resp.Body)}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(resp.Body))
	}

	return resp.Body, nil
}

// Post performs an HTTP POST with form-encoded data and returns the response body.
func (c *Client) Post(ctx context.Context, rawURL string, form url.Values) ([]byte, error) {
	if c.testHTTP != nil {
		return c.postWithStdHTTP(ctx, rawURL, form)
	}

	headers := azuretls.OrderedHeaders{
		{"Referer", c.baseURL + "/nyscef/CaseSearch"},
		{"Content-Type", "application/x-www-form-urlencoded"},
		{"Origin", c.baseURL},
		{"Sec-Fetch-Dest", "document"},
		{"Sec-Fetch-Mode", "navigate"},
		{"Sec-Fetch-Site", "same-origin"},
	}

	resp, err := c.session.Post(rawURL, strings.NewReader(form.Encode()), headers)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{}
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, &BlockedError{StatusCode: resp.StatusCode, Body: string(resp.Body)}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, truncateBody(resp.Body))
	}

	return resp.Body, nil
}

// getWithStdHTTP uses the standard http.Client (for tests only).
func (c *Client) getWithStdHTTP(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html")

	resp, err := c.testHTTP.Do(req)
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

// postWithStdHTTP uses the standard http.Client for POST (for tests only).
func (c *Client) postWithStdHTTP(ctx context.Context, rawURL string, form url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")

	resp, err := c.testHTTP.Do(req)
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

// RateLimitError is returned when NYSCEF responds with HTTP 429.
type RateLimitError struct{}

func (e *RateLimitError) Error() string {
	return "nyscef rate limit exceeded; try again later"
}

// BlockedError is returned when NYSCEF blocks the request (HTTP 403).
type BlockedError struct {
	StatusCode int
	Body       string
}

func (e *BlockedError) Error() string {
	return fmt.Sprintf("nyscef blocked request (HTTP %d); Chrome TLS fingerprint may need updating", e.StatusCode)
}

// truncateBody returns the first 200 chars of a response body for error messages.
func truncateBody(body []byte) string {
	s := string(body)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
