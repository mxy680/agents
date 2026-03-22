package auth

import (
	"os"
	"testing"
)

func TestNewCanvasSession_Valid(t *testing.T) {
	t.Setenv("CANVAS_BASE_URL", "https://canvas.test.edu")
	t.Setenv("CANVAS_COOKIES", "_normandy_session=abc123; _csrf_token=xyz789")

	sess, err := NewCanvasSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.BaseURL != "https://canvas.test.edu" {
		t.Errorf("BaseURL = %q, want %q", sess.BaseURL, "https://canvas.test.edu")
	}
	if sess.Cookies != "_normandy_session=abc123; _csrf_token=xyz789" {
		t.Errorf("Cookies = %q, want full cookie string", sess.Cookies)
	}
	if sess.UserAgent != defaultCanvasUserAgent {
		t.Errorf("UserAgent = %q, want default", sess.UserAgent)
	}
}

func TestNewCanvasSession_WithCustomUserAgent(t *testing.T) {
	t.Setenv("CANVAS_BASE_URL", "https://canvas.test.edu")
	t.Setenv("CANVAS_COOKIES", "_normandy_session=abc")
	t.Setenv("CANVAS_USER_AGENT", "CustomAgent/1.0")

	sess, err := NewCanvasSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.UserAgent != "CustomAgent/1.0" {
		t.Errorf("UserAgent = %q, want %q", sess.UserAgent, "CustomAgent/1.0")
	}
}

func TestNewCanvasSession_MissingBaseURL(t *testing.T) {
	os.Unsetenv("CANVAS_BASE_URL")
	t.Setenv("CANVAS_COOKIES", "_normandy_session=abc")

	_, err := NewCanvasSession()
	if err == nil {
		t.Fatal("expected error for missing CANVAS_BASE_URL, got nil")
	}
}

func TestNewCanvasSession_MissingCookies(t *testing.T) {
	t.Setenv("CANVAS_BASE_URL", "https://canvas.test.edu")
	os.Unsetenv("CANVAS_COOKIES")

	_, err := NewCanvasSession()
	if err == nil {
		t.Fatal("expected error for missing CANVAS_COOKIES, got nil")
	}
}

func TestCanvasSession_CookieString(t *testing.T) {
	sess := &CanvasSession{
		Cookies: "_normandy_session=abc; _csrf_token=xyz; log_session_id=123",
	}
	got := sess.CookieString()
	want := "_normandy_session=abc; _csrf_token=xyz; log_session_id=123"
	if got != want {
		t.Errorf("CookieString() = %q, want %q", got, want)
	}
}

func TestNewCanvasSession_TrimsTrailingSlash(t *testing.T) {
	t.Setenv("CANVAS_BASE_URL", "https://canvas.test.edu/")
	t.Setenv("CANVAS_COOKIES", "_normandy_session=abc")

	sess, err := NewCanvasSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.BaseURL != "https://canvas.test.edu" {
		t.Errorf("BaseURL = %q, want trailing slash stripped", sess.BaseURL)
	}
}
