package auth

import (
	"os"
	"testing"
)

func TestNewInstagramSession_Valid(t *testing.T) {
	t.Setenv("INSTAGRAM_SESSION_ID", "abc123session")
	t.Setenv("INSTAGRAM_CSRF_TOKEN", "csrf456token")
	t.Setenv("INSTAGRAM_DS_USER_ID", "9876543210")

	sess, err := NewInstagramSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.SessionID != "abc123session" {
		t.Errorf("SessionID = %q, want %q", sess.SessionID, "abc123session")
	}
	if sess.CSRFToken != "csrf456token" {
		t.Errorf("CSRFToken = %q, want %q", sess.CSRFToken, "csrf456token")
	}
	if sess.DSUserID != "9876543210" {
		t.Errorf("DSUserID = %q, want %q", sess.DSUserID, "9876543210")
	}
	// Optional fields absent: should be empty string.
	if sess.Mid != "" {
		t.Errorf("Mid = %q, want empty", sess.Mid)
	}
	if sess.IgDid != "" {
		t.Errorf("IgDid = %q, want empty", sess.IgDid)
	}
	// UserAgent should default.
	if sess.UserAgent != defaultInstagramUserAgent {
		t.Errorf("UserAgent = %q, want default", sess.UserAgent)
	}
}

func TestNewInstagramSession_WithOptionals(t *testing.T) {
	t.Setenv("INSTAGRAM_SESSION_ID", "sid")
	t.Setenv("INSTAGRAM_CSRF_TOKEN", "csrf")
	t.Setenv("INSTAGRAM_DS_USER_ID", "uid")
	t.Setenv("INSTAGRAM_MID", "mid-value")
	t.Setenv("INSTAGRAM_IG_DID", "did-value")
	t.Setenv("INSTAGRAM_USER_AGENT", "CustomAgent/1.0")

	sess, err := NewInstagramSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.Mid != "mid-value" {
		t.Errorf("Mid = %q, want %q", sess.Mid, "mid-value")
	}
	if sess.IgDid != "did-value" {
		t.Errorf("IgDid = %q, want %q", sess.IgDid, "did-value")
	}
	if sess.UserAgent != "CustomAgent/1.0" {
		t.Errorf("UserAgent = %q, want %q", sess.UserAgent, "CustomAgent/1.0")
	}
}

func TestNewInstagramSession_MissingSessionID(t *testing.T) {
	os.Unsetenv("INSTAGRAM_SESSION_ID")
	t.Setenv("INSTAGRAM_CSRF_TOKEN", "csrf")
	t.Setenv("INSTAGRAM_DS_USER_ID", "uid")

	_, err := NewInstagramSession()
	if err == nil {
		t.Fatal("expected error for missing INSTAGRAM_SESSION_ID, got nil")
	}
}

func TestNewInstagramSession_MissingCSRFToken(t *testing.T) {
	t.Setenv("INSTAGRAM_SESSION_ID", "sid")
	os.Unsetenv("INSTAGRAM_CSRF_TOKEN")
	t.Setenv("INSTAGRAM_DS_USER_ID", "uid")

	_, err := NewInstagramSession()
	if err == nil {
		t.Fatal("expected error for missing INSTAGRAM_CSRF_TOKEN, got nil")
	}
}

func TestNewInstagramSession_MissingDSUserID(t *testing.T) {
	t.Setenv("INSTAGRAM_SESSION_ID", "sid")
	t.Setenv("INSTAGRAM_CSRF_TOKEN", "csrf")
	os.Unsetenv("INSTAGRAM_DS_USER_ID")

	_, err := NewInstagramSession()
	if err == nil {
		t.Fatal("expected error for missing INSTAGRAM_DS_USER_ID, got nil")
	}
}

func TestInstagramSession_CookieString_RequiredOnly(t *testing.T) {
	sess := &InstagramSession{
		SessionID: "mysession",
		CSRFToken: "mycsrf",
		DSUserID:  "myuid",
	}
	got := sess.CookieString()
	want := "sessionid=mysession; csrftoken=mycsrf; ds_user_id=myuid"
	if got != want {
		t.Errorf("CookieString() = %q, want %q", got, want)
	}
}

func TestInstagramSession_CookieString_WithOptionals(t *testing.T) {
	sess := &InstagramSession{
		SessionID: "s",
		CSRFToken: "c",
		DSUserID:  "u",
		Mid:       "m",
		IgDid:     "d",
	}
	got := sess.CookieString()
	want := "sessionid=s; csrftoken=c; ds_user_id=u; mid=m; ig_did=d"
	if got != want {
		t.Errorf("CookieString() = %q, want %q", got, want)
	}
}

func TestRedactSession(t *testing.T) {
	sess := &InstagramSession{
		SessionID: "abc123456789",
		CSRFToken: "csrf987654321",
		DSUserID:  "uid111222333",
	}
	got := redactSession(sess)
	// Should not contain the full values.
	for _, full := range []string{"abc123456789", "csrf987654321", "uid111222333"} {
		for _, substr := range []string{full} {
			if len(got) > 0 && containsSubstring(got, substr) {
				t.Errorf("redactSession output %q should not contain full value %q", got, substr)
			}
		}
	}
	// Should contain the last 4 chars.
	for _, tail := range []string{"6789", "4321", "2333"} {
		if !containsSubstring(got, tail) {
			t.Errorf("redactSession output %q should contain tail %q", got, tail)
		}
	}
}

// containsSubstring is a simple helper to avoid importing strings in tests.
func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
