package auth

import (
	"os"
	"testing"
)

func TestNewLinkedInSession_Valid(t *testing.T) {
	t.Setenv("LINKEDIN_LI_AT", "abc123session")
	t.Setenv("LINKEDIN_JSESSIONID", "ajax:9876543210")

	sess, err := NewLinkedInSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.LiAt != "abc123session" {
		t.Errorf("LiAt = %q, want %q", sess.LiAt, "abc123session")
	}
	if sess.JSESSIONID != "ajax:9876543210" {
		t.Errorf("JSESSIONID = %q, want %q", sess.JSESSIONID, "ajax:9876543210")
	}
	// UserAgent should default.
	if sess.UserAgent != defaultLinkedInUserAgent {
		t.Errorf("UserAgent = %q, want default", sess.UserAgent)
	}
}

func TestNewLinkedInSession_WithOptionals(t *testing.T) {
	t.Setenv("LINKEDIN_LI_AT", "li_at_val")
	t.Setenv("LINKEDIN_JSESSIONID", "jsid_val")
	t.Setenv("LINKEDIN_USER_AGENT", "CustomAgent/1.0")

	sess, err := NewLinkedInSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.UserAgent != "CustomAgent/1.0" {
		t.Errorf("UserAgent = %q, want %q", sess.UserAgent, "CustomAgent/1.0")
	}
}

func TestNewLinkedInSession_MissingLiAt(t *testing.T) {
	os.Unsetenv("LINKEDIN_LI_AT")
	t.Setenv("LINKEDIN_JSESSIONID", "jsid")

	_, err := NewLinkedInSession()
	if err == nil {
		t.Fatal("expected error for missing LINKEDIN_LI_AT, got nil")
	}
}

func TestNewLinkedInSession_MissingJSESSIONID(t *testing.T) {
	t.Setenv("LINKEDIN_LI_AT", "li_at")
	os.Unsetenv("LINKEDIN_JSESSIONID")

	_, err := NewLinkedInSession()
	if err == nil {
		t.Fatal("expected error for missing LINKEDIN_JSESSIONID, got nil")
	}
}

func TestLinkedInSession_CookieString(t *testing.T) {
	sess := &LinkedInSession{
		LiAt:       "mysession",
		JSESSIONID: "ajax:12345",
	}
	got := sess.CookieString()
	want := `li_at=mysession; JSESSIONID="ajax:12345"`
	if got != want {
		t.Errorf("CookieString() = %q, want %q", got, want)
	}
}

func TestRedactLinkedInSession(t *testing.T) {
	sess := &LinkedInSession{
		LiAt:       "abc123456789",
		JSESSIONID: "ajax:987654321",
	}
	got := redactLinkedInSession(sess)
	// Should not contain the full values.
	for _, full := range []string{"abc123456789", "ajax:987654321"} {
		if containsSubstring(got, full) {
			t.Errorf("redactLinkedInSession output %q should not contain full value %q", got, full)
		}
	}
	// Should contain the last 4 chars.
	for _, tail := range []string{"6789", "4321"} {
		if !containsSubstring(got, tail) {
			t.Errorf("redactLinkedInSession output %q should contain tail %q", got, tail)
		}
	}
}
