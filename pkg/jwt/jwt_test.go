package jwt

import (
	"testing"
	"time"
)

func TestSignAndVerify(t *testing.T) {
	svc := NewService("test-secret", 168*time.Hour)
	token, err := svc.Sign("user123")
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	userID, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if userID != "user123" {
		t.Errorf("want user123, got %s", userID)
	}
}

func TestVerifyExpired(t *testing.T) {
	svc := NewService("secret", -time.Hour) // negative = already expired
	token, _ := svc.Sign("user123")
	_, err := svc.Verify(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerifyWrongSecret(t *testing.T) {
	a := NewService("secret-a", time.Hour)
	b := NewService("secret-b", time.Hour)
	token, _ := a.Sign("user123")
	_, err := b.Verify(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}
