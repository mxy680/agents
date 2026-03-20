package auth

import (
	"os"
	"testing"
)

func TestNewFramerCredentials_Valid(t *testing.T) {
	t.Setenv("FRAMER_API_KEY", "framer-key-abc123")
	t.Setenv("FRAMER_PROJECT_URL", "https://framer.com/projects/Website--aabbccddeeff")

	creds, err := NewFramerCredentials()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.APIKey != "framer-key-abc123" {
		t.Errorf("APIKey = %q, want %q", creds.APIKey, "framer-key-abc123")
	}
	if creds.ProjectURL != "https://framer.com/projects/Website--aabbccddeeff" {
		t.Errorf("ProjectURL = %q, want %q", creds.ProjectURL, "https://framer.com/projects/Website--aabbccddeeff")
	}
}

func TestNewFramerCredentials_MissingAPIKey(t *testing.T) {
	os.Unsetenv("FRAMER_API_KEY")
	t.Setenv("FRAMER_PROJECT_URL", "https://framer.com/projects/Website--aabbccddeeff")

	_, err := NewFramerCredentials()
	if err == nil {
		t.Fatal("expected error for missing FRAMER_API_KEY, got nil")
	}
}

func TestNewFramerCredentials_MissingProjectURL(t *testing.T) {
	t.Setenv("FRAMER_API_KEY", "framer-key-abc123")
	os.Unsetenv("FRAMER_PROJECT_URL")

	_, err := NewFramerCredentials()
	if err == nil {
		t.Fatal("expected error for missing FRAMER_PROJECT_URL, got nil")
	}
}

func TestNewFramerCredentials_BothMissing(t *testing.T) {
	os.Unsetenv("FRAMER_API_KEY")
	os.Unsetenv("FRAMER_PROJECT_URL")

	_, err := NewFramerCredentials()
	if err == nil {
		t.Fatal("expected error when both env vars are missing, got nil")
	}
}

func TestRedactFramerCredentials(t *testing.T) {
	creds := &FramerCredentials{
		APIKey:     "framer-key-abc123456789",
		ProjectURL: "https://framer.com/projects/Website--aabbccddeeff",
	}
	got := redactFramerCredentials(creds)
	// Should not contain the full API key.
	if containsSubstring(got, "framer-key-abc123456789") {
		t.Errorf("redactFramerCredentials output %q should not contain full API key", got)
	}
	// Should contain last 4 chars of the key.
	if !containsSubstring(got, "6789") {
		t.Errorf("redactFramerCredentials output %q should contain tail of API key", got)
	}
}
