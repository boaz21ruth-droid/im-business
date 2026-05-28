// Package totp wraps pquerna/otp with the issuer/window defaults this app uses.
package totp

import (
	"time"

	"github.com/pquerna/otp/totp"
)

const (
	issuer = "IM Wallet"
	// Allow ±1 step (±30s) clock skew between server and authenticator app.
	skew = 1
)

// Setup returns a new TOTP secret (base32) and an otpauth:// URL for QR rendering.
// accountName should be human-readable (email/phone/userID); it's what the
// Authenticator app shows next to the issuer.
func Setup(accountName string) (secret, otpauthURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// Validate returns true if code matches secret within the configured skew window.
func Validate(secret, code string) bool {
	ok, _ := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period: 30,
		Skew:   skew,
		Digits: 6,
	})
	return ok
}
