package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

// GitHubEnvConfig holds the environment variable names for GitHub OAuth credentials.
var GitHubEnvConfig = struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
	BaseURL      string
}{
	ClientID:     "GITHUB_CLIENT_ID",
	ClientSecret: "GITHUB_CLIENT_SECRET",
	AccessToken:  "GITHUB_ACCESS_TOKEN",
	RefreshToken: "GITHUB_REFRESH_TOKEN",
	BaseURL:      "GITHUB_API_BASE_URL",
}

// gitHubEndpoint is the OAuth2 endpoint for GitHub.
var gitHubEndpoint = oauth2.Endpoint{
	AuthURL:  "https://github.com/login/oauth/authorize",
	TokenURL: "https://github.com/login/oauth/access_token",
}

// GitHubBaseURL returns the GitHub API base URL from the environment,
// defaulting to https://api.github.com.
func GitHubBaseURL() string {
	if u := os.Getenv(GitHubEnvConfig.BaseURL); u != "" {
		return u
	}
	return "https://api.github.com"
}

// githubHeaderTransport wraps an http.RoundTripper to inject GitHub API headers.
type githubHeaderTransport struct {
	base http.RoundTripper
}

func (t *githubHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return t.base.RoundTrip(req)
}

// newGitHubOAuthConfig builds an oauth2.Config from environment variables.
func newGitHubOAuthConfig() (*oauth2.Config, error) {
	clientID, err := readEnv(GitHubEnvConfig.ClientID)
	if err != nil {
		return nil, err
	}
	clientSecret, err := readEnv(GitHubEnvConfig.ClientSecret)
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     gitHubEndpoint,
		Scopes:       []string{"repo", "gist", "read:org", "workflow"},
	}, nil
}

// newGitHubToken builds an oauth2.Token from environment variables.
// RefreshToken is optional — GitHub OAuth Apps issue non-expiring tokens.
func newGitHubToken() (*oauth2.Token, error) {
	accessToken, err := readEnv(GitHubEnvConfig.AccessToken)
	if err != nil {
		return nil, err
	}
	refreshToken := os.Getenv(GitHubEnvConfig.RefreshToken) // optional

	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// NewGitHubClient creates an authenticated HTTP client for the GitHub API.
// It uses OAuth2 with automatic token refresh and injects required GitHub headers.
func NewGitHubClient(ctx context.Context) (*http.Client, error) {
	config, err := newGitHubOAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("github oauth config: %w", err)
	}
	token, err := newGitHubToken()
	if err != nil {
		return nil, fmt.Errorf("github oauth token: %w", err)
	}

	baseSource := config.TokenSource(ctx, token)
	notifySource := &tokenNotifySource{
		base:         baseSource,
		lastToken:    token.AccessToken,
		refreshToken: token.RefreshToken,
	}
	oauthClient := oauth2.NewClient(ctx, notifySource)

	// Wrap the OAuth client's transport with GitHub-specific headers
	oauthClient.Transport = &githubHeaderTransport{base: oauthClient.Transport}

	return oauthClient, nil
}
