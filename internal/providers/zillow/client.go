package zillow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	azuretls "github.com/Noooste/azuretls-client"
)

const (
	zillowBaseURL   = "https://www.zillow.com"
	zillowStaticURL = "https://www.zillowstatic.com"
	mortgageAPIURL  = "https://mortgageapi.zillow.com"
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

	// Chrome 131 JA3 fingerprint
	chromeJA3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,45-13-43-0-16-65281-51-18-11-27-35-23-10-5-17613-21,29-23-24,0"

	// Chrome HTTP/2 SETTINGS fingerprint
	chromeHTTP2 = "1:65536;2:0;3:1000;4:6291456;6:262144|15663105|0|m,s,a,p"
)

// Client wraps azuretls to impersonate Chrome's TLS + HTTP/2 fingerprint.
// Zillow validates that the TLS handshake matches a real browser — Go's
// default net/http gets blocked with HTTP 403.
type Client struct {
	session   *azuretls.Session
	baseURL   string // https://www.zillow.com
	staticURL string // https://www.zillowstatic.com
	mortURL   string // https://mortgageapi.zillow.com
	userAgent string
	cookies   string // raw cookie string from Playwright session capture
	testHTTP  *http.Client // only for tests (bypasses azuretls)
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory using Chrome TLS impersonation.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		userAgent := os.Getenv("ZILLOW_USER_AGENT")
		if userAgent == "" {
			userAgent = defaultUserAgent
		}

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

		// Optional proxy support
		if proxyURL := os.Getenv("ZILLOW_PROXY_URL"); proxyURL != "" {
			if err := session.SetProxy(proxyURL); err != nil {
				return nil, fmt.Errorf("set proxy: %w", err)
			}
		}

		// Set default headers to match Chrome
		session.OrderedHeaders = azuretls.OrderedHeaders{
			{"User-Agent", userAgent},
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
			baseURL:   zillowBaseURL,
			staticURL: zillowStaticURL,
			mortURL:   mortgageAPIURL,
			userAgent: userAgent,
			cookies:   os.Getenv("ZILLOW_COOKIES"),
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(httpClient *http.Client, base string) *Client {
	return &Client{
		baseURL:   base,
		staticURL: base,
		mortURL:   base,
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
		{"Accept", "application/json"},
		{"Referer", c.baseURL + "/"},
		{"Origin", c.baseURL},
	}
	if c.cookies != "" {
		headers = append(headers, azuretls.OrderedHeaders{{"Cookie", c.cookies}}...)
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

// PutJSON performs an HTTP PUT with a JSON body and returns the response body.
func (c *Client) PutJSON(ctx context.Context, rawURL string, payload any) ([]byte, error) {
	if c.testHTTP != nil {
		return c.putJSONWithStdHTTP(ctx, rawURL, payload)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	headers := azuretls.OrderedHeaders{
		{"Content-Type", "application/json"},
		{"Accept", "application/json"},
		{"Referer", c.baseURL + "/"},
		{"Origin", c.baseURL},
	}
	if c.cookies != "" {
		headers = append(headers, azuretls.OrderedHeaders{{"Cookie", c.cookies}}...)
	}

	resp, err := c.session.Put(rawURL, string(data), headers)
	if err != nil {
		return nil, fmt.Errorf("http put: %w", err)
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
	c.applyHeaders(req)

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

// putJSONWithStdHTTP uses the standard http.Client for PUT (tests only).
func (c *Client) putJSONWithStdHTTP(ctx context.Context, rawURL string, payload any) ([]byte, error) {
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

	resp, err := c.testHTTP.Do(req)
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

// applyHeaders is kept for backward compatibility but not used with azuretls.
func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("Origin", c.baseURL)
	if c.cookies != "" {
		req.Header.Set("Cookie", c.cookies)
	}
}
