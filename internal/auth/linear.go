package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// LinearEnvConfig holds the environment variable names for Linear API key auth.
var LinearEnvConfig = struct {
	APIKey  string
	BaseURL string
}{
	APIKey:  "LINEAR_API_KEY",
	BaseURL: "LINEAR_API_BASE_URL",
}

// LinearBaseURL returns the Linear GraphQL endpoint from the environment,
// defaulting to https://api.linear.app/graphql.
func LinearBaseURL() string {
	if u := os.Getenv(LinearEnvConfig.BaseURL); u != "" {
		return u
	}
	return "https://api.linear.app/graphql"
}

// linearBearerTransport injects the Linear Bearer token into every request.
type linearBearerTransport struct {
	base   http.RoundTripper
	apiKey string
}

func (t *linearBearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", t.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

// NewLinearClient creates an authenticated HTTP client for the Linear GraphQL API.
// Reads LINEAR_API_KEY (required) from the environment.
// The returned client has the token injected via a custom RoundTripper.
func NewLinearClient(_ context.Context) (*http.Client, error) {
	apiKey, err := readEnv(LinearEnvConfig.APIKey)
	if err != nil {
		return nil, fmt.Errorf("linear auth: %w", err)
	}

	return &http.Client{
		Transport: &linearBearerTransport{
			base:   http.DefaultTransport,
			apiKey: apiKey,
		},
	}, nil
}
