package auth

import (
	"fmt"
	"os"
)

const (
	defaultYelpUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
)

// YelpEnvConfig holds the environment variable names for Yelp session auth.
var YelpEnvConfig = struct {
	BSE       string
	ZSS       string
	CSRFToken string
	UserAgent string
}{
	BSE:       "YELP_BSE",
	ZSS:       "YELP_ZSS",
	CSRFToken: "YELP_CSRF_TOKEN",
	UserAgent: "YELP_USER_AGENT",
}

// YelpSession holds the cookie-based credentials required to authenticate
// requests to the Yelp web API.
type YelpSession struct {
	BSE       string // bse cookie (primary session token)
	ZSS       string // zss cookie (secondary session token)
	CSRFToken string // csrftok cookie
	UserAgent string
}

// NewYelpSession reads Yelp session credentials from environment variables.
// Required: YELP_BSE.
// Optional: YELP_ZSS, YELP_CSRF_TOKEN, YELP_USER_AGENT.
func NewYelpSession() (*YelpSession, error) {
	bse, err := readEnv(YelpEnvConfig.BSE)
	if err != nil {
		return nil, err
	}

	zss := os.Getenv(YelpEnvConfig.ZSS)
	csrfToken := os.Getenv(YelpEnvConfig.CSRFToken)

	userAgent := os.Getenv(YelpEnvConfig.UserAgent)
	if userAgent == "" {
		userAgent = defaultYelpUserAgent
	}

	return &YelpSession{
		BSE:       bse,
		ZSS:       zss,
		CSRFToken: csrfToken,
		UserAgent: userAgent,
	}, nil
}

// CookieString builds the Cookie header value from the session fields.
func (s *YelpSession) CookieString() string {
	cookie := "bse=" + s.BSE
	if s.ZSS != "" {
		cookie += "; zss=" + s.ZSS
	}
	if s.CSRFToken != "" {
		cookie += "; csrftok=" + s.CSRFToken
	}
	return cookie
}

// redactYelpSession returns a log-safe representation of the session.
func redactYelpSession(s *YelpSession) string {
	return fmt.Sprintf(
		"YelpSession{bse=%s, zss=%s}",
		redact(s.BSE),
		redact(s.ZSS),
	)
}
