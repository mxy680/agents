package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
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
	AccessToken:  "GOOGLE_ACCESS_TOKEN",
	RefreshToken: "GOOGLE_REFRESH_TOKEN",
}

// readEnv reads a required environment variable or returns an error.
func readEnv(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", fmt.Errorf("required environment variable %s is not set", name)
	}
	return val, nil
}

// readEnvWithFallback reads an environment variable with a fallback name.
// Returns the value from the primary name if set, otherwise the fallback name.
func readEnvWithFallback(primary, fallback string) (string, error) {
	val := os.Getenv(primary)
	if val != "" {
		return val, nil
	}
	val = os.Getenv(fallback)
	if val != "" {
		return val, nil
	}
	return "", fmt.Errorf("required environment variable %s (or %s) is not set", primary, fallback)
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
			gmail.MailGoogleComScope,
			gmail.GmailSettingsBasicScope,
			gmail.GmailSettingsSharingScope,
			sheets.SpreadsheetsScope,
			drive.DriveFileScope,
		},
	}, nil
}

// NewToken builds an oauth2.Token from environment variables.
// Reads GOOGLE_ACCESS_TOKEN / GOOGLE_REFRESH_TOKEN with fallback to
// GMAIL_ACCESS_TOKEN / GMAIL_REFRESH_TOKEN for backward compatibility.
func NewToken() (*oauth2.Token, error) {
	accessToken, err := readEnvWithFallback(EnvConfig.AccessToken, "GMAIL_ACCESS_TOKEN")
	if err != nil {
		return nil, err
	}
	refreshToken, err := readEnvWithFallback(EnvConfig.RefreshToken, "GMAIL_REFRESH_TOKEN")
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

// tokenNotifySource wraps a token source and logs refresh events to stderr.
// Token values are redacted — only the event is logged, never the credential.
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
		fmt.Fprintln(os.Stderr, "TOKEN_REFRESHED: access_token refreshed")
		if tok.RefreshToken != "" && tok.RefreshToken != t.refreshToken {
			t.refreshToken = tok.RefreshToken
			fmt.Fprintln(os.Stderr, "TOKEN_REFRESHED: refresh_token rotated")
		}
	}
	return tok, nil
}

// newAuthenticatedClient creates an OAuth2 HTTP client from environment variables.
func newAuthenticatedClient(ctx context.Context) (*http.Client, error) {
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
	return oauth2.NewClient(ctx, notifySource), nil
}

// NewGmailService creates an authenticated Gmail API service from environment variables.
func NewGmailService(ctx context.Context) (*gmail.Service, error) {
	client, err := newAuthenticatedClient(ctx)
	if err != nil {
		return nil, err
	}
	return gmail.NewService(ctx, option.WithHTTPClient(client))
}

// NewSheetsService creates an authenticated Google Sheets API service from environment variables.
func NewSheetsService(ctx context.Context) (*sheets.Service, error) {
	client, err := newAuthenticatedClient(ctx)
	if err != nil {
		return nil, err
	}
	return sheets.NewService(ctx, option.WithHTTPClient(client))
}

// NewDriveService creates an authenticated Google Drive API service from environment variables.
func NewDriveService(ctx context.Context) (*drive.Service, error) {
	client, err := newAuthenticatedClient(ctx)
	if err != nil {
		return nil, err
	}
	return drive.NewService(ctx, option.WithHTTPClient(client))
}
