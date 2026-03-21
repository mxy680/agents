package auth

import (
	"os"
	"testing"
)

func TestNewBlueBubblesCredentials_Valid(t *testing.T) {
	t.Setenv("BLUEBUBBLES_URL", "https://my-mac.ngrok.io")
	t.Setenv("BLUEBUBBLES_PASSWORD", "supersecret123")

	creds, err := NewBlueBubblesCredentials()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.URL != "https://my-mac.ngrok.io" {
		t.Errorf("URL = %q, want %q", creds.URL, "https://my-mac.ngrok.io")
	}
	if creds.Password != "supersecret123" {
		t.Errorf("Password = %q, want %q", creds.Password, "supersecret123")
	}
}

func TestNewBlueBubblesCredentials_MissingURL(t *testing.T) {
	os.Unsetenv("BLUEBUBBLES_URL")
	t.Setenv("BLUEBUBBLES_PASSWORD", "pass")

	_, err := NewBlueBubblesCredentials()
	if err == nil {
		t.Fatal("expected error for missing BLUEBUBBLES_URL, got nil")
	}
}

func TestNewBlueBubblesCredentials_MissingPassword(t *testing.T) {
	t.Setenv("BLUEBUBBLES_URL", "https://example.com")
	os.Unsetenv("BLUEBUBBLES_PASSWORD")

	_, err := NewBlueBubblesCredentials()
	if err == nil {
		t.Fatal("expected error for missing BLUEBUBBLES_PASSWORD, got nil")
	}
}

func TestRedactBlueBubblesCredentials(t *testing.T) {
	creds := &BlueBubblesCredentials{
		URL:      "https://my-mac.ngrok.io",
		Password: "supersecret123",
	}
	got := redactBlueBubblesCredentials(creds)
	// Should not contain full password.
	if containsSubstring(got, "supersecret123") {
		t.Errorf("redact output %q should not contain full password", got)
	}
	// Should contain tail of password.
	if !containsSubstring(got, "t123") {
		t.Errorf("redact output %q should contain tail %q", got, "t123")
	}
}
