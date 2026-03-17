package crypto

import (
	"crypto/rand"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"short string", "hello"},
		{"token-like", "ya29.a0AfH6SMBx-example-access-token-value"},
		{"unicode", "日本語テスト"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(key, tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if encrypted == tt.plaintext {
				t.Error("encrypted text should differ from plaintext")
			}

			decrypted, err := Decrypt(key, encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptDifferentCiphertexts(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	e1, _ := Encrypt(key, "same")
	e2, _ := Encrypt(key, "same")

	if e1 == e2 {
		t.Error("same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	rand.Read(key1)
	rand.Read(key2)

	encrypted, _ := Encrypt(key1, "secret")
	_, err := Decrypt(key2, encrypted)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
}

func TestInvalidKeyLength(t *testing.T) {
	shortKey := make([]byte, 16)

	_, err := Encrypt(shortKey, "test")
	if err == nil {
		t.Error("expected error for short key on encrypt")
	}

	_, err = Decrypt(shortKey, "dGVzdA==")
	if err == nil {
		t.Error("expected error for short key on decrypt")
	}
}
