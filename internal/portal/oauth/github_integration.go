package oauth

import "golang.org/x/oauth2"

// gitHubIntegrationEndpoint is the OAuth2 endpoint for GitHub.
var gitHubIntegrationEndpoint = oauth2.Endpoint{
	AuthURL:  "https://github.com/login/oauth/authorize",
	TokenURL: "https://github.com/login/oauth/access_token",
}

// NewGitHubIntegrationConfig creates an OAuth2 config for GitHub integration
// with scopes covering repos, gists, org read, and workflow access.
func NewGitHubIntegrationConfig(clientID, clientSecret, baseURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     gitHubIntegrationEndpoint,
		RedirectURL:  baseURL + "/integrations/github/callback",
		Scopes:       []string{"repo", "gist", "read:org", "workflow"},
	}
}
