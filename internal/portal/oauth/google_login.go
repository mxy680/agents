package oauth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// NewGoogleLoginConfig creates an OAuth2 config for Google login with
// narrow scopes: openid, profile, and email.
func NewGoogleLoginConfig(clientID, clientSecret, baseURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  baseURL + "/auth/google/callback",
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}
}
