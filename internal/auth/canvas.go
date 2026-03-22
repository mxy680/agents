package auth

import (
	"fmt"
	"os"
	"strings"
)

const (
	defaultCanvasUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// CanvasEnvConfig holds the environment variable names for Canvas LMS session auth.
var CanvasEnvConfig = struct {
	BaseURL       string
	SessionCookie string
	CSRFToken     string
	LogSessionID  string
	UserAgent     string
}{
	BaseURL:       "CANVAS_BASE_URL",
	SessionCookie: "CANVAS_SESSION_COOKIE",
	CSRFToken:     "CANVAS_CSRF_TOKEN",
	LogSessionID:  "CANVAS_LOG_SESSION_ID",
	UserAgent:     "CANVAS_USER_AGENT",
}

// CanvasSession holds the cookie-based credentials required to authenticate
// requests to the Canvas LMS API.
type CanvasSession struct {
	BaseURL       string // Canvas instance URL, e.g. https://canvas.university.edu
	SessionCookie string // _normandy_session cookie (primary session token)
	CSRFToken     string // _csrf_token cookie (for mutating requests)
	LogSessionID  string // log_session_id cookie (optional, for logging)
	UserAgent     string
}

// NewCanvasSession reads Canvas session credentials from environment variables.
// Required: CANVAS_BASE_URL, CANVAS_SESSION_COOKIE, CANVAS_CSRF_TOKEN.
// Optional: CANVAS_LOG_SESSION_ID, CANVAS_USER_AGENT.
func NewCanvasSession() (*CanvasSession, error) {
	baseURL, err := readEnv(CanvasEnvConfig.BaseURL)
	if err != nil {
		return nil, err
	}
	sessionCookie, err := readEnv(CanvasEnvConfig.SessionCookie)
	if err != nil {
		return nil, err
	}
	csrfToken, err := readEnv(CanvasEnvConfig.CSRFToken)
	if err != nil {
		return nil, err
	}

	logSessionID := os.Getenv(CanvasEnvConfig.LogSessionID)
	userAgent := os.Getenv(CanvasEnvConfig.UserAgent)
	if userAgent == "" {
		userAgent = defaultCanvasUserAgent
	}

	return &CanvasSession{
		BaseURL:       strings.TrimRight(baseURL, "/"),
		SessionCookie: sessionCookie,
		CSRFToken:     csrfToken,
		LogSessionID:  logSessionID,
		UserAgent:     userAgent,
	}, nil
}

// CookieString builds the Cookie header value from the session fields.
func (s *CanvasSession) CookieString() string {
	cookies := fmt.Sprintf("_normandy_session=%s; _csrf_token=%s", s.SessionCookie, s.CSRFToken)
	if s.LogSessionID != "" {
		cookies += fmt.Sprintf("; log_session_id=%s", s.LogSessionID)
	}
	return cookies
}

// redactCanvasSession returns a log-safe representation of the session.
func redactCanvasSession(s *CanvasSession) string {
	return fmt.Sprintf(
		"CanvasSession{base_url=%s, session=%s, csrf=%s}",
		redact(s.BaseURL),
		redact(s.SessionCookie),
		redact(s.CSRFToken),
	)
}
