package auth

import (
	"fmt"
	"os"
)

const (
	defaultXUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
)

// XEnvConfig holds the environment variable names for X (Twitter) session auth.
var XEnvConfig = struct {
	AuthToken string
	CSRFToken string
	UserAgent string
}{
	AuthToken: "X_AUTH_TOKEN",
	CSRFToken: "X_CSRF_TOKEN",
	UserAgent: "X_USER_AGENT",
}

// XSession holds the cookie-based credentials required to authenticate
// requests to the X (Twitter) internal API.
type XSession struct {
	AuthToken string // auth_token cookie (primary session token)
	CSRFToken string // ct0 cookie (also sent as X-CSRF-Token header)
	UserAgent string
}

// NewXSession reads X session credentials from environment variables.
// Required: X_AUTH_TOKEN, X_CSRF_TOKEN.
// Optional: X_USER_AGENT.
func NewXSession() (*XSession, error) {
	authToken, err := readEnv(XEnvConfig.AuthToken)
	if err != nil {
		return nil, err
	}
	csrfToken, err := readEnv(XEnvConfig.CSRFToken)
	if err != nil {
		return nil, err
	}

	userAgent := os.Getenv(XEnvConfig.UserAgent)
	if userAgent == "" {
		userAgent = defaultXUserAgent
	}

	return &XSession{
		AuthToken: authToken,
		CSRFToken: csrfToken,
		UserAgent: userAgent,
	}, nil
}

// CookieString builds the Cookie header value from the session fields.
func (s *XSession) CookieString() string {
	return fmt.Sprintf("auth_token=%s; ct0=%s", s.AuthToken, s.CSRFToken)
}

// redactXSession returns a log-safe representation of the session.
func redactXSession(s *XSession) string {
	return fmt.Sprintf(
		"XSession{auth_token=%s, ct0=%s}",
		redact(s.AuthToken),
		redact(s.CSRFToken),
	)
}
