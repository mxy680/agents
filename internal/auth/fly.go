package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// FlyEnvConfig holds the environment variable names for Fly.io API token auth.
var FlyEnvConfig = struct {
	Token   string
	BaseURL string
}{
	Token:   "FLY_API_TOKEN",
	BaseURL: "FLY_API_BASE_URL",
}

// FlyBaseURL returns the Fly.io Machines API base URL from the environment,
// defaulting to https://api.machines.dev.
func FlyBaseURL() string {
	if u := os.Getenv(FlyEnvConfig.BaseURL); u != "" {
		return u
	}
	return "https://api.machines.dev"
}

// flyBearerTransport injects the Fly.io Bearer token into every request.
type flyBearerTransport struct {
	base  http.RoundTripper
	token string
}

func (t *flyBearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

// NewFlyClient creates an authenticated HTTP client for the Fly.io API.
// Reads FLY_API_TOKEN (required) from the environment.
// The returned client has the token injected via a custom RoundTripper.
func NewFlyClient(_ context.Context) (*http.Client, error) {
	token, err := readEnv(FlyEnvConfig.Token)
	if err != nil {
		return nil, fmt.Errorf("fly auth: %w", err)
	}

	return &http.Client{
		Transport: &flyBearerTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}, nil
}
