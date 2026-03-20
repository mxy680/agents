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
	BCookie    string
	Lidc       string
	LiMc       string
	UserAgent  string
}{
	LiAt:       "LINKEDIN_LI_AT",
	JSESSIONID: "LINKEDIN_JSESSIONID",
	BCookie:    "LINKEDIN_BCOOKIE",
	Lidc:       "LINKEDIN_LIDC",
	LiMc:       "LINKEDIN_LI_MC",
	UserAgent:  "LINKEDIN_USER_AGENT",
}

// LinkedInSession holds the cookie-based credentials required to authenticate
// requests to the LinkedIn Voyager API.
type LinkedInSession struct {
	LiAt       string // li_at cookie (primary session token)
	JSESSIONID string // JSESSIONID cookie (also used as CSRF token)
	BCookie    string // bcookie (browser ID, optional but reduces bot detection)
	Lidc       string // lidc cookie (optional, reduces bot detection)
	LiMc       string // li_mc cookie (optional, reduces bot detection)
	UserAgent  string
}

// NewLinkedInSession reads LinkedIn session credentials from environment variables.
// Required: LINKEDIN_LI_AT, LINKEDIN_JSESSIONID.
// Optional: LINKEDIN_BCOOKIE, LINKEDIN_LIDC, LINKEDIN_LI_MC, LINKEDIN_USER_AGENT.
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
		BCookie:    os.Getenv(LinkedInEnvConfig.BCookie),
		Lidc:       os.Getenv(LinkedInEnvConfig.Lidc),
		LiMc:       os.Getenv(LinkedInEnvConfig.LiMc),
		UserAgent:  userAgent,
	}, nil
}

// CookieString builds the Cookie header value from the session fields.
func (s *LinkedInSession) CookieString() string {
	cookie := fmt.Sprintf("li_at=%s; JSESSIONID=\"%s\"", s.LiAt, s.JSESSIONID)
	if s.BCookie != "" {
		cookie += fmt.Sprintf("; bcookie=\"%s\"", s.BCookie)
	}
	if s.Lidc != "" {
		cookie += fmt.Sprintf("; lidc=\"%s\"", s.Lidc)
	}
	if s.LiMc != "" {
		cookie += fmt.Sprintf("; li_mc=\"%s\"", s.LiMc)
	}
	return cookie
}

// redactLinkedInSession returns a log-safe representation of the session.
func redactLinkedInSession(s *LinkedInSession) string {
	return fmt.Sprintf(
		"LinkedInSession{li_at=%s, jsessionid=%s}",
		redact(s.LiAt),
		redact(s.JSESSIONID),
	)
}
