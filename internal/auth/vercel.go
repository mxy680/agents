package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// VercelEnvConfig holds the environment variable names for Vercel API token auth.
var VercelEnvConfig = struct {
	Token    string
	TeamID   string
	BaseURL  string
}{
	Token:   "VERCEL_TOKEN",
	TeamID:  "VERCEL_TEAM_ID",
	BaseURL: "VERCEL_API_BASE_URL",
}

// VercelBaseURL returns the Vercel API base URL from the environment,
// defaulting to https://api.vercel.com.
func VercelBaseURL() string {
	if u := os.Getenv(VercelEnvConfig.BaseURL); u != "" {
		return u
	}
	return "https://api.vercel.com"
}

// vercelBearerTransport injects the Vercel Bearer token into every request.
type vercelBearerTransport struct {
	base  http.RoundTripper
	token string
}

func (t *vercelBearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

// NewVercelClient creates an authenticated HTTP client for the Vercel API.
// Reads VERCEL_TOKEN (required) and VERCEL_TEAM_ID (optional) from the environment.
// The returned client has the token injected via a custom RoundTripper.
func NewVercelClient(_ context.Context) (*http.Client, error) {
	token, err := readEnv(VercelEnvConfig.Token)
	if err != nil {
		return nil, fmt.Errorf("vercel auth: %w", err)
	}

	return &http.Client{
		Transport: &vercelBearerTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}, nil
}
