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
		wantFmt   string // prefix format to verify structure
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
			// SHA1 produces a 40-char hex string
			hash := got[len(tc.wantFmt):]
			if len(hash) != 40 {
				t.Errorf("SHA1 hex should be 40 chars, got %d in %q", len(hash), hash)
			}
		})
	}
}

func TestComputeSAPISIDHash_Deterministic(t *testing.T) {
	// Same inputs must produce the same output
	ts := int64(1700000000)
	sapisid := "test_sapisid_value"
	h1 := computeSAPISIDHash(ts, sapisid)
	h2 := computeSAPISIDHash(ts, sapisid)
	if h1 != h2 {
		t.Errorf("expected deterministic output: %q != %q", h1, h2)
	}
}

func TestComputeSAPISIDHash_DifferentInputs(t *testing.T) {
	// Different inputs must produce different hashes
	ts := int64(1700000000)
	h1 := computeSAPISIDHash(ts, "sapisid_a")
	h2 := computeSAPISIDHash(ts, "sapisid_b")
	if h1 == h2 {
		t.Error("different SAPISIDs should produce different hashes")
	}
}

func TestGCPConsoleSession_CookieString(t *testing.T) {
	tests := []struct {
		name    string
		session GCPConsoleSession
		want    string
	}{
		{
			name: "required cookies only",
			session: GCPConsoleSession{
				SAPISID: "sapisid_val",
				SID:     "sid_val",
			},
			want: "SAPISID=sapisid_val; SID=sid_val",
		},
		{
			name: "all cookies",
			session: GCPConsoleSession{
				SAPISID: "sapisid_val",
				SID:     "sid_val",
				HSID:    "hsid_val",
				SSID:    "ssid_val",
				APISID:  "apisid_val",
			},
			want: "SAPISID=sapisid_val; SID=sid_val; HSID=hsid_val; SSID=ssid_val; APISID=apisid_val",
		},
		{
			name: "with HSID only",
			session: GCPConsoleSession{
				SAPISID: "s",
				SID:     "d",
				HSID:    "h",
			},
			want: "SAPISID=s; SID=d; HSID=h",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.session.CookieString()
			if got != tc.want {
				t.Errorf("CookieString() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGCPConsoleSession_SAPISIDHash_Format(t *testing.T) {
	session := &GCPConsoleSession{SAPISID: "test_sapisid", SID: "test_sid"}
	hash := session.SAPISIDHash()
	// Must start with "SAPISIDHASH "
	if len(hash) < 12 || hash[:12] != "SAPISIDHASH " {
		t.Errorf("SAPISIDHash() should start with 'SAPISIDHASH ', got %q", hash)
	}
	// Must contain an underscore separating timestamp from hex
	var ts int64
	var hexPart string
	if _, err := fmt.Sscanf(hash[12:], "%d_%s", &ts, &hexPart); err != nil {
		t.Errorf("SAPISIDHash() format invalid: %q, parse error: %v", hash, err)
	}
	if len(hexPart) != 40 {
		t.Errorf("SHA1 hex should be 40 chars, got %d", len(hexPart))
	}
}

func TestNewGCPConsoleSession_MissingRequired(t *testing.T) {
	t.Setenv("GCP_CONSOLE_SAPISID", "")
	t.Setenv("GCP_CONSOLE_SID", "")

	_, err := NewGCPConsoleSession()
	if err == nil {
		t.Error("expected error when required env vars are missing")
	}
}

func TestNewGCPConsoleSession_MissingSID(t *testing.T) {
	t.Setenv("GCP_CONSOLE_SAPISID", "some_sapisid")
	t.Setenv("GCP_CONSOLE_SID", "")

	_, err := NewGCPConsoleSession()
	if err == nil {
		t.Error("expected error when GCP_CONSOLE_SID is missing")
	}
}

func TestNewGCPConsoleSession_Success(t *testing.T) {
	t.Setenv("GCP_CONSOLE_SAPISID", "my_sapisid")
	t.Setenv("GCP_CONSOLE_SID", "my_sid")
	t.Setenv("GCP_CONSOLE_HSID", "my_hsid")
	t.Setenv("GCP_CONSOLE_SSID", "")
	t.Setenv("GCP_CONSOLE_APISID", "")

	session, err := NewGCPConsoleSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.SAPISID != "my_sapisid" {
		t.Errorf("SAPISID = %q, want %q", session.SAPISID, "my_sapisid")
	}
	if session.SID != "my_sid" {
		t.Errorf("SID = %q, want %q", session.SID, "my_sid")
	}
	if session.HSID != "my_hsid" {
		t.Errorf("HSID = %q, want %q", session.HSID, "my_hsid")
	}
}
