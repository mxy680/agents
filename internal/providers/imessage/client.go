package imessage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emdash-projects/agents/internal/auth"
)

// Client is an HTTP client wrapper for the BlueBubbles REST API.
type Client struct {
	http     *http.Client
	baseURL  string
	password string
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory that reads real credentials
// from environment variables.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		creds, err := auth.NewBlueBubblesCredentials()
		if err != nil {
			return nil, fmt.Errorf("bluebubbles auth: %w", err)
		}
		return &Client{
			http:     &http.Client{Timeout: 30 * time.Second},
			baseURL:  strings.TrimRight(creds.URL, "/"),
			password: creds.Password,
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
func newClientWithBase(httpClient *http.Client, base, password string) *Client {
	return &Client{
		http:     httpClient,
		baseURL:  strings.TrimRight(base, "/"),
		password: password,
	}
}

// buildURL constructs a full API URL with the password query parameter.
func (c *Client) buildURL(path string, params url.Values) string {
	if params == nil {
		params = url.Values{}
	}
	params.Set("password", c.password)
	return fmt.Sprintf("%s/api/v1/%s?%s", c.baseURL, strings.TrimPrefix(path, "/"), params.Encode())
}

// Get performs a GET request to the BlueBubbles API.
func (c *Client) Get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL(path, params), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return c.do(req)
}

// Post performs a POST request with a JSON body to the BlueBubbles API.
func (c *Client) Post(ctx context.Context, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL(path, nil), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.buildURL(path, nil), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.buildURL(path, nil), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return c.do(req)
}

// Download performs a GET request and writes the response body to a writer.
func (c *Client) Download(ctx context.Context, path string, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL(path, nil), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

// do executes an HTTP request and returns the response body.
func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// APIResponse is the standard BlueBubbles API response envelope.
type APIResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   json.RawMessage `json:"error,omitempty"`
}

// ParseResponse extracts the data field from a BlueBubbles API response.
func ParseResponse(body []byte) (json.RawMessage, error) {
	var resp APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("API error %d: %s", resp.Status, string(resp.Error))
	}
	return resp.Data, nil
}

// ParseResponseRaw returns the raw body without extracting data field.
// Used for endpoints that don't follow the standard envelope.
func ParseResponseRaw(body []byte) (json.RawMessage, error) {
	return json.RawMessage(body), nil
}
