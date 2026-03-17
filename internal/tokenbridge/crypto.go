package tokenbridge

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	nonceLength = 12
	tagLength   = 16
)

// Decrypt decodes raw AES-256-GCM encrypted bytes (from bytea column).
// Wire format: nonce [12 bytes] || ciphertext || auth_tag [16 bytes]
func Decrypt(data []byte, hexKey string) (string, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", fmt.Errorf("invalid hex key: %w", err)
	}
	if len(key) != 32 {
		return "", errors.New("key must be 32 bytes (64 hex characters)")
	}

	if len(data) < nonceLength+tagLength {
		return "", errors.New("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes.NewCipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("cipher.NewGCM: %w", err)
	}

	nonce := data[:nonceLength]
	ciphertext := data[nonceLength:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("gcm.Open: %w", err)
	}

	return string(plaintext), nil
}
