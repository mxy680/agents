package auth

import (
	"os"
	"testing"
)

func TestNewXSession_Valid(t *testing.T) {
	t.Setenv("X_AUTH_TOKEN", "abc123authtoken")
	t.Setenv("X_CSRF_TOKEN", "csrf456token")

	sess, err := NewXSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.AuthToken != "abc123authtoken" {
		t.Errorf("AuthToken = %q, want %q", sess.AuthToken, "abc123authtoken")
	}
	if sess.CSRFToken != "csrf456token" {
		t.Errorf("CSRFToken = %q, want %q", sess.CSRFToken, "csrf456token")
	}
	// UserAgent should default.
	if sess.UserAgent != defaultXUserAgent {
		t.Errorf("UserAgent = %q, want default", sess.UserAgent)
	}
}

func TestNewXSession_WithCustomUserAgent(t *testing.T) {
	t.Setenv("X_AUTH_TOKEN", "auth")
	t.Setenv("X_CSRF_TOKEN", "csrf")
	t.Setenv("X_USER_AGENT", "CustomAgent/1.0")

	sess, err := NewXSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.UserAgent != "CustomAgent/1.0" {
		t.Errorf("UserAgent = %q, want %q", sess.UserAgent, "CustomAgent/1.0")
	}
}

func TestNewXSession_MissingAuthToken(t *testing.T) {
	os.Unsetenv("X_AUTH_TOKEN")
	t.Setenv("X_CSRF_TOKEN", "csrf")

	_, err := NewXSession()
	if err == nil {
		t.Fatal("expected error for missing X_AUTH_TOKEN, got nil")
	}
}

func TestNewXSession_MissingCSRFToken(t *testing.T) {
	t.Setenv("X_AUTH_TOKEN", "auth")
	os.Unsetenv("X_CSRF_TOKEN")

	_, err := NewXSession()
	if err == nil {
		t.Fatal("expected error for missing X_CSRF_TOKEN, got nil")
	}
}

func TestXSession_CookieString(t *testing.T) {
	sess := &XSession{
		AuthToken: "myauth",
		CSRFToken: "mycsrf",
	}
	got := sess.CookieString()
	want := "auth_token=myauth; ct0=mycsrf"
	if got != want {
		t.Errorf("CookieString() = %q, want %q", got, want)
	}
}

func TestRedactXSession(t *testing.T) {
	sess := &XSession{
		AuthToken: "abc123456789",
		CSRFToken: "csrf987654321",
	}
	got := redactXSession(sess)
	// Should not contain the full values.
	for _, full := range []string{"abc123456789", "csrf987654321"} {
		if containsSubstring(got, full) {
			t.Errorf("redactXSession output %q should not contain full value %q", got, full)
		}
	}
	// Should contain the last 4 chars.
	for _, tail := range []string{"6789", "4321"} {
		if !containsSubstring(got, tail) {
			t.Errorf("redactXSession output %q should contain tail %q", got, tail)
		}
	}
}
