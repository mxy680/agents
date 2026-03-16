package auth

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// EnvConfig holds the environment variable names for Google OAuth credentials.
var EnvConfig = struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
}{
	ClientID:     "GOOGLE_DESKTOP_CLIENT_ID",
	ClientSecret: "GOOGLE_DESKTOP_CLIENT_SECRET",
	AccessToken:  "GMAIL_ACCESS_TOKEN",
	RefreshToken: "GMAIL_REFRESH_TOKEN",
}

// readEnv reads a required environment variable or returns an error.
func readEnv(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", fmt.Errorf("required environment variable %s is not set", name)
	}
	return val, nil
}

// NewOAuthConfig builds an oauth2.Config from environment variables.
func NewOAuthConfig() (*oauth2.Config, error) {
	clientID, err := readEnv(EnvConfig.ClientID)
	if err != nil {
		return nil, err
	}
	clientSecret, err := readEnv(EnvConfig.ClientSecret)
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			gmail.GmailModifyScope,
			gmail.GmailSettingsBasicScope,
			gmail.GmailSettingsSharingScope,
		},
	}, nil
}

// NewToken builds an oauth2.Token from environment variables.
func NewToken() (*oauth2.Token, error) {
	accessToken, err := readEnv(EnvConfig.AccessToken)
	if err != nil {
		return nil, err
	}
	refreshToken, err := readEnv(EnvConfig.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		// Leave Expiry zero so the token source will attempt refresh immediately
		// if the access token is expired. The oauth2 library handles this.
	}, nil
}

// tokenNotifySource wraps a token source and prints new tokens to stderr.
type tokenNotifySource struct {
	base         oauth2.TokenSource
	lastToken    string
	refreshToken string
}

func (t *tokenNotifySource) Token() (*oauth2.Token, error) {
	tok, err := t.base.Token()
	if err != nil {
		return nil, err
	}
	if tok.AccessToken != t.lastToken {
		t.lastToken = tok.AccessToken
		fmt.Fprintf(os.Stderr, "TOKEN_REFRESHED: access_token=%s\n", tok.AccessToken)
		if tok.RefreshToken != "" && tok.RefreshToken != t.refreshToken {
			t.refreshToken = tok.RefreshToken
			fmt.Fprintf(os.Stderr, "TOKEN_REFRESHED: refresh_token=%s\n", tok.RefreshToken)
		}
	}
	return tok, nil
}

// NewGmailService creates an authenticated Gmail API service from environment variables.
func NewGmailService(ctx context.Context) (*gmail.Service, error) {
	config, err := NewOAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("oauth config: %w", err)
	}

	token, err := NewToken()
	if err != nil {
		return nil, fmt.Errorf("oauth token: %w", err)
	}

	baseSource := config.TokenSource(ctx, token)
	notifySource := &tokenNotifySource{
		base:         baseSource,
		lastToken:    token.AccessToken,
		refreshToken: token.RefreshToken,
	}

	client := oauth2.NewClient(ctx, notifySource)
	return gmail.NewService(ctx, option.WithHTTPClient(client))
}
