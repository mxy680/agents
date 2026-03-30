package linear

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/emdash-projects/agents/internal/auth"
)

// Client is a GraphQL client for the Linear API.
type Client struct {
	http    *http.Client
	baseURL string // https://api.linear.app/graphql
}

// ClientFactory creates an authenticated Linear Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// NewClient creates a Client using the real Linear API.
func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := auth.NewLinearClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		http:    httpClient,
		baseURL: auth.LinearBaseURL(),
	}, nil
}

// LinearError represents a single error returned by the Linear GraphQL API.
type LinearError struct {
	Message string `json:"message"`
}

func (e *LinearError) Error() string {
	return fmt.Sprintf("Linear API error: %s", e.Message)
}

// graphQL sends a GraphQL query/mutation and unmarshals response.data into result.
// If the response contains errors, the first error is returned.
func (c *Client) graphQL(ctx context.Context, query string, variables map[string]any, result any) error {
	body := map[string]any{
		"query": query,
	}
	if len(variables) > 0 {
		body["variables"] = variables
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding GraphQL request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating GraphQL request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("executing GraphQL request: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading GraphQL response: %w", err)
	}

	// Parse the envelope: {"data": {...}, "errors": [...]}
	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []LinearError   `json:"errors"`
	}
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return fmt.Errorf("decoding GraphQL envelope: %w", err)
	}

	if len(envelope.Errors) > 0 {
		msgs := make([]string, 0, len(envelope.Errors))
		for _, e := range envelope.Errors {
			msgs = append(msgs, e.Message)
		}
		return fmt.Errorf("Linear API errors: %s", strings.Join(msgs, "; "))
	}

	if result == nil || len(envelope.Data) == 0 {
		return nil
	}

	if err := json.Unmarshal(envelope.Data, result); err != nil {
		return fmt.Errorf("decoding GraphQL data: %w", err)
	}

	return nil
}

// --- shared stderr helper ---

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", args...)
}
