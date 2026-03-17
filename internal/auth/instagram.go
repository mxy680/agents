package auth

import (
	"fmt"
	"os"
)

const (
	defaultInstagramUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
)

// InstagramEnvConfig holds the environment variable names for Instagram session auth.
var InstagramEnvConfig = struct {
	SessionID   string
	CSRFToken   string
	DSUserID    string
	Mid         string
	IgDid       string
	UserAgent   string
}{
	SessionID: "INSTAGRAM_SESSION_ID",
	CSRFToken: "INSTAGRAM_CSRF_TOKEN",
	DSUserID:  "INSTAGRAM_DS_USER_ID",
	Mid:       "INSTAGRAM_MID",
	IgDid:     "INSTAGRAM_IG_DID",
	UserAgent: "INSTAGRAM_USER_AGENT",
}

// InstagramSession holds the cookie-based credentials required to authenticate
// requests to the Instagram web API.
type InstagramSession struct {
	SessionID string
	CSRFToken string
	DSUserID  string
	Mid       string // optional
	IgDid     string // optional
	UserAgent string
}

// NewInstagramSession reads Instagram session credentials from environment variables.
// Required: INSTAGRAM_SESSION_ID, INSTAGRAM_CSRF_TOKEN, INSTAGRAM_DS_USER_ID.
// Optional: INSTAGRAM_MID, INSTAGRAM_IG_DID, INSTAGRAM_USER_AGENT.
func NewInstagramSession() (*InstagramSession, error) {
	sessionID, err := readEnv(InstagramEnvConfig.SessionID)
	if err != nil {
		return nil, err
	}
	csrfToken, err := readEnv(InstagramEnvConfig.CSRFToken)
	if err != nil {
		return nil, err
	}
	dsUserID, err := readEnv(InstagramEnvConfig.DSUserID)
	if err != nil {
		return nil, err
	}

	mid := os.Getenv(InstagramEnvConfig.Mid)
	igDid := os.Getenv(InstagramEnvConfig.IgDid)
	userAgent := os.Getenv(InstagramEnvConfig.UserAgent)
	if userAgent == "" {
		userAgent = defaultInstagramUserAgent
	}

	return &InstagramSession{
		SessionID: sessionID,
		CSRFToken: csrfToken,
		DSUserID:  dsUserID,
		Mid:       mid,
		IgDid:     igDid,
		UserAgent: userAgent,
	}, nil
}

// CookieString builds the Cookie header value from the session fields.
func (s *InstagramSession) CookieString() string {
	cookie := fmt.Sprintf("sessionid=%s; csrftoken=%s; ds_user_id=%s", s.SessionID, s.CSRFToken, s.DSUserID)
	if s.Mid != "" {
		cookie += "; mid=" + s.Mid
	}
	if s.IgDid != "" {
		cookie += "; ig_did=" + s.IgDid
	}
	return cookie
}

// redactSession returns a log-safe representation of the session showing only
// partial values so credentials are not leaked to stderr/logs.
func redactSession(s *InstagramSession) string {
	return fmt.Sprintf(
		"InstagramSession{session_id=%s, csrf_token=%s, ds_user_id=%s}",
		redact(s.SessionID),
		redact(s.CSRFToken),
		redact(s.DSUserID),
	)
}
