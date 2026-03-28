package fly

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

const (
	defaultBaseURL    = "https://api.machines.dev"
	defaultGraphQLURL = "https://api.fly.io/graphql"
)

// Client is a thin HTTP wrapper for the Fly.io REST and GraphQL APIs.
type Client struct {
	http       *http.Client
	baseURL    string
	graphqlURL string
}

// ClientFactory creates an authenticated Fly.io Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// NewClient creates a Client using the real Fly.io API.
func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := auth.NewFlyClient(ctx)
	if err != nil {
		return nil, err
	}

	baseURL := auth.FlyBaseURL()
	graphqlURL := os.Getenv("FLY_GRAPHQL_URL")
	if graphqlURL == "" {
		graphqlURL = defaultGraphQLURL
	}

	return &Client{
		http:       httpClient,
		baseURL:    baseURL,
		graphqlURL: graphqlURL,
	}, nil
}

// FlyError represents an error response from the Fly.io API.
type FlyError struct {
	StatusCode int
	Message    string `json:"error"`
}

func (e *FlyError) Error() string {
	return fmt.Sprintf("Fly.io API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// do executes an HTTP request against the Fly.io REST API and returns the raw
// response body. The caller is responsible for parsing. Non-2xx responses
// are returned as *FlyError.
func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	url := c.baseURL + path

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
		apiErr := &FlyError{StatusCode: resp.StatusCode}
		// Fly.io may return {"error": "message"} or plain text
		if jsonErr := json.Unmarshal(rawBody, apiErr); jsonErr != nil || apiErr.Message == "" {
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

// graphQL executes a GraphQL query or mutation against the Fly.io GraphQL API.
func (c *Client) graphQL(ctx context.Context, query string, variables map[string]any, result any) error {
	payload := map[string]any{
		"query":     query,
		"variables": variables,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encoding graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.graphqlURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating graphql request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("executing graphql request: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading graphql response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &FlyError{StatusCode: resp.StatusCode, Message: string(rawBody)}
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return fmt.Errorf("decoding graphql envelope: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return fmt.Errorf("graphql error: %s", envelope.Errors[0].Message)
	}
	if result == nil || len(envelope.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(envelope.Data, result); err != nil {
		return fmt.Errorf("decoding graphql data: %w", err)
	}
	return nil
}

// --- shared stderr helper ---

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", args...)
}
