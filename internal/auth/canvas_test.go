package auth

import (
	"os"
	"testing"
)

func TestNewCanvasSession_Valid(t *testing.T) {
	t.Setenv("CANVAS_BASE_URL", "https://canvas.university.edu")
	t.Setenv("CANVAS_SESSION_COOKIE", "abc123session")
	t.Setenv("CANVAS_CSRF_TOKEN", "csrf456token")

	sess, err := NewCanvasSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.BaseURL != "https://canvas.university.edu" {
		t.Errorf("BaseURL = %q, want %q", sess.BaseURL, "https://canvas.university.edu")
	}
	if sess.SessionCookie != "abc123session" {
		t.Errorf("SessionCookie = %q, want %q", sess.SessionCookie, "abc123session")
	}
	if sess.CSRFToken != "csrf456token" {
		t.Errorf("CSRFToken = %q, want %q", sess.CSRFToken, "csrf456token")
	}
	// Optional field absent: should be empty string.
	if sess.LogSessionID != "" {
		t.Errorf("LogSessionID = %q, want empty", sess.LogSessionID)
	}
	// UserAgent should default.
	if sess.UserAgent != defaultCanvasUserAgent {
		t.Errorf("UserAgent = %q, want default", sess.UserAgent)
	}
}

func TestNewCanvasSession_WithCustomUserAgent(t *testing.T) {
	t.Setenv("CANVAS_BASE_URL", "https://canvas.university.edu")
	t.Setenv("CANVAS_SESSION_COOKIE", "session")
	t.Setenv("CANVAS_CSRF_TOKEN", "csrf")
	t.Setenv("CANVAS_USER_AGENT", "CustomAgent/1.0")

	sess, err := NewCanvasSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.UserAgent != "CustomAgent/1.0" {
		t.Errorf("UserAgent = %q, want %q", sess.UserAgent, "CustomAgent/1.0")
	}
}

func TestNewCanvasSession_WithLogSessionID(t *testing.T) {
	t.Setenv("CANVAS_BASE_URL", "https://canvas.university.edu")
	t.Setenv("CANVAS_SESSION_COOKIE", "session")
	t.Setenv("CANVAS_CSRF_TOKEN", "csrf")
	t.Setenv("CANVAS_LOG_SESSION_ID", "logsession123")

	sess, err := NewCanvasSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.LogSessionID != "logsession123" {
		t.Errorf("LogSessionID = %q, want %q", sess.LogSessionID, "logsession123")
	}
}

func TestNewCanvasSession_MissingBaseURL(t *testing.T) {
	os.Unsetenv("CANVAS_BASE_URL")
	t.Setenv("CANVAS_SESSION_COOKIE", "session")
	t.Setenv("CANVAS_CSRF_TOKEN", "csrf")

	_, err := NewCanvasSession()
	if err == nil {
		t.Fatal("expected error for missing CANVAS_BASE_URL, got nil")
	}
}

func TestNewCanvasSession_MissingSessionCookie(t *testing.T) {
	t.Setenv("CANVAS_BASE_URL", "https://canvas.university.edu")
	os.Unsetenv("CANVAS_SESSION_COOKIE")
	t.Setenv("CANVAS_CSRF_TOKEN", "csrf")

	_, err := NewCanvasSession()
	if err == nil {
		t.Fatal("expected error for missing CANVAS_SESSION_COOKIE, got nil")
	}
}

func TestNewCanvasSession_MissingCSRFToken(t *testing.T) {
	t.Setenv("CANVAS_BASE_URL", "https://canvas.university.edu")
	t.Setenv("CANVAS_SESSION_COOKIE", "session")
	os.Unsetenv("CANVAS_CSRF_TOKEN")

	_, err := NewCanvasSession()
	if err == nil {
		t.Fatal("expected error for missing CANVAS_CSRF_TOKEN, got nil")
	}
}

func TestCanvasSession_CookieString(t *testing.T) {
	t.Run("without log_session_id", func(t *testing.T) {
		sess := &CanvasSession{
			SessionCookie: "mysession",
			CSRFToken:     "mycsrf",
		}
		got := sess.CookieString()
		want := "_normandy_session=mysession; _csrf_token=mycsrf"
		if got != want {
			t.Errorf("CookieString() = %q, want %q", got, want)
		}
	})

	t.Run("with log_session_id", func(t *testing.T) {
		sess := &CanvasSession{
			SessionCookie: "mysession",
			CSRFToken:     "mycsrf",
			LogSessionID:  "mylogsession",
		}
		got := sess.CookieString()
		want := "_normandy_session=mysession; _csrf_token=mycsrf; log_session_id=mylogsession"
		if got != want {
			t.Errorf("CookieString() = %q, want %q", got, want)
		}
	})
}

func TestRedactCanvasSession(t *testing.T) {
	sess := &CanvasSession{
		BaseURL:       "https://canvas.university.edu",
		SessionCookie: "abc123456789",
		CSRFToken:     "csrf987654321",
	}
	got := redactCanvasSession(sess)
	// Should not contain the full values.
	for _, full := range []string{"abc123456789", "csrf987654321"} {
		if containsSubstring(got, full) {
			t.Errorf("redactCanvasSession output %q should not contain full value %q", got, full)
		}
	}
	// Should contain the last 4 chars.
	for _, tail := range []string{"6789", "4321"} {
		if !containsSubstring(got, tail) {
			t.Errorf("redactCanvasSession output %q should contain tail %q", got, tail)
		}
	}
}
