package auth

import (
	"os"
	"testing"
)

func TestNewYelpSession_MissingBSE(t *testing.T) {
	os.Unsetenv("YELP_BSE")
	_, err := NewYelpSession()
	if err == nil {
		t.Fatal("expected error when YELP_BSE is missing")
	}
}

func TestNewYelpSession_OK(t *testing.T) {
	t.Setenv("YELP_BSE", "test-bse-value")
	t.Setenv("YELP_ZSS", "test-zss-value")
	t.Setenv("YELP_CSRF_TOKEN", "test-csrf")

	sess, err := NewYelpSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.BSE != "test-bse-value" {
		t.Errorf("BSE = %q, want %q", sess.BSE, "test-bse-value")
	}
	if sess.ZSS != "test-zss-value" {
		t.Errorf("ZSS = %q, want %q", sess.ZSS, "test-zss-value")
	}
	if sess.CSRFToken != "test-csrf" {
		t.Errorf("CSRFToken = %q, want %q", sess.CSRFToken, "test-csrf")
	}
	if sess.UserAgent == "" {
		t.Error("expected default user agent")
	}
}

func TestYelpSession_CookieString(t *testing.T) {
	sess := &YelpSession{
		BSE:       "bse123",
		ZSS:       "zss456",
		CSRFToken: "csrf789",
	}
	got := sess.CookieString()
	want := "bse=bse123; zss=zss456; csrftok=csrf789"
	if got != want {
		t.Errorf("CookieString() = %q, want %q", got, want)
	}
}

func TestYelpSession_CookieString_BSEOnly(t *testing.T) {
	sess := &YelpSession{
		BSE: "bse123",
	}
	got := sess.CookieString()
	want := "bse=bse123"
	if got != want {
		t.Errorf("CookieString() = %q, want %q", got, want)
	}
}

func TestRedactYelpSession(t *testing.T) {
	sess := &YelpSession{BSE: "abcdef1234567890", ZSS: "zss1234567890"}
	s := redactYelpSession(sess)
	if s == "" {
		t.Error("expected non-empty redacted string")
	}
}
