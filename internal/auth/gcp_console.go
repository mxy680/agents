package auth

import (
	"crypto/sha1" //nolint:gosec // SAPISIDHASH uses SHA1 by design (Google's protocol)
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// GCPConsoleEnvConfig holds the environment variable names for GCP Console cookie auth.
var GCPConsoleEnvConfig = struct {
	AllCookies string
	SAPISID    string
}{
	AllCookies: "GCP_CONSOLE_ALL_COOKIES",
	SAPISID:    "GCP_CONSOLE_SAPISID",
}

// GCPConsoleSession holds the cookie-based credentials required to authenticate
// requests to the GCP Console internal API via SAPISIDHASH.
type GCPConsoleSession struct {
	SAPISID    string // required: used to compute the SAPISIDHASH
	AllCookies string // required: full cookie string including HttpOnly cookies
}

// NewGCPConsoleSession reads GCP Console cookie credentials from environment variables.
// Requires GCP_CONSOLE_ALL_COOKIES (full cookie string from extension).
// SAPISID can be set separately or extracted from AllCookies.
func NewGCPConsoleSession() (*GCPConsoleSession, error) {
	allCookies := os.Getenv(GCPConsoleEnvConfig.AllCookies)
	sapisid := os.Getenv(GCPConsoleEnvConfig.SAPISID)

	// Extract SAPISID from AllCookies if not set separately
	if sapisid == "" && allCookies != "" {
		for _, part := range strings.Split(allCookies, "; ") {
			if strings.HasPrefix(part, "SAPISID=") {
				sapisid = strings.TrimPrefix(part, "SAPISID=")
				break
			}
		}
	}

	if sapisid == "" {
		return nil, fmt.Errorf("gcp console auth: SAPISID not found — set GCP_CONSOLE_SAPISID or include SAPISID in GCP_CONSOLE_ALL_COOKIES")
	}

	if allCookies == "" {
		return nil, fmt.Errorf("gcp console auth: GCP_CONSOLE_ALL_COOKIES is required (capture cookies from console.cloud.google.com via the extension)")
	}

	return &GCPConsoleSession{
		SAPISID:    sapisid,
		AllCookies: allCookies,
	}, nil
}

// SAPISIDHash computes the Authorization header value for the GCP Console API.
// Format: SAPISIDHASH {timestamp}_{sha1("{timestamp} {SAPISID} {origin}")}
func (s *GCPConsoleSession) SAPISIDHash() string {
	ts := time.Now().Unix()
	return computeSAPISIDHash(ts, s.SAPISID)
}

// computeSAPISIDHash is the pure function used for both production and testing.
func computeSAPISIDHash(timestamp int64, sapisid string) string {
	origin := "https://console.cloud.google.com"
	input := fmt.Sprintf("%d %s %s", timestamp, sapisid, origin)
	//nolint:gosec // SHA1 is required by Google's SAPISIDHASH protocol — not a security choice
	h := sha1.Sum([]byte(input))
	return fmt.Sprintf("SAPISIDHASH %d_%x", timestamp, h)
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
	req.Header.Set("Cookie", t.session.AllCookies)
	req.Header.Set("X-Goog-AuthUser", "0")
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
