package auth

import (
	"os"
	"strings"
)

const (
	defaultCanvasUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// CanvasEnvConfig holds the environment variable names for Canvas LMS session auth.
var CanvasEnvConfig = struct {
	BaseURL   string
	Cookies   string
	UserAgent string
}{
	BaseURL:   "CANVAS_BASE_URL",
	Cookies:   "CANVAS_COOKIES",
	UserAgent: "CANVAS_USER_AGENT",
}

// CanvasSession holds the cookie-based credentials required to authenticate
// requests to the Canvas LMS API.
type CanvasSession struct {
	BaseURL   string // Canvas instance URL, e.g. https://canvas.university.edu
	Cookies   string // Full cookie string (all cookies as "name=value; name2=value2; ...")
	UserAgent string
}

// NewCanvasSession reads Canvas session credentials from environment variables.
// Required: CANVAS_BASE_URL, CANVAS_COOKIES.
// Optional: CANVAS_USER_AGENT.
func NewCanvasSession() (*CanvasSession, error) {
	baseURL, err := readEnv(CanvasEnvConfig.BaseURL)
	if err != nil {
		return nil, err
	}
	cookies, err := readEnv(CanvasEnvConfig.Cookies)
	if err != nil {
		return nil, err
	}

	userAgent := os.Getenv(CanvasEnvConfig.UserAgent)
	if userAgent == "" {
		userAgent = defaultCanvasUserAgent
	}

	return &CanvasSession{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		Cookies:   cookies,
		UserAgent: userAgent,
	}, nil
}

// CookieString returns the full cookie string to send in the Cookie header.
func (s *CanvasSession) CookieString() string {
	return s.Cookies
}
