package vercel

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

const defaultBaseURL = "https://api.vercel.com"

// Client is a thin HTTP wrapper for the Vercel REST API.
type Client struct {
	http    *http.Client
	baseURL string
	teamID  string
}

// ClientFactory creates an authenticated Vercel Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// NewClient creates a Client using the real Vercel API.
func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := auth.NewVercelClient(ctx)
	if err != nil {
		return nil, err
	}

	baseURL := auth.VercelBaseURL()
	teamID := os.Getenv(auth.VercelEnvConfig.TeamID) // optional

	return &Client{
		http:    httpClient,
		baseURL: baseURL,
		teamID:  teamID,
	}, nil
}

// VercelError represents an error response from the Vercel API.
type VercelError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *VercelError) Error() string {
	return fmt.Sprintf("Vercel API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// addTeamID appends teamId query param if the client has a team configured.
func (c *Client) addTeamID(path string) string {
	if c.teamID == "" {
		return path
	}
	// Determine separator: path may already have a query string
	sep := "?"
	for _, ch := range path {
		if ch == '?' {
			sep = "&"
			break
		}
	}
	return path + sep + "teamId=" + c.teamID
}

// do executes an HTTP request against the Vercel API and returns the raw
// response body. The caller is responsible for parsing. Non-2xx responses
// are returned as *VercelError.
func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	url := c.baseURL + c.addTeamID(path)

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
		apiErr := &VercelError{StatusCode: resp.StatusCode}
		// Vercel wraps errors inside an "error" key
		var wrapper struct {
			Error *VercelError `json:"error"`
		}
		if jsonErr := json.Unmarshal(rawBody, &wrapper); jsonErr == nil && wrapper.Error != nil {
			apiErr.Code = wrapper.Error.Code
			apiErr.Message = wrapper.Error.Message
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

// --- shared stderr helper ---

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "WARNING: "+format+"\n", args...)
}
