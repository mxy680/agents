package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// GCPEnvConfig holds the environment variable names for GCP bearer token auth.
var GCPEnvConfig = struct {
	AccessToken string
	ProjectID   string
}{
	AccessToken: "GCP_ACCESS_TOKEN",
	ProjectID:   "GCP_PROJECT_ID",
}

// gcpBearerTransport injects the GCP Bearer token into every request.
type gcpBearerTransport struct {
	base  http.RoundTripper
	token string
}

func (t *gcpBearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

// NewGCPClient creates an authenticated HTTP client for the GCP REST APIs.
// Reads GCP_ACCESS_TOKEN (required) from the environment. Accepts a Google
// OAuth access token or a service account token obtained via gcloud.
func NewGCPClient(_ context.Context) (*http.Client, error) {
	token, err := readEnv(GCPEnvConfig.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("gcp auth: %w", err)
	}

	return &http.Client{
		Transport: &gcpBearerTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}, nil
}

// GCPDefaultProject returns the default GCP project ID from the environment.
func GCPDefaultProject() string {
	return os.Getenv(GCPEnvConfig.ProjectID)
}
