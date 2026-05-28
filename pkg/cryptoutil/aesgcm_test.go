package cryptoutil

import (
	"strings"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key, err := DecodeKey(strings.Repeat("ab", 32)) // 32 bytes
	if err != nil {
		t.Fatal(err)
	}
	plain := "JBSWY3DPEHPK3PXP" // example base32 TOTP secret

	enc, err := EncryptAESGCM(key, plain)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if enc == plain {
		t.Fatal("ciphertext equals plaintext")
	}

	dec, err := DecryptAESGCM(key, enc)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if dec != plain {
		t.Fatalf("roundtrip mismatch: got %q", dec)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	keyA, _ := DecodeKey(strings.Repeat("ab", 32))
	keyB, _ := DecodeKey(strings.Repeat("cd", 32))
	enc, _ := EncryptAESGCM(keyA, "hello")
	if _, err := DecryptAESGCM(keyB, enc); err == nil {
		t.Fatal("expected decrypt to fail with wrong key")
	}
}

func TestDecodeKeyInvalidLength(t *testing.T) {
	if _, err := DecodeKey("aabb"); err == nil {
		t.Fatal("expected error for 2-byte key")
	}
}

func TestEncryptDifferentNonceEachCall(t *testing.T) {
	key, _ := DecodeKey(strings.Repeat("ab", 32))
	a, _ := EncryptAESGCM(key, "same")
	b, _ := EncryptAESGCM(key, "same")
	if a == b {
		t.Fatal("nonce should be random — two encryptions of same plaintext should differ")
	}
}
