package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/web1/im-business/internal/repo"
	"github.com/web1/im-business/pkg/cryptoutil"
	totppkg "github.com/web1/im-business/pkg/totp"
)

// TotpService manages TOTP setup, enable, disable, and per-transfer verification.
// The secret is stored AES-GCM-encrypted in the users table; the plain secret
// only ever exists transiently in memory during validation.
type TotpService struct {
	users      *repo.UserRepo
	rdb        *redis.Client
	encryptKey []byte
}

// Setup writes the temporary secret here (so a user can re-load the QR within
// 10 minutes without re-generating). Indexed by userID.
const (
	setupKeyFmt  = "im:totp:setup:%s"
	setupTTL     = 10 * time.Minute
	failKeyFmt   = "im:totp:fail:%s"
	failWindow   = 5 * time.Minute
	failMaxCount = 5
	lockTTL      = 15 * time.Minute
)

var (
	ErrTotpAlreadyEnabled = errors.New("totp already enabled")
	ErrTotpNotEnabled     = errors.New("totp not enabled")
	ErrTotpInvalidCode    = errors.New("invalid totp code")
	ErrTotpLocked         = errors.New("totp verification locked")
	ErrTotpSetupExpired   = errors.New("setup expired, request a new QR code")
)

func NewTotpService(users *repo.UserRepo, rdb *redis.Client, encryptKey []byte) *TotpService {
	return &TotpService{users: users, rdb: rdb, encryptKey: encryptKey}
}

// Setup generates a fresh secret for a user and caches it in Redis pending Enable.
// Re-calling Setup overwrites the previous pending secret.
func (s *TotpService) Setup(ctx context.Context, userID, accountName string) (secret, otpauthURL string, err error) {
	u, err := s.users.FindByUserID(userID)
	if err != nil {
		return "", "", err
	}
	if u == nil {
		return "", "", errors.New("user not found")
	}
	if u.TotpEnabled {
		return "", "", ErrTotpAlreadyEnabled
	}
	if accountName == "" {
		accountName = userID
	}
	secret, otpauthURL, err = totppkg.Setup(accountName)
	if err != nil {
		return "", "", fmt.Errorf("generate: %w", err)
	}
	if err := s.rdb.Set(ctx, fmt.Sprintf(setupKeyFmt, userID), secret, setupTTL).Err(); err != nil {
		return "", "", fmt.Errorf("cache: %w", err)
	}
	return secret, otpauthURL, nil
}

// Enable persists the pending secret and flips enabled=true, after verifying
// the user can produce a valid code from their Authenticator app.
func (s *TotpService) Enable(ctx context.Context, userID, code string) error {
	u, err := s.users.FindByUserID(userID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("user not found")
	}
	if u.TotpEnabled {
		return ErrTotpAlreadyEnabled
	}
	secret, err := s.rdb.Get(ctx, fmt.Sprintf(setupKeyFmt, userID)).Result()
	if errors.Is(err, redis.Nil) {
		return ErrTotpSetupExpired
	}
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	if !totppkg.Validate(secret, code) {
		return ErrTotpInvalidCode
	}
	enc, err := cryptoutil.EncryptAESGCM(s.encryptKey, secret)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}
	if err := s.users.Update(userID, map[string]any{
		"totp_secret":  enc,
		"totp_enabled": true,
	}); err != nil {
		return err
	}
	s.rdb.Del(ctx, fmt.Sprintf(setupKeyFmt, userID))
	return nil
}

// Disable verifies the current code, then clears the secret and disables 2FA.
func (s *TotpService) Disable(ctx context.Context, userID, code string) error {
	u, err := s.users.FindByUserID(userID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("user not found")
	}
	if !u.TotpEnabled || u.TotpSecret == "" {
		return ErrTotpNotEnabled
	}
	secret, err := cryptoutil.DecryptAESGCM(s.encryptKey, u.TotpSecret)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}
	if !totppkg.Validate(secret, code) {
		return ErrTotpInvalidCode
	}
	return s.users.Update(userID, map[string]any{
		"totp_secret":  "",
		"totp_enabled": false,
	})
}

// Status returns whether the user has TOTP enabled.
func (s *TotpService) Status(ctx context.Context, userID string) (bool, error) {
	u, err := s.users.FindByUserID(userID)
	if err != nil {
		return false, err
	}
	if u == nil {
		return false, errors.New("user not found")
	}
	return u.TotpEnabled, nil
}

// Verify is the gating call used by the wallet transfer flow.
// Users without TOTP enabled always pass — callers should still gate by Status
// when displaying UI, but the API is forgiving so the client can blindly call.
// Tracks consecutive failures in Redis; locks the user out for lockTTL after
// failMaxCount failures within failWindow.
func (s *TotpService) Verify(ctx context.Context, userID, code string) error {
	u, err := s.users.FindByUserID(userID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("user not found")
	}
	if !u.TotpEnabled || u.TotpSecret == "" {
		return nil
	}

	failKey := fmt.Sprintf(failKeyFmt, userID)
	count, err := s.rdb.Get(ctx, failKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("redis: %w", err)
	}
	if count >= failMaxCount {
		return ErrTotpLocked
	}

	secret, err := cryptoutil.DecryptAESGCM(s.encryptKey, u.TotpSecret)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}
	if !totppkg.Validate(secret, code) {
		// On the first failure, set the key with TTL.
		// Subsequent failures INCR; once count hits failMaxCount, escalate TTL to lockTTL.
		newCount, _ := s.rdb.Incr(ctx, failKey).Result()
		if newCount == 1 {
			s.rdb.Expire(ctx, failKey, failWindow)
		}
		if newCount >= failMaxCount {
			s.rdb.Expire(ctx, failKey, lockTTL)
		}
		return ErrTotpInvalidCode
	}
	s.rdb.Del(ctx, failKey)
	return nil
}
