package auth

import (
	"fmt"
	"testing"
)

func TestComputeSAPISIDHash(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		sapisid   string
		wantFmt   string
	}{
		{
			name:      "basic hash",
			timestamp: 1700000000,
			sapisid:   "abc123",
			wantFmt:   "SAPISIDHASH 1700000000_",
		},
		{
			name:      "different timestamp",
			timestamp: 1234567890,
			sapisid:   "xyz789",
			wantFmt:   "SAPISIDHASH 1234567890_",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeSAPISIDHash(tc.timestamp, tc.sapisid)
			if len(got) < len(tc.wantFmt) {
				t.Fatalf("hash too short: got %q", got)
			}
			if got[:len(tc.wantFmt)] != tc.wantFmt {
				t.Errorf("prefix mismatch: want %q, got %q", tc.wantFmt, got[:len(tc.wantFmt)])
			}
			hash := got[len(tc.wantFmt):]
			if len(hash) != 40 {
				t.Errorf("SHA1 hex should be 40 chars, got %d in %q", len(hash), hash)
			}
		})
	}
}

func TestComputeSAPISIDHash_Deterministic(t *testing.T) {
	ts := int64(1700000000)
	sapisid := "test_sapisid_value"
	h1 := computeSAPISIDHash(ts, sapisid)
	h2 := computeSAPISIDHash(ts, sapisid)
	if h1 != h2 {
		t.Errorf("expected deterministic output: %q != %q", h1, h2)
	}
}

func TestComputeSAPISIDHash_DifferentInputs(t *testing.T) {
	ts := int64(1700000000)
	h1 := computeSAPISIDHash(ts, "sapisid_a")
	h2 := computeSAPISIDHash(ts, "sapisid_b")
	if h1 == h2 {
		t.Error("different SAPISIDs should produce different hashes")
	}
}

func TestGCPConsoleSession_SAPISIDHash_Format(t *testing.T) {
	session := &GCPConsoleSession{SAPISID: "test_sapisid", AllCookies: "SAPISID=test_sapisid; SID=test_sid"}
	hash := session.SAPISIDHash()
	if len(hash) < 12 || hash[:12] != "SAPISIDHASH " {
		t.Errorf("SAPISIDHash() should start with 'SAPISIDHASH ', got %q", hash)
	}
	var ts int64
	var hexPart string
	if _, err := fmt.Sscanf(hash[12:], "%d_%s", &ts, &hexPart); err != nil {
		t.Errorf("SAPISIDHash() format invalid: %q, parse error: %v", hash, err)
	}
	if len(hexPart) != 40 {
		t.Errorf("SHA1 hex should be 40 chars, got %d", len(hexPart))
	}
}

func TestNewGCPConsoleSession_MissingAll(t *testing.T) {
	t.Setenv("GCP_CONSOLE_SAPISID", "")
	t.Setenv("GCP_CONSOLE_ALL_COOKIES", "")

	_, err := NewGCPConsoleSession()
	if err == nil {
		t.Error("expected error when all env vars are missing")
	}
}

func TestNewGCPConsoleSession_MissingAllCookies(t *testing.T) {
	t.Setenv("GCP_CONSOLE_SAPISID", "some_sapisid")
	t.Setenv("GCP_CONSOLE_ALL_COOKIES", "")

	_, err := NewGCPConsoleSession()
	if err == nil {
		t.Error("expected error when GCP_CONSOLE_ALL_COOKIES is missing")
	}
}

func TestNewGCPConsoleSession_ExtractSAPISIDFromCookies(t *testing.T) {
	t.Setenv("GCP_CONSOLE_SAPISID", "")
	t.Setenv("GCP_CONSOLE_ALL_COOKIES", "SID=abc; SAPISID=extracted_value; HSID=def")

	session, err := NewGCPConsoleSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.SAPISID != "extracted_value" {
		t.Errorf("SAPISID = %q, want %q", session.SAPISID, "extracted_value")
	}
}

func TestNewGCPConsoleSession_ExplicitSAPISID(t *testing.T) {
	t.Setenv("GCP_CONSOLE_SAPISID", "explicit_value")
	t.Setenv("GCP_CONSOLE_ALL_COOKIES", "SID=abc; SAPISID=cookie_value")

	session, err := NewGCPConsoleSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.SAPISID != "explicit_value" {
		t.Errorf("SAPISID = %q, want %q (explicit should take precedence)", session.SAPISID, "explicit_value")
	}
}

func TestNewGCPConsoleSession_AllCookiesUsedDirectly(t *testing.T) {
	cookies := "SID=abc; SAPISID=val; __Secure-1PSID=xyz"
	t.Setenv("GCP_CONSOLE_SAPISID", "")
	t.Setenv("GCP_CONSOLE_ALL_COOKIES", cookies)

	session, err := NewGCPConsoleSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.AllCookies != cookies {
		t.Errorf("AllCookies = %q, want %q", session.AllCookies, cookies)
	}
}
