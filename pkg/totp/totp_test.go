package totp

import (
	"strings"
	"testing"
	"time"

	pqotp "github.com/pquerna/otp/totp"
)

func TestSetupGeneratesOtpauthURL(t *testing.T) {
	secret, url, err := Setup("alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if secret == "" {
		t.Fatal("empty secret")
	}
	if !strings.HasPrefix(url, "otpauth://totp/") {
		t.Fatalf("bad url: %q", url)
	}
	if !strings.Contains(url, "IM%20Wallet") && !strings.Contains(url, "IM+Wallet") {
		t.Fatalf("expected issuer 'IM Wallet' in url: %q", url)
	}
}

func TestValidateRoundtrip(t *testing.T) {
	secret, _, err := Setup("bob@example.com")
	if err != nil {
		t.Fatal(err)
	}
	code, err := pqotp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !Validate(secret, code) {
		t.Fatal("expected current code to validate")
	}
}

func TestValidateRejectsWrongCode(t *testing.T) {
	secret, _, _ := Setup("c@x.com")
	if Validate(secret, "000000") {
		t.Fatal("000000 should not validate against a fresh secret")
	}
}
