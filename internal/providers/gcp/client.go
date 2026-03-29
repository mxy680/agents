package gcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/emdash-projects/agents/internal/auth"
)

// Base URLs for the various GCP REST APIs.
const (
	resourceManagerBaseURL = "https://cloudresourcemanager.googleapis.com/v3"
	serviceUsageBaseURL    = "https://serviceusage.googleapis.com/v1"
	iamBaseURL             = "https://iam.googleapis.com/v1"
	iapBaseURL             = "https://iap.googleapis.com/v1"
)

// Client is a thin HTTP wrapper for the GCP REST APIs.
// It operates on full URLs (not relative paths) since different resource
// types are served from different base URLs.
type Client struct {
	http      *http.Client
	projectID string
}

// ClientFactory creates an authenticated GCP Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// NewClient creates a Client using the real GCP APIs.
func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := auth.NewGCPClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		http:      httpClient,
		projectID: auth.GCPDefaultProject(),
	}, nil
}

// GCPError represents an error response from a GCP API.
type GCPError struct {
	StatusCode int
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Status     string `json:"status"`
}

func (e *GCPError) Error() string {
	return fmt.Sprintf("GCP API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// do executes an HTTP request against the given full URL and returns the raw
// response body. Non-2xx responses are returned as *GCPError.
func (c *Client) do(ctx context.Context, method, url string, body any) ([]byte, error) {
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
		apiErr := &GCPError{StatusCode: resp.StatusCode}
		// GCP wraps errors inside an "error" key.
		var wrapper struct {
			Error *GCPError `json:"error"`
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
func (c *Client) doJSON(ctx context.Context, method, url string, body any, result any) error {
	raw, err := c.do(ctx, method, url, body)
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

// Operation represents a GCP long-running operation.
type Operation struct {
	Name     string          `json:"name"`
	Done     bool            `json:"done"`
	Error    *GCPError       `json:"error,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
}

// waitForOperation polls a long-running operation until it completes or fails.
// It uses simple exponential backoff capped at 10 seconds.
func (c *Client) waitForOperation(ctx context.Context, op *Operation) (*Operation, error) {
	if op.Done {
		if op.Error != nil {
			return nil, op.Error
		}
		return op, nil
	}

	// Determine the base URL for the operation type.
	// Operation names look like: "operations/xxx" or "projects/{id}/operations/{id}"
	// Use cloudresourcemanager for project operations, serviceusage for service ops.
	baseURL := resourceManagerBaseURL
	delay := 2 * time.Second

	for !op.Done {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}

		url := baseURL + "/" + op.Name
		var updated Operation
		if err := c.doJSON(ctx, http.MethodGet, url, nil, &updated); err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = &updated

		if delay < 10*time.Second {
			delay *= 2
		}
	}

	if op.Error != nil {
		return nil, op.Error
	}
	return op, nil
}

// project returns the configured project ID, falling back to the environment.
func (c *Client) project() string {
	return c.projectID
}

// resolveProject returns the provided project ID if non-empty, or falls back to the client default.
func (c *Client) resolveProject(flagProject string) (string, error) {
	if flagProject != "" {
		return flagProject, nil
	}
	if c.projectID != "" {
		return c.projectID, nil
	}
	return "", fmt.Errorf("no project specified: use --project or set GCP_PROJECT_ID")
}

// --- shared stderr helper ---

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", args...)
}
