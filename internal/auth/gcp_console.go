package auth

import (
	"crypto/sha1" //nolint:gosec // SAPISIDHASH uses SHA1 by design (Google's protocol)
	"fmt"
	"net/http"
	"os"
	"time"
)

// GCPConsoleEnvConfig holds the environment variable names for GCP Console cookie auth.
var GCPConsoleEnvConfig = struct {
	SAPISID    string
	SID        string
	HSID       string
	SSID       string
	APISID     string
	AllCookies string
}{
	SAPISID:    "GCP_CONSOLE_SAPISID",
	SID:        "GCP_CONSOLE_SID",
	HSID:       "GCP_CONSOLE_HSID",
	SSID:       "GCP_CONSOLE_SSID",
	APISID:     "GCP_CONSOLE_APISID",
	AllCookies: "GCP_CONSOLE_ALL_COOKIES",
}

// GCPConsoleSession holds the cookie-based credentials required to authenticate
// requests to the GCP Console internal API via SAPISIDHASH.
type GCPConsoleSession struct {
	SAPISID    string // required: used to compute the SAPISIDHASH
	SID        string // required: included in Cookie header
	HSID       string // optional
	SSID       string // optional
	APISID     string // optional
	AllCookies string // optional: full cookie string (overrides individual cookies)
}

// NewGCPConsoleSession reads GCP Console cookie credentials from environment variables.
// Required: GCP_CONSOLE_SAPISID, GCP_CONSOLE_SID.
// Optional: GCP_CONSOLE_HSID, GCP_CONSOLE_SSID, GCP_CONSOLE_APISID.
func NewGCPConsoleSession() (*GCPConsoleSession, error) {
	sapisid, err := readEnv(GCPConsoleEnvConfig.SAPISID)
	if err != nil {
		return nil, fmt.Errorf("gcp console auth: %w", err)
	}
	sid, err := readEnv(GCPConsoleEnvConfig.SID)
	if err != nil {
		return nil, fmt.Errorf("gcp console auth: %w", err)
	}

	return &GCPConsoleSession{
		SAPISID:    sapisid,
		SID:        sid,
		HSID:       os.Getenv(GCPConsoleEnvConfig.HSID),
		SSID:       os.Getenv(GCPConsoleEnvConfig.SSID),
		APISID:     os.Getenv(GCPConsoleEnvConfig.APISID),
		AllCookies: os.Getenv(GCPConsoleEnvConfig.AllCookies),
	}, nil
}

// SAPISIDHash computes the Authorization header value for the GCP Console API.
// Format: SAPISIDHASH {timestamp}_{sha1("{timestamp} {SAPISID} https://console.cloud.google.com")}
func (s *GCPConsoleSession) SAPISIDHash() string {
	ts := time.Now().Unix()
	return computeSAPISIDHash(ts, s.SAPISID)
}

// computeSAPISIDHash is the pure function used for both production and testing.
// Google expects multiple hash variants in the Authorization header.
func computeSAPISIDHash(timestamp int64, sapisid string) string {
	origin := "https://console.cloud.google.com"
	input := fmt.Sprintf("%d %s %s", timestamp, sapisid, origin)
	//nolint:gosec // SHA1 is required by Google's SAPISIDHASH protocol — not a security choice
	h := sha1.Sum([]byte(input))
	hash := fmt.Sprintf("%d_%x", timestamp, h)
	return fmt.Sprintf("SAPISIDHASH %s SAPISID1PHASH %s SAPISID3PHASH %s", hash, hash, hash)
}

// CookieString builds the Cookie header value from the session fields.
// If AllCookies is set (full cookie string from extension), it's used directly.
func (s *GCPConsoleSession) CookieString() string {
	if s.AllCookies != "" {
		return s.AllCookies
	}
	cookie := fmt.Sprintf("SAPISID=%s; SID=%s", s.SAPISID, s.SID)
	if s.HSID != "" {
		cookie += "; HSID=" + s.HSID
	}
	if s.SSID != "" {
		cookie += "; SSID=" + s.SSID
	}
	if s.APISID != "" {
		cookie += "; APISID=" + s.APISID
	}
	return cookie
}

// gcpConsoleTransport injects SAPISIDHASH auth headers into every request.
// The hash is computed fresh per request since it embeds the current timestamp.
type gcpConsoleTransport struct {
	base    http.RoundTripper
	session *GCPConsoleSession
}

func (t *gcpConsoleTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", t.session.SAPISIDHash())
	req.Header.Set("Cookie", t.session.CookieString())
	req.Header.Set("Origin", "https://console.cloud.google.com")
	req.Header.Set("Referer", "https://console.cloud.google.com/")
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

// NewGCPConsoleClient creates an authenticated HTTP client for the GCP Console internal API.
// Reads cookie credentials from environment variables.
func NewGCPConsoleClient() (*http.Client, error) {
	session, err := NewGCPConsoleSession()
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &gcpConsoleTransport{
			base:    http.DefaultTransport,
			session: session,
		},
	}, nil
}
