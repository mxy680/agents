package gcpconsole

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/emdash-projects/agents/internal/auth"
)

const (
	defaultBaseURL = "https://clientauthconfig.clients6.google.com/v1"

	// gcpConsoleAPIKey is the public GCP Console API key — same for all users.
	gcpConsoleAPIKey = "AIzaSyCI-zsRP85UVOi0DjtiCwWBwQ1djDy741g"
)

// Client is a thin HTTP wrapper for the GCP Console internal API.
type Client struct {
	http    *http.Client
	baseURL string
}

// ClientFactory creates an authenticated GCP Console Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// NewClient creates a Client using the real GCP Console API.
func NewClient(_ context.Context) (*Client, error) {
	httpClient, err := auth.NewGCPConsoleClient()
	if err != nil {
		return nil, fmt.Errorf("gcp-console client: %w", err)
	}

	return &Client{
		http:    httpClient,
		baseURL: defaultBaseURL,
	}, nil
}

// GCPConsoleError represents an error response from the GCP Console API.
type GCPConsoleError struct {
	StatusCode int
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Status     string `json:"status"`
}

func (e *GCPConsoleError) Error() string {
	return fmt.Sprintf("GCP Console API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// addAPIKey appends the public API key query parameter to the path.
func addAPIKey(path string) string {
	if strings.Contains(path, "?") {
		return path + "&key=" + gcpConsoleAPIKey
	}
	return path + "?key=" + gcpConsoleAPIKey
}

// do executes an HTTP request against the GCP Console API and returns the raw
// response body. Non-2xx responses are returned as *GCPConsoleError.
func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	url := c.baseURL + addAPIKey(path)

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &GCPConsoleError{StatusCode: resp.StatusCode}
		var wrapper struct {
			Error *GCPConsoleError `json:"error"`
		}
		if jsonErr := json.Unmarshal(rawBody, &wrapper); jsonErr == nil && wrapper.Error != nil {
			apiErr.Code = wrapper.Error.Code
			apiErr.Message = wrapper.Error.Message
			apiErr.Status = wrapper.Error.Status
		}
		if apiErr.Message == "" {
			apiErr.Message = string(rawBody)
		}
		return nil, apiErr
	}

	return rawBody, nil
}

// doJSON executes an HTTP request and JSON-decodes the response into result.
func (c *Client) doJSON(ctx context.Context, method, path string, body any, result any) error {
	raw, err := c.do(ctx, method, path, body)
	if err != nil {
		return err
	}
	if result == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// newClientWithBase creates a Client that targets a custom base URL (used in tests).
func newClientWithBase(httpClient *http.Client, base string) *Client {
	return &Client{
		http:    httpClient,
		baseURL: base,
	}
}

// defaultHTTPClient returns an HTTP client with a reasonable timeout, used in tests.
func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}
