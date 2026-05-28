// Package cryptoutil provides small encryption helpers used across services.
package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// DecodeKey parses a hex-encoded AES key. Length must be 16, 24, or 32 bytes
// (AES-128/192/256). 32 bytes (= 64 hex chars) is recommended.
func DecodeKey(hexKey string) ([]byte, error) {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("hex decode: %w", err)
	}
	switch len(b) {
	case 16, 24, 32:
		return b, nil
	default:
		return nil, fmt.Errorf("aes key must be 16/24/32 bytes, got %d", len(b))
	}
}

// EncryptAESGCM seals plaintext with AES-GCM. Output format:
//
//	base64( nonce(12) || ciphertext || gcm_tag(16) )
func EncryptAESGCM(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	out := make([]byte, 0, len(nonce)+len(sealed))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptAESGCM reverses EncryptAESGCM. Returns an error if the auth tag fails
// (tampering, wrong key, or corrupted ciphertext).
func DecryptAESGCM(key []byte, ciphertext string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce := raw[:gcm.NonceSize()]
	body := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, body, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
