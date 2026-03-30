package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/emdash-projects/agents/internal/auth"
)

const defaultBaseURL = "https://api.cloudflare.com/client/v4"

// Client is a thin HTTP wrapper for the Cloudflare REST API.
type Client struct {
	http      *http.Client
	baseURL   string
	accountID string
}

// ClientFactory creates an authenticated Cloudflare Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// NewClient creates a Client using the real Cloudflare API.
func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := auth.NewCloudflareClient(ctx)
	if err != nil {
		return nil, err
	}

	baseURL := auth.CloudflareBaseURL()
	accountID := os.Getenv(auth.CloudflareEnvConfig.AccountID) // optional

	return &Client{
		http:      httpClient,
		baseURL:   baseURL,
		accountID: accountID,
	}, nil
}

// CloudflareAPIError represents an error entry from the Cloudflare API.
type CloudflareAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CloudflareError is returned when the Cloudflare API responds with success=false.
type CloudflareError struct {
	StatusCode int
	Errors     []CloudflareAPIError
}

func (e *CloudflareError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("Cloudflare API error (HTTP %d): %s", e.StatusCode, e.Errors[0].Message)
	}
	return fmt.Sprintf("Cloudflare API error (HTTP %d)", e.StatusCode)
}

// do executes an HTTP request against the Cloudflare API and returns the raw
// bytes of the "result" field from the Cloudflare response envelope.
// Non-2xx responses or success=false are returned as *CloudflareError.
func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	var contentType string
	switch v := body.(type) {
	case nil:
		// no body
	case rawBody:
		bodyReader = bytes.NewReader(v.data)
		contentType = v.contentType
	default:
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
		contentType = "application/json"
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Cloudflare always wraps responses in {"success": bool, "result": ..., "errors": [...]}
	var envelope struct {
		Success bool                 `json:"success"`
		Result  json.RawMessage      `json:"result"`
		Errors  []CloudflareAPIError `json:"errors"`
	}

	if jsonErr := json.Unmarshal(rawBytes, &envelope); jsonErr != nil {
		// Not JSON — treat HTTP status as the error indicator
		if resp.StatusCode >= 400 {
			return nil, &CloudflareError{StatusCode: resp.StatusCode}
		}
		return rawBytes, nil
	}

	if !envelope.Success || resp.StatusCode >= 400 {
		return nil, &CloudflareError{
			StatusCode: resp.StatusCode,
			Errors:     envelope.Errors,
		}
	}

	return envelope.Result, nil
}

// doJSON executes an HTTP request and JSON-decodes the result field into result.
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

// accountPath builds an account-scoped path, returning an error if no account ID is set.
func (c *Client) accountPath(suffix string) (string, error) {
	if c.accountID == "" {
		return "", fmt.Errorf("CLOUDFLARE_ACCOUNT_ID is required for this command")
	}
	return fmt.Sprintf("/accounts/%s%s", c.accountID, suffix), nil
}

// rawBody wraps raw bytes with an explicit Content-Type for non-JSON uploads.
type rawBody struct {
	data        []byte
	contentType string
}

// --- shared stderr helper ---

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", args...)
}
