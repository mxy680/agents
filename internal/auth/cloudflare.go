package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// CloudflareEnvConfig holds the environment variable names for Cloudflare API token auth.
var CloudflareEnvConfig = struct {
	Token     string
	AccountID string
	BaseURL   string
}{
	Token:     "CLOUDFLARE_API_TOKEN",
	AccountID: "CLOUDFLARE_ACCOUNT_ID",
	BaseURL:   "CLOUDFLARE_API_BASE_URL",
}

// CloudflareBaseURL returns the Cloudflare API base URL from the environment,
// defaulting to https://api.cloudflare.com/client/v4.
func CloudflareBaseURL() string {
	if u := os.Getenv(CloudflareEnvConfig.BaseURL); u != "" {
		return u
	}
	return "https://api.cloudflare.com/client/v4"
}

// cloudflareBearerTransport injects the Cloudflare Bearer token into every request.
type cloudflareBearerTransport struct {
	base  http.RoundTripper
	token string
}

func (t *cloudflareBearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

// NewCloudflareClient creates an authenticated HTTP client for the Cloudflare API.
// Reads CLOUDFLARE_API_TOKEN (required) from the environment.
// The returned client has the token injected via a custom RoundTripper.
func NewCloudflareClient(_ context.Context) (*http.Client, error) {
	token, err := readEnv(CloudflareEnvConfig.Token)
	if err != nil {
		return nil, fmt.Errorf("cloudflare auth: %w", err)
	}

	return &http.Client{
		Transport: &cloudflareBearerTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}, nil
}
