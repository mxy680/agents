package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
)

// GCPEnvConfig holds the environment variable names for GCP auth.
var GCPEnvConfig = struct {
	AccessToken        string
	ServiceAccountJSON string
	ProjectID          string
}{
	AccessToken:        "GCP_ACCESS_TOKEN",
	ServiceAccountJSON: "GCP_SERVICE_ACCOUNT_JSON",
	ProjectID:          "GCP_PROJECT_ID",
}

// NewGCPClient creates an authenticated HTTP client for the GCP REST APIs.
// Priority:
//  1. GCP_SERVICE_ACCOUNT_JSON — generates tokens automatically (preferred)
//  2. GCP_ACCESS_TOKEN — uses a pre-generated bearer token (short-lived)
func NewGCPClient(ctx context.Context) (*http.Client, error) {
	// Try service account JSON first
	saJSON := os.Getenv(GCPEnvConfig.ServiceAccountJSON)
	if saJSON != "" {
		creds, err := google.CredentialsFromJSON(ctx, []byte(saJSON),
			"https://www.googleapis.com/auth/cloud-platform",
		)
		if err != nil {
			return nil, fmt.Errorf("gcp auth: parse service account: %w", err)
		}
		return &http.Client{
			Transport: &gcpTokenTransport{
				base:  http.DefaultTransport,
				creds: creds,
			},
		}, nil
	}

	// Fall back to static access token
	token := os.Getenv(GCPEnvConfig.AccessToken)
	if token == "" {
		return nil, fmt.Errorf("gcp auth: set GCP_SERVICE_ACCOUNT_JSON or GCP_ACCESS_TOKEN")
	}

	return &http.Client{
		Transport: &gcpBearerTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}, nil
}

// gcpTokenTransport auto-refreshes tokens from service account credentials.
type gcpTokenTransport struct {
	base  http.RoundTripper
	creds *google.Credentials
}

func (t *gcpTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tok, err := t.creds.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("gcp auth: get token: %w", err)
	}
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

// gcpBearerTransport injects a static bearer token.
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

// GCPDefaultProject returns the default GCP project ID from the environment.
// Falls back to parsing it from the service account JSON.
func GCPDefaultProject() string {
	if p := os.Getenv(GCPEnvConfig.ProjectID); p != "" {
		return p
	}
	saJSON := os.Getenv(GCPEnvConfig.ServiceAccountJSON)
	if saJSON != "" {
		var sa struct {
			ProjectID string `json:"project_id"`
		}
		if json.Unmarshal([]byte(saJSON), &sa) == nil && sa.ProjectID != "" {
			return sa.ProjectID
		}
	}
	return ""
}
