package auth

import (
	"fmt"
	"os"
)

const (
	defaultLinkedInUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
)

// LinkedInEnvConfig holds the environment variable names for LinkedIn session auth.
var LinkedInEnvConfig = struct {
	LiAt       string
	JSESSIONID string
	UserAgent  string
}{
	LiAt:       "LINKEDIN_LI_AT",
	JSESSIONID: "LINKEDIN_JSESSIONID",
	UserAgent:  "LINKEDIN_USER_AGENT",
}

// LinkedInSession holds the cookie-based credentials required to authenticate
// requests to the LinkedIn Voyager API.
type LinkedInSession struct {
	LiAt       string // li_at cookie (primary session token)
	JSESSIONID string // JSESSIONID cookie (also used as CSRF token)
	UserAgent  string
}

// NewLinkedInSession reads LinkedIn session credentials from environment variables.
// Required: LINKEDIN_LI_AT, LINKEDIN_JSESSIONID.
// Optional: LINKEDIN_USER_AGENT.
func NewLinkedInSession() (*LinkedInSession, error) {
	liAt, err := readEnv(LinkedInEnvConfig.LiAt)
	if err != nil {
		return nil, err
	}
	jsessionID, err := readEnv(LinkedInEnvConfig.JSESSIONID)
	if err != nil {
		return nil, err
	}

	userAgent := os.Getenv(LinkedInEnvConfig.UserAgent)
	if userAgent == "" {
		userAgent = defaultLinkedInUserAgent
	}

	return &LinkedInSession{
		LiAt:       liAt,
		JSESSIONID: jsessionID,
		UserAgent:  userAgent,
	}, nil
}

// CookieString builds the Cookie header value from the session fields.
func (s *LinkedInSession) CookieString() string {
	return fmt.Sprintf("li_at=%s; JSESSIONID=\"%s\"", s.LiAt, s.JSESSIONID)
}

// redactLinkedInSession returns a log-safe representation of the session.
func redactLinkedInSession(s *LinkedInSession) string {
	return fmt.Sprintf(
		"LinkedInSession{li_at=%s, jsessionid=%s}",
		redact(s.LiAt),
		redact(s.JSESSIONID),
	)
}
