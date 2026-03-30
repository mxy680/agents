package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// SupabaseEnvConfig holds the environment variable names for Supabase OAuth credentials.
var SupabaseEnvConfig = struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
	BaseURL      string
}{
	ClientID:     "SUPABASE_INTEGRATION_CLIENT_ID",
	ClientSecret: "SUPABASE_INTEGRATION_CLIENT_SECRET",
	AccessToken:  "SUPABASE_ACCESS_TOKEN",
	RefreshToken: "SUPABASE_REFRESH_TOKEN",
	BaseURL:      "SUPABASE_API_BASE_URL",
}

// supabaseEndpoint is the OAuth2 endpoint for Supabase Management API.
var supabaseEndpoint = oauth2.Endpoint{
	AuthURL:  "https://api.supabase.com/v1/oauth/authorize",
	TokenURL: "https://api.supabase.com/v1/oauth/token",
}

// SupabaseBaseURL returns the Supabase Management API base URL from the environment,
// defaulting to https://api.supabase.com.
func SupabaseBaseURL() string {
	if u := os.Getenv(SupabaseEnvConfig.BaseURL); u != "" {
		return u
	}
	return "https://api.supabase.com"
}

// supabaseHeaderTransport wraps an http.RoundTripper to inject Supabase API headers.
type supabaseHeaderTransport struct {
	base http.RoundTripper
}

func (t *supabaseHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Accept", "application/json")
	return t.base.RoundTrip(req)
}

// newSupabaseOAuthConfig builds an oauth2.Config from environment variables.
func newSupabaseOAuthConfig() (*oauth2.Config, error) {
	clientID, err := readEnv(SupabaseEnvConfig.ClientID)
	if err != nil {
		return nil, err
	}
	clientSecret, err := readEnv(SupabaseEnvConfig.ClientSecret)
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     supabaseEndpoint,
		Scopes:       []string{"all"},
	}, nil
}

// newSupabaseToken builds an oauth2.Token from environment variables.
func newSupabaseToken() (*oauth2.Token, error) {
	accessToken, err := readEnv(SupabaseEnvConfig.AccessToken)
	if err != nil {
		return nil, err
	}
	refreshToken := os.Getenv(SupabaseEnvConfig.RefreshToken) // optional

	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		// Set Expiry to the past so the oauth2 library always refreshes on first use.
		Expiry: time.Now().Add(-time.Minute),
	}, nil
}

// NewSupabaseClient creates an authenticated HTTP client for the Supabase Management API.
// It uses OAuth2 with automatic token refresh and injects required headers.
func NewSupabaseClient(ctx context.Context) (*http.Client, error) {
	config, err := newSupabaseOAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("supabase oauth config: %w", err)
	}
	token, err := newSupabaseToken()
	if err != nil {
		return nil, fmt.Errorf("supabase oauth token: %w", err)
	}

	baseSource := config.TokenSource(ctx, token)
	notifySource := &tokenNotifySource{
		base:         baseSource,
		lastToken:    token.AccessToken,
		refreshToken: token.RefreshToken,
	}
	oauthClient := oauth2.NewClient(ctx, notifySource)

	// Wrap the OAuth client's transport with Supabase-specific headers
	oauthClient.Transport = &supabaseHeaderTransport{base: oauthClient.Transport}

	return oauthClient, nil
}
